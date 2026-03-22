#!/bin/sh
set -e

REPO="nattergabriel/reseed"
INSTALL_DIR="/usr/local/bin"
BINARY="reseed"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  linux) OS="linux" ;;
  darwin) OS="darwin" ;;
  *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

# Get latest version
VERSION=$(curl -sSf "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d '"' -f 4)
if [ -z "$VERSION" ]; then
  echo "Failed to fetch latest version" >&2
  exit 1
fi

VERSION_NUM="${VERSION#v}"
ARCHIVE="${BINARY}_${VERSION_NUM}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"

TMPDIR_PATH=$(mktemp -d)
trap 'rm -rf "$TMPDIR_PATH"' EXIT

echo "Downloading ${BINARY} ${VERSION} for ${OS}/${ARCH}..."
curl -sSfL "$URL" -o "${TMPDIR_PATH}/${ARCHIVE}"
tar -xzf "${TMPDIR_PATH}/${ARCHIVE}" -C "$TMPDIR_PATH"

echo "Installing to ${INSTALL_DIR}/${BINARY}..."
if ! install -m 755 "${TMPDIR_PATH}/${BINARY}" "${INSTALL_DIR}/${BINARY}" 2>/dev/null; then
  sudo install -m 755 "${TMPDIR_PATH}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
fi

echo "${BINARY} ${VERSION} installed successfully"
