package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"xcp/internal/downloader"
	"xcp/internal/github"
)

// MockDownloader for testing
type MockDownloader struct {
	Source *github.GitHubSource
	Target string
	Opts   downloader.DownloadOptions
	Err    error
}

// Download implements the Downloader interface
func (m *MockDownloader) Download(source *github.GitHubSource, target string, opts downloader.DownloadOptions) error {
	m.Source = source
	m.Target = target
	m.Opts = opts
	return m.Err
}

func TestCLI_Version(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	cli := New(Options{
		Args:   []string{"-v"},
		Stdout: stdout,
		Stderr: stderr,
	})

	err := cli.Run([]string{"-v"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "xcp version") {
		t.Errorf("Expected version output to contain 'xcp version', got %q", out)
	}
}

func TestCLI_Help(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	cli := New(Options{
		Args:   []string{"-h"},
		Stdout: stdout,
		Stderr: stderr,
	})

	err := cli.Run([]string{"-h"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	out := stderr.String()
	if !strings.Contains(out, "Usage:") {
		t.Errorf("Expected help output to contain 'Usage:', got %q", out)
	}
}

func TestCLI_MissingSource(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	cli := New(Options{
		Args:   []string{},
		Stdout: stdout,
		Stderr: stderr,
	})

	err := cli.Run([]string{})
	if err != ErrMissingSource {
		t.Errorf("Expected ErrMissingSource, got %v", err)
	}

	out := stderr.String()
	if !strings.Contains(out, "Usage:") {
		t.Errorf("Expected help output to contain 'Usage:', got %q", out)
	}
}

func TestCLI_InvalidSource(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	cli := New(Options{
		Args:   []string{"invalid-source"},
		Stdout: stdout,
		Stderr: stderr,
	})

	err := cli.Run([]string{"invalid-source"})
	if err == nil {
		t.Errorf("Expected error for invalid source")
	}
}

func TestCLI_ParseArgs(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		downloaderErr   error
		expectError     bool
		expectSource    string
		expectTarget    string
		expectToStdout  bool
		expectOverwrite bool
	}{
		{
			name:           "Valid source and target",
			args:           []string{"github:owner/repo", "/target/path"},
			expectError:    false,
			expectSource:   "owner/repo",
			expectTarget:   "/target/path",
			expectToStdout: false,
		},
		{
			name:           "Valid source only (directory)",
			args:           []string{"github:owner/repo"},
			expectError:    false,
			expectSource:   "owner/repo",
			expectTarget:   ".",
			expectToStdout: false,
		},
		{
			name:           "Valid source with file path",
			args:           []string{"github:owner/repo/file.txt"},
			expectError:    false,
			expectSource:   "owner/repo/file.txt",
			expectTarget:   "",
			expectToStdout: true,
		},
		{
			name:            "Overwrite flag",
			args:            []string{"-f", "github:owner/repo", "/target/path"},
			expectError:     false,
			expectSource:    "owner/repo",
			expectTarget:    "/target/path",
			expectOverwrite: true,
		},
		{
			name:          "Downloader error",
			args:          []string{"github:owner/repo"},
			downloaderErr: errors.New("download failed"),
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout := new(bytes.Buffer)
			stderr := new(bytes.Buffer)
			mock := &MockDownloader{Err: tt.downloaderErr}

			cli := New(Options{
				Args:       tt.args,
				Stdout:     stdout,
				Stderr:     stderr,
				Downloader: mock,
			})

			err := cli.Run(tt.args)

			if tt.expectError && err == nil {
				t.Errorf("Expected error, got nil")
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if tt.expectError {
				return
			}

			// Verify the source
			if tt.expectSource != "" {
				if mock.Source == nil {
					t.Errorf("Expected source to be set, got nil")
					return
				}

				gotSource := mock.Source.Owner + "/" + mock.Source.Repo
				if mock.Source.Path != "" {
					gotSource += "/" + mock.Source.Path
				}

				if !strings.HasPrefix(gotSource, tt.expectSource) {
					t.Errorf("Expected source %q, got %q", tt.expectSource, gotSource)
				}
			}

			// Verify the target
			if tt.expectTarget != "" && mock.Target != tt.expectTarget {
				t.Errorf("Expected target %q, got %q", tt.expectTarget, mock.Target)
			}

			// Verify stdout option
			if tt.expectToStdout != mock.Opts.OutputToStdout {
				t.Errorf("Expected OutputToStdout %v, got %v", tt.expectToStdout, mock.Opts.OutputToStdout)
			}

			// Verify overwrite option
			if tt.expectOverwrite != mock.Opts.Overwrite {
				t.Errorf("Expected Overwrite %v, got %v", tt.expectOverwrite, mock.Opts.Overwrite)
			}
		})
	}
}
