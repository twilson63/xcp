# xcp Windows Installation Script
# This script automatically detects architecture, downloads the appropriate
# binary from GitHub releases, and installs it to the system PATH.

param(
    [string]$InstallDir = "$env:LOCALAPPDATA\xcp\bin",
    [string]$Version = "latest",
    [switch]$Verbose = $false,
    [switch]$Help = $false
)

# Configuration
$GitHubRepo = "twilson63/xcp"
$BinaryName = "xcp.exe"
$TempDir = [System.IO.Path]::GetTempPath() + [System.Guid]::NewGuid().ToString()

# Create temp directory
New-Item -ItemType Directory -Path $TempDir -Force | Out-Null

# Cleanup function
function Cleanup {
    if (Test-Path $TempDir) {
        Remove-Item -Path $TempDir -Recurse -Force -ErrorAction SilentlyContinue
        Write-Verbose "Cleaned up temporary directory: $TempDir"
    }
}

# Set cleanup on exit
trap { Cleanup }

# Logging functions
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

function Write-VerboseLog {
    param([string]$Message)
    if ($Verbose) {
        Write-Host "[DEBUG] $Message" -ForegroundColor Cyan
    }
}

# Platform detection
function Get-Platform {
    $os = "windows"
    
    # Detect architecture
    $arch = switch ($env:PROCESSOR_ARCHITECTURE) {
        "AMD64" { "amd64" }
        "x86" { "386" }
        "ARM64" { "arm64" }
        default { 
            Write-Error "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE"
            exit 1
        }
    }
    
    return "$os-$arch"
}

# Get latest release version from GitHub API
function Get-LatestVersion {
    $apiUrl = "https://api.github.com/repos/$GitHubRepo/releases/latest"
    Write-VerboseLog "Fetching latest version from: $apiUrl"
    
    try {
        $response = Invoke-RestMethod -Uri $apiUrl -ErrorAction Stop
        return $response.tag_name
    }
    catch {
        # Check if it's a rate limiting error
        if ($_.Exception.Message -like "*rate limit*" -or $_.Exception.Message -like "*403*") {
            Write-Warning "GitHub API rate limit exceeded. Using fallback version v2.0.0"
            Write-Info "You can specify a version explicitly with: -Version v2.0.0"
            return "v2.0.0"
        }
        
        Write-Warning "GitHub API error. Using fallback version v2.0.0"
        Write-VerboseLog "API error: $($_.Exception.Message)"
        return "v2.0.0"
    }
}

# Download and verify binary
function Download-Binary {
    param(
        [string]$Platform,
        [string]$Version
    )
    
    $binaryName = "xcp-$Platform.exe"
    $downloadUrl = "https://github.com/$GitHubRepo/releases/download/$Version/$binaryName"
    $checksumUrl = "https://github.com/$GitHubRepo/releases/download/$Version/checksums.txt"
    $binaryPath = Join-Path $TempDir $BinaryName
    $checksumPath = Join-Path $TempDir "checksums.txt"
    
    Write-Info "Downloading xcp $Version for $Platform..."
    Write-VerboseLog "Download URL: $downloadUrl"
    
    # Download binary
    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile $binaryPath -ErrorAction Stop
        Write-VerboseLog "Binary downloaded to: $binaryPath"
    }
    catch {
        Write-Error "Failed to download binary from $downloadUrl"
        Write-Error $_.Exception.Message
        exit 1
    }
    
    # Download and verify checksums
    Write-VerboseLog "Downloading checksums for verification..."
    try {
        Invoke-WebRequest -Uri $checksumUrl -OutFile $checksumPath -ErrorAction Stop
        
        # Verify checksum
        Write-Info "Verifying checksum..."
        $checksumContent = Get-Content $checksumPath
        $expectedHash = ($checksumContent | Where-Object { $_ -match $binaryName } | ForEach-Object { $_.Split(' ')[0] })
        
        if ($expectedHash) {
            $actualHash = (Get-FileHash -Path $binaryPath -Algorithm SHA256).Hash.ToLower()
            
            if ($expectedHash -ne $actualHash) {
                Write-Error "Checksum verification failed!"
                Write-Error "Expected: $expectedHash"
                Write-Error "Actual:   $actualHash"
                exit 1
            }
            Write-Success "Checksum verification passed."
        }
        else {
            Write-Warning "Could not find checksum for $binaryName. Skipping verification."
        }
    }
    catch {
        Write-Warning "Could not download checksums file. Skipping verification."
        Write-VerboseLog $_.Exception.Message
    }
    
    return $binaryPath
}

# Install binary
function Install-Binary {
    param([string]$BinaryPath)
    
    $installPath = Join-Path $InstallDir $BinaryName
    
    # Test binary
    Write-Info "Testing downloaded binary..."
    try {
        $versionOutput = & $BinaryPath --version 2>$null
        Write-VerboseLog "Binary test successful: $versionOutput"
    }
    catch {
        Write-Error "Downloaded binary failed to execute. It may be corrupted."
        exit 1
    }
    
    # Create install directory
    if (-not (Test-Path $InstallDir)) {
        Write-Info "Creating installation directory: $InstallDir"
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }
    
    # Install binary
    Write-Info "Installing to $installPath..."
    try {
        Copy-Item -Path $BinaryPath -Destination $installPath -Force
        Write-VerboseLog "Binary installed to: $installPath"
    }
    catch {
        Write-Error "Failed to install binary to $installPath"
        Write-Error $_.Exception.Message
        exit 1
    }
}

# Add to PATH
function Add-ToPath {
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    
    if ($currentPath -notlike "*$InstallDir*") {
        Write-Info "Adding $InstallDir to user PATH..."
        $newPath = "$InstallDir;$currentPath"
        [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
        Write-Success "Added to PATH. Please restart your terminal or run: refreshenv"
    }
    else {
        Write-VerboseLog "Installation directory already in PATH"
    }
}

# Verify installation
function Test-Installation {
    Write-Info "Verifying installation..."
    
    # Refresh current session PATH
    $env:PATH = "$InstallDir;$env:PATH"
    
    try {
        $versionOutput = & xcp --version 2>$null
        if (-not $versionOutput) {
            $versionOutput = & xcp version 2>$null
        }
        if (-not $versionOutput) {
            $versionOutput = "version unknown"
        }
        
        Write-Success "xcp installed successfully!"
        Write-Info "Version: $versionOutput"
        Write-Info "Location: $(Join-Path $InstallDir $BinaryName)"
    }
    catch {
        Write-Warning "xcp was installed but verification failed. You may need to:"
        Write-Warning "  1. Restart your terminal"
        Write-Warning "  2. Or run: refreshenv (if using Chocolatey)"
        Write-Warning "  3. Or manually add $InstallDir to your PATH"
    }
}

# Usage information
function Show-Usage {
    @"
xcp Windows Installation Script

USAGE:
    .\install.ps1 [OPTIONS]

OPTIONS:
    -InstallDir DIR     Installation directory (default: $env:LOCALAPPDATA\xcp\bin)
    -Version VER        Specific version to install (default: latest)
    -Verbose           Enable verbose output
    -Help              Show this help message

EXAMPLES:
    # Install latest version to default location
    .\install.ps1
    
    # Install to custom directory
    .\install.ps1 -InstallDir "C:\tools\xcp"
    
    # Install specific version
    .\install.ps1 -Version "v1.0.0"
    
    # Verbose installation
    .\install.ps1 -Verbose

"@
}

# Main installation process
function Main {
    if ($Help) {
        Show-Usage
        return
    }
    
    Write-Info "Starting xcp installation..."
    
    # Detect platform
    $platform = Get-Platform
    Write-VerboseLog "Detected platform: $platform"
    
    # Get version
    $targetVersion = $Version
    if ($targetVersion -eq "latest") {
        $targetVersion = Get-LatestVersion
        if (-not $targetVersion) {
            Write-Error "Failed to determine latest version"
            exit 1
        }
    }
    Write-VerboseLog "Target version: $targetVersion"
    
    # Download binary
    $binaryPath = Download-Binary -Platform $platform -Version $targetVersion
    
    # Install binary
    Install-Binary -BinaryPath $binaryPath
    
    # Add to PATH
    Add-ToPath
    
    # Verify installation
    Test-Installation
    
    # Cleanup
    Cleanup
    
    Write-Success "xcp installation completed successfully!"
    Write-Info ""
    Write-Info "Get started with: xcp --help"
    Write-Info "Note: You may need to restart your terminal for the PATH changes to take effect."
}

# Run main function
Main