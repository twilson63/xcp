package downloader

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"xcp/internal/github"
	xtest "xcp/internal/testing"
)

func TestDownloadFile(t *testing.T) {
	// Create mock client
	mockClient := xtest.NewMockGitHubClient()

	// Add test file
	owner := "testowner"
	repo := "testrepo"
	path := "testfile.txt"
	content := []byte("test file content")
	mockClient.AddFile(owner, repo, path, content)
	mockClient.AddRepository(owner, repo, true)

	// Create mock stdout/stderr
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	// Create downloader
	dl := NewDownloader(mockClient, stdout, stderr)

	// Test downloading to file
	source := &github.GitHubSource{
		Owner:  owner,
		Repo:   repo,
		Path:   path,
		IsFile: true,
	}

	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "downloaded.txt")

	// Download file
	err := dl.DownloadFile(source, destPath, DownloadOptions{
		OutputToStdout: false,
		Overwrite:      true,
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify file was written
	gotContent, err := os.ReadFile(destPath)
	if err != nil {
		t.Errorf("Failed to read downloaded file: %v", err)
	}

	if !bytes.Equal(gotContent, content) {
		t.Errorf("Downloaded content doesn't match.\nExpected: %s\nGot: %s", content, gotContent)
	}

	// Test downloading to stdout
	stdout.Reset()
	stderr.Reset()

	err = dl.DownloadFile(source, "", DownloadOptions{
		OutputToStdout: true,
		Overwrite:      false,
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !bytes.Equal(stdout.Bytes(), content) {
		t.Errorf("Stdout content doesn't match.\nExpected: %s\nGot: %s", content, stdout.Bytes())
	}

	// Test error cases
	notFoundSource := &github.GitHubSource{
		Owner:  owner,
		Repo:   repo,
		Path:   "nonexistent.txt",
		IsFile: true,
	}

	err = dl.DownloadFile(notFoundSource, destPath, DownloadOptions{})
	if err == nil {
		t.Errorf("Expected error for non-existent file")
	}

	// Test repository not found
	unknownRepoSource := &github.GitHubSource{
		Owner:  "unknown",
		Repo:   "unknown",
		Path:   path,
		IsFile: true,
	}

	err = dl.DownloadFile(unknownRepoSource, destPath, DownloadOptions{})
	if err == nil {
		t.Errorf("Expected error for unknown repository")
	}

	// Test GitHub API failure
	mockClient.FailGetFileContent = true
	err = dl.DownloadFile(source, destPath, DownloadOptions{})
	if err == nil {
		t.Errorf("Expected error for GitHub API failure")
	}
	mockClient.FailGetFileContent = false
}

func TestDownloadDirectory(t *testing.T) {
	// Create mock client
	mockClient := xtest.NewMockGitHubClient()

	// Add test repository with directory structure
	owner := "testowner"
	repo := "testrepo"
	dirPath := "testdir"

	// Add mock directory contents
	dirContents := github.DirectoryContents{
		{
			Type: github.FileContent,
			Name: "file1.txt",
			Path: "testdir/file1.txt",
			Size: 10,
		},
		{
			Type: github.FileContent,
			Name: "file2.txt",
			Path: "testdir/file2.txt",
			Size: 20,
		},
		{
			Type: github.DirectoryContent,
			Name: "subdir",
			Path: "testdir/subdir",
			Size: 0,
		},
	}

	subdirContents := github.DirectoryContents{
		{
			Type: github.FileContent,
			Name: "file3.txt",
			Path: "testdir/subdir/file3.txt",
			Size: 30,
		},
	}

	// Add directory structure to mock client
	mockClient.AddDirectory(owner, repo, dirPath, dirContents)
	mockClient.AddDirectory(owner, repo, dirPath+"/subdir", subdirContents)
	mockClient.AddRepository(owner, repo, true)

	// Add file contents
	mockClient.AddFile(owner, repo, "testdir/file1.txt", []byte("file1 content"))
	mockClient.AddFile(owner, repo, "testdir/file2.txt", []byte("file2 content"))
	mockClient.AddFile(owner, repo, "testdir/subdir/file3.txt", []byte("file3 content"))

	// Create mock stdout/stderr
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	// Create downloader
	dl := NewDownloader(mockClient, stdout, stderr)

	// Test downloading directory
	source := &github.GitHubSource{
		Owner:  owner,
		Repo:   repo,
		Path:   dirPath,
		IsFile: false,
	}

	tempDir := t.TempDir()

	// Download directory
	err := dl.DownloadDirectory(source, tempDir, DownloadOptions{
		OutputToStdout: false, // Cannot output directory to stdout
		Overwrite:      true,
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify directory structure was created
	expectedFiles := map[string][]byte{
		filepath.Join(tempDir, "file1.txt"):           []byte("file1 content"),
		filepath.Join(tempDir, "file2.txt"):           []byte("file2 content"),
		filepath.Join(tempDir, "subdir", "file3.txt"): []byte("file3 content"),
	}

	for path, expectedContent := range expectedFiles {
		gotContent, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("Failed to read downloaded file %s: %v", path, err)
			continue
		}

		if !bytes.Equal(gotContent, expectedContent) {
			t.Errorf("Downloaded content for %s doesn't match.\nExpected: %s\nGot: %s",
				path, expectedContent, gotContent)
		}
	}

	// Test error cases

	// Test directory not found
	notFoundSource := &github.GitHubSource{
		Owner:  owner,
		Repo:   repo,
		Path:   "nonexistent",
		IsFile: false,
	}

	err = dl.DownloadDirectory(notFoundSource, tempDir, DownloadOptions{})
	if err == nil {
		t.Errorf("Expected error for non-existent directory")
	}

	// Test attempting to output directory to stdout
	err = dl.DownloadDirectory(source, "", DownloadOptions{
		OutputToStdout: true,
	})
	if err == nil {
		t.Errorf("Expected error for attempting to output directory to stdout")
	}

	// Test GitHub API failure
	mockClient.FailGetDirContent = true
	err = dl.DownloadDirectory(source, tempDir, DownloadOptions{})
	if err == nil {
		t.Errorf("Expected error for GitHub API failure")
	}
	mockClient.FailGetDirContent = false
}

func TestDownload(t *testing.T) {
	// Create mock client
	mockClient := xtest.NewMockGitHubClient()

	// Add test repository
	owner := "testowner"
	repo := "testrepo"
	mockClient.AddRepository(owner, repo, true)

	// Add file
	filePath := "testfile.txt"
	fileContent := []byte("test file content")
	mockClient.AddFile(owner, repo, filePath, fileContent)

	// Add directory
	dirPath := "testdir"
	dirContents := github.DirectoryContents{
		{
			Type: github.FileContent,
			Name: "file1.txt",
			Path: "testdir/file1.txt",
			Size: 10,
		},
	}
	mockClient.AddDirectory(owner, repo, dirPath, dirContents)
	mockClient.AddFile(owner, repo, "testdir/file1.txt", []byte("file1 content"))

	// Create mock stdout/stderr
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	// Create downloader
	dl := NewDownloader(mockClient, stdout, stderr)

	// Test cases
	tests := []struct {
		name        string
		source      *github.GitHubSource
		destPath    string
		options     DownloadOptions
		expectError bool
	}{
		{
			name: "Download file",
			source: &github.GitHubSource{
				Owner:  owner,
				Repo:   repo,
				Path:   filePath,
				IsFile: true,
			},
			destPath: "downloaded.txt",
			options: DownloadOptions{
				OutputToStdout: false,
				Overwrite:      true,
			},
			expectError: false,
		},
		{
			name: "Download directory",
			source: &github.GitHubSource{
				Owner:  owner,
				Repo:   repo,
				Path:   dirPath,
				IsFile: false,
			},
			destPath: "downloaded-dir",
			options: DownloadOptions{
				OutputToStdout: false,
				Overwrite:      true,
			},
			expectError: false,
		},
		{
			name: "File to stdout",
			source: &github.GitHubSource{
				Owner:  owner,
				Repo:   repo,
				Path:   filePath,
				IsFile: true,
			},
			destPath: "",
			options: DownloadOptions{
				OutputToStdout: true,
				Overwrite:      false,
			},
			expectError: false,
		},
		{
			name: "Directory to stdout (should fail)",
			source: &github.GitHubSource{
				Owner:  owner,
				Repo:   repo,
				Path:   dirPath,
				IsFile: false,
			},
			destPath: "",
			options: DownloadOptions{
				OutputToStdout: true,
				Overwrite:      false,
			},
			expectError: true,
		},
		{
			name: "Unknown repository",
			source: &github.GitHubSource{
				Owner:  "unknown",
				Repo:   "unknown",
				Path:   filePath,
				IsFile: true,
			},
			destPath: "downloaded.txt",
			options: DownloadOptions{
				OutputToStdout: false,
				Overwrite:      true,
			},
			expectError: true,
		},
		{
			name: "No destination path",
			source: &github.GitHubSource{
				Owner:  owner,
				Repo:   repo,
				Path:   filePath,
				IsFile: false,
			},
			destPath: "",
			options: DownloadOptions{
				OutputToStdout: false,
				Overwrite:      true,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset buffers
			stdout.Reset()
			stderr.Reset()

			// Create a temp directory for this test
			tempDir := t.TempDir()
			var destPath string
			if tt.destPath != "" {
				destPath = filepath.Join(tempDir, tt.destPath)
			} else {
				destPath = tt.destPath
			}

			// Run the download
			err := dl.Download(tt.source, destPath, tt.options)

			if tt.expectError && err == nil {
				t.Errorf("Expected error, got nil")
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// For successful file downloads to stdout, verify content
			if !tt.expectError && tt.source.IsFile && tt.options.OutputToStdout {
				if !bytes.Equal(stdout.Bytes(), fileContent) {
					t.Errorf("Stdout content doesn't match.\nExpected: %s\nGot: %s",
						fileContent, stdout.Bytes())
				}
			}
		})
	}
}
