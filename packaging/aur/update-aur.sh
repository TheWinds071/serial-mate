#!/bin/bash
# Script to update AUR PKGBUILD files with new version and checksums

set -e

VERSION="${1:-}"
if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 1.0.0"
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEMP_DIR=$(mktemp -d)

cleanup() {
    rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

echo "Updating AUR PKGBUILD files for version $VERSION"

# Download source tarball to calculate checksum
echo "Downloading source tarball..."
SOURCE_URL="https://github.com/TheWinds071/serial-mate/archive/v${VERSION}.tar.gz"
wget -q "$SOURCE_URL" -O "$TEMP_DIR/serial-mate-${VERSION}.tar.gz"

# Calculate SHA256 checksum
echo "Calculating SHA256 checksum..."
SHA256SUM=$(sha256sum "$TEMP_DIR/serial-mate-${VERSION}.tar.gz" | awk '{print $1}')
echo "SHA256: $SHA256SUM"

# Update PKGBUILD
echo "Updating PKGBUILD..."
sed -i "s/^pkgver=.*/pkgver=${VERSION}/" "$SCRIPT_DIR/PKGBUILD"
sed -i "s/^sha256sums=.*/sha256sums=('${SHA256SUM}')/" "$SCRIPT_DIR/PKGBUILD"

# Generate .SRCINFO
echo "Generating .SRCINFO..."
cd "$SCRIPT_DIR"
if command -v makepkg >/dev/null 2>&1; then
    makepkg --printsrcinfo > .SRCINFO
    echo "✓ .SRCINFO generated"
else
    echo "⚠ makepkg not found, skipping .SRCINFO generation"
    echo "  Run 'makepkg --printsrcinfo > .SRCINFO' manually on an Arch system"
fi

echo ""
echo "✓ AUR PKGBUILD files updated successfully!"
echo ""
echo "Next steps:"
echo "1. Review the changes: git diff packaging/aur/"
echo "2. Test the PKGBUILD: cd packaging/aur && makepkg -si"
echo "3. Commit and publish to AUR"
