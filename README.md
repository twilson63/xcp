# xcp - External Copy Program

A CLI tool for copying files from remote GitHub repositories to local directories.

## Overview

`xcp` is a command-line utility that allows users to copy files or directories from any GitHub repository to a local target path. It functions similar to a copy command but works with remote GitHub sources.

## Use Cases

```bash
# Copy files from a GitHub repository to current directory
xcp github:twilson63/qa

# Copy a specific file and pipe to another command
xcp github:twilson63/foo/data.json | jq

# Copy to a specific target directory
xcp github:twilson63/qa ./target/path

# Copy a specific file to a target directory
xcp github:twilson63/foo/data.json ./target/path
```

## Requirements

### Core Functionality

1. **Source Parameter**
   - Format: `github:{owner}/{repo}[/{path}]`
   - Required parameter
   - Supports optional path within the repository

2. **Target Parameter**
   - Optional parameter
   - If not provided, defaults to current working directory
   - If provided, specifies the local destination path

3. **Download Capabilities**
   - Fetch files from public GitHub repositories
   - Handle both individual files and directories
   - Preserve directory structure when copying directories
   - Support for piping file contents to other commands

### Technical Requirements

1. **CLI Interface**
   - Simple, intuitive command structure
   - Proper error handling and user feedback
   - Support for help (`--help`, `-h`) flag

2. **GitHub Integration**
   - Use GitHub API for fetching repository contents
   - Handle rate limiting appropriately
   - Support for large files if possible

3. **File Operations**
   - Create directories as needed
   - Overwrite existing files with confirmation
   - Preserve file permissions where possible

### Implementation Considerations

1. **Error Handling**
   - Invalid repository URLs
   - Repository not found
   - File/directory not found in repository
   - Network connectivity issues
   - Permission denied errors

2. **Performance**
   - Efficient downloading of files
   - Progress indicators for large downloads
   - Caching mechanisms for repeated requests

3. **Security**
   - Validate repository URLs
   - Sanitize file paths to prevent directory traversal
   - Handle GitHub tokens for private repositories (future enhancement)

## Future Enhancements

1. Support for private repositories with authentication
2. Support for other Git hosting platforms (GitLab, Bitbucket)
3. Support for specific branches or tags
4. Dry-run mode to preview what would be copied
5. Filtering options for selecting specific files