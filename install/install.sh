#!/usr/bin/env bash
set -euo pipefail

REPO="${LANDROP_REPO:-wingoo/landrop}"
VERSION="${VERSION:-latest}"

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

case "$OS" in
  darwin)
    ASSET="landrop_darwin_${ARCH}.tar.gz"
    ;;
  linux)
    ASSET="landrop_linux_${ARCH}.tar.gz"
    ;;
  *)
    echo "Unsupported OS: $OS"
    exit 1
    ;;
esac

if [[ "$VERSION" == "latest" ]]; then
  DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/${ASSET}"
  CHECKSUM_URL="https://github.com/${REPO}/releases/latest/download/checksums.txt"
else
  DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET}"
  CHECKSUM_URL="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"
fi

echo "Resolving release: $REPO ($VERSION)"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

ARCHIVE_PATH="$TMP_DIR/$ASSET"
CHECKSUM_PATH="$TMP_DIR/checksums.txt"
if ! curl -fL "$DOWNLOAD_URL" -o "$ARCHIVE_PATH"; then
  echo "Could not download asset: ${ASSET} (${DOWNLOAD_URL})"
  exit 1
fi
if ! curl -fL "$CHECKSUM_URL" -o "$CHECKSUM_PATH"; then
  echo "Could not download checksums.txt (${CHECKSUM_URL})"
  exit 1
fi

EXPECTED_SHA="$(grep -E "[[:space:]]${ASSET}\$" "$CHECKSUM_PATH" | awk '{print $1}' || true)"
if [[ -z "$EXPECTED_SHA" ]]; then
  echo "Could not find checksum entry for ${ASSET}"
  exit 1
fi

if command -v sha256sum >/dev/null 2>&1; then
  ACTUAL_SHA="$(sha256sum "$ARCHIVE_PATH" | awk '{print $1}')"
elif command -v shasum >/dev/null 2>&1; then
  ACTUAL_SHA="$(shasum -a 256 "$ARCHIVE_PATH" | awk '{print $1}')"
else
  echo "No SHA-256 tool found (need sha256sum or shasum)"
  exit 1
fi

if [[ "$ACTUAL_SHA" != "$EXPECTED_SHA" ]]; then
  echo "Checksum verification failed for ${ASSET}"
  exit 1
fi

tar -xzf "$ARCHIVE_PATH" -C "$TMP_DIR"

BIN_SRC="$TMP_DIR/landrop"
if [[ ! -f "$BIN_SRC" ]]; then
  echo "landrop binary not found in archive"
  exit 1
fi

INSTALL_DIR="/usr/local/bin"
if [[ ! -w "$INSTALL_DIR" ]]; then
  INSTALL_DIR="$HOME/.local/bin"
  mkdir -p "$INSTALL_DIR"
fi

install -m 0755 "$BIN_SRC" "$INSTALL_DIR/landrop"

echo "Installed: $INSTALL_DIR/landrop"
if [[ "$INSTALL_DIR" == "$HOME/.local/bin" ]]; then
  echo "Add to PATH if needed: export PATH=\"$HOME/.local/bin:\$PATH\""
fi

echo "Run: landrop --help"
