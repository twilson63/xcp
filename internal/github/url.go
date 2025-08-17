package github

import (
	"errors"
	"fmt"
	"strings"
)

// ParsedURL represents a fully parsed GitHub repository URL with ref support
type ParsedURL struct {
	Owner string
	Repo  string
	Path  string
	Ref   string
}

// GitHubSource represents a parsed GitHub repository source (for backward compatibility)
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

// ParseGitHubURL parses a GitHub URL in the format "github:owner/repo/path" or "github:owner/repo@ref/path"
func ParseGitHubURL(url string) (*GitHubSource, error) {
	parsed, err := ParseGitHubURLWithRef(url)
	if err != nil {
		return nil, err
	}

	// Convert to legacy GitHubSource for backward compatibility
	return &GitHubSource{
		Owner:  parsed.Owner,
		Repo:   parsed.Repo,
		Path:   parsed.Path,
		IsFile: parsed.IsFile(),
	}, nil
}

// ParseGitHubURLWithRef parses a GitHub URL with full ref support
// Supported formats:
//   - github:owner/repo
//   - github:owner/repo/path/to/file
//   - github:owner/repo@branch
//   - github:owner/repo@tag
//   - github:owner/repo@commit
//   - github:owner/repo@ref/path/to/file
//   - github:owner/repo/path@ref
func ParseGitHubURLWithRef(url string) (*ParsedURL, error) {
	if !strings.HasPrefix(url, "github:") {
		return nil, ErrInvalidURL
	}

	// Remove prefix
	urlPart := strings.TrimPrefix(url, "github:")

	// Split by @ to separate owner/repo/path from ref
	var ownerRepoPart, refPart string
	atIndex := strings.Index(urlPart, "@")

	if atIndex == -1 {
		// No @ found, default ref to "main"
		ownerRepoPart = urlPart
		refPart = "main"
	} else {
		ownerRepoPart = urlPart[:atIndex]
		refPart = urlPart[atIndex+1:]

		// Handle case where path comes after @ref
		// e.g., github:owner/repo/path@branch
		slashInRef := strings.Index(refPart, "/")
		if slashInRef != -1 {
			// Path is after the ref, move it to ownerRepoPart
			pathAfterRef := refPart[slashInRef:]
			refPart = refPart[:slashInRef]
			ownerRepoPart = ownerRepoPart + pathAfterRef
		}
	}

	// Parse owner/repo/path
	parts := strings.SplitN(ownerRepoPart, "/", 3)
	if len(parts) < 2 {
		return nil, ErrInvalidURL
	}

	owner := parts[0]
	repo := parts[1]
	path := ""

	if len(parts) > 2 {
		path = parts[2]
	}

	if owner == "" {
		return nil, ErrMissingOwner
	}

	if repo == "" {
		return nil, ErrMissingRepo
	}

	if refPart == "" {
		refPart = "main"
	}

	return &ParsedURL{
		Owner: owner,
		Repo:  repo,
		Path:  path,
		Ref:   refPart,
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

// ZipURL returns the GitHub zip download URL for this parsed URL
func (p *ParsedURL) ZipURL() string {
	return fmt.Sprintf("https://github.com/%s/%s/archive/%s.zip", p.Owner, p.Repo, p.Ref)
}

// IsFile returns true if the path appears to be a file (has an extension or doesn't end with /)
func (p *ParsedURL) IsFile() bool {
	if p.Path == "" {
		return false
	}

	// If path ends with /, it's definitely a directory
	if strings.HasSuffix(p.Path, "/") {
		return false
	}

	// If path contains a file extension, it's likely a file
	lastSlash := strings.LastIndex(p.Path, "/")
	fileName := p.Path
	if lastSlash != -1 {
		fileName = p.Path[lastSlash+1:]
	}

	// Consider it a file if it has an extension
	return strings.Contains(fileName, ".")
}

// IsDirectory returns true if the path appears to be a directory
func (p *ParsedURL) IsDirectory() bool {
	return !p.IsFile()
}

// FullRepoName returns the full repository name (owner/repo)
func (p *ParsedURL) FullRepoName() string {
	return p.Owner + "/" + p.Repo
}

// APIPath returns the path suitable for GitHub API calls
func (p *ParsedURL) APIPath() string {
	return p.Path
}

// String returns a string representation of the parsed URL
func (p *ParsedURL) String() string {
	base := fmt.Sprintf("github:%s/%s", p.Owner, p.Repo)

	if p.Path != "" && p.Ref != "main" {
		return fmt.Sprintf("%s@%s/%s", base, p.Ref, p.Path)
	} else if p.Path != "" {
		return fmt.Sprintf("%s/%s", base, p.Path)
	} else if p.Ref != "main" {
		return fmt.Sprintf("%s@%s", base, p.Ref)
	}

	return base
}
