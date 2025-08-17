package downloader

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"xcp/internal/github"
)

var (
	ErrFailedToCreateDir  = errors.New("failed to create directory")
	ErrFailedToWriteFile  = errors.New("failed to write file")
	ErrNoContentToWrite   = errors.New("no content to write")
	ErrInvalidDestination = errors.New("invalid destination path")
)

// GitHubClient interface for GitHub API operations
type GitHubClient interface {
	GetFileContent(owner, repo, path string) ([]byte, error)
	GetDirectoryContents(owner, repo, path string) (github.DirectoryContents, error)
	RepositoryExists(owner, repo string) (bool, error)
}

// Downloader is responsible for downloading files from GitHub
type Downloader struct {
	client GitHubClient
	stdout io.Writer
	stderr io.Writer
}

// DownloadOptions configures how files are downloaded
type DownloadOptions struct {
	OutputToStdout bool
	Overwrite      bool
}

// NewDownloader creates a new Downloader
func NewDownloader(client GitHubClient, stdout, stderr io.Writer) *Downloader {
	return &Downloader{
		client: client,
		stdout: stdout,
		stderr: stderr,
	}
}

// DownloadFile downloads a single file from GitHub
func (d *Downloader) DownloadFile(source *github.GitHubSource, destPath string, opts DownloadOptions) error {
	// Get file content from GitHub
	content, err := d.client.GetFileContent(source.Owner, source.Repo, source.Path)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	if opts.OutputToStdout {
		_, err := d.stdout.Write(content)
		return err
	}

	// Create destination directory if it doesn't exist
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("%w: %s: %v", ErrFailedToCreateDir, destDir, err)
	}

	// Check if file exists
	if !opts.Overwrite {
		if _, err := os.Stat(destPath); err == nil {
			return fmt.Errorf("file already exists: %s", destPath)
		}
	}

	// Write file to destination
	if err := os.WriteFile(destPath, content, 0644); err != nil {
		return fmt.Errorf("%w: %s: %v", ErrFailedToWriteFile, destPath, err)
	}

	fmt.Fprintf(d.stderr, "Downloaded %s to %s\n", source.Path, destPath)
	return nil
}

// DownloadDirectory recursively downloads a directory from GitHub
func (d *Downloader) DownloadDirectory(source *github.GitHubSource, destPath string, opts DownloadOptions) error {
	// Don't allow stdout for directories
	if opts.OutputToStdout {
		return errors.New("cannot output directory to stdout")
	}

	// Get directory contents from GitHub
	contents, err := d.client.GetDirectoryContents(source.Owner, source.Repo, source.Path)
	if err != nil {
		return fmt.Errorf("failed to list directory contents: %w", err)
	}

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return fmt.Errorf("%w: %s: %v", ErrFailedToCreateDir, destPath, err)
	}

	for _, item := range contents {
		itemDestPath := filepath.Join(destPath, item.Name)

		switch item.Type {
		case github.FileContent:
			// Create a new source for each file
			fileSource := &github.GitHubSource{
				Owner:  source.Owner,
				Repo:   source.Repo,
				Path:   item.Path,
				IsFile: true,
			}

			if err := d.DownloadFile(fileSource, itemDestPath, opts); err != nil {
				return err
			}

		case github.DirectoryContent:
			// Create a new source for each directory
			dirSource := &github.GitHubSource{
				Owner:  source.Owner,
				Repo:   source.Repo,
				Path:   item.Path,
				IsFile: false,
			}

			if err := d.DownloadDirectory(dirSource, itemDestPath, opts); err != nil {
				return err
			}

		default:
			fmt.Fprintf(d.stderr, "Skipping unknown content type: %s for %s\n", item.Type, item.Path)
		}
	}

	return nil
}

// Download handles downloading either a file or directory based on the source
func (d *Downloader) Download(source *github.GitHubSource, destPath string, opts DownloadOptions) error {
	// Validate destination path
	if destPath == "" && !opts.OutputToStdout {
		return ErrInvalidDestination
	}

	// Check if the repository exists
	exists, err := d.client.RepositoryExists(source.Owner, source.Repo)
	if err != nil {
		return fmt.Errorf("failed to check repository: %w", err)
	}

	if !exists {
		return fmt.Errorf("repository not found: %s/%s", source.Owner, source.Repo)
	}

	// If path is empty, download the entire repository
	if source.Path == "" {
		return d.DownloadDirectory(source, destPath, opts)
	}

	// Try to download as a file first
	fileErr := d.DownloadFile(source, destPath, opts)
	if fileErr == nil {
		return nil
	}

	// If it's not a file, try to download as a directory
	if errors.Is(fileErr, github.ErrFileNotFound) {
		return d.DownloadDirectory(source, destPath, opts)
	}

	// Return the original error
	return fileErr
}
