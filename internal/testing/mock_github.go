// Package testing provides test utilities and mocks
package testing

import (
	"errors"
	"xcp/internal/github"
)

// MockGitHubClient is a mock implementation of the GitHub API client for testing
type MockGitHubClient struct {
	FileContents       map[string][]byte
	DirectoryContents  map[string]github.DirectoryContents
	ExistingRepos      map[string]bool
	FailGetFileContent bool
	FailGetDirContent  bool
	FailRepoExists     bool
}

// NewMockGitHubClient creates a new mock GitHub client
func NewMockGitHubClient() *MockGitHubClient {
	return &MockGitHubClient{
		FileContents:      make(map[string][]byte),
		DirectoryContents: make(map[string]github.DirectoryContents),
		ExistingRepos:     make(map[string]bool),
	}
}

// GetFileContent mocks fetching a file's content
func (m *MockGitHubClient) GetFileContent(owner, repo, path string) ([]byte, error) {
	if m.FailGetFileContent {
		return nil, errors.New("mock file content failure")
	}

	key := owner + "/" + repo + "/" + path
	content, exists := m.FileContents[key]
	if !exists {
		return nil, github.ErrFileNotFound
	}

	return content, nil
}

// GetDirectoryContents mocks fetching directory contents
func (m *MockGitHubClient) GetDirectoryContents(owner, repo, path string) (github.DirectoryContents, error) {
	if m.FailGetDirContent {
		return nil, errors.New("mock directory content failure")
	}

	key := owner + "/" + repo + "/" + path
	content, exists := m.DirectoryContents[key]
	if !exists {
		return nil, github.ErrDirectoryNotFound
	}

	return content, nil
}

// RepositoryExists mocks checking if a repository exists
func (m *MockGitHubClient) RepositoryExists(owner, repo string) (bool, error) {
	if m.FailRepoExists {
		return false, errors.New("mock repository exists failure")
	}

	key := owner + "/" + repo
	exists, found := m.ExistingRepos[key]
	if !found {
		return false, nil
	}

	return exists, nil
}

// AddFile adds a mock file
func (m *MockGitHubClient) AddFile(owner, repo, path string, content []byte) {
	key := owner + "/" + repo + "/" + path
	m.FileContents[key] = content
}

// AddDirectory adds a mock directory
func (m *MockGitHubClient) AddDirectory(owner, repo, path string, contents github.DirectoryContents) {
	key := owner + "/" + repo + "/" + path
	m.DirectoryContents[key] = contents
}

// AddRepository adds a mock repository
func (m *MockGitHubClient) AddRepository(owner, repo string, exists bool) {
	key := owner + "/" + repo
	m.ExistingRepos[key] = exists
}
