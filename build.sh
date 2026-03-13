#!/bin/bash

# Cross-platform build script for sodamusic-downloader
# Produces static single-file executables for Windows, macOS, and Linux

set -e

# Configuration
APP_NAME="sodamusic-downloader"
OUTPUT_DIR="build"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Building ${APP_NAME} for multiple platforms${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Create output directory
mkdir -p "${OUTPUT_DIR}"

# Clean old builds
echo -e "${YELLOW}Cleaning old builds...${NC}"
rm -f "${OUTPUT_DIR}"/*
echo ""

# Function to build for a specific platform
build_platform() {
    local os=$1
    local arch=$2
    local ext=""
    local output_name=""
    
    if [ "$os" = "windows" ]; then
        ext=".exe"
        output_name="${APP_NAME}-${os}-${arch}${ext}"
    elif [ "$os" = "darwin" ]; then
        output_name="${APP_NAME}-${os}-${arch}"
    else
        output_name="${APP_NAME}-${os}-${arch}"
    fi
    
    echo -e "${YELLOW}Building for ${os}/${arch}...${NC}"
    
    # Set environment variables and build
    GOOS=${os} \
    GOARCH=${arch} \
    CGO_ENABLED=0 \
    go build \
        -ldflags="-s -w" \
        -trimpath \
        -o "${OUTPUT_DIR}/${output_name}" \
        .
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Successfully built ${output_name}${NC}"
        
        # Show file size
        if [ "$os" = "windows" ]; then
            size=$(powershell -Command "(Get-Item '${OUTPUT_DIR}/${output_name}').Length / 1MB")
            echo -e "  Size: ${size} MB"
        else
            size=$(ls -lh "${OUTPUT_DIR}/${output_name}" | awk '{print $5}')
            echo -e "  Size: ${size}"
        fi
        echo ""
    else
        echo -e "${RED}✗ Failed to build for ${os}/${arch}${NC}"
        exit 1
    fi
}

# Build for Windows (x64)
build_platform "windows" "amd64"

# Build for Windows (ARM64)
build_platform "windows" "arm64"

# Build for macOS (x64)
build_platform "darwin" "amd64"

# Build for macOS (ARM64/M1/M2)
build_platform "darwin" "arm64"

# Build for Linux (x64)
build_platform "linux" "amd64"

# Build for Linux (ARM64)
build_platform "linux" "arm64"

# Create checksums
echo -e "${YELLOW}Generating checksums...${NC}"
cd "${OUTPUT_DIR}"

if command -v sha256sum &> /dev/null; then
    sha256sum * > CHECKSUMS_SHA256.txt
    echo -e "${GREEN}✓ SHA256 checksums generated${NC}"
elif command -v shasum &> /dev/null; then
    shasum -a 256 * > CHECKSUMS_SHA256.txt
    echo -e "${GREEN}✓ SHA256 checksums generated${NC}"
else
    echo -e "${YELLOW}⚠ Warning: Could not generate checksums (sha256sum not found)${NC}"
fi

cd ..

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}✓ All builds completed successfully!${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "Build outputs are in the ${GREEN}${OUTPUT_DIR}${NC} directory:"
ls -la "${OUTPUT_DIR}"
