#!/bin/bash
# Upload sdtop source to Launchpad PPA (run from repository root)
# Launchpad will build the actual .deb packages

set -e

VERSION="1.0.1"
MAINTAINER="Yash Saini <yashsaini99@example.com>"
DISTROS=("noble" "jammy" "focal")  # Ubuntu 24.04, 22.04, 20.04

# Get GPG key ID
GPG_KEY=$(gpg --list-secret-keys --keyid-format LONG | grep sec | awk '{print $2}' | cut -d'/' -f2 | head -n1)

if [ -z "$GPG_KEY" ]; then
    echo "❌ No GPG key found. Run: gpg --full-generate-key"
    exit 1
fi

echo "Using GPG key: $GPG_KEY"
echo "Building source packages for Ubuntu releases..."

# Create temporary build directory
BUILD_DIR=$(mktemp -d)
trap "rm -rf $BUILD_DIR" EXIT

for DISTRO in "${DISTROS[@]}"; do
    echo ""
    echo "📦 Building for Ubuntu $DISTRO..."
    
    # Copy source to build directory
    PACKAGE_DIR="$BUILD_DIR/sdtop-${VERSION}"
    mkdir -p "$PACKAGE_DIR"
    
    # Create orig tarball
    git archive --format=tar --prefix=sdtop-${VERSION}/ v${VERSION} | gzip > "$BUILD_DIR/sdtop_${VERSION}.orig.tar.gz"
    
    # Extract and add debian directory
    tar -xzf "$BUILD_DIR/sdtop_${VERSION}.orig.tar.gz" -C "$BUILD_DIR"
    cp -r packaging/debian "$PACKAGE_DIR/"
    
    # Update changelog for this Ubuntu release
    cd "$PACKAGE_DIR"
    
    # Create changelog entry
    cat > debian/changelog << EOF
sdtop (${VERSION}-0ubuntu1~${DISTRO}1) ${DISTRO}; urgency=medium

  * Release for Ubuntu ${DISTRO}
  * Fix COPR RPM build issues
  * Add git-core dependency for Go modules
  * Disable debug package for Go builds

 -- ${MAINTAINER}  $(date -R)

sdtop (1.0.0-1) unstable; urgency=medium

  * Initial release

 -- ${MAINTAINER}  Mon, 01 Jan 2024 00:00:00 +0000
EOF
    
    # Build source package (not binary - Launchpad will build that)
    debuild -S -sa -k${GPG_KEY}
    
    cd "$BUILD_DIR"
    
    # Upload to PPA
    echo "Uploading to PPA..."
    dput ppa:yashsaini99/sdtop "sdtop_${VERSION}-0ubuntu1~${DISTRO}1_source.changes"
    
    echo "✅ Uploaded for $DISTRO"
done

echo ""
echo "🎉 All uploads complete!"
echo ""
echo "Check build status: https://launchpad.net/~yashsaini99/+archive/ubuntu/sdtop"
echo ""
echo "Users can install with:"
echo "  sudo add-apt-repository ppa:yashsaini99/sdtop"
echo "  sudo apt update"
echo "  sudo apt install sdtop"
