#!/bin/bash

# Build script for wistia-downloader
# Generates binaries for macOS, Linux, and Windows on both ARM64 and AMD64 architectures

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Project info
APP_NAME="wistia-downloader"
VERSION=${VERSION:-$(date +%Y%m%d-%H%M%S)}
BUILD_DIR="build"
SRC_DIR="src"

echo -e "${BLUE}Building ${APP_NAME} v${VERSION}${NC}"

# Clean and create build directory
rm -rf ${BUILD_DIR}
mkdir -p ${BUILD_DIR}

# Define target platforms and architectures
declare -a PLATFORMS=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
    "windows/arm64"
)

# Build for each platform
for PLATFORM in "${PLATFORMS[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "${PLATFORM}"
    
    # Map darwin to macos for user-friendly naming
    PLATFORM_NAME=${GOOS}
    if [ ${GOOS} = "darwin" ]; then
        PLATFORM_NAME="macos"
    fi
    
    # Set binary filename (always wistia-downloader with .exe for Windows)
    BINARY_NAME="${APP_NAME}"
    if [ ${GOOS} = "windows" ]; then
        BINARY_NAME="${BINARY_NAME}.exe"
    fi
    
    OUTPUT_PATH="${BUILD_DIR}/${BINARY_NAME}"
    
    echo -e "${YELLOW}Building for ${PLATFORM_NAME}/${GOARCH}...${NC}"
    
    # Build the binary
    env GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags="-s -w" -o ${OUTPUT_PATH} ./${SRC_DIR}
    
    if [ $? -eq 0 ]; then
        # Get file size
        SIZE=$(du -h ${OUTPUT_PATH} | cut -f1)
        echo -e "${GREEN}✓ Built ${BINARY_NAME} for ${PLATFORM_NAME}/${GOARCH} (${SIZE})${NC}"
        
        # Create archive immediately
        ARCHIVE_NAME="${APP_NAME}-${PLATFORM_NAME}-${GOARCH}-${VERSION}"
        
        cd ${BUILD_DIR}
        if [ ${GOOS} = "windows" ]; then
            # Create zip for Windows
            zip -q "${ARCHIVE_NAME}.zip" "${BINARY_NAME}"
            echo -e "${GREEN}✓ Created ${ARCHIVE_NAME}.zip${NC}"
        else
            # Create tar.gz for Unix-like systems
            tar -czf "${ARCHIVE_NAME}.tar.gz" "${BINARY_NAME}"
            echo -e "${GREEN}✓ Created ${ARCHIVE_NAME}.tar.gz${NC}"
        fi
        
        # Remove the binary file to keep only archives
        rm "${BINARY_NAME}"
        cd ..
        
    else
        echo -e "${RED}✗ Failed to build for ${PLATFORM_NAME}/${GOARCH}${NC}"
        exit 1
    fi
done

echo -e "${GREEN}Build complete! All platform archives created in ${BUILD_DIR}/${NC}"

# Display final summary
echo -e "${BLUE}Distribution files:${NC}"
ls -la ${BUILD_DIR}/*.{tar.gz,zip} 2>/dev/null || true

# Display summary
echo -e "${BLUE}Build Summary:${NC}"
echo "  Version: ${VERSION}"
echo "  Platforms: macOS (Intel/ARM), Linux (Intel/ARM), Windows (Intel/ARM)"
echo "  Binary name: ${APP_NAME} (wistia-downloader.exe for Windows)"
echo "  Location: ${BUILD_DIR}/"
echo -e "${YELLOW}Note: All binaries are named '${APP_NAME}' when extracted from archives${NC}"
echo -e "${YELLOW}Tip: Set VERSION environment variable to customize version (e.g., VERSION=1.0.0 ./build.sh)${NC}"
