#!/bin/sh
# gg installer - https://github.com/cyclecore-dev/gg
# Usage: curl -fsSL https://raw.githubusercontent.com/cyclecore-dev/gg/main/gg.sh | sh
set -e

VERSION="v0.9.2"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

echo "Installing gg $VERSION — the 2-letter agent-native git client"

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)  ARCH="x86_64" ;;
    aarch64) ARCH="aarch64" ;;
    arm64)   ARCH="aarch64" ;;
    *)       echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
    linux|darwin) ;;
    *)            echo "Unsupported OS: $OS"; exit 1 ;;
esac

BINARY="gg_${OS}_${ARCH}"
URL="https://github.com/cyclecore-dev/gg/releases/download/${VERSION}/${BINARY}"

echo "Downloading $BINARY..."
curl -fsSL "$URL" -o /tmp/gg || { echo "Download failed"; exit 1; }

# Install with sudo if needed
if [ -w "$INSTALL_DIR" ]; then
    mv /tmp/gg "$INSTALL_DIR/gg"
    chmod +x "$INSTALL_DIR/gg"
else
    sudo mv /tmp/gg "$INSTALL_DIR/gg"
    sudo chmod +x "$INSTALL_DIR/gg"
fi

echo "✓ gg $VERSION installed to $INSTALL_DIR/gg"
echo "  Run 'gg init' to configure, or 'gg maaza' for status"
