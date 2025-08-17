# Release Process

This document outlines the process for creating and publishing releases of xcp.

## Prerequisites

- Push access to the repository
- Git configured with signing key (recommended)
- All tests passing on main branch

## Release Steps

### 1. Prepare Release

1. **Update version information** in relevant files
2. **Update CHANGELOG.md** with new features, bug fixes, and breaking changes
3. **Ensure all tests pass**:
   ```bash
   make test
   make lint
   ```
4. **Test cross-compilation**:
   ```bash
   make build-all
   ```

### 2. Create Release

1. **Create and push tag**:
   ```bash
   # Replace v1.0.0 with your version
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

2. **GitHub Actions will automatically**:
   - Run all tests
   - Build binaries for all platforms
   - Generate checksums
   - Create GitHub release with artifacts

### 3. Verify Release

1. **Check GitHub Actions** completed successfully
2. **Verify all artifacts** are attached to the release:
   - `xcp-linux-amd64`
   - `xcp-linux-arm64`
   - `xcp-darwin-amd64`
   - `xcp-darwin-arm64`
   - `xcp-windows-amd64.exe`
   - `checksums.txt`

3. **Test installation** using the install script:
   ```bash
   curl -fsSL https://raw.githubusercontent.com/twilson63/xcp/main/scripts/install.sh | bash -s -- --version v1.0.0
   ```

### 4. Post-Release Tasks

1. **Update package managers** (when available):
   - Homebrew formula
   - Chocolatey package
   - Docker images

2. **Announce release** (if applicable):
   - Blog post
   - Social media
   - Community forums

## Manual Release (Emergency)

If GitHub Actions fails, you can create a release manually:

```bash
# Build all binaries
make build-all

# Generate checksums
make checksums

# Create release using GitHub CLI
gh release create v1.0.0 \
  --title "Release v1.0.0" \
  --notes "Release notes here" \
  dist/*
```

## Version Scheme

This project follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version for incompatible API changes
- **MINOR** version for backwards-compatible functionality
- **PATCH** version for backwards-compatible bug fixes

Examples:
- `v1.0.0` - Initial release
- `v1.1.0` - New features, backwards compatible
- `v1.1.1` - Bug fixes
- `v2.0.0` - Breaking changes

## Rollback

If a release has critical issues:

1. **Delete the problematic tag**:
   ```bash
   git tag -d v1.0.0
   git push origin :refs/tags/v1.0.0
   ```

2. **Delete the GitHub release** through the web interface

3. **Fix issues and create a new release**

## Build Targets

Current build targets:

| Platform | Architecture | Binary Name |
|----------|-------------|-------------|
| Linux | amd64 | xcp-linux-amd64 |
| Linux | arm64 | xcp-linux-arm64 |
| macOS | amd64 | xcp-darwin-amd64 |
| macOS | arm64 | xcp-darwin-arm64 |
| Windows | amd64 | xcp-windows-amd64.exe |

## Security

- All binaries are built with stripped debug information (`-ldflags="-s -w"`)
- SHA256 checksums are provided for integrity verification
- Future releases may include code signing

## Troubleshooting Releases

### GitHub Actions Fails

1. Check the Actions tab for error details
2. Common issues:
   - Test failures
   - Build environment changes
   - Permission issues

### Missing Artifacts

If some artifacts are missing from the release:

1. Check if the build completed for all platforms
2. Verify the `dist/` directory contains all expected files
3. Re-run the failed GitHub Action job

### Installation Script Issues

If users report installation problems:

1. Test the script locally
2. Check for platform-specific issues
3. Verify download URLs are correct