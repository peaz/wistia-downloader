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
    
    # Set output filename
    OUTPUT_NAME="${APP_NAME}-${PLATFORM_NAME}-${GOARCH}"
    if [ ${GOOS} = "windows" ]; then
        OUTPUT_NAME="${OUTPUT_NAME}.exe"
    fi
    
    OUTPUT_PATH="${BUILD_DIR}/${OUTPUT_NAME}"
    
    echo -e "${YELLOW}Building for ${PLATFORM_NAME}/${GOARCH}...${NC}"
    
    # Build the binary
    env GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags="-s -w" -o ${OUTPUT_PATH} ./${SRC_DIR}
    
    if [ $? -eq 0 ]; then
        # Get file size
        SIZE=$(du -h ${OUTPUT_PATH} | cut -f1)
        echo -e "${GREEN}✓ Built ${OUTPUT_NAME} (${SIZE})${NC}"
    else
        echo -e "${RED}✗ Failed to build for ${PLATFORM_NAME}/${GOARCH}${NC}"
        exit 1
    fi
done

echo -e "${GREEN}Build complete! Binaries created in ${BUILD_DIR}/${NC}"
echo -e "${BLUE}Built binaries:${NC}"
ls -la ${BUILD_DIR}/

# Create archives for distribution
echo -e "${BLUE}Creating distribution archives...${NC}"
cd ${BUILD_DIR}

for file in *; do
    if [ -f "$file" ]; then
        # Extract platform info from filename
        if [[ $file =~ ${APP_NAME}-(.+)-(.+)(\..+)?$ ]]; then
            PLATFORM_NAME="${BASH_REMATCH[1]}"
            GOARCH="${BASH_REMATCH[2]}"
            EXT="${BASH_REMATCH[3]}"
            
            ARCHIVE_NAME="${APP_NAME}-${PLATFORM_NAME}-${GOARCH}-${VERSION}"
            
            if [ ${PLATFORM_NAME} = "windows" ]; then
                # Create zip for Windows
                zip -q "${ARCHIVE_NAME}.zip" "$file"
                echo -e "${GREEN}✓ Created ${ARCHIVE_NAME}.zip${NC}"
            else
                # Create tar.gz for Unix-like systems
                tar -czf "${ARCHIVE_NAME}.tar.gz" "$file"
                echo -e "${GREEN}✓ Created ${ARCHIVE_NAME}.tar.gz${NC}"
            fi
        fi
    fi
done

cd ..

echo -e "${GREEN}All builds completed successfully!${NC}"
echo -e "${BLUE}Distribution files:${NC}"
ls -la ${BUILD_DIR}/*.{tar.gz,zip} 2>/dev/null || true

# Display summary
echo -e "${BLUE}Build Summary:${NC}"
echo "  Version: ${VERSION}"
echo "  Platforms: macOS (Intel/ARM), Linux (Intel/ARM), Windows (Intel/ARM)"
echo "  Location: ${BUILD_DIR}/"
echo -e "${YELLOW}Tip: Set VERSION environment variable to customize version (e.g., VERSION=1.0.0 ./build.sh)${NC}"
