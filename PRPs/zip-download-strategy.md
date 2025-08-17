# xcp Zip Download Strategy

## Overview

This document outlines a strategic change to xcp's file download mechanism to address GitHub API rate limit issues by switching from individual file downloads via the GitHub API to downloading repository zip archives and extracting specific files.

## Problem Statement

### Current Issues
- **GitHub API Rate Limits**: The current implementation downloads each file individually via the GitHub API, which quickly exhausts the rate limit (60 requests/hour for unauthenticated, 5000/hour for authenticated)
- **Performance**: Multiple API calls for copying directories with many files
- **Reliability**: Users frequently encounter "GitHub API rate limit exceeded" errors
- **User Experience**: Inconsistent behavior depending on API usage

### Current Implementation Analysis
The existing implementation in `internal/downloader/downloader.go` uses:
1. GitHub API to get file metadata
2. Individual API calls for each file content
3. Direct file-by-file copying to target directory

## Proposed Solution

### Zip-Based Download Strategy
Switch to downloading the entire repository as a zip archive and extracting only the required files/directories.

### Key Benefits
1. **Rate Limit Avoidance**: Zip downloads don't count against GitHub API rate limits
2. **Performance**: Single download vs. multiple API calls
3. **Reliability**: More consistent behavior regardless of repository size
4. **Offline Capability**: Could cache zip files for repeated operations
5. **Bandwidth Efficiency**: Compressed downloads

### Architecture Changes

#### New Download Flow
1. **Parse Source**: Extract owner, repo, path, and ref from source URL
2. **Download Zip**: Download repository zip from `https://github.com/{owner}/{repo}/archive/{ref}.zip`
3. **Extract**: Unzip to temporary directory
4. **Filter & Copy**: Copy only the requested path(s) to target directory
5. **Cleanup**: Remove temporary files

#### URL Patterns Supported
```
github:owner/repo                    # Entire repository (main branch)
github:owner/repo/path/to/file       # Specific file
github:owner/repo/path/to/dir        # Specific directory
github:owner/repo@branch             # Specific branch
github:owner/repo@tag                # Specific tag
github:owner/repo@commit             # Specific commit
github:owner/repo/path@ref           # Path at specific ref
```

## Technical Implementation

### Required Dependencies
```go
// Add to go.mod
"archive/zip"           // Built-in Go package
"path/filepath"         // Built-in Go package
"net/http"             // Built-in Go package (already used)
"io"                   // Built-in Go package (already used)
```

### Core Components

#### 1. Zip Downloader (`internal/downloader/zip.go`)
```go
type ZipDownloader struct {
    client *http.Client
    tempDir string
}

type DownloadRequest struct {
    Owner    string
    Repo     string
    Path     string  // Optional: specific path within repo
    Ref      string  // Branch, tag, or commit (default: main)
    Target   string  // Local target directory
}

func (zd *ZipDownloader) Download(req DownloadRequest) error
func (zd *ZipDownloader) downloadZip(url string) (string, error)
func (zd *ZipDownloader) extractPath(zipPath, sourcePath, targetPath string) error
```

#### 2. URL Parser Enhancement (`internal/github/url.go`)
```go
type ParsedURL struct {
    Owner   string
    Repo    string
    Path    string  // Empty for root, "path/to/file" for specific paths
    Ref     string  // Branch, tag, or commit
}

func ParseGitHubURL(source string) (*ParsedURL, error)
func (p *ParsedURL) ZipURL() string
func (p *ParsedURL) IsFile() bool
func (p *ParsedURL) IsDirectory() bool
```

#### 3. File Type Detection
```go
type PathInfo struct {
    IsFile      bool
    IsDirectory bool
    Exists      bool
}

func DetectPathType(zipPath, targetPath string) (*PathInfo, error)
```

### Implementation Details

#### Zip Download URLs
GitHub provides zip downloads for any ref:
```
https://github.com/{owner}/{repo}/archive/{ref}.zip
```

Examples:
- `https://github.com/twilson63/qa/archive/main.zip`
- `https://github.com/twilson63/qa/archive/v1.0.0.zip`
- `https://github.com/twilson63/qa/archive/feature-branch.zip`

#### Zip Structure
GitHub zip files have a top-level directory named `{repo}-{ref}/`:
```
qa-main/
├── README.md
├── src/
│   └── data.json
└── docs/
    └── guide.md
```

#### Path Resolution
1. **Root Copy**: `github:owner/repo` → Copy all contents from `{repo}-{ref}/`
2. **File Copy**: `github:owner/repo/file.txt` → Copy `{repo}-{ref}/file.txt`
3. **Directory Copy**: `github:owner/repo/src` → Copy `{repo}-{ref}/src/` and contents

#### Error Handling
- **Repository Not Found**: HTTP 404 on zip download
- **Invalid Reference**: HTTP 404 on zip download
- **Path Not Found**: Path doesn't exist in extracted zip
- **Network Issues**: Timeout, connection failures
- **Disk Space**: Insufficient space for zip download/extraction
- **Permission Issues**: Cannot write to target directory

### Backward Compatibility

#### Configuration Options
```go
type DownloadConfig struct {
    Method          string // "zip" (default), "api" (fallback)
    UseAuthentication bool // For API fallback
    TempDir         string // Custom temp directory
    PreserveZip     bool   // Keep zip for debugging
}
```

#### Fallback Strategy
1. **Primary**: Zip download method
2. **Fallback**: Current API method (for edge cases)
3. **Configuration**: Allow users to choose method

### File Structure Changes

#### New Files
```
internal/
├── downloader/
│   ├── downloader.go      # Main downloader interface
│   ├── zip.go            # NEW: Zip-based downloader
│   ├── api.go            # RENAMED: Current API downloader
│   └── config.go         # NEW: Configuration options
├── github/
│   ├── url.go            # ENHANCED: Better URL parsing
│   └── types.go          # NEW: Common types
└── archive/              # NEW: Archive handling utilities
    ├── zip.go
    └── extractor.go
```

#### Modified Files
```
cmd/xcp/main.go           # Update to use new downloader
internal/cli/cli.go       # Add configuration options
```

## Performance Analysis

### Current Implementation
- **Small file**: 1 API call (metadata) + 1 API call (content) = 2 calls
- **Directory with 10 files**: 1 + 10 = 11 API calls
- **Large repository**: Potentially hundreds of API calls

### Proposed Implementation
- **Any size**: 1 HTTP request for zip download
- **Rate limit**: No API calls used
- **Network**: Single compressed download

### Trade-offs

#### Advantages
✅ **No rate limits**: Zip downloads are unrestricted  
✅ **Better performance**: Single download vs. multiple API calls  
✅ **Reliable**: Consistent behavior regardless of repo size  
✅ **Compressed**: Smaller downloads due to compression  
✅ **Atomic**: Complete success or failure (no partial downloads)  

#### Disadvantages
❌ **Disk space**: Requires temporary storage for entire repository  
❌ **Network**: Downloads entire repo even for single files  
❌ **Memory**: Large repositories require more memory for extraction  
❌ **Complexity**: More complex implementation  

### Use Case Analysis

#### Optimal Cases
- **Large directories**: Significant API call reduction
- **Multiple files**: Single download vs. many API calls
- **Repeated access**: Could cache zip files
- **Public repositories**: No authentication concerns

#### Sub-optimal Cases
- **Single small file**: Downloads entire repository
- **Very large repositories**: Disk space and bandwidth overhead
- **Network-constrained environments**: Large zip downloads

## Implementation Plan

### Phase 1: Core Implementation (Week 1)
- [ ] Create zip downloader implementation
- [ ] Enhance URL parsing for better ref handling
- [ ] Implement zip extraction and filtering
- [ ] Add comprehensive error handling
- [ ] Create unit tests for core functionality

### Phase 2: Integration (Week 2)
- [ ] Integrate zip downloader into main CLI
- [ ] Update command-line interface
- [ ] Add configuration options
- [ ] Implement fallback to API method
- [ ] Add integration tests

### Phase 3: Testing & Polish (Week 3)
- [ ] Comprehensive testing with various repository types
- [ ] Performance benchmarking
- [ ] Error scenario testing
- [ ] Documentation updates
- [ ] User experience improvements

### Phase 4: Release (Week 4)
- [ ] Code review and refinement
- [ ] Update README and usage examples
- [ ] Create migration guide
- [ ] Release as minor version update
- [ ] Monitor user feedback

## Testing Strategy

### Unit Tests
```go
// Test cases for zip downloader
func TestZipDownloader_Download()
func TestZipDownloader_ExtractPath()
func TestZipDownloader_HandleErrors()

// Test cases for URL parsing
func TestParseGitHubURL()
func TestParseURLWithRef()
func TestParseURLWithPath()

// Test cases for path detection
func TestDetectPathType()
func TestPathExists()
```

### Integration Tests
```go
// Test real GitHub repositories
func TestDownloadPublicRepo()
func TestDownloadPrivateRepo()
func TestDownloadSpecificPath()
func TestDownloadWithDifferentRefs()
```

### Performance Tests
```go
// Compare zip vs. API performance
func BenchmarkZipDownload()
func BenchmarkAPIDownload()
func TestRateLimitAvoidance()
```

### Edge Cases
- Very large repositories (>100MB)
- Repositories with special characters in names
- Binary files and various file types
- Symbolic links and special files
- Empty repositories and directories
- Invalid references and paths

## Configuration & CLI Changes

### New Command-Line Options
```bash
# Method selection
xcp --method=zip github:owner/repo
xcp --method=api github:owner/repo  # Fallback

# Temporary directory
xcp --temp-dir=/custom/tmp github:owner/repo

# Debug options
xcp --preserve-temp github:owner/repo  # Keep temp files
xcp --verbose github:owner/repo        # Detailed logging
```

### Environment Variables
```bash
export XCP_DOWNLOAD_METHOD=zip
export XCP_TEMP_DIR=/tmp/xcp
export XCP_PRESERVE_TEMP=false
export XCP_GITHUB_TOKEN=token  # For API fallback
```

### Configuration File Support
```json
{
  "download": {
    "method": "zip",
    "tempDir": "/tmp/xcp",
    "preserveTemp": false,
    "fallbackToAPI": true
  },
  "github": {
    "token": "optional-token"
  }
}
```

## Error Handling & User Experience

### Improved Error Messages
```
Current: "GitHub API rate limit exceeded"
Proposed: "Unable to download repository. Retrying with zip method..."

Current: "Failed to download file: file.txt"
Proposed: "File 'file.txt' not found in repository github:owner/repo"
```

### Progress Indication
```bash
Downloading repository archive... [####      ] 45%
Extracting files... [##########] 100%
Copying to target directory... [##########] 100%
Successfully copied 15 files to ./target
```

### Verbose Output
```bash
$ xcp --verbose github:twilson63/qa ./target
[INFO] Parsing URL: github:twilson63/qa
[INFO] Detected: owner=twilson63, repo=qa, ref=main, path=
[INFO] Downloading zip: https://github.com/twilson63/qa/archive/main.zip
[INFO] Downloaded 2.3MB to /tmp/xcp-123456/qa-main.zip
[INFO] Extracting archive...
[INFO] Copying files from qa-main/ to ./target/
[INFO] Copied 15 files successfully
[INFO] Cleaned up temporary files
```

## Security Considerations

### Temporary File Security
- Use secure temporary directories with proper permissions
- Clean up temporary files even on failure
- Avoid predictable temporary file names

### Path Traversal Protection
- Validate extracted file paths
- Prevent extraction outside target directory
- Handle symbolic links safely

### Network Security
- Validate GitHub URLs
- Use HTTPS for all downloads
- Implement request timeouts

## Monitoring & Metrics

### Success Metrics
- Reduction in GitHub API rate limit errors
- Improved download success rates
- Performance improvements for large repositories

### Monitoring Points
- Download success/failure rates
- Average download times
- Temporary disk space usage
- User adoption of new method

## Migration Strategy

### Default Behavior
- **v1.x**: Current API method (backward compatibility)
- **v2.0**: Zip method as default with API fallback
- **v3.0**: Zip method only (remove API method)

### User Communication
1. **Announcement**: Blog post explaining benefits
2. **Documentation**: Update all examples and guides
3. **CLI Help**: Emphasize new capabilities
4. **Error Messages**: Guide users to new method

## Future Enhancements

### Caching Layer
- Cache downloaded zip files locally
- Smart cache invalidation based on repository updates
- Configurable cache size and location

### Incremental Downloads
- Compare local cache with remote repository
- Download only changed files
- Git-like synchronization

### Parallel Processing
- Concurrent extraction for large archives
- Parallel file copying
- Streaming extraction for very large files

## Conclusion

The zip-based download strategy addresses the fundamental rate limiting issues with the current implementation while providing better performance and reliability. The implementation plan ensures backward compatibility during transition while laying the groundwork for future enhancements.

This change transforms xcp from a rate-limited tool to a reliable, high-performance file copying utility that can handle repositories of any size without GitHub API restrictions.

## Success Criteria

### Must Have
- [ ] Zero GitHub API rate limit errors for zip downloads
- [ ] Successful download of repositories up to 1GB
- [ ] Backward compatibility with existing commands
- [ ] Complete test coverage for new functionality

### Should Have
- [ ] 50%+ performance improvement for multi-file operations
- [ ] Clear progress indication for large downloads
- [ ] Comprehensive error handling and user guidance
- [ ] Configurable download methods

### Could Have
- [ ] Local caching of downloaded repositories
- [ ] Resume capability for interrupted downloads
- [ ] Compression ratio reporting
- [ ] Integration with Git authentication

The implementation of this strategy will significantly improve xcp's reliability and user experience while positioning it for future enhancements in repository synchronization and caching.