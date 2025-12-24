#!/bin/bash
# Build and upload sdtop to Ubuntu PPA
# Usage: ./build-ppa.sh <version> <ubuntu-series>
# Example: ./build-ppa.sh 1.0.1 noble

set -e

VERSION=${1:-1.0.1}
SERIES=${2:-noble}  # noble (24.04), mantic (23.10), jammy (22.04), focal (20.04)

echo "Building sdtop ${VERSION} for Ubuntu ${SERIES}"

# Update changelog for specific Ubuntu release
cd packaging/debian
dch -v "${VERSION}-1~${SERIES}1" -D ${SERIES} "Release for Ubuntu ${SERIES}"

# Build source package
cd ../..
debuild -S -sa -k$(gpg --list-secret-keys --keyid-format LONG | grep sec | awk '{print $2}' | cut -d'/' -f2 | head -n1)

# Upload to PPA (adjust PPA name as needed)
cd ..
dput ppa:yashsaini99/sdtop sdtop_${VERSION}-1~${SERIES}1_source.changes

echo "✅ Uploaded to PPA. Check status at: https://launchpad.net/~yashsaini99/+archive/ubuntu/sdtop"
echo "Build will take 10-30 minutes. Users can install with:"
echo "  sudo add-apt-repository ppa:yashsaini99/sdtop"
echo "  sudo apt update"
echo "  sudo apt install sdtop"
