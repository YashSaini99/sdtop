#!/bin/bash
set -e

VERSION=${1:-1.0.0}
PLATFORMS=("linux/amd64" "linux/arm64" "linux/arm/v7")

echo "Building sdtop v${VERSION} for multiple platforms..."

# Create build directory
mkdir -p dist

for PLATFORM in "${PLATFORMS[@]}"; do
    OS=$(echo $PLATFORM | cut -d'/' -f1)
    ARCH=$(echo $PLATFORM | cut -d'/' -f2)
    VARIANT=$(echo $PLATFORM | cut -d'/' -f3)
    
    OUTPUT_NAME="sdtop-${VERSION}-${OS}-${ARCH}"
    if [ -n "$VARIANT" ]; then
        OUTPUT_NAME="${OUTPUT_NAME}-${VARIANT}"
    fi
    
    echo "Building for ${PLATFORM}..."
    
    GOOS=$OS GOARCH=$ARCH CGO_ENABLED=0 \
        go build -ldflags="-s -w -X main.Version=${VERSION}" \
        -o "dist/${OUTPUT_NAME}" \
        ./cmd/main.go
    
    # Create tarball
    tar -czf "dist/${OUTPUT_NAME}.tar.gz" -C dist "${OUTPUT_NAME}"
    
    # Calculate checksums
    sha256sum "dist/${OUTPUT_NAME}.tar.gz" >> "dist/checksums.txt"
    
    echo "✓ Built ${OUTPUT_NAME}"
done

echo ""
echo "Build complete! Files in dist/ directory:"
ls -lh dist/
