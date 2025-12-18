#!/bin/bash

# CloudSlash Release Builder
# Run this on your local machine to build binaries for all platforms.

set -e

echo "🚀 Starting CloudSlash Build..."

# Create output directory
mkdir -p dist
rm -f dist/*

# Check for Go
if ! command -v go &> /dev/null; then
    echo "❌ Error: 'go' is not installed. Please install Go first (brew install go)."
    exit 1
fi

# 1. Linux (AMD64)
echo "📦 Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o dist/cloudslash_linux_amd64 cmd/cloudslash/main.go

# 2. Windows (AMD64)
echo "📦 Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o dist/cloudslash_windows_amd64.exe cmd/cloudslash/main.go

# 3. macOS (Intel)
echo "📦 Building for macOS (Intel)..."
GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o dist/cloudslash_darwin_amd64 cmd/cloudslash/main.go

# 4. macOS (Apple Silicon / M1 / M2 / M3)
echo "📦 Building for macOS (Apple Silicon)..."
GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o dist/cloudslash_darwin_arm64 cmd/cloudslash/main.go

echo ""
echo "✅ Build Complete! binaries are in the 'dist' folder:"
ls -lh dist

echo ""
echo "NEXT STEP: Upload these files to: https://github.com/DrSkyle/CloudSlash/releases/new"
