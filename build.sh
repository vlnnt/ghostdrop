
#!/usr/bin/env bash

set -e

cd "$(dirname "$0")"

echo "Building GhostDrop for Linux..."

mkdir -p ghostdrop-linux

GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-s -w" \
    -o ghostdrop-linux/ghostdrop \
    ./cmd/ghostdrop

echo
echo "Build completed:"
echo "  ghostdrop-linux/ghostdrop"
