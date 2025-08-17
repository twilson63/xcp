package downloader

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"xcp/internal/github"
)

var (
	ErrZipDownloadFailed     = errors.New("failed to download zip archive")
	ErrZipExtractFailed      = errors.New("failed to extract zip archive")
	ErrPathNotFoundInZip     = errors.New("path not found in zip archive")
	ErrInvalidZipPath        = errors.New("invalid path in zip archive")
	ErrDiskSpaceInsufficient = errors.New("insufficient disk space")
)

// ZipDownloader downloads GitHub repositories as zip archives
type ZipDownloader struct {
	httpClient *http.Client
	tempDir    string
	stdout     io.Writer
	stderr     io.Writer
}

// DownloadRequest contains the parameters for a zip download
type DownloadRequest struct {
	Owner  string
	Repo   string
	Path   string // Optional: specific path within repo
	Ref    string // Branch, tag, or commit (default: main)
	Target string // Local target directory
}

// NewZipDownloader creates a new ZipDownloader
func NewZipDownloader(stdout, stderr io.Writer) *ZipDownloader {
	return &ZipDownloader{
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // Longer timeout for large repositories
		},
		tempDir: os.TempDir(),
		stdout:  stdout,
		stderr:  stderr,
	}
}

// NewZipDownloaderWithTempDir creates a new ZipDownloader with custom temp directory
func NewZipDownloaderWithTempDir(tempDir string, stdout, stderr io.Writer) *ZipDownloader {
	return &ZipDownloader{
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
		tempDir: tempDir,
		stdout:  stdout,
		stderr:  stderr,
	}
}

// Download downloads a repository using the zip method
func (zd *ZipDownloader) Download(req DownloadRequest) error {
	// Default ref to main if not specified
	if req.Ref == "" {
		req.Ref = "main"
	}

	// Build zip URL
	zipURL := fmt.Sprintf("https://github.com/%s/%s/archive/%s.zip", req.Owner, req.Repo, req.Ref)

	// Download zip file
	zipPath, err := zd.downloadZip(zipURL)
	if err != nil {
		return fmt.Errorf("failed to download repository zip: %w", err)
	}

	// Ensure cleanup
	defer func() {
		if err := os.Remove(zipPath); err != nil {
			fmt.Fprintf(zd.stderr, "Warning: failed to clean up zip file %s: %v\n", zipPath, err)
		}
	}()

	// Extract specific path or entire repository
	repoPrefix := fmt.Sprintf("%s-%s", req.Repo, req.Ref)
	sourcePath := req.Path
	if sourcePath != "" {
		sourcePath = filepath.Join(repoPrefix, req.Path)
	} else {
		sourcePath = repoPrefix
	}

	err = zd.extractPath(zipPath, sourcePath, req.Target)
	if err != nil {
		return fmt.Errorf("failed to extract path from zip: %w", err)
	}

	fmt.Fprintf(zd.stderr, "Successfully downloaded %s/%s to %s\n", req.Owner, req.Repo, req.Target)
	return nil
}

// downloadZip downloads a zip file from the given URL and returns the local path
func (zd *ZipDownloader) downloadZip(url string) (string, error) {
	// Create HTTP request
	resp, err := zd.httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("%w: network error: %v", ErrZipDownloadFailed, err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("%w: repository or reference not found (404)", ErrZipDownloadFailed)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: unexpected status code %d", ErrZipDownloadFailed, resp.StatusCode)
	}

	// Create temporary file
	tempFile, err := os.CreateTemp(zd.tempDir, "xcp-download-*.zip")
	if err != nil {
		return "", fmt.Errorf("%w: failed to create temp file: %v", ErrZipDownloadFailed, err)
	}
	defer tempFile.Close()

	// Check available disk space (simple heuristic)
	if resp.ContentLength > 0 {
		if err := zd.checkDiskSpace(tempFile.Name(), resp.ContentLength); err != nil {
			os.Remove(tempFile.Name())
			return "", err
		}
	}

	// Copy response body to file
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("%w: failed to write zip file: %v", ErrZipDownloadFailed, err)
	}

	return tempFile.Name(), nil
}

// extractPath extracts a specific path from the zip archive to the target directory
func (zd *ZipDownloader) extractPath(zipPath, sourcePath, targetPath string) error {
	// Open zip file
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("%w: failed to open zip file: %v", ErrZipExtractFailed, err)
	}
	defer reader.Close()

	// Ensure target directory exists
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return fmt.Errorf("%w: failed to create target directory: %v", ErrZipExtractFailed, err)
	}

	found := false
	extractedCount := 0

	// Process each file in the zip
	for _, file := range reader.File {
		// Check if this file matches our source path
		if !zd.pathMatches(file.Name, sourcePath) {
			continue
		}

		found = true

		// Calculate relative path from source to target
		relPath, err := zd.getRelativePath(file.Name, sourcePath)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidZipPath, err)
		}

		// Skip if this is the source directory itself (not its contents)
		if relPath == "" && file.FileInfo().IsDir() {
			continue
		}

		// Build target file path - handle the case where we're extracting a single file
		var targetFilePath string
		if relPath == "" {
			// This is the exact file we want to extract
			targetFilePath = filepath.Join(targetPath, filepath.Base(file.Name))
		} else {
			targetFilePath = filepath.Join(targetPath, relPath)
		}

		// Validate path to prevent zip slip attacks
		cleanTarget := filepath.Clean(targetPath)
		cleanTargetFile := filepath.Clean(targetFilePath)
		if !strings.HasPrefix(cleanTargetFile, cleanTarget+string(os.PathSeparator)) &&
			cleanTargetFile != cleanTarget {
			return fmt.Errorf("%w: path traversal attempt: %s", ErrInvalidZipPath, file.Name)
		}

		// Extract file or directory
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetFilePath, file.FileInfo().Mode()); err != nil {
				return fmt.Errorf("%w: failed to create directory %s: %v", ErrZipExtractFailed, targetFilePath, err)
			}
		} else {
			if err := zd.extractFile(file, targetFilePath); err != nil {
				return fmt.Errorf("%w: failed to extract file %s: %v", ErrZipExtractFailed, file.Name, err)
			}
			extractedCount++
		}
	}

	if !found {
		return fmt.Errorf("%w: path '%s' not found in repository", ErrPathNotFoundInZip, sourcePath)
	}

	if extractedCount > 0 {
		fmt.Fprintf(zd.stderr, "Extracted %d files\n", extractedCount)
	}

	return nil
}

// pathMatches checks if a zip file path matches the source path we want to extract
func (zd *ZipDownloader) pathMatches(zipPath, sourcePath string) bool {
	// Normalize paths
	zipPath = filepath.ToSlash(zipPath)
	sourcePath = filepath.ToSlash(sourcePath)

	// Exact match
	if zipPath == sourcePath {
		return true
	}

	// Check if zipPath is under sourcePath (for directory extraction)
	if strings.HasPrefix(zipPath, sourcePath+"/") {
		return true
	}

	return false
}

// getRelativePath calculates the relative path from sourcePath to zipPath
func (zd *ZipDownloader) getRelativePath(zipPath, sourcePath string) (string, error) {
	// Normalize paths
	zipPath = filepath.ToSlash(zipPath)
	sourcePath = filepath.ToSlash(sourcePath)

	// If exact match, return empty (this is the source itself)
	if zipPath == sourcePath {
		return "", nil
	}

	// If zipPath is under sourcePath, return the relative part
	if strings.HasPrefix(zipPath, sourcePath+"/") {
		return strings.TrimPrefix(zipPath, sourcePath+"/"), nil
	}

	return "", fmt.Errorf("path %s is not under source path %s", zipPath, sourcePath)
}

// extractFile extracts a single file from the zip archive
func (zd *ZipDownloader) extractFile(file *zip.File, targetPath string) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %v", err)
	}

	// Open file in zip
	rc, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file in zip: %v", err)
	}
	defer rc.Close()

	// Create target file
	outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.FileInfo().Mode())
	if err != nil {
		return fmt.Errorf("failed to create target file: %v", err)
	}
	defer outFile.Close()

	// Copy content
	_, err = io.Copy(outFile, rc)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %v", err)
	}

	return nil
}

// checkDiskSpace performs a basic check for available disk space
func (zd *ZipDownloader) checkDiskSpace(filePath string, requiredBytes int64) error {
	// Get file system stats
	var stat os.FileInfo
	var err error

	// Try to get stats of the directory
	dir := filepath.Dir(filePath)
	stat, err = os.Stat(dir)
	if err != nil {
		return fmt.Errorf("failed to check disk space: %v", err)
	}

	// This is a basic check - in a real implementation, you might want to use
	// syscalls to get actual available space. For now, we'll just check if we
	// can create the file.
	_ = stat

	// Simple heuristic: if the required size is very large (>1GB), warn the user
	if requiredBytes > 1<<30 {
		fmt.Fprintf(zd.stderr, "Warning: downloading large repository (%.1f MB)\n", float64(requiredBytes)/(1<<20))
	}

	return nil
}

// DownloadFromSource downloads using a GitHubSource (adapter for existing interface)
func (zd *ZipDownloader) DownloadFromSource(source *github.GitHubSource, targetPath string, ref string) error {
	req := DownloadRequest{
		Owner:  source.Owner,
		Repo:   source.Repo,
		Path:   source.Path,
		Ref:    ref,
		Target: targetPath,
	}

	return zd.Download(req)
}
