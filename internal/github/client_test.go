package github

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// testClient creates a new GitHub client that uses the given test server
func testClient(server *httptest.Server) *Client {
	client := NewClient()
	client.httpClient = server.Client()
	return client
}

// makeTestURL creates a URL for a test request based on the server URL
func makeTestURL(serverURL, path string) string {
	return strings.Replace(path, "https://api.github.com", serverURL, 1)
}

func TestGetFileContent(t *testing.T) {
	// Set up a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the path from the URL
		path := r.URL.Path

		switch path {
		case "/repos/owner/repo/contents/file.txt":
			// Return a file content response
			content := "Hello, World!"
			encoded := base64.StdEncoding.EncodeToString([]byte(content))
			resp := ContentResponse{
				Type:        FileContent,
				Name:        "file.txt",
				Path:        "file.txt",
				Sha:         "abc123",
				Size:        len(content),
				Content:     encoded,
				Encoding:    "base64",
				DownloadURL: "https://example.com/download/file.txt",
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)

		case "/repos/owner/repo/contents/not-found.txt":
			// Return a 404 response
			w.WriteHeader(http.StatusNotFound)

		case "/repos/owner/repo/contents/rate-limit":
			// Return a rate limit exceeded response
			w.WriteHeader(http.StatusForbidden)
			w.Header().Set("X-RateLimit-Remaining", "0")

		default:
			// Return a 404 response for any other path
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create a client that uses the test server
	client := testClient(server)

	// Use a custom makeRequest method to point to our test server
	originalGetFunc := getContentsURL
	getContentsURL = func(owner, repo, path string) string {
		return server.URL + "/repos/" + owner + "/" + repo + "/contents/" + path
	}
	defer func() { getContentsURL = originalGetFunc }()

	// Test getting a valid file
	content, err := client.GetFileContent("owner", "repo", "file.txt")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if string(content) != "Hello, World!" {
		t.Errorf("Expected content 'Hello, World!', got '%s'", string(content))
	}

	// Test getting a non-existent file
	_, err = client.GetFileContent("owner", "repo", "not-found.txt")
	if err != ErrFileNotFound {
		t.Errorf("Expected ErrFileNotFound, got %v", err)
	}

	// Test rate limit exceeded
	_, err = client.GetFileContent("owner", "repo", "rate-limit")
	if err != ErrRateLimitExceeded {
		t.Errorf("Expected ErrRateLimitExceeded, got %v", err)
	}
}

func TestGetDirectoryContents(t *testing.T) {
	// Set up a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the path from the URL
		path := r.URL.Path

		switch path {
		case "/repos/owner/repo/contents/dir":
			// Return a directory content response
			contents := DirectoryContents{
				{
					Type:        FileContent,
					Name:        "file1.txt",
					Path:        "dir/file1.txt",
					Sha:         "abc123",
					Size:        10,
					DownloadURL: "https://example.com/download/dir/file1.txt",
				},
				{
					Type:        FileContent,
					Name:        "file2.txt",
					Path:        "dir/file2.txt",
					Sha:         "def456",
					Size:        20,
					DownloadURL: "https://example.com/download/dir/file2.txt",
				},
				{
					Type:        DirectoryContent,
					Name:        "subdir",
					Path:        "dir/subdir",
					Sha:         "ghi789",
					DownloadURL: "",
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(contents)

		case "/repos/owner/repo/contents/empty-dir":
			// Return an empty directory
			contents := DirectoryContents{}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(contents)

		case "/repos/owner/repo/contents/not-found-dir":
			// Return a 404 response
			w.WriteHeader(http.StatusNotFound)

		default:
			// Return a 404 response for any other path
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create a client that uses the test server
	client := testClient(server)

	// Use a custom makeRequest method to point to our test server
	originalGetFunc := getContentsURL
	getContentsURL = func(owner, repo, path string) string {
		return server.URL + "/repos/" + owner + "/" + repo + "/contents/" + path
	}
	defer func() { getContentsURL = originalGetFunc }()

	// Test getting a valid directory
	contents, err := client.GetDirectoryContents("owner", "repo", "dir")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(contents) != 3 {
		t.Errorf("Expected 3 items, got %d", len(contents))
	}

	// Test getting an empty directory
	contents, err = client.GetDirectoryContents("owner", "repo", "empty-dir")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(contents) != 0 {
		t.Errorf("Expected 0 items, got %d", len(contents))
	}

	// Test getting a non-existent directory
	_, err = client.GetDirectoryContents("owner", "repo", "not-found-dir")
	if err != ErrDirectoryNotFound {
		t.Errorf("Expected ErrDirectoryNotFound, got %v", err)
	}
}

func TestRepositoryExists(t *testing.T) {
	// Set up a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the path from the URL
		path := r.URL.Path

		switch path {
		case "/repos/owner/existing-repo":
			// Return a success response
			w.WriteHeader(http.StatusOK)

		case "/repos/owner/non-existing-repo":
			// Return a 404 response
			w.WriteHeader(http.StatusNotFound)

		case "/repos/rate-limited/repo":
			// Return a rate limit exceeded response
			w.WriteHeader(http.StatusForbidden)
			w.Header().Set("X-RateLimit-Remaining", "0")

		default:
			// Return a 404 response for any other path
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create a client that uses the test server
	client := testClient(server)

	// Use a custom makeRequest method to point to our test server
	originalGetFunc := getRepoURL
	getRepoURL = func(owner, repo string) string {
		return server.URL + "/repos/" + owner + "/" + repo
	}
	defer func() { getRepoURL = originalGetFunc }()

	// Test checking an existing repository
	exists, err := client.RepositoryExists("owner", "existing-repo")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !exists {
		t.Errorf("Expected repository to exist")
	}

	// Test checking a non-existent repository
	exists, err = client.RepositoryExists("owner", "non-existing-repo")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if exists {
		t.Errorf("Expected repository to not exist")
	}

	// Test rate limit exceeded
	_, err = client.RepositoryExists("rate-limited", "repo")
	if err != ErrRateLimitExceeded {
		t.Errorf("Expected ErrRateLimitExceeded, got %v", err)
	}
}
