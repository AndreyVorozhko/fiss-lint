#!/usr/bin/env sh
# fiss-lint quick install script for Linux and macOS
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/AndreyVorozhko/fiss-lint/main/scripts/install.sh | bash
#   curl -fsSL https://raw.githubusercontent.com/AndreyVorozhko/fiss-lint/main/scripts/install.sh | bash -s -- v1.0.0
#   VERSION=v1.0.0 curl -fsSL ... | bash
#   INSTALL_DIR=/usr/local/bin curl -fsSL ... | bash

set -eu

REPO="${REPO:-AndreyVorozhko/fiss-lint}"
BIN_NAME="fiss-lint"

# Determine requested version
VERSION="${1:-${VERSION:-latest}}"
case "$VERSION" in
  latest)
    ;;
  [0-9]*)
    VERSION="v${VERSION}"
    ;;
  v*)
    ;;
  *)
    echo "Error: Invalid version format '$VERSION'." >&2
    exit 1
    ;;
esac

# Detect Operating System
OS_RAW="$(uname -s)"
case "$OS_RAW" in
  Linux|linux)
    OS="linux"
    ;;
  Darwin|darwin)
    OS="darwin"
    ;;
  *)
    echo "Error: Unsupported operating system '$OS_RAW'." >&2
    echo "fiss-lint installer currently supports Linux and macOS (Darwin)." >&2
    exit 1
    ;;
esac

# Detect Architecture
ARCH_RAW="$(uname -m)"
case "$ARCH_RAW" in
  x86_64|amd64)
    ARCH="amd64"
    ;;
  aarch64|arm64)
    ARCH="arm64"
    ;;
  *)
    echo "Error: Unsupported architecture '$ARCH_RAW'." >&2
    echo "fiss-lint installer currently supports amd64 (x86_64) and arm64 (aarch64)." >&2
    exit 1
    ;;
esac

ASSET_NAME="${BIN_NAME}-${OS}-${ARCH}"

# Determine installation directory
if [ -z "${INSTALL_DIR:-}" ]; then
  if [ -w "/usr/local/bin" ] && [ "$(id -u)" -eq 0 ]; then
    INSTALL_DIR="/usr/local/bin"
  elif [ -w "/usr/local/bin" ] && [ -d "/usr/local/bin" ]; then
    INSTALL_DIR="/usr/local/bin"
  else
    INSTALL_DIR="${HOME}/.local/bin"
  fi
fi

# Detect download tool
if command -v curl >/dev/null 2>&1; then
  DOWNLOAD_TOOL="curl"
elif command -v wget >/dev/null 2>&1; then
  DOWNLOAD_TOOL="wget"
else
  echo "Error: Neither 'curl' nor 'wget' was found. Please install curl or wget." >&2
  exit 1
fi

# Construct download URLs
if [ "$VERSION" = "latest" ]; then
  DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/${ASSET_NAME}"
  CHECKSUM_URL="https://github.com/${REPO}/releases/latest/download/sha256sums"
else
  DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET_NAME}"
  CHECKSUM_URL="https://github.com/${REPO}/releases/download/${VERSION}/sha256sums"
fi

# Create temporary directory
TMP_DIR="$(mktemp -d 2>/dev/null || mktemp -d -t 'fiss-lint-install')"
cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT INT TERM

TMP_BIN="${TMP_DIR}/${ASSET_NAME}"
TMP_CHECKSUMS="${TMP_DIR}/sha256sums"

echo "=== fiss-lint installer ==="
echo "Target OS:       ${OS}"
echo "Target Arch:     ${ARCH}"
echo "Requested:       ${VERSION}"
echo "Install path:    ${INSTALL_DIR}/${BIN_NAME}"
echo "Downloading binary from ${DOWNLOAD_URL}..."

# Download binary
if [ "$DOWNLOAD_TOOL" = "curl" ]; then
  if ! curl -fsSL "$DOWNLOAD_URL" -o "$TMP_BIN"; then
    echo "Error: Failed to download ${DOWNLOAD_URL}" >&2
    echo "Please verify that version '${VERSION}' and asset '${ASSET_NAME}' exist." >&2
    exit 1
  fi
else
  if ! wget -qO "$TMP_BIN" "$DOWNLOAD_URL"; then
    echo "Error: Failed to download ${DOWNLOAD_URL}" >&2
    echo "Please verify that version '${VERSION}' and asset '${ASSET_NAME}' exist." >&2
    exit 1
  fi
fi

# Optional SHA256 verification if sha256sums file is present in release
CHECKSUM_FOUND=0
if [ "$DOWNLOAD_TOOL" = "curl" ]; then
  if curl -fsSL "$CHECKSUM_URL" -o "$TMP_CHECKSUMS" 2>/dev/null; then
    CHECKSUM_FOUND=1
  fi
else
  if wget -qO "$TMP_CHECKSUMS" "$CHECKSUM_URL" 2>/dev/null; then
    CHECKSUM_FOUND=1
  fi
fi

if [ "$CHECKSUM_FOUND" -eq 1 ] && [ -s "$TMP_CHECKSUMS" ]; then
  echo "Verifying SHA256 checksum..."
  if command -v sha256sum >/dev/null 2>&1; then
    if ! (cd "$TMP_DIR" && grep -E "[[:space:]]${ASSET_NAME}\$" sha256sums | sha256sum -c --status 2>/dev/null); then
      echo "Error: SHA256 checksum verification failed for ${ASSET_NAME}!" >&2
      exit 1
    fi
    echo "Checksum verified successfully."
  elif command -v shasum >/dev/null 2>&1; then
    if ! (cd "$TMP_DIR" && grep -E "[[:space:]]${ASSET_NAME}\$" sha256sums | shasum -a 256 -c --status 2>/dev/null); then
      echo "Error: SHA256 checksum verification failed for ${ASSET_NAME}!" >&2
      exit 1
    fi
    echo "Checksum verified successfully."
  fi
fi

# Ensure target directory exists
mkdir -p "$INSTALL_DIR"

# Install binary
chmod +x "$TMP_BIN"
mv "$TMP_BIN" "${INSTALL_DIR}/${BIN_NAME}"

# Verify installation
if ! "${INSTALL_DIR}/${BIN_NAME}" --version >/dev/null 2>&1; then
  echo "Error: Installed binary failed execution test: ${INSTALL_DIR}/${BIN_NAME} --version" >&2
  exit 1
fi

# Check if INSTALL_DIR is in PATH
case ":${PATH}:" in
  *":${INSTALL_DIR}:"*)
    IN_PATH=1
    ;;
  *)
    IN_PATH=0
    ;;
esac

echo ""
echo "fiss-lint was successfully installed!"
"${INSTALL_DIR}/${BIN_NAME}" --version
echo ""

if [ "$IN_PATH" -eq 0 ]; then
  echo "Notice: '${INSTALL_DIR}' is not in your PATH."
  echo "To add it to your PATH, add the following line to your shell profile (~/.bashrc, ~/.zshrc, or ~/.profile):"
  echo ""
  echo "    export PATH=\"${INSTALL_DIR}:\$PATH\""
  echo ""
  echo "Then run: export PATH=\"${INSTALL_DIR}:\$PATH\""
else
  echo "You can now run 'fiss-lint' from your terminal."
fi
