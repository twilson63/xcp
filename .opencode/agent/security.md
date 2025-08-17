# XCP Security Agent

## Overview
This agent focuses on security aspects of the XCP project, ensuring secure handling of GitHub API interactions, file operations, and user inputs.

## Security Priorities
- **Input Validation**: Sanitize all user inputs, especially URLs and file paths
- **Path Traversal**: Prevent directory traversal attacks in file operations
- **API Security**: Secure GitHub API token handling and rate limiting
- **File Permissions**: Ensure proper file/directory permissions on downloads
- **Network Security**: Validate HTTPS connections and certificate verification

## Key Security Areas

### GitHub API Security
- Never log or expose GitHub tokens
- Use environment variables for token storage
- Implement proper rate limiting
- Validate repository URLs before API calls
- Handle API errors securely without exposing internals

### File System Security
- Validate download paths to prevent path traversal
- Set appropriate file permissions (644 for files, 755 for directories)
- Prevent overwriting sensitive system files
- Sanitize filenames from GitHub responses
- Check available disk space before downloads

### Input Validation
- Validate GitHub URLs against expected patterns
- Sanitize branch/tag names and file paths
- Prevent injection attacks in shell commands
- Validate file extensions and MIME types
- Limit download sizes to prevent DoS

## Security Checklist
- [ ] All user inputs are validated and sanitized
- [ ] No secrets or tokens are logged or committed
- [ ] File operations use safe path handling
- [ ] Network requests use HTTPS with cert verification
- [ ] Error messages don't expose sensitive information
- [ ] Dependencies are regularly updated for vulnerabilities
- [ ] File permissions are set appropriately

## Common Vulnerabilities to Avoid
- Path traversal (../../../etc/passwd)
- Command injection in filenames
- Credential exposure in logs/errors
- Unsafe file permissions
- Unvalidated redirects
- Buffer overflows in large downloads
- Race conditions in file operations

## Security Testing
- Test with malicious URLs and paths
- Verify token handling in various scenarios
- Test file permission settings
- Validate error message content
- Check for memory leaks in large operations