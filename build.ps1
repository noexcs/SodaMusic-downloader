# Cross-platform build script for sodamusic-downloader (PowerShell)
# Produces static single-file executables for Windows, macOS, and Linux

$ErrorActionPreference = "Stop"

# Configuration
$AppName = "sodamusic-downloader"
$OutputDir = "build"

# Colors
function Write-Color {
    param(
        [string]$Text,
        [string]$Color = "White"
    )
    switch ($Color) {
        "Red" { Write-Host $Text -ForegroundColor Red }
        "Green" { Write-Host $Text -ForegroundColor Green }
        "Yellow" { Write-Host $Text -ForegroundColor Yellow }
        "Blue" { Write-Host $Text -ForegroundColor Blue }
        default { Write-Host $Text }
    }
}

Write-Color "========================================" "Blue"
Write-Color "Building $AppName for multiple platforms" "Blue"
Write-Color "========================================" "Blue"
Write-Host ""

# Create output directory
if (!(Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir | Out-Null
}

# Clean old builds
Write-Color "Cleaning old builds..." "Yellow"
Get-ChildItem -Path $OutputDir -File | Remove-Item -Force
Write-Host ""

# Platform configurations
$platforms = @(
    @{OS="windows"; Arch="amd64"; Ext=".exe"},
    @{OS="windows"; Arch="arm64"; Ext=".exe"},
    @{OS="darwin"; Arch="amd64"; Ext=""},
    @{OS="darwin"; Arch="arm64"; Ext=""},
    @{OS="linux"; Arch="amd64"; Ext=""},
    @{OS="linux"; Arch="arm64"; Ext=""}
)

# Build function
function Build-Platform {
    param(
        [string]$os,
        [string]$arch,
        [string]$ext
    )
    
    $outputName = "$AppName-$os-$arch$ext"
    
    Write-Color "Building for $os/$arch..." "Yellow"
    
    $env:GOOS = $os
    $env:GOARCH = $arch
    $env:CGO_ENABLED = "0"
    
    try {
        go build `
            -ldflags "-s -w" `
            -trimpath `
            -o "$OutputDir\$outputName" `
            .
        
        if ($LASTEXITCODE -eq 0) {
            Write-Color "✓ Successfully built $outputName" "Green"
            
            # Show file size
            $fileSize = (Get-Item "$OutputDir\$outputName").Length / 1MB
            Write-Host "  Size: $([math]::Round($fileSize, 2)) MB"
            Write-Host ""
        } else {
            Write-Color "✗ Failed to build for $os/$arch" "Red"
            exit 1
        }
    } catch {
        Write-Color "✗ Error building for $os/$arch : $_" "Red"
        exit 1
    } finally {
        # Clear environment variables
        Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
        Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue
        Remove-Item Env:\CGO_ENABLED -ErrorAction SilentlyContinue
    }
}

# Build all platforms
foreach ($platform in $platforms) {
    Build-Platform -os $platform.OS -arch $platform.Arch -ext $platform.Ext
}

# Generate checksums
Write-Color "Generating checksums..." "Yellow"
try {
    Set-Location $OutputDir
    $files = Get-ChildItem -File | Where-Object { $_.Name -ne "CHECKSUMS_SHA256.txt" }
    $checksums = @()
    
    foreach ($file in $files) {
        $hash = Get-FileHash -Path $file.Name -Algorithm SHA256
        $checksums += "$($hash.Hash) $($file.Name)"
    }
    
    $checksums | Set-Content "CHECKSUMS_SHA256.txt"
    Write-Color "✓ SHA256 checksums generated" "Green"
    Set-Location ..
} catch {
    Write-Color "⚠ Warning: Could not generate checksums" "Yellow"
}

Write-Host ""
Write-Color "========================================" "Blue"
Write-Color "✓ All builds completed successfully!" "Green"
Write-Color "========================================" "Blue"
Write-Host ""
Write-Color "Build outputs are in the $OutputDir directory:" "Green"
Get-ChildItem -Path $OutputDir | Format-Table Name, Length, LastWriteTime -AutoSize
