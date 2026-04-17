#!/usr/bin/env bash
set -euo pipefail

REPO="n0bisuke/youtube-cli"
INSTALL_DIR="/usr/local/bin"

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$OS" in
  darwin) OS="darwin" ;;
  linux)  OS="linux" ;;
  *)
    echo "Error: unsupported OS: $OS"
    exit 1
    ;;
esac

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)
    echo "Error: unsupported architecture: $ARCH"
    exit 1
    ;;
esac

BINARY="ysh-${OS}-${ARCH}"
if [ "$OS" = "windows" ]; then
  BINARY="${BINARY}.exe"
fi

# Get latest release tag from GitHub API
echo "Checking latest release..."
LATEST_TAG="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | head -1 | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')"

if [ -z "$LATEST_TAG" ]; then
  echo "Error: could not find latest release"
  exit 1
fi

DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_TAG}/${BINARY}"

echo "Downloading ysh ${LATEST_TAG} for ${OS}/${ARCH}..."
curl -fsSL -o /tmp/ysh "${DOWNLOAD_URL}"
chmod +x /tmp/ysh

echo "Installing to ${INSTALL_DIR}/ysh..."
if [ -w "$INSTALL_DIR" ]; then
  mv /tmp/ysh "${INSTALL_DIR}/ysh"
else
  echo "Need sudo to install to ${INSTALL_DIR}:"
  sudo mv /tmp/ysh "${INSTALL_DIR}/ysh"
fi

echo "Installed: $(ysh --version 2>/dev/null || echo "ysh ${LATEST_TAG}")"
echo "Done! Run: ysh"