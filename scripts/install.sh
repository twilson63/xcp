#!/bin/bash
set -e

# xcp Universal Installation Script
# This script automatically detects platform and architecture, downloads the appropriate
# binary from GitHub releases, and installs it to the system PATH.

# Configuration
GITHUB_REPO="twilson63/xcp"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="xcp"
TEMP_DIR=$(mktemp -d)
VERSION="${VERSION:-latest}"
VERBOSE="${VERBOSE:-false}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

log_verbose() {
    if [[ "$VERBOSE" == "true" ]]; then
        echo -e "${BLUE}[DEBUG]${NC} $1"
    fi
}

# Cleanup function
cleanup() {
    log_verbose "Cleaning up temporary directory: $TEMP_DIR"
    rm -rf "$TEMP_DIR"
}

# Set trap for cleanup
trap cleanup EXIT

# Platform detection
detect_platform() {
    local os
    local arch
    
    # Detect OS
    case "$(uname -s)" in
        Linux*)     os="linux" ;;
        Darwin*)    os="darwin" ;;
        CYGWIN*|MINGW*|MSYS*) os="windows" ;;
        *)          
            log_error "Unsupported operating system: $(uname -s)"
            exit 1
            ;;
    esac
    
    # Detect architecture
    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64" ;;
        arm64|aarch64)  arch="arm64" ;;
        i386|i686)      arch="386" ;;
        *)              
            log_error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac
    
    echo "${os}-${arch}"
}

# Get latest release version from GitHub API
get_latest_version() {
    local api_url="https://api.github.com/repos/${GITHUB_REPO}/releases/latest"
    log_verbose "Fetching latest version from: $api_url"
    
    local response
    if command -v curl >/dev/null 2>&1; then
        response=$(curl -s "$api_url")
    elif command -v wget >/dev/null 2>&1; then
        response=$(wget -qO- "$api_url")
    else
        log_error "Neither curl nor wget is available. Please install one of them."
        exit 1
    fi
    
    # Check for API rate limiting
    if echo "$response" | grep -q "API rate limit exceeded"; then
        log_warning "GitHub API rate limit exceeded. Using fallback version v2.0.0"
        log_info "You can specify a version explicitly with: --version v2.0.0"
        echo "v2.0.0"
        return 0
    fi
    
    # Check for other API errors
    if echo "$response" | grep -q '"message"'; then
        log_warning "GitHub API error. Using fallback version v2.0.0"
        log_verbose "API response: $response"
        echo "v2.0.0"
        return 0
    fi
    
    # Extract version from response
    local version
    version=$(echo "$response" | grep '"tag_name":' | sed -E 's/.*"tag_name": "([^"]+)".*/\1/')
    
    if [[ -z "$version" ]]; then
        log_warning "Could not parse version from API response. Using fallback version v2.0.0"
        echo "v2.0.0"
    else
        echo "$version"
    fi
}

# Download and verify binary
download_binary() {
    local platform="$1"
    local version="$2"
    local binary_name="xcp-${platform}"
    local download_url
    local checksum_url
    
    if [[ "$platform" == *"windows"* ]]; then
        binary_name="${binary_name}.exe"
    fi
    
    download_url="https://github.com/${GITHUB_REPO}/releases/download/${version}/${binary_name}"
    checksum_url="https://github.com/${GITHUB_REPO}/releases/download/${version}/checksums.txt"
    
    log_info "Downloading xcp ${version} for ${platform}..."
    log_verbose "Download URL: $download_url"
    
    # Download binary
    if command -v curl >/dev/null 2>&1; then
        curl -L -o "$TEMP_DIR/$BINARY_NAME" "$download_url" || {
            log_error "Failed to download binary from $download_url"
            exit 1
        }
    elif command -v wget >/dev/null 2>&1; then
        wget -O "$TEMP_DIR/$BINARY_NAME" "$download_url" || {
            log_error "Failed to download binary from $download_url"
            exit 1
        }
    else
        log_error "Neither curl nor wget is available. Please install one of them."
        exit 1
    fi
    
    # Download and verify checksums
    log_verbose "Downloading checksums for verification..."
    if command -v curl >/dev/null 2>&1; then
        curl -L -o "$TEMP_DIR/checksums.txt" "$checksum_url" 2>/dev/null || {
            log_warning "Could not download checksums file. Skipping verification."
            return 0
        }
    elif command -v wget >/dev/null 2>&1; then
        wget -O "$TEMP_DIR/checksums.txt" "$checksum_url" 2>/dev/null || {
            log_warning "Could not download checksums file. Skipping verification."
            return 0
        }
    fi
    
    # Verify checksum
    if [[ -f "$TEMP_DIR/checksums.txt" ]]; then
        log_info "Verifying checksum..."
        cd "$TEMP_DIR"
        
        local expected_hash
        expected_hash=$(grep "$binary_name" checksums.txt | cut -d' ' -f1) || {
            log_warning "Could not find checksum for $binary_name. Skipping verification."
            return 0
        }
        
        if command -v sha256sum >/dev/null 2>&1; then
            local actual_hash
            actual_hash=$(sha256sum "$BINARY_NAME" | cut -d' ' -f1)
            if [[ "$expected_hash" != "$actual_hash" ]]; then
                log_error "Checksum verification failed!"
                log_error "Expected: $expected_hash"
                log_error "Actual:   $actual_hash"
                exit 1
            fi
            log_success "Checksum verification passed."
        elif command -v shasum >/dev/null 2>&1; then
            local actual_hash
            actual_hash=$(shasum -a 256 "$BINARY_NAME" | cut -d' ' -f1)
            if [[ "$expected_hash" != "$actual_hash" ]]; then
                log_error "Checksum verification failed!"
                log_error "Expected: $expected_hash"
                log_error "Actual:   $actual_hash"
                exit 1
            fi
            log_success "Checksum verification passed."
        else
            log_warning "Neither sha256sum nor shasum available. Skipping checksum verification."
        fi
        cd - >/dev/null
    fi
}

# Install binary
install_binary() {
    local binary_path="$TEMP_DIR/$BINARY_NAME"
    local install_path="$INSTALL_DIR/$BINARY_NAME"
    
    # Make binary executable
    chmod +x "$binary_path"
    
    # Test binary
    log_info "Testing downloaded binary..."
    if ! "$binary_path" --version >/dev/null 2>&1; then
        log_error "Downloaded binary failed to execute. It may be corrupted."
        exit 1
    fi
    
    # Check if we need sudo for installation
    if [[ ! -w "$INSTALL_DIR" ]]; then
        log_info "Installing to $install_path (requires sudo)..."
        sudo cp "$binary_path" "$install_path" || {
            log_error "Failed to install binary to $install_path"
            exit 1
        }
    else
        log_info "Installing to $install_path..."
        cp "$binary_path" "$install_path" || {
            log_error "Failed to install binary to $install_path"
            exit 1
        }
    fi
}

# Verify installation
verify_installation() {
    log_info "Verifying installation..."
    
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        local version_output
        version_output=$("$BINARY_NAME" --version 2>/dev/null || "$BINARY_NAME" version 2>/dev/null || echo "version unknown")
        log_success "xcp installed successfully!"
        log_info "Version: $version_output"
        log_info "Location: $(which $BINARY_NAME)"
    else
        log_warning "xcp was installed but is not in PATH. You may need to:"
        log_warning "  1. Add $INSTALL_DIR to your PATH"
        log_warning "  2. Restart your shell"
        log_warning "  3. Or run: export PATH=\"$INSTALL_DIR:\$PATH\""
    fi
}

# Usage information
show_usage() {
    cat << EOF
xcp Installation Script

USAGE:
    $0 [OPTIONS]

OPTIONS:
    -d, --dir DIR       Installation directory (default: /usr/local/bin)
    -v, --version VER   Specific version to install (default: latest)
    --verbose           Enable verbose output
    -h, --help          Show this help message

EXAMPLES:
    # Install latest version to default location
    $0
    
    # Install to custom directory
    $0 --dir ~/.local/bin
    
    # Install specific version
    $0 --version v1.0.0
    
    # Verbose installation
    $0 --verbose

ENVIRONMENT VARIABLES:
    INSTALL_DIR         Installation directory
    VERSION             Version to install
    VERBOSE             Enable verbose output (true/false)

EOF
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -d|--dir)
                INSTALL_DIR="$2"
                shift 2
                ;;
            -v|--version)
                VERSION="$2"
                shift 2
                ;;
            --verbose)
                VERBOSE="true"
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
}

# Main installation process
main() {
    log_info "Starting xcp installation..."
    
    # Parse arguments
    parse_args "$@"
    
    # Detect platform
    local platform
    platform=$(detect_platform)
    log_verbose "Detected platform: $platform"
    
    # Get version
    local version="$VERSION"
    if [[ "$version" == "latest" ]]; then
        version=$(get_latest_version)
        if [[ -z "$version" ]]; then
            log_error "Failed to determine latest version"
            exit 1
        fi
    fi
    log_verbose "Target version: $version"
    
    # Create install directory if it doesn't exist
    if [[ ! -d "$INSTALL_DIR" ]]; then
        log_info "Creating installation directory: $INSTALL_DIR"
        if [[ ! -w "$(dirname "$INSTALL_DIR")" ]]; then
            sudo mkdir -p "$INSTALL_DIR"
        else
            mkdir -p "$INSTALL_DIR"
        fi
    fi
    
    # Download binary
    download_binary "$platform" "$version"
    
    # Install binary
    install_binary
    
    # Verify installation
    verify_installation
    
    log_success "xcp installation completed successfully!"
    log_info ""
    log_info "Get started with: xcp --help"
}

# Run main function with all arguments
main "$@"