#!/bin/bash
# Build script for Repo Onboarding Copilot

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
APP_NAME="repo-onboarding-copilot"
BUILD_DIR="build"
DIST_DIR="dist"

# Get version info
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0")
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS="-ldflags \"-X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE} -X main.CommitHash=${COMMIT_HASH}\""

echo -e "${GREEN}Building ${APP_NAME}...${NC}"
echo "Version: ${VERSION}"
echo "Build Date: ${BUILD_DATE}"
echo "Commit: ${COMMIT_HASH}"
echo ""

# Clean previous builds
echo -e "${YELLOW}Cleaning previous builds...${NC}"
rm -rf ${BUILD_DIR} ${DIST_DIR}

# Install dependencies
echo -e "${YELLOW}Installing dependencies...${NC}"
go mod download
go mod tidy

# Build for current platform
echo -e "${YELLOW}Building for current platform...${NC}"
mkdir -p ${BUILD_DIR}
eval "go build -trimpath ${LDFLAGS} -o ${BUILD_DIR}/${APP_NAME} ./cmd"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Build completed successfully${NC}"
    echo "Binary location: ${BUILD_DIR}/${APP_NAME}"
else
    echo -e "${RED}✗ Build failed${NC}"
    exit 1
fi