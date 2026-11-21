#!/bin/bash

# Build script for tayara-api v0.2.0
# Builds binaries for multiple platforms

VERSION="v0.2.0"
APP_NAME="tayara-api"
BUILD_DIR="build"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Building ${APP_NAME} ${VERSION}${NC}"

# Create build directory
mkdir -p ${BUILD_DIR}

# Platforms to build for
PLATFORMS=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
)

# Build for each platform
for PLATFORM in "${PLATFORMS[@]}"; do
    IFS='/' read -r -a PARTS <<< "$PLATFORM"
    GOOS="${PARTS[0]}"
    GOARCH="${PARTS[1]}"

    OUTPUT_NAME="${BUILD_DIR}/${APP_NAME}-${GOOS}-${GOARCH}"

    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME+='.exe'
    fi

    echo -e "${GREEN}Building for ${GOOS}/${GOARCH}...${NC}"

    GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w -X main.Version=${VERSION}" -o $OUTPUT_NAME main.go

    if [ $? -ne 0 ]; then
        echo "Error building for ${GOOS}/${GOARCH}"
        exit 1
    fi
done

# Generate checksums
echo -e "${BLUE}Generating checksums...${NC}"
cd ${BUILD_DIR}
shasum -a 256 ${APP_NAME}-* > checksums.txt
cd ..

# Package swagger docs
echo -e "${BLUE}Packaging swagger docs...${NC}"
tar -czf ${BUILD_DIR}/swagger-docs.tar.gz docs/

echo -e "${GREEN}Build complete! Artifacts in ${BUILD_DIR}/${NC}"
ls -lh ${BUILD_DIR}/
