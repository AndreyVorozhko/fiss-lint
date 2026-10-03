<#
.SYNOPSIS
    Installs fiss-lint on Windows.
.DESCRIPTION
    Autonomous quick installer script for fiss-lint on Windows.
    Downloads the pre-compiled binary matching the system architecture from GitHub Releases
    and adds the installation directory to the User PATH environment variable.
.EXAMPLE
    irm https://raw.githubusercontent.com/AndreyVorozhko/fiss-lint/main/scripts/install.ps1 | iex
.EXAMPLE
    & .\scripts\install.ps1 -Version v1.0.0
#>

[CmdletBinding()]
param (
    [Parameter(Position = 0)]
    [string]$Version = $env:VERSION,

    [Parameter()]
    [string]$InstallDir = $env:INSTALL_DIR,

    [Parameter()]
    [string]$Repo = "AndreyVorozhko/fiss-lint"
)

$ErrorActionPreference = 'Stop'

# Ensure TLS 1.2+ for secure download on legacy Windows PowerShell 5.1
try {
    [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor [System.Net.SecurityProtocolType]::Tls12
} catch {
    # Ignore if not supported / already modern
}

# Determine requested version
if (-not $Version) {
    $Version = "latest"
} elseif ($Version -match '^\d') {
    $Version = "v$Version"
}

# Detect architecture
$arch = $env:PROCESSOR_ARCHITECTURE
switch -Regex ($arch) {
    'AMD64|x86_64' { $targetArch = 'amd64' }
    'ARM64|aarch64' { $targetArch = 'arm64' }
    default {
        Write-Error "Unsupported processor architecture '$arch'. fiss-lint supports amd64 and arm64."
        exit 1
    }
}

$binName = "fiss-lint.exe"
$assetName = "fiss-lint-windows-$targetArch.exe"

# Determine installation directory
if (-not $InstallDir) {
    if ($env:LOCALAPPDATA) {
        $InstallDir = Join-Path $env:LOCALAPPDATA "Programs\fiss-lint"
    } else {
        $InstallDir = Join-Path $HOME "bin"
    }
}

# Construct download URLs
if ($Version -eq 'latest') {
    $downloadUrl = "https://github.com/$Repo/releases/latest/download/$assetName"
    $checksumUrl = "https://github.com/$Repo/releases/latest/download/sha256sums"
} else {
    $downloadUrl = "https://github.com/$Repo/releases/download/$Version/$assetName"
    $checksumUrl = "https://github.com/$Repo/releases/download/$Version/sha256sums"
}

Write-Host "=== fiss-lint installer (Windows) ===" -ForegroundColor Cyan
Write-Host "Target Arch:  $targetArch"
Write-Host "Requested:    $Version"
Write-Host "Install path: $(Join-Path $InstallDir $binName)"
Write-Host "Downloading binary from $downloadUrl..."

# Create temporary directory
$tmpDir = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null

try {
    $tmpBin = Join-Path $tmpDir $assetName
    Invoke-WebRequest -Uri $downloadUrl -OutFile $tmpBin -UseBasicParsing

    # Optional SHA256 checksum verification
    $tmpChecksums = Join-Path $tmpDir "sha256sums"
    $checksumFound = $false
    try {
        Invoke-WebRequest -Uri $checksumUrl -OutFile $tmpChecksums -UseBasicParsing -ErrorAction SilentlyContinue
        if (Test-Path $tmpChecksums) {
            $checksumFound = $true
        }
    } catch {
        $checksumFound = $false
    }

    if ($checksumFound -and (Test-Path $tmpChecksums)) {
        Write-Host "Verifying SHA256 checksum..."
        $hashResult = (Get-FileHash -Path $tmpBin -Algorithm SHA256).Hash.ToLower()
        $expectedLine = Get-Content -Path $tmpChecksums | Where-Object { $_ -match "$assetName$" }
        if ($expectedLine) {
            $expectedHash = ($expectedLine -split '\s+')[0].ToLower()
            if ($hashResult -ne $expectedHash) {
                Write-Error "SHA256 checksum verification failed for $assetName! Expected: $expectedHash, Actual: $hashResult"
                exit 1
            }
            Write-Host "Checksum verified successfully." -ForegroundColor Green
        }
    }

    # Ensure installation directory exists
    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }

    $targetFile = Join-Path $InstallDir $binName
    Move-Item -Path $tmpBin -Destination $targetFile -Force

    # Verify execution
    try {
        $installedVersionOutput = & "$targetFile" --version 2>&1
    } catch {
        Write-Error "Installed binary failed execution test: $targetFile --version"
        exit 1
    }

    # Update User Environment PATH if needed
    $userPath = [System.Environment]::GetEnvironmentVariable("Path", [System.EnvironmentVariableTarget]::User)
    $pathList = ($userPath -split ';') | Where-Object { $_ -ne '' }
    $inPath = $pathList -contains $InstallDir

    if (-not $inPath) {
        $newUserPath = ($pathList + $InstallDir) -join ';'
        [System.Environment]::SetEnvironmentVariable("Path", $newUserPath, [System.EnvironmentVariableTarget]::User)
        $env:Path = "$env:Path;$InstallDir"
        Write-Host "Added '$InstallDir' to your User PATH environment variable." -ForegroundColor Yellow
    }

    Write-Host "`nfiss-lint was successfully installed!" -ForegroundColor Green
    Write-Host $installedVersionOutput
    Write-Host ""

    if (-not $inPath) {
        Write-Host "Please restart your terminal or PowerShell session for PATH changes to take effect." -ForegroundColor Yellow
    } else {
        Write-Host "You can now run 'fiss-lint' directly from your terminal." -ForegroundColor Green
    }
} finally {
    if (Test-Path $tmpDir) {
        Remove-Item -Path $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}
