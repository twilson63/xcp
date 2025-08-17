package downloader

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestZipDownloader_pathMatches(t *testing.T) {
	zd := &ZipDownloader{}

	tests := []struct {
		name       string
		zipPath    string
		sourcePath string
		expected   bool
	}{
		{
			name:       "Exact match",
			zipPath:    "repo-main/file.txt",
			sourcePath: "repo-main/file.txt",
			expected:   true,
		},
		{
			name:       "Directory match",
			zipPath:    "repo-main/src/file.go",
			sourcePath: "repo-main/src",
			expected:   true,
		},
		{
			name:       "No match",
			zipPath:    "repo-main/other/file.txt",
			sourcePath: "repo-main/src",
			expected:   false,
		},
		{
			name:       "Root directory match",
			zipPath:    "repo-main/anything.txt",
			sourcePath: "repo-main",
			expected:   true,
		},
		{
			name:       "Partial name match should not match",
			zipPath:    "repo-main/source-file.txt",
			sourcePath: "repo-main/src",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := zd.pathMatches(tt.zipPath, tt.sourcePath)
			if result != tt.expected {
				t.Errorf("pathMatches(%q, %q) = %v, expected %v", tt.zipPath, tt.sourcePath, result, tt.expected)
			}
		})
	}
}

func TestZipDownloader_getRelativePath(t *testing.T) {
	zd := &ZipDownloader{}

	tests := []struct {
		name        string
		zipPath     string
		sourcePath  string
		expected    string
		expectError bool
	}{
		{
			name:        "Exact match returns empty",
			zipPath:     "repo-main/src",
			sourcePath:  "repo-main/src",
			expected:    "",
			expectError: false,
		},
		{
			name:        "File under directory",
			zipPath:     "repo-main/src/file.go",
			sourcePath:  "repo-main/src",
			expected:    "file.go",
			expectError: false,
		},
		{
			name:        "Nested path",
			zipPath:     "repo-main/src/internal/pkg/file.go",
			sourcePath:  "repo-main/src",
			expected:    "internal/pkg/file.go",
			expectError: false,
		},
		{
			name:        "Path not under source",
			zipPath:     "repo-main/other/file.go",
			sourcePath:  "repo-main/src",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := zd.getRelativePath(tt.zipPath, tt.sourcePath)

			if tt.expectError {
				if err == nil {
					t.Errorf("getRelativePath(%q, %q) expected error, got nil", tt.zipPath, tt.sourcePath)
				}
				return
			}

			if err != nil {
				t.Errorf("getRelativePath(%q, %q) unexpected error: %v", tt.zipPath, tt.sourcePath, err)
				return
			}

			if result != tt.expected {
				t.Errorf("getRelativePath(%q, %q) = %q, expected %q", tt.zipPath, tt.sourcePath, result, tt.expected)
			}
		})
	}
}

func TestZipDownloader_extractPath(t *testing.T) {
	// Create a test zip file
	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "test.zip")

	// Create zip file with test structure
	createTestZip(t, zipPath)

	// Create stdout/stderr buffers
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	zd := NewZipDownloader(stdout, stderr)

	tests := []struct {
		name        string
		sourcePath  string
		expectError bool
		expectFiles []string
	}{
		{
			name:        "Extract entire repository",
			sourcePath:  "repo-main",
			expectError: false,
			expectFiles: []string{"README.md", "src/main.go", "docs/guide.md"},
		},
		{
			name:        "Extract specific directory",
			sourcePath:  "repo-main/src",
			expectError: false,
			expectFiles: []string{"main.go"},
		},
		{
			name:        "Extract specific file",
			sourcePath:  "repo-main/README.md",
			expectError: false,
			expectFiles: []string{"README.md"},
		},
		{
			name:        "Extract non-existent path",
			sourcePath:  "repo-main/nonexistent",
			expectError: true,
			expectFiles: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh target directory for each test
			targetDir := filepath.Join(tempDir, "target-"+strings.ReplaceAll(tt.name, " ", "-"))

			err := zd.extractPath(zipPath, tt.sourcePath, targetDir)

			if tt.expectError {
				if err == nil {
					t.Errorf("extractPath expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("extractPath unexpected error: %v", err)
				return
			}

			// Verify expected files exist
			for _, expectedFile := range tt.expectFiles {
				filePath := filepath.Join(targetDir, expectedFile)
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					t.Errorf("Expected file %s does not exist", expectedFile)
				}
			}
		})
	}
}

func TestDownloadRequest_validation(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	zd := NewZipDownloader(stdout, stderr)

	tempDir := t.TempDir()

	tests := []struct {
		name        string
		req         DownloadRequest
		expectError bool
	}{
		{
			name: "Valid request",
			req: DownloadRequest{
				Owner:  "owner",
				Repo:   "repo",
				Path:   "",
				Ref:    "main",
				Target: tempDir,
			},
			expectError: true, // Will fail because we're not actually downloading from GitHub
		},
		{
			name: "Default ref",
			req: DownloadRequest{
				Owner:  "owner",
				Repo:   "repo",
				Path:   "",
				Ref:    "", // Should default to "main"
				Target: tempDir,
			},
			expectError: true, // Will fail because we're not actually downloading from GitHub
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := zd.Download(tt.req)

			if tt.expectError {
				if err == nil {
					t.Errorf("Download expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Download unexpected error: %v", err)
				}
			}
		})
	}
}

func TestZipDownloader_checkDiskSpace(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	zd := NewZipDownloader(stdout, stderr)

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.zip")

	// Test with small file size (should not error)
	err := zd.checkDiskSpace(filePath, 1024)
	if err != nil {
		t.Errorf("checkDiskSpace with small size unexpected error: %v", err)
	}

	// Test with large file size (should warn but not error)
	stderr.Reset()
	err = zd.checkDiskSpace(filePath, 2<<30) // 2GB
	if err != nil {
		t.Errorf("checkDiskSpace with large size unexpected error: %v", err)
	}

	// Should have printed a warning
	if !strings.Contains(stderr.String(), "Warning") {
		t.Errorf("Expected warning for large file, got: %s", stderr.String())
	}
}

func TestNewZipDownloader(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	zd := NewZipDownloader(stdout, stderr)
	if zd == nil {
		t.Error("NewZipDownloader returned nil")
	}

	if zd.stdout != stdout {
		t.Error("stdout not set correctly")
	}

	if zd.stderr != stderr {
		t.Error("stderr not set correctly")
	}

	if zd.tempDir != os.TempDir() {
		t.Errorf("tempDir = %s, expected %s", zd.tempDir, os.TempDir())
	}
}

func TestNewZipDownloaderWithTempDir(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	customTempDir := "/tmp/custom"

	zd := NewZipDownloaderWithTempDir(customTempDir, stdout, stderr)
	if zd == nil {
		t.Error("NewZipDownloaderWithTempDir returned nil")
	}

	if zd.tempDir != customTempDir {
		t.Errorf("tempDir = %s, expected %s", zd.tempDir, customTempDir)
	}
}

// createTestZip creates a test zip file with a predictable structure
func createTestZip(t *testing.T, zipPath string) {
	t.Helper()

	// Create zip file
	file, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}
	defer file.Close()

	writer := zip.NewWriter(file)
	defer writer.Close()

	// Add test files to zip
	files := map[string]string{
		"repo-main/README.md":     "# Test Repository\n",
		"repo-main/src/main.go":   "package main\n\nfunc main() {}\n",
		"repo-main/docs/guide.md": "# Guide\n\nThis is a guide.\n",
	}

	for name, content := range files {
		fw, err := writer.Create(name)
		if err != nil {
			t.Fatalf("Failed to create file in zip: %v", err)
		}

		if !strings.HasSuffix(name, "/") { // Only write content for files, not directories
			_, err = io.WriteString(fw, content)
			if err != nil {
				t.Fatalf("Failed to write file content: %v", err)
			}
		}
	}
}
