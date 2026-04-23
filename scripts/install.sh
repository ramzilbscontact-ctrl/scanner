#!/usr/bin/env bash
# Agenzia Scanner — one-line installer for macOS / Linux
# Usage: curl -sSL https://api.getagenzia.fr/scanner/install.sh | sh
set -eu

REPO="ramzilbscontact-ctrl/scanner"
VERSION="${AGENZIA_VERSION:-latest}"

color() { printf '\033[%sm%s\033[0m' "$1" "$2"; }
info() { echo "$(color 36 "ℹ") $*"; }
ok()   { echo "$(color 32 "✓") $*"; }
err()  { echo "$(color 31 "✗") $*" >&2; exit 1; }

# ── Detect platform ───────────────────────────────
os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "$os" in
  linux)  platform="linux" ;;
  darwin) platform="darwin" ;;
  *) err "Unsupported OS: $os (only macOS and Linux supported here — Windows users: use install.ps1)" ;;
esac
case "$arch" in
  x86_64|amd64) archtag="x86_64" ;;
  arm64|aarch64) archtag="arm64" ;;
  *) err "Unsupported architecture: $arch" ;;
esac

info "Detected $platform/$archtag"

# ── Resolve version ──────────────────────────────
if [ "$VERSION" = "latest" ]; then
  VERSION="$(curl -sSL "https://api.github.com/repos/$REPO/releases/latest" \
    | grep '"tag_name":' \
    | head -1 \
    | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')"
  [ -z "$VERSION" ] && err "Could not resolve latest version"
fi
info "Installing $REPO@$VERSION"

# ── Download + extract ───────────────────────────
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
archive="agenzia-scan_${VERSION#v}_${platform}_${archtag}.tar.gz"
url="https://github.com/$REPO/releases/download/$VERSION/$archive"

info "Downloading $archive"
curl -fsSL "$url" -o "$tmp/archive.tar.gz" || err "Download failed — check version tag"
tar -xzf "$tmp/archive.tar.gz" -C "$tmp"

# ── Install ──────────────────────────────────────
dest="${AGENZIA_INSTALL_DIR:-/usr/local/bin}"
if [ ! -w "$dest" ]; then
  info "Installing to $dest (requires sudo)"
  sudo mv "$tmp/agenzia-scan" "$dest/agenzia-scan"
else
  mv "$tmp/agenzia-scan" "$dest/agenzia-scan"
fi
chmod +x "$dest/agenzia-scan"

ok "Installed $dest/agenzia-scan"
echo
info "Run a first scan now:"
echo "  $ agenzia-scan"
echo
info "Optional: upload to dashboard.agenzia.uk for continuous monitoring:"
echo "  $ agenzia-scan --upload --api-key \$AGENZIA_API_KEY"
echo
info "Star us on GitHub if this helps 🙏 https://github.com/$REPO"
