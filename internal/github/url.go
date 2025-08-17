package github

import (
	"errors"
	"strings"
)

// GitHubSource represents a parsed GitHub repository source
type GitHubSource struct {
	Owner  string
	Repo   string
	Path   string
	IsFile bool
}

var (
	ErrInvalidURL   = errors.New("invalid GitHub URL format")
	ErrMissingOwner = errors.New("GitHub owner is required")
	ErrMissingRepo  = errors.New("GitHub repository is required")
)

// ParseGitHubURL parses a GitHub URL in the format "github:owner/repo/path"
func ParseGitHubURL(url string) (*GitHubSource, error) {
	if !strings.HasPrefix(url, "github:") {
		return nil, ErrInvalidURL
	}

	// Remove prefix
	path := strings.TrimPrefix(url, "github:")
	parts := strings.SplitN(path, "/", 3)

	if len(parts) < 2 {
		return nil, ErrInvalidURL
	}

	owner := parts[0]
	repo := parts[1]
	filePath := ""

	if len(parts) > 2 {
		filePath = parts[2]
	}

	if owner == "" {
		return nil, ErrMissingOwner
	}

	if repo == "" {
		return nil, ErrMissingRepo
	}

	// Determine if the path is likely a file or directory
	// This is a guess - we'll know for sure when we call the GitHub API
	isFile := false
	if filePath != "" {
		isFile = !strings.HasSuffix(filePath, "/")
	}

	return &GitHubSource{
		Owner:  owner,
		Repo:   repo,
		Path:   filePath,
		IsFile: isFile,
	}, nil
}

// APIPath returns the GitHub API path for this source
func (s *GitHubSource) APIPath() string {
	if s.Path == "" {
		return ""
	}
	return s.Path
}

// FullRepoName returns the full repository name (owner/repo)
func (s *GitHubSource) FullRepoName() string {
	return s.Owner + "/" + s.Repo
}
