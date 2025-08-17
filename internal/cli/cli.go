package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"xcp/internal/downloader"
	"xcp/internal/github"
)

const (
	version = "0.1.0"
)

var (
	ErrMissingSource = errors.New("source parameter is required")
	ErrInvalidArgs   = errors.New("invalid command-line arguments")
)

// Downloader interface for downloading content
type Downloader interface {
	Download(source *github.GitHubSource, destPath string, opts downloader.DownloadOptions) error
}

// CLI represents the command-line interface
type CLI struct {
	flagSet    *flag.FlagSet
	stdout     io.Writer
	stderr     io.Writer
	downloader Downloader

	// Command-line flags
	showVersion bool
	showHelp    bool
	overwrite   bool
	method      string
	tempDir     string
	verbose     bool
}

// Options for configuring the CLI
type Options struct {
	Args       []string
	Stdout     io.Writer
	Stderr     io.Writer
	Downloader Downloader
}

// New creates a new CLI instance
func New(opts Options) *CLI {
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}

	cli := &CLI{
		flagSet:    flag.NewFlagSet("xcp", flag.ContinueOnError),
		stdout:     opts.Stdout,
		stderr:     opts.Stderr,
		downloader: opts.Downloader,
	}

	cli.flagSet.SetOutput(opts.Stderr)
	cli.flagSet.BoolVar(&cli.showVersion, "version", false, "Show version information")
	cli.flagSet.BoolVar(&cli.showVersion, "v", false, "Show version information (shorthand)")
	cli.flagSet.BoolVar(&cli.showHelp, "help", false, "Show help information")
	cli.flagSet.BoolVar(&cli.showHelp, "h", false, "Show help information (shorthand)")
	cli.flagSet.BoolVar(&cli.overwrite, "overwrite", false, "Overwrite existing files")
	cli.flagSet.BoolVar(&cli.overwrite, "f", false, "Overwrite existing files (shorthand)")
	cli.flagSet.StringVar(&cli.method, "method", "zip", "Download method: zip (default) or api")
	cli.flagSet.StringVar(&cli.tempDir, "temp-dir", "", "Custom temporary directory for zip extraction")
	cli.flagSet.BoolVar(&cli.verbose, "verbose", false, "Enable verbose output")

	return cli
}

// Run executes the CLI with the provided arguments
func (c *CLI) Run(args []string) error {
	if err := c.flagSet.Parse(args); err != nil {
		return err
	}

	if c.showVersion {
		fmt.Fprintf(c.stdout, "xcp version %s\n", version)
		return nil
	}

	if c.showHelp {
		c.printHelp()
		return nil
	}

	// Get non-flag arguments
	args = c.flagSet.Args()
	if len(args) == 0 {
		c.printHelp()
		return ErrMissingSource
	}

	// First argument is always the source
	sourceURL := args[0]

	// Parse GitHub URL
	source, err := github.ParseGitHubURL(sourceURL)
	if err != nil {
		return fmt.Errorf("invalid source URL: %w", err)
	}

	// Also parse with enhanced parser for zip downloader
	parsedURL, err := github.ParseGitHubURLWithRef(sourceURL)
	if err != nil {
		return fmt.Errorf("invalid source URL: %w", err)
	}

	// Determine target path
	var targetPath string
	outputToStdout := false

	if len(args) > 1 {
		targetPath = args[1]
	} else {
		// If no target path is provided and it's a file, output to stdout
		if source.IsFile {
			outputToStdout = true
		} else {
			// For directories, use current directory
			targetPath = "."
		}
	}

	// If target path is provided and the source is a file, append filename if target is a directory
	if targetPath != "" && source.IsFile {
		if stat, err := os.Stat(targetPath); err == nil && stat.IsDir() {
			// Target is an existing directory, append filename
			filename := filepath.Base(source.Path)
			targetPath = filepath.Join(targetPath, filename)
		}
	}

	// Set download options
	opts := downloader.DownloadOptions{
		OutputToStdout: outputToStdout,
		Overwrite:      c.overwrite,
	}

	// Use zip downloader for new method (only if no custom downloader provided)
	if c.method == "zip" && c.downloader == nil {
		var zipDownloader *downloader.ZipDownloader
		if c.tempDir != "" {
			zipDownloader = downloader.NewZipDownloaderWithTempDir(c.tempDir, c.stdout, c.stderr)
		} else {
			zipDownloader = downloader.NewZipDownloader(c.stdout, c.stderr)
		}

		// Create download request from parsed URL
		req := downloader.DownloadRequest{
			Owner:  parsedURL.Owner,
			Repo:   parsedURL.Repo,
			Path:   parsedURL.Path,
			Ref:    parsedURL.Ref,
			Target: targetPath,
		}

		return zipDownloader.Download(req)
	}

	// Create default API downloader if none provided
	if c.downloader == nil {
		client := github.NewClient()
		c.downloader = downloader.NewDownloader(client, c.stdout, c.stderr)
	}

	// Use the provided downloader (for tests) or fallback to API downloader
	return c.downloader.Download(source, targetPath, opts)
}

// printHelp displays the help information
func (c *CLI) printHelp() {
	fmt.Fprintln(c.stderr, "xcp - External Copy Program")
	fmt.Fprintln(c.stderr, "Copy files from GitHub repositories to local directories")
	fmt.Fprintln(c.stderr)
	fmt.Fprintln(c.stderr, "Usage:")
	fmt.Fprintln(c.stderr, "  xcp [options] <source> [target]")
	fmt.Fprintln(c.stderr)
	fmt.Fprintln(c.stderr, "Arguments:")
	fmt.Fprintln(c.stderr, "  source:  github:owner/repo/path[@ref]")
	fmt.Fprintln(c.stderr, "  target:  local directory or file (defaults to current directory)")
	fmt.Fprintln(c.stderr)
	fmt.Fprintln(c.stderr, "Options:")
	c.flagSet.PrintDefaults()
	fmt.Fprintln(c.stderr)
	fmt.Fprintln(c.stderr, "Examples:")
	fmt.Fprintln(c.stderr, "  xcp github:twilson63/qa")
	fmt.Fprintln(c.stderr, "  xcp github:twilson63/qa@main")
	fmt.Fprintln(c.stderr, "  xcp github:twilson63/qa@v1.0.0")
	fmt.Fprintln(c.stderr, "  xcp github:twilson63/foo/data.json | jq")
	fmt.Fprintln(c.stderr, "  xcp github:twilson63/qa ./target/path")
	fmt.Fprintln(c.stderr, "  xcp --method=api github:twilson63/qa")
	fmt.Fprintln(c.stderr, "  xcp --verbose --temp-dir=/tmp github:twilson63/qa")
}
