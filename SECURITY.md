# XCP Security Implementation

## Overview

XCP implements comprehensive security measures for the zip download strategy to protect against various attack vectors including path traversal, zip bombs, and other malicious inputs.

## Security Features Implemented

### 1. Path Traversal Protection

**Location**: `internal/archive/zip.go`, `internal/security/validate.go`

**Protection Against**:
- `../../../etc/passwd` style attacks
- Absolute path exploitation
- Symlink attacks
- Directory escape attempts

**Implementation**:
```go
// Validates paths to prevent traversal
func ValidateFilePath(path string) error
func ValidateExtractedPath(extractPath, targetDir string) error
func SanitizeFilePath(path, baseDir string) (string, error)
```

**Key Features**:
- Strict path validation using `filepath.Clean`
- Absolute path detection and rejection
- Directory boundary enforcement
- Symlink detection and blocking

### 2. Zip Bomb Protection

**Location**: `internal/archive/zip.go`

**Protection Against**:
- Decompression bombs (high compression ratio files)
- Excessive file count attacks
- Memory exhaustion attacks
- Disk space exhaustion

**Implementation**:
```go
const (
    maxFileSize      = 100 * 1024 * 1024  // 100MB per file
    maxTotalSize     = 1024 * 1024 * 1024 // 1GB total extraction
    maxFileCount     = 10000              // Maximum files to extract
    compressionRatio = 100                // Max compression ratio
)
```

**Key Features**:
- Pre-extraction size validation
- Compression ratio analysis
- File count limits
- Real-time size monitoring during extraction

### 3. Secure Temporary File Handling

**Location**: `internal/security/tempfile.go`

**Protection Against**:
- Temporary file leakage
- Insecure permissions
- Predictable file names
- Race conditions

**Implementation**:
```go
// Creates secure temporary directories with restricted permissions
func SecureTempDir(prefix string) (string, error)
func SecureTempFile(dir, prefix string) (*os.File, error)
```

**Key Features**:
- Cryptographically secure random names
- Restricted file permissions (0600 for files, 0700 for directories)
- Automatic cleanup on failure
- Safe temporary directory creation

### 4. Input Validation

**Location**: `internal/security/validate.go`

**Protection Against**:
- Malicious URLs
- Invalid GitHub sources
- Unsafe filenames
- Reserved system names

**Implementation**:
```go
func ValidateGitHubURL(rawURL string) error
func ValidateGitHubSource(owner, repo, ref string) error
func ValidateFileName(filename string) error
```

**Key Features**:
- HTTPS-only URL validation
- GitHub domain verification
- Filename safety checks
- Windows reserved name detection

### 5. Network Security

**Location**: `internal/downloader/zip.go`

**Protection Against**:
- Man-in-the-middle attacks
- Insecure connections
- Malicious redirects
- Oversized downloads

**Implementation**:
```go
// TLS 1.2+ with strong cipher suites
TLSClientConfig: &tls.Config{
    MinVersion: tls.VersionTLS12,
    CipherSuites: []uint16{
        tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
        tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
        // ...
    },
}
```

**Key Features**:
- TLS 1.2+ enforcement
- Strong cipher suite selection
- Redirect validation to GitHub domains only
- Download size limits
- Connection timeouts

## Security Testing

### Automated Security Tests

Run the security test suite:

```bash
# Test path traversal protection
go test ./internal/security/ -run TestPathTraversal -v

# Test zip bomb protection
go test ./internal/archive/ -run TestZipBomb -v

# Test input validation
go test ./internal/security/ -run TestValidation -v
```

### Manual Security Testing

#### 1. Path Traversal Tests

```bash
# Test with malicious paths
xcp github:owner/repo/../../../etc/passwd ./target  # Should fail
xcp github:owner/repo/../../root/.ssh ./target     # Should fail
```

#### 2. Large File Tests

```bash
# Test with repositories containing large files
xcp github:owner/large-repo ./target  # Should respect size limits
```

#### 3. URL Validation Tests

```bash
# Test with malicious URLs
xcp http://malicious.com/repo.zip ./target      # Should fail (HTTP)
xcp https://evil.com/fake-repo.zip ./target     # Should fail (non-GitHub)
```

## Security Configuration

### Environment Variables

```bash
# Maximum download size (default: 2GB)
export XCP_MAX_DOWNLOAD_SIZE=2147483648

# Maximum extraction size (default: 1GB)
export XCP_MAX_EXTRACT_SIZE=1073741824

# Maximum file count (default: 10000)
export XCP_MAX_FILE_COUNT=10000

# Temporary directory (default: system temp)
export XCP_TEMP_DIR=/secure/temp/path
```

### Command-Line Options

```bash
# Use custom temporary directory
xcp --temp-dir=/secure/temp github:owner/repo ./target

# Verbose security logging
xcp --verbose github:owner/repo ./target
```

## Security Monitoring

### Log Analysis

XCP logs security-relevant events:

```
WARN: Large repository detected (150.2 MB)
INFO: Validated GitHub URL: https://github.com/owner/repo/archive/main.zip
INFO: Created secure temporary directory: /tmp/xcp-secure-abc123
INFO: Extracted 1,234 files to ./target (total: 45.6 MB)
INFO: Cleaned up temporary files
```

### Metrics to Monitor

- Download sizes and extraction ratios
- Temporary file cleanup success/failure
- Path validation failures
- Network connection security

## Incident Response

### If Security Issue Detected

1. **Immediate Actions**:
   - Stop the download/extraction process
   - Clean up any temporary files
   - Log the security event

2. **Investigation**:
   - Check logs for attack patterns
   - Verify file integrity
   - Assess potential damage

3. **Recovery**:
   - Remove compromised files
   - Update security configurations
   - Report incident if necessary

## Security Best Practices

### For Users

1. **Always verify repository sources**:
   ```bash
   # Good: Official repository
   xcp github:golang/go ./go-source
   
   # Suspicious: Unknown source
   xcp github:unknown-user/suspicious-repo ./target
   ```

2. **Use restricted target directories**:
   ```bash
   # Good: Isolated directory
   xcp github:owner/repo ./isolated/target
   
   # Bad: System directory
   xcp github:owner/repo /usr/local/bin
   ```

3. **Monitor resource usage**:
   - Check available disk space before large downloads
   - Monitor extraction progress
   - Set reasonable timeouts

### For Developers

1. **Input Validation**:
   - Always validate user inputs
   - Use security package functions
   - Never trust external data

2. **Error Handling**:
   - Fail securely on errors
   - Clean up resources on failure
   - Log security events

3. **Testing**:
   - Include security tests in CI/CD
   - Test with malicious inputs
   - Verify cleanup procedures

## Security Limitations

### Known Limitations

1. **Resource Exhaustion**:
   - Large numbers of small files can still consume resources
   - Network bandwidth consumption is not limited

2. **Timing Attacks**:
   - Path validation timing may leak information
   - Consider constant-time validation for sensitive use cases

3. **System Dependencies**:
   - Relies on OS file system security
   - Temporary directory security depends on OS configuration

### Future Improvements

1. **Enhanced Monitoring**:
   - Real-time resource usage tracking
   - Anomaly detection for download patterns

2. **Additional Validation**:
   - Content-based file validation
   - Malware scanning integration

3. **Performance Optimization**:
   - Streaming validation for large files
   - Parallel security checks

## Contact

For security issues or questions:
- Create an issue in the repository
- Follow responsible disclosure practices
- Include minimal reproduction steps

## Compliance

This implementation follows:
- OWASP Secure Coding Practices
- NIST Cybersecurity Framework guidelines
- Industry best practices for file handling security