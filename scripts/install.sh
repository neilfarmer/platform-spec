#!/bin/bash
# Installation script for platform-spec
# Usage:
#   curl -sSL https://raw.githubusercontent.com/neilfarmer/platform-spec/main/scripts/install.sh | bash
#   curl -sSL https://raw.githubusercontent.com/neilfarmer/platform-spec/main/scripts/install.sh | bash -s -- --dir ~/.bin

set -e

# Default installation directory
INSTALL_DIR="/usr/local/bin"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --dir|--prefix)
      INSTALL_DIR="$2"
      shift 2
      ;;
    -h|--help)
      echo "Usage: $0 [--dir DIRECTORY]"
      echo ""
      echo "Options:"
      echo "  --dir, --prefix DIRECTORY    Install binary to specified directory (default: /usr/local/bin)"
      echo "  -h, --help                   Show this help message"
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      echo "Use --help for usage information"
      exit 1
      ;;
  esac
done

# Expand tilde in path
INSTALL_DIR="${INSTALL_DIR/#\~/$HOME}"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
  darwin)
    OS="darwin"
    ;;
  linux)
    OS="linux"
    ;;
  *)
    echo "Unsupported OS: $OS"
    exit 1
    ;;
esac

case "$ARCH" in
  x86_64|amd64)
    ARCH="amd64"
    ;;
  aarch64|arm64)
    ARCH="arm64"
    ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

# Get latest release info from GitHub API
echo "Fetching latest release..."
RELEASE_URL="https://api.github.com/repos/neilfarmer/platform-spec/releases/latest"
DOWNLOAD_URL=$(curl -s "$RELEASE_URL" | grep "browser_download_url.*${OS}_${ARCH}" | cut -d '"' -f 4)

if [ -z "$DOWNLOAD_URL" ]; then
  echo "Error: Could not find download URL for ${OS}_${ARCH}"
  exit 1
fi

echo "Downloading platform-spec from: $DOWNLOAD_URL"

# Create temporary directory
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

# Download and extract
if [[ "$DOWNLOAD_URL" == *.zip ]]; then
  curl -sSL "$DOWNLOAD_URL" -o platform-spec.zip
  unzip -q platform-spec.zip
  rm platform-spec.zip
elif [[ "$DOWNLOAD_URL" == *.tar.gz ]]; then
  curl -sSL "$DOWNLOAD_URL" | tar xz
fi

# Create installation directory if it doesn't exist
if [ ! -d "$INSTALL_DIR" ]; then
  echo "Creating directory: $INSTALL_DIR"
  mkdir -p "$INSTALL_DIR"
fi

# Install
echo "Installing platform-spec to $INSTALL_DIR..."
if [ -w "$INSTALL_DIR" ]; then
  mv platform-spec "$INSTALL_DIR/platform-spec"
else
  sudo mv platform-spec "$INSTALL_DIR/platform-spec"
fi

# Cleanup
cd - > /dev/null
rm -rf "$TMP_DIR"

echo "✓ platform-spec installed successfully to $INSTALL_DIR/platform-spec"
echo ""

# Check if install directory is in PATH
if [[ ":$PATH:" == *":$INSTALL_DIR:"* ]]; then
  echo "Verifying installation..."
  platform-spec version
else
  echo "⚠  Warning: $INSTALL_DIR is not in your PATH"
  echo "   Add it to your PATH by adding this line to your ~/.bashrc or ~/.zshrc:"
  echo "   export PATH=\"\$PATH:$INSTALL_DIR\""
  echo ""
  echo "   Or run the binary directly:"
  echo "   $INSTALL_DIR/platform-spec version"
fi
