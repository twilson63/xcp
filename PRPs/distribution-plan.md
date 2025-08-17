# xcp Distribution Plan

## Overview

This document outlines the comprehensive distribution strategy for the xcp CLI tool, including cross-platform binary builds, installation scripts, and release automation.

## Target Platforms

### Primary Platforms
Based on Go's cross-compilation capabilities and market demand:

1. **Linux**
   - `linux/amd64` - Primary Linux desktop/server platform
   - `linux/arm64` - ARM-based servers and modern devices
   - `linux/386` - Legacy 32-bit systems (optional)

2. **macOS**
   - `darwin/amd64` - Intel-based Macs
   - `darwin/arm64` - Apple Silicon Macs (M1/M2/M3)

3. **Windows**
   - `windows/amd64` - 64-bit Windows systems
   - `windows/386` - 32-bit Windows systems (optional)
   - `windows/arm64` - ARM-based Windows devices (future)

### Priority Matrix
| Platform | Priority | Justification |
|----------|----------|---------------|
| linux/amd64 | High | Most common server/developer platform |
| darwin/amd64 | High | Intel Mac developer workstations |
| darwin/arm64 | High | Apple Silicon Mac adoption |
| windows/amd64 | High | Windows developer workstations |
| linux/arm64 | Medium | Growing ARM server adoption |
| windows/386 | Low | Legacy systems, declining usage |
| linux/386 | Low | Legacy systems, minimal demand |

## Build Strategy

### Cross-Compilation Setup
Go provides excellent cross-compilation support through environment variables:

```bash
# Build for Linux (amd64)
GOOS=linux GOARCH=amd64 go build -o dist/xcp-linux-amd64 ./cmd/xcp

# Build for macOS (amd64)
GOOS=darwin GOARCH=amd64 go build -o dist/xcp-darwin-amd64 ./cmd/xcp

# Build for macOS (arm64)
GOOS=darwin GOARCH=arm64 go build -o dist/xcp-darwin-arm64 ./cmd/xcp

# Build for Windows (amd64)
GOOS=windows GOARCH=amd64 go build -o dist/xcp-windows-amd64.exe ./cmd/xcp
```

### Automated Build Process

#### Makefile Targets
Extend the existing Makefile with cross-compilation targets:

```make
.PHONY: build-all build-linux build-darwin build-windows clean-dist

# Build all platforms
build-all: build-linux build-darwin build-windows

# Linux builds
build-linux:
	mkdir -p dist
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/xcp-linux-amd64 ./cmd/xcp
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o dist/xcp-linux-arm64 ./cmd/xcp

# macOS builds
build-darwin:
	mkdir -p dist
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/xcp-darwin-amd64 ./cmd/xcp
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/xcp-darwin-arm64 ./cmd/xcp

# Windows builds
build-windows:
	mkdir -p dist
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/xcp-windows-amd64.exe ./cmd/xcp

# Clean distribution directory
clean-dist:
	rm -rf dist/
```

#### Build Optimization
- Use `-ldflags="-s -w"` to strip debug information and reduce binary size
- Consider UPX compression for further size reduction (optional)
- Implement version injection via ldflags

### GitHub Actions Workflow

Create `.github/workflows/release.yml` for automated releases:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Run tests
      run: go test ./...
    
    - name: Build binaries
      run: make build-all
    
    - name: Create release
      uses: softprops/action-gh-release@v1
      with:
        files: dist/*
        generate_release_notes: true
```

## Installation Methods

### 1. Direct Binary Download
**Target Users**: Advanced users, CI/CD systems

**Implementation**:
- Host binaries on GitHub Releases
- Provide direct download links
- Include checksums for verification

**Usage**:
```bash
# Linux/macOS
curl -L https://github.com/owner/xcp/releases/latest/download/xcp-linux-amd64 -o xcp
chmod +x xcp
sudo mv xcp /usr/local/bin/
```

### 2. Installation Script (Recommended)
**Target Users**: General developers, quick setup

**Features**:
- Auto-detect platform and architecture
- Download appropriate binary
- Install to system PATH
- Verify installation
- Handle permissions

**Script Location**: `scripts/install.sh`

### 3. Package Managers (Future)
**Target Users**: Platform-specific workflows

**Planned Support**:
- **Homebrew** (macOS/Linux): `brew install xcp`
- **Chocolatey** (Windows): `choco install xcp`
- **Snap** (Linux): `snap install xcp`
- **AUR** (Arch Linux): Community package

### 4. Container Images (Optional)
**Target Users**: Containerized environments

**Implementation**:
- Multi-arch Docker images
- Minimal Alpine-based images
- Available on Docker Hub and GitHub Container Registry

## Installation Script Design

### Universal Install Script (`scripts/install.sh`)

**Features**:
1. **Platform Detection**
   - Automatically detect OS (Linux/macOS/Windows via WSL)
   - Detect architecture (amd64/arm64/386)
   - Handle edge cases and unsupported platforms

2. **Download Management**
   - Fetch latest release information from GitHub API
   - Download appropriate binary with progress indication
   - Verify checksums (SHA256)
   - Handle network failures gracefully

3. **Installation Process**
   - Create temporary directory for download
   - Extract/rename binary as needed
   - Install to `/usr/local/bin` or user-specified location
   - Handle permission requirements (sudo prompt)
   - Create symlinks if necessary

4. **Verification**
   - Test installed binary
   - Display version information
   - Provide usage instructions

5. **Error Handling**
   - Clear error messages
   - Cleanup on failure
   - Support for verbose/debug mode

### Windows Installation Script (`scripts/install.ps1`)

**PowerShell script for Windows users**:
- Similar functionality to Unix script
- Handle Windows-specific paths and permissions
- Support for both PowerShell Core and Windows PowerShell

### Usage Examples

```bash
# Standard installation
curl -fsSL https://raw.githubusercontent.com/owner/xcp/main/scripts/install.sh | bash

# Custom installation directory
curl -fsSL https://raw.githubusercontent.com/owner/xcp/main/scripts/install.sh | bash -s -- --dir=/custom/path

# Specific version
curl -fsSL https://raw.githubusercontent.com/owner/xcp/main/scripts/install.sh | bash -s -- --version=v1.0.0
```

## Release Process

### Version Management
- **Semantic Versioning**: Follow semver (v1.0.0, v1.1.0, v2.0.0)
- **Git Tags**: Tag releases in git
- **Changelog**: Maintain CHANGELOG.md with release notes

### Release Checklist
1. **Pre-release**
   - [ ] Update version in code
   - [ ] Update CHANGELOG.md
   - [ ] Run full test suite
   - [ ] Test cross-compilation builds
   - [ ] Update documentation

2. **Release**
   - [ ] Create git tag
   - [ ] Push tag to trigger GitHub Actions
   - [ ] Verify all binaries build successfully
   - [ ] Test installation script with new release

3. **Post-release**
   - [ ] Update package manager definitions
   - [ ] Announce release (if applicable)
   - [ ] Monitor for issues

### Automation
- **GitHub Actions**: Trigger builds on tag push
- **Release Notes**: Auto-generate from commit messages
- **Asset Upload**: Automatically attach binaries to release

## Distribution Channels

### Primary
1. **GitHub Releases** - Main distribution channel
2. **Installation Script** - Recommended for users
3. **Direct Download** - Power users and CI/CD

### Secondary (Future)
1. **Homebrew** - macOS/Linux package manager
2. **Chocolatey** - Windows package manager
3. **Docker Hub** - Container images
4. **Linux Package Repositories** - APT, YUM, etc.

## Security Considerations

### Binary Integrity
- **Checksums**: SHA256 hashes for all binaries
- **Code Signing**: 
  - macOS: Apple Developer certificate
  - Windows: Authenticode signing
- **Supply Chain**: Reproducible builds

### Installation Security
- **HTTPS**: All downloads over encrypted connections
- **Verification**: Checksum validation in install scripts
- **Permissions**: Minimal required permissions
- **Cleanup**: Remove temporary files securely

## Metrics and Analytics

### Download Tracking
- GitHub Releases download statistics
- Installation script usage metrics (non-PII)
- Platform distribution analysis

### Success Metrics
- Download counts by platform
- Installation success rates
- User feedback and issues

## Implementation Timeline

### Phase 1: Core Distribution (Week 1-2)
- [ ] Extend Makefile with cross-compilation targets
- [ ] Create universal installation script
- [ ] Set up GitHub Actions for release automation
- [ ] Test full release process

### Phase 2: Enhanced Distribution (Week 3-4)
- [ ] Windows PowerShell installation script
- [ ] Binary signing setup
- [ ] Checksum generation and verification
- [ ] Documentation and usage examples

### Phase 3: Package Managers (Month 2)
- [ ] Homebrew formula
- [ ] Chocolatey package
- [ ] Container images
- [ ] Linux package repository setup

## Maintenance

### Regular Tasks
- Monitor GitHub Actions for build failures
- Update dependencies and Go version
- Respond to installation issues
- Update package manager definitions

### Quarterly Reviews
- Analyze download metrics
- Review platform priorities
- Update distribution strategy
- Security audit of build process

## Conclusion

This distribution plan provides a comprehensive approach to making xcp easily accessible across all major platforms. The phased implementation allows for quick initial distribution while building toward a more robust ecosystem of installation methods.

The focus on automation, security, and user experience ensures that xcp can be easily adopted by developers while maintaining high standards for reliability and trust.