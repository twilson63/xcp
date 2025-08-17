# xcp - External Copy Program

A high-performance CLI tool for copying files from GitHub repositories to local directories without API rate limits.

## Overview

`xcp` is a command-line utility that downloads files or directories from any GitHub repository to a local target path. Version 2.0 introduces a revolutionary zip-based download strategy that eliminates GitHub API rate limits while providing 50%+ performance improvements for multi-file operations.

## Quick Start

```bash
# Copy entire repository
xcp github:twilson63/qa

# Copy specific branch or tag  
xcp github:twilson63/qa@v1.0.0
xcp github:twilson63/qa@main

# Copy specific file and pipe to command
xcp github:twilson63/foo/data.json | jq

# Copy to specific target directory
xcp github:twilson63/qa ./target/path

# Copy specific file with custom settings
xcp --verbose --method=zip github:twilson63/foo/data.json ./target/
```

## ✨ What's New in v2.0

### 🚀 Zip Download Strategy
- **No Rate Limits**: Download repositories without GitHub API restrictions
- **50%+ Faster**: Significant performance improvement for multi-file operations  
- **More Reliable**: Consistent behavior regardless of repository size
- **Automatic**: Zip method is now the default (with API fallback)

### 🎯 Enhanced URL Syntax
```bash
github:owner/repo@branch          # Specific branch
github:owner/repo@v1.0.0          # Specific tag
github:owner/repo@abc123          # Specific commit  
github:owner/repo@ref/path/file   # Path at specific reference
```

### ⚙️ New CLI Options
- `--method zip|api` - Choose download method (zip is default)
- `--temp-dir DIR` - Custom temporary directory for extraction
- `--verbose` - Detailed progress output with download statistics

## 📋 Features

### Core Functionality
1. **Source Parameter**: `github:{owner}/{repo}[@ref][/{path}]`
   - Required parameter specifying GitHub repository
   - Optional branch/tag/commit with `@ref` syntax  
   - Optional path within repository

2. **Target Parameter**: Local destination (optional)
   - Defaults to current directory if not specified
   - Creates directories as needed

3. **Download Capabilities**
   - Download entire repositories or specific files/directories
   - Support for all Git references (branches, tags, commits)
   - Preserve directory structure and file permissions
   - Stream file contents to stdout for piping

### Performance & Reliability
- **Zip-based downloads**: No GitHub API rate limits
- **Streaming**: Memory-efficient processing for large repositories
- **Concurrent extraction**: Parallel processing where beneficial
- **Comprehensive error handling**: Clear, actionable error messages

## 📦 Installation

### Quick Install Scripts

**Unix/Linux/macOS:**
```bash
curl -sSL https://raw.githubusercontent.com/twilson63/xcp/main/scripts/install.sh | bash
```

**Windows (PowerShell):**
```powershell
iwr -useb https://raw.githubusercontent.com/twilson63/xcp/main/scripts/install.ps1 | iex
```

### Manual Installation

1. Download the appropriate binary from [releases](https://github.com/twilson63/xcp/releases)
2. Make it executable: `chmod +x xcp-*`
3. Move to a directory in your PATH: `mv xcp-* /usr/local/bin/xcp`

### Supported Platforms
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

## 🎯 Usage Examples

### Basic Operations
```bash
# Download entire repository (main branch)
xcp github:facebook/react

# Download specific branch
xcp github:facebook/react@v18.2.0

# Download specific directory
xcp github:facebook/react/packages/react

# Download to custom location
xcp github:facebook/react ./my-react-copy
```

### Advanced Usage
```bash
# Use API method for small files
xcp --method=api github:owner/repo/single-file.txt

# Verbose output with progress
xcp --verbose github:large/repository

# Custom temp directory
xcp --temp-dir=/tmp/custom github:owner/repo

# Stream file content
xcp github:owner/repo/data.json | jq '.key'
```

### URL Format Reference
```
github:owner/repo                    # Entire repository (main branch)
github:owner/repo@branch             # Specific branch  
github:owner/repo@v1.0.0             # Specific tag
github:owner/repo@abc123             # Specific commit
github:owner/repo/path/to/file       # Specific file
github:owner/repo/path/to/dir/       # Specific directory
github:owner/repo@ref/path           # Path at specific ref
```

## 🔧 CLI Options

```
Usage: xcp [options] <source> [target]

Options:
  -h, --help              Show help information
  -v, --version          Show version information
  -f, --overwrite        Overwrite existing files
  --method string        Download method: zip (default) or api
  --temp-dir string      Custom temporary directory for zip extraction
  --verbose              Enable verbose output

Arguments:
  source                 github:owner/repo[@ref][/path]
  target                 Local directory or file (optional)
```

## 🛡️ Security

xcp implements comprehensive security measures:
- **Path traversal protection**: Prevents zip slip attacks
- **Input validation**: Validates all URLs and file paths
- **Secure temp files**: Uses cryptographically secure temporary files
- **Zip bomb protection**: Limits file sizes and counts during extraction

## ⚡ Performance

- **No Rate Limits**: Zip downloads bypass GitHub API limits
- **50%+ Faster**: Significant improvement for repositories with multiple files
- **Memory Efficient**: Streaming processing for large repositories
- **Concurrent Processing**: Parallel extraction where beneficial

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch: `git checkout -b feature/amazing-feature`
3. Commit your changes: `git commit -m 'Add amazing feature'`
4. Push to the branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🚀 What's Next

- Local caching of downloaded repositories
- Resume capability for interrupted downloads
- Git-like synchronization features
- Integration with Git authentication for private repositories