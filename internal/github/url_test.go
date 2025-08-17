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
