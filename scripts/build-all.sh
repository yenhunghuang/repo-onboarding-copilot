#!/bin/bash
# Cross-platform build script for Repo Onboarding Copilot

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

APP_NAME="repo-onboarding-copilot"
DIST_DIR="dist"

# Get version info
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0")
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

LDFLAGS="-ldflags \"-X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE} -X main.CommitHash=${COMMIT_HASH}\""

echo -e "${GREEN}Cross-platform build for ${APP_NAME}...${NC}"
echo "Version: ${VERSION}"
echo ""

# Clean and prepare
rm -rf ${DIST_DIR}
go mod download

# Build targets
declare -a platforms=("darwin/amd64" "darwin/arm64" "linux/amd64" "linux/arm64" "windows/amd64" "windows/arm64")

for platform in "${platforms[@]}"; do
    IFS='/' read -r -a platform_split <<< "$platform"
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    
    output_name="${APP_NAME}"
    if [ $GOOS = "windows" ]; then
        output_name+=".exe"
    fi
    
    output_path="${DIST_DIR}/${GOOS}-${GOARCH}/${output_name}"
    
    echo -e "${YELLOW}Building ${GOOS}/${GOARCH}...${NC}"
    mkdir -p "$(dirname "$output_path")"
    
    env GOOS=$GOOS GOARCH=$GOARCH go build -trimpath $LDFLAGS -o "$output_path" ./cmd
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ ${GOOS}/${GOARCH} build completed${NC}"
    else
        echo -e "${RED}✗ ${GOOS}/${GOARCH} build failed${NC}"
        exit 1
    fi
done

echo -e "${GREEN}All builds completed successfully!${NC}"