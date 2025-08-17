package github

import (
	"testing"
)

func TestParseGitHubURL(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		expectedOwner string
		expectedRepo  string
		expectedPath  string
		expectedFile  bool
		expectedErr   error
	}{
		{
			name:          "Valid URL with path",
			url:           "github:twilson63/foo/data.json",
			expectedOwner: "twilson63",
			expectedRepo:  "foo",
			expectedPath:  "data.json",
			expectedFile:  true,
			expectedErr:   nil,
		},
		{
			name:          "Valid URL without path",
			url:           "github:twilson63/foo",
			expectedOwner: "twilson63",
			expectedRepo:  "foo",
			expectedPath:  "",
			expectedFile:  false,
			expectedErr:   nil,
		},
		{
			name:          "Valid URL with directory path",
			url:           "github:twilson63/foo/dir/",
			expectedOwner: "twilson63",
			expectedRepo:  "foo",
			expectedPath:  "dir/",
			expectedFile:  false,
			expectedErr:   nil,
		},
		{
			name:          "Invalid URL format",
			url:           "githubtwilight/foo",
			expectedOwner: "",
			expectedRepo:  "",
			expectedPath:  "",
			expectedFile:  false,
			expectedErr:   ErrInvalidURL,
		},
		{
			name:          "Missing repo",
			url:           "github:twilight",
			expectedOwner: "",
			expectedRepo:  "",
			expectedPath:  "",
			expectedFile:  false,
			expectedErr:   ErrInvalidURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, err := ParseGitHubURL(tt.url)

			if tt.expectedErr != nil {
				if err == nil {
					t.Errorf("Expected error %v, got nil", tt.expectedErr)
				} else if err != tt.expectedErr {
					t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if source.Owner != tt.expectedOwner {
				t.Errorf("Expected owner %s, got %s", tt.expectedOwner, source.Owner)
			}

			if source.Repo != tt.expectedRepo {
				t.Errorf("Expected repo %s, got %s", tt.expectedRepo, source.Repo)
			}

			if source.Path != tt.expectedPath {
				t.Errorf("Expected path %s, got %s", tt.expectedPath, source.Path)
			}

			if source.IsFile != tt.expectedFile {
				t.Errorf("Expected isFile %v, got %v", tt.expectedFile, source.IsFile)
			}
		})
	}
}

func TestParseGitHubURLWithRef(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		expectedOwner string
		expectedRepo  string
		expectedPath  string
		expectedRef   string
		expectedErr   error
	}{
		{
			name:          "Simple repo",
			url:           "github:twilson63/qa",
			expectedOwner: "twilson63",
			expectedRepo:  "qa",
			expectedPath:  "",
			expectedRef:   "main",
			expectedErr:   nil,
		},
		{
			name:          "Repo with branch",
			url:           "github:twilson63/qa@develop",
			expectedOwner: "twilson63",
			expectedRepo:  "qa",
			expectedPath:  "",
			expectedRef:   "develop",
			expectedErr:   nil,
		},
		{
			name:          "Repo with tag",
			url:           "github:twilson63/qa@v1.0.0",
			expectedOwner: "twilson63",
			expectedRepo:  "qa",
			expectedPath:  "",
			expectedRef:   "v1.0.0",
			expectedErr:   nil,
		},
		{
			name:          "Repo with commit hash",
			url:           "github:twilson63/qa@abc123def456",
			expectedOwner: "twilson63",
			expectedRepo:  "qa",
			expectedPath:  "",
			expectedRef:   "abc123def456",
			expectedErr:   nil,
		},
		{
			name:          "File with branch",
			url:           "github:twilson63/qa@develop/src/data.json",
			expectedOwner: "twilson63",
			expectedRepo:  "qa",
			expectedPath:  "src/data.json",
			expectedRef:   "develop",
			expectedErr:   nil,
		},
		{
			name:          "Directory with branch",
			url:           "github:twilson63/qa@main/src/",
			expectedOwner: "twilson63",
			expectedRepo:  "qa",
			expectedPath:  "src/",
			expectedRef:   "main",
			expectedErr:   nil,
		},
		{
			name:          "Path with ref at end",
			url:           "github:twilson63/qa/src/data.json@feature-branch",
			expectedOwner: "twilson63",
			expectedRepo:  "qa",
			expectedPath:  "src/data.json",
			expectedRef:   "feature-branch",
			expectedErr:   nil,
		},
		{
			name:          "Invalid URL format",
			url:           "githubtwilight/foo",
			expectedOwner: "",
			expectedRepo:  "",
			expectedPath:  "",
			expectedRef:   "",
			expectedErr:   ErrInvalidURL,
		},
		{
			name:          "Missing repo",
			url:           "github:twilight",
			expectedOwner: "",
			expectedRepo:  "",
			expectedPath:  "",
			expectedRef:   "",
			expectedErr:   ErrInvalidURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := ParseGitHubURLWithRef(tt.url)

			if tt.expectedErr != nil {
				if err == nil {
					t.Errorf("Expected error %v, got nil", tt.expectedErr)
				} else if err != tt.expectedErr {
					t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if parsed.Owner != tt.expectedOwner {
				t.Errorf("Expected owner %s, got %s", tt.expectedOwner, parsed.Owner)
			}

			if parsed.Repo != tt.expectedRepo {
				t.Errorf("Expected repo %s, got %s", tt.expectedRepo, parsed.Repo)
			}

			if parsed.Path != tt.expectedPath {
				t.Errorf("Expected path %s, got %s", tt.expectedPath, parsed.Path)
			}

			if parsed.Ref != tt.expectedRef {
				t.Errorf("Expected ref %s, got %s", tt.expectedRef, parsed.Ref)
			}
		})
	}
}

func TestParsedURL_ZipURL(t *testing.T) {
	tests := []struct {
		name        string
		parsed      *ParsedURL
		expectedURL string
	}{
		{
			name: "Main branch",
			parsed: &ParsedURL{
				Owner: "twilson63",
				Repo:  "qa",
				Ref:   "main",
			},
			expectedURL: "https://github.com/twilson63/qa/archive/main.zip",
		},
		{
			name: "Feature branch",
			parsed: &ParsedURL{
				Owner: "twilson63",
				Repo:  "qa",
				Ref:   "feature-branch",
			},
			expectedURL: "https://github.com/twilson63/qa/archive/feature-branch.zip",
		},
		{
			name: "Tag",
			parsed: &ParsedURL{
				Owner: "twilson63",
				Repo:  "qa",
				Ref:   "v1.0.0",
			},
			expectedURL: "https://github.com/twilson63/qa/archive/v1.0.0.zip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := tt.parsed.ZipURL()
			if url != tt.expectedURL {
				t.Errorf("Expected URL %s, got %s", tt.expectedURL, url)
			}
		})
	}
}

func TestParsedURL_IsFile(t *testing.T) {
	tests := []struct {
		name     string
		parsed   *ParsedURL
		expected bool
	}{
		{
			name: "No path",
			parsed: &ParsedURL{
				Path: "",
			},
			expected: false,
		},
		{
			name: "Directory with trailing slash",
			parsed: &ParsedURL{
				Path: "src/",
			},
			expected: false,
		},
		{
			name: "File with extension",
			parsed: &ParsedURL{
				Path: "src/data.json",
			},
			expected: true,
		},
		{
			name: "File without extension",
			parsed: &ParsedURL{
				Path: "README",
			},
			expected: false,
		},
		{
			name: "Hidden file",
			parsed: &ParsedURL{
				Path: ".gitignore",
			},
			expected: true,
		},
		{
			name: "Directory path",
			parsed: &ParsedURL{
				Path: "src/components",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.parsed.IsFile()
			if result != tt.expected {
				t.Errorf("Expected IsFile() %v, got %v for path %s", tt.expected, result, tt.parsed.Path)
			}
		})
	}
}

func TestParsedURL_IsDirectory(t *testing.T) {
	tests := []struct {
		name     string
		parsed   *ParsedURL
		expected bool
	}{
		{
			name: "No path",
			parsed: &ParsedURL{
				Path: "",
			},
			expected: true,
		},
		{
			name: "Directory with trailing slash",
			parsed: &ParsedURL{
				Path: "src/",
			},
			expected: true,
		},
		{
			name: "File with extension",
			parsed: &ParsedURL{
				Path: "src/data.json",
			},
			expected: false,
		},
		{
			name: "Directory path",
			parsed: &ParsedURL{
				Path: "src/components",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.parsed.IsDirectory()
			if result != tt.expected {
				t.Errorf("Expected IsDirectory() %v, got %v for path %s", tt.expected, result, tt.parsed.Path)
			}
		})
	}
}
