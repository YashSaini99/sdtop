#!/bin/bash
# Simple PPA upload script without debuild dependency
set -e

VERSION="1.0.1"
MAINTAINER="Yash Saini <yashsaini99@example.com>"
GPG_KEY="EC8347BF52E517BD"

echo "📦 Building source package for Ubuntu..."

# Create build directory
BUILD_DIR="/tmp/sdtop-ppa-$$"
mkdir -p "$BUILD_DIR"
cd "$BUILD_DIR"

# Create orig tarball from git
cd /home/yash/Documents/sdtop
git archive --format=tar --prefix=sdtop-${VERSION}/ HEAD | gzip > "$BUILD_DIR/sdtop_${VERSION}.orig.tar.gz"

# Extract tarball
cd "$BUILD_DIR"
tar -xzf "sdtop_${VERSION}.orig.tar.gz"
cd "sdtop-${VERSION}"

# Copy debian directory
cp -r /home/yash/Documents/sdtop/packaging/debian debian/

# Update changelog for Ubuntu noble (24.04)
cat > debian/changelog << EOF
sdtop (${VERSION}-0ubuntu1) noble; urgency=medium

  * Release for Ubuntu Noble (24.04)
  * Fix COPR RPM build issues
  * Add git-core dependency for Go modules
  * Disable debug package for Go builds

 -- ${MAINTAINER}  $(date -R)
EOF

# Build source package
dpkg-source -b .

cd "$BUILD_DIR"

# Create .changes file manually
cat > "sdtop_${VERSION}-0ubuntu1_source.changes" << EOF
Format: 1.8
Date: $(date -R)
Source: sdtop
Binary: sdtop
Architecture: source
Version: ${VERSION}-0ubuntu1
Distribution: noble
Urgency: medium
Maintainer: ${MAINTAINER}
Changed-By: ${MAINTAINER}
Description:
 sdtop - Terminal-based systemd service manager
Changes:
 sdtop (${VERSION}-0ubuntu1) noble; urgency=medium
 .
   * Release for Ubuntu Noble (24.04)
Checksums-Sha1:
 $(sha1sum sdtop_${VERSION}.orig.tar.gz | awk '{print $1" "$2" sdtop_${VERSION}.orig.tar.gz"}')
 $(sha1sum sdtop_${VERSION}-0ubuntu1.debian.tar.xz | awk '{print $1" "$2" sdtop_${VERSION}-0ubuntu1.debian.tar.xz"}')
 $(sha1sum sdtop_${VERSION}-0ubuntu1.dsc | awk '{print $1" "$2" sdtop_${VERSION}-0ubuntu1.dsc"}')
Checksums-Sha256:
 $(sha256sum sdtop_${VERSION}.orig.tar.gz | awk '{print $1" "$2" sdtop_${VERSION}.orig.tar.gz"}')
 $(sha256sum sdtop_${VERSION}-0ubuntu1.debian.tar.xz | awk '{print $1" "$2" sdtop_${VERSION}-0ubuntu1.debian.tar.xz"}')
 $(sha256sum sdtop_${VERSION}-0ubuntu1.dsc | awk '{print $1" "$2" sdtop_${VERSION}-0ubuntu1.dsc"}')
Files:
 $(md5sum sdtop_${VERSION}.orig.tar.gz | awk '{print $1" "$2" sdtop_${VERSION}.orig.tar.gz"}')
 $(md5sum sdtop_${VERSION}-0ubuntu1.debian.tar.xz | awk '{print $1" "$2" sdtop_${VERSION}-0ubuntu1.debian.tar.xz"}')
 $(md5sum sdtop_${VERSION}-0ubuntu1.dsc | awk '{print $1" "$2" sdtop_${VERSION}-0ubuntu1.dsc"}')
EOF

# Sign the changes file
gpg --armor --sign --detach-sign -u $GPG_KEY "sdtop_${VERSION}-0ubuntu1_source.changes"

echo ""
echo "✅ Source package created in: $BUILD_DIR"
echo ""
echo "To upload to PPA, run:"
echo "  dput ppa:yashsaini99/sdtop $BUILD_DIR/sdtop_${VERSION}-0ubuntu1_source.changes"
echo ""
echo "Or install dput and run:"
echo "  cd $BUILD_DIR && dput ppa sdtop_${VERSION}-0ubuntu1_source.changes"
