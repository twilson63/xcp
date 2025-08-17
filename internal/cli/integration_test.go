package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestCLI_ZipDownloadIntegration tests the zip download functionality with a real (small) repository
func TestCLI_ZipDownloadIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "xcp-integration-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	// Test downloading a small public repository using zip method
	cli := New(Options{
		Stdout: stdout,
		Stderr: stderr,
	})

	// Use a small, stable repository for testing
	// This repo should exist and be small to avoid test flakiness
	targetPath := filepath.Join(tmpDir, "test-repo")
	args := []string{"--method=zip", "--verbose", "github:octocat/Hello-World", targetPath}

	err = cli.Run(args)
	if err != nil {
		t.Logf("stderr: %s", stderr.String())
		t.Logf("stdout: %s", stdout.String())
		t.Fatalf("CLI run failed: %v", err)
	}

	// Verify that files were downloaded
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Errorf("Target directory was not created")
	}

	// Check for at least one file in the downloaded directory
	entries, err := os.ReadDir(targetPath)
	if err != nil {
		t.Fatalf("Failed to read target directory: %v", err)
	}

	if len(entries) == 0 {
		t.Errorf("No files were downloaded to target directory")
	}

	t.Logf("Successfully downloaded %d entries", len(entries))
	t.Logf("stderr output: %s", stderr.String())
}

// TestCLI_ZipDownloadWithRef tests downloading from a specific branch/tag
func TestCLI_ZipDownloadWithRef(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "xcp-ref-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	cli := New(Options{
		Stdout: stdout,
		Stderr: stderr,
	})

	// Test with a specific branch/ref - using main branch explicitly
	targetPath := filepath.Join(tmpDir, "test-repo-main")
	args := []string{"--method=zip", "github:octocat/Hello-World@main", targetPath}

	err = cli.Run(args)
	if err != nil {
		t.Logf("stderr: %s", stderr.String())
		t.Fatalf("CLI run with ref failed: %v", err)
	}

	// Verify that files were downloaded
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Errorf("Target directory was not created")
	}

	t.Logf("Successfully downloaded with ref @main")
}

// TestCLI_APIFallback tests the API fallback method
func TestCLI_APIFallback(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "xcp-api-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	cli := New(Options{
		Stdout: stdout,
		Stderr: stderr,
	})

	// Test using API method explicitly
	targetPath := filepath.Join(tmpDir, "test-repo-api")
	args := []string{"--method=api", "github:octocat/Hello-World", targetPath}

	err = cli.Run(args)
	if err != nil {
		t.Logf("stderr: %s", stderr.String())
		t.Fatalf("CLI run with API method failed: %v", err)
	}

	// Verify that files were downloaded
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Errorf("Target directory was not created")
	}

	t.Logf("Successfully downloaded with API method")
}
