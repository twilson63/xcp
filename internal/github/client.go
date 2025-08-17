package github

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	apiBaseURL     = "https://api.github.com"
	defaultTimeout = 30 * time.Second
)

// URL generators for API endpoints
var (
	// getContentsURL generates the URL for fetching repository contents
	getContentsURL = func(owner, repo, path string) string {
		return fmt.Sprintf("%s/repos/%s/%s/contents/%s", apiBaseURL, owner, repo, url.PathEscape(path))
	}

	// getRepoURL generates the URL for checking repository existence
	getRepoURL = func(owner, repo string) string {
		return fmt.Sprintf("%s/repos/%s/%s", apiBaseURL, owner, repo)
	}
)

var (
	ErrFileNotFound       = errors.New("file not found in repository")
	ErrDirectoryNotFound  = errors.New("directory not found in repository")
	ErrRateLimitExceeded  = errors.New("GitHub API rate limit exceeded")
	ErrRepositoryNotFound = errors.New("GitHub repository not found")
	ErrNetworkFailure     = errors.New("network failure when contacting GitHub API")
)

// ContentType represents the type of content returned by the GitHub API
type ContentType string

const (
	FileContent      ContentType = "file"
	DirectoryContent ContentType = "dir"
)

// Client is a GitHub API client
type Client struct {
	httpClient *http.Client
	token      string // For future authentication support
}

// ContentResponse represents the response from the GitHub contents API
type ContentResponse struct {
	Type        ContentType `json:"type"`
	Name        string      `json:"name"`
	Path        string      `json:"path"`
	Sha         string      `json:"sha"`
	Size        int         `json:"size"`
	URL         string      `json:"url"`
	DownloadURL string      `json:"download_url"`
	Content     string      `json:"content"`
	Encoding    string      `json:"encoding"`
}

// DirectoryContents represents a list of contents in a directory
type DirectoryContents []ContentResponse

// NewClient creates a new GitHub API client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// GetFileContent fetches the content of a file from a GitHub repository
func (c *Client) GetFileContent(owner, repo, path string) ([]byte, error) {
	apiURL := getContentsURL(owner, repo, path)

	resp, err := c.httpClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNetworkFailure, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrFileNotFound
	}

	if resp.StatusCode == http.StatusForbidden {
		return nil, ErrRateLimitExceeded
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var content ContentResponse
	if err := json.Unmarshal(body, &content); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if content.Type != FileContent {
		return nil, fmt.Errorf("expected file content, got %s", content.Type)
	}

	// Decode base64 content
	if content.Encoding == "base64" {
		decoded, err := base64.StdEncoding.DecodeString(content.Content)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 content: %w", err)
		}
		return decoded, nil
	}

	return []byte(content.Content), nil
}

// GetDirectoryContents fetches the contents of a directory from a GitHub repository
func (c *Client) GetDirectoryContents(owner, repo, path string) (DirectoryContents, error) {
	apiURL := getContentsURL(owner, repo, path)

	resp, err := c.httpClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNetworkFailure, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrDirectoryNotFound
	}

	if resp.StatusCode == http.StatusForbidden {
		return nil, ErrRateLimitExceeded
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var contents DirectoryContents
	if err := json.Unmarshal(body, &contents); err != nil {
		// If it's not a directory, it might be a file
		var singleContent ContentResponse
		if err := json.Unmarshal(body, &singleContent); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		if singleContent.Type == FileContent {
			return nil, fmt.Errorf("expected directory, got file: %s", singleContent.Path)
		}

		return nil, fmt.Errorf("failed to parse directory contents: %w", err)
	}

	return contents, nil
}

// RepositoryExists checks if a repository exists
func (c *Client) RepositoryExists(owner, repo string) (bool, error) {
	apiURL := getRepoURL(owner, repo)

	resp, err := c.httpClient.Get(apiURL)
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrNetworkFailure, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	if resp.StatusCode == http.StatusForbidden {
		return false, ErrRateLimitExceeded
	}

	return resp.StatusCode == http.StatusOK, nil
}
