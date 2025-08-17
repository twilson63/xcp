# Installation Guide

This guide provides multiple ways to install xcp on your system.

## Quick Install (Recommended)

### Unix/Linux/macOS
```bash
curl -fsSL https://raw.githubusercontent.com/twilson63/xcp/main/scripts/install.sh | bash
```

### Windows (PowerShell)
```powershell
iwr -useb https://raw.githubusercontent.com/twilson63/xcp/main/scripts/install.ps1 | iex
```

## Manual Installation

### 1. Download Binary

Visit the [releases page](https://github.com/twilson63/xcp/releases) and download the appropriate binary for your platform:

- **Linux (64-bit)**: `xcp-linux-amd64`
- **Linux (ARM64)**: `xcp-linux-arm64`
- **macOS (Intel)**: `xcp-darwin-amd64`
- **macOS (Apple Silicon)**: `xcp-darwin-arm64`
- **Windows (64-bit)**: `xcp-windows-amd64.exe`

### 2. Make Executable and Install

#### Linux/macOS
```bash
# Download (replace URL with actual release URL)
curl -L https://github.com/twilson63/xcp/releases/latest/download/xcp-linux-amd64 -o xcp

# Make executable
chmod +x xcp

# Move to PATH
sudo mv xcp /usr/local/bin/
```

#### Windows
1. Download the `.exe` file
2. Place it in a directory that's in your PATH
3. Or add the directory to your PATH environment variable

## Installation Script Options

### Unix/Linux/macOS Script

```bash
# Install to custom directory
curl -fsSL https://raw.githubusercontent.com/twilson63/xcp/main/scripts/install.sh | bash -s -- --dir ~/.local/bin

# Install specific version
curl -fsSL https://raw.githubusercontent.com/twilson63/xcp/main/scripts/install.sh | bash -s -- --version v1.0.0

# Verbose output
curl -fsSL https://raw.githubusercontent.com/twilson63/xcp/main/scripts/install.sh | bash -s -- --verbose
```

### Windows PowerShell Script

```powershell
# Install to custom directory
iwr -useb https://raw.githubusercontent.com/twilson63/xcp/main/scripts/install.ps1 | iex -InstallDir "C:\tools\xcp"

# Install specific version
iwr -useb https://raw.githubusercontent.com/twilson63/xcp/main/scripts/install.ps1 | iex -Version "v1.0.0"

# Verbose output
iwr -useb https://raw.githubusercontent.com/twilson63/xcp/main/scripts/install.ps1 | iex -Verbose
```

## Package Managers (Coming Soon)

### Homebrew (macOS/Linux)
```bash
brew install xcp
```

### Chocolatey (Windows)
```powershell
choco install xcp
```

### Snap (Linux)
```bash
snap install xcp
```

## Verification

After installation, verify xcp is working:

```bash
xcp --version
xcp --help
```

## Security

### Checksum Verification

Each release includes SHA256 checksums. To verify your download:

1. Download the binary and `checksums.txt`
2. Verify the checksum:

```bash
# Linux/macOS
sha256sum -c checksums.txt

# Windows (PowerShell)
Get-FileHash xcp-windows-amd64.exe -Algorithm SHA256
```

### HTTPS Downloads

All downloads are served over HTTPS to ensure integrity during transport.

## Troubleshooting

### Permission Denied

If you get permission errors during installation:

```bash
# Linux/macOS - try without sudo first
curl -fsSL https://raw.githubusercontent.com/twilson63/xcp/main/scripts/install.sh | bash -s -- --dir ~/.local/bin

# Then add to PATH
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### Command Not Found

If `xcp` is not found after installation:

1. Check if the installation directory is in your PATH
2. Restart your terminal
3. On Windows, run `refreshenv` if using Chocolatey

### Binary Won't Execute

If the binary fails to run:

1. Ensure it's executable: `chmod +x xcp`
2. Check architecture compatibility
3. Verify the binary wasn't corrupted during download

## Building from Source

If you prefer to build from source:

```bash
git clone https://github.com/twilson63/xcp.git
cd xcp
go build -o xcp ./cmd/xcp
```

## Uninstall

To remove xcp:

```bash
# If installed to /usr/local/bin
sudo rm /usr/local/bin/xcp

# If installed to ~/.local/bin
rm ~/.local/bin/xcp
```

## Support

If you encounter issues:

1. Check the [troubleshooting section](#troubleshooting)
2. Search existing [issues](https://github.com/twilson63/xcp/issues)
3. Create a new issue with details about your system and the error