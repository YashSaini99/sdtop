#!/bin/bash
# sdtop installation script
# Usage: curl -sL https://raw.githubusercontent.com/YashSaini99/sdtop/master/install.sh | bash

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
REPO="YashSaini99/sdtop"
VERSION="${SDTOP_VERSION:-latest}"  # Use env var or latest
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$OS" != "linux" ]; then
    echo -e "${RED}✗ Error: This script only supports Linux${NC}"
    echo "  Detected OS: $OS"
    exit 1
fi

case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    armv7l|armv7*)
        ARCH="arm-v7"
        ;;
    *)
        echo -e "${RED}✗ Error: Unsupported architecture: $ARCH${NC}"
        echo "  Supported: x86_64, aarch64, armv7l"
        exit 1
        ;;
esac

echo -e "${GREEN}sdtop installer${NC}"
echo "=================="
echo "OS: $OS"
echo "Architecture: $ARCH"
echo "Install directory: $INSTALL_DIR"
echo ""

# Get latest version if not specified
if [ "$VERSION" = "latest" ]; then
    echo -e "${YELLOW}→${NC} Fetching latest version..."
    VERSION=$(curl -sL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/')
    if [ -z "$VERSION" ]; then
        echo -e "${RED}✗ Error: Could not determine latest version${NC}"
        exit 1
    fi
fi

echo -e "${YELLOW}→${NC} Version: v$VERSION"

# Construct download URL
FILENAME="sdtop-${VERSION}-${OS}-${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/v${VERSION}/${FILENAME}"

echo -e "${YELLOW}→${NC} Downloading from: $URL"

# Create temporary directory
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

# Download
if ! curl -fsSL "$URL" -o "$TMP_DIR/$FILENAME"; then
    echo -e "${RED}✗ Error: Download failed${NC}"
    echo "  URL: $URL"
    exit 1
fi

echo -e "${GREEN}✓${NC} Downloaded successfully"

# Extract
echo -e "${YELLOW}→${NC} Extracting..."
tar -xzf "$TMP_DIR/$FILENAME" -C "$TMP_DIR"

# Find the binary (it might be named differently after extraction)
BINARY=$(find "$TMP_DIR" -type f -name "sdtop*" -executable | head -n 1)
if [ -z "$BINARY" ]; then
    # If not found, try without executable flag (some systems don't preserve permissions in tar)
    BINARY=$(find "$TMP_DIR" -type f -name "sdtop*" | head -n 1)
fi

if [ -z "$BINARY" ]; then
    echo -e "${RED}✗ Error: Binary not found in archive${NC}"
    exit 1
fi

# Install
echo -e "${YELLOW}→${NC} Installing to $INSTALL_DIR..."

if [ -w "$INSTALL_DIR" ]; then
    # Directory is writable
    mv "$BINARY" "$INSTALL_DIR/sdtop"
    chmod +x "$INSTALL_DIR/sdtop"
else
    # Need sudo
    if ! command -v sudo &> /dev/null; then
        echo -e "${RED}✗ Error: $INSTALL_DIR is not writable and sudo is not available${NC}"
        exit 1
    fi
    sudo mv "$BINARY" "$INSTALL_DIR/sdtop"
    sudo chmod +x "$INSTALL_DIR/sdtop"
fi

echo -e "${GREEN}✓${NC} Installed successfully!"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}✓ sdtop v$VERSION installed!${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Run:"
echo -e "  ${GREEN}sdtop${NC}"
echo ""
echo "For help:"
echo -e "  ${GREEN}sdtop --help${NC}"
echo ""

# Verify installation
if command -v sdtop &> /dev/null; then
    VERSION_OUTPUT=$(sdtop --version 2>&1 || echo "unknown")
    echo "Installed version: $VERSION_OUTPUT"
else
    echo -e "${YELLOW}⚠${NC} Warning: sdtop not found in PATH"
    echo "  You may need to add $INSTALL_DIR to your PATH"
fi
