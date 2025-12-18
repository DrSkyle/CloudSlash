#!/bin/bash
set -e

echo "🚀 Setting up Isolated Build Environment..."

# 1. Create directory
mkdir -p /tmp/astra_build/cloudslash
echo "✅ Created /tmp/astra_build/cloudslash"

# 2. Sync Code (Excluding heavy/irrelevant files)
# We exclude 'cloudslash-out' and 'builds' to avoid circular syncs or syncing artifacts
rsync -avq --delete \
    --exclude='node_modules' \
    --exclude='.git' \
    --exclude='cloudslash-out' \
    --exclude='builds' \
    /home/astra/antigravity/Share/Portfolio/CloudSlash/ /tmp/astra_build/cloudslash/

echo "✅ Synced code to /tmp/astra_build/cloudslash"

# 3. Bootstrap Go (if not found in /tmp/go)
# FORCE REINSTALL IF PREVIOUS FAILED (check via go version command)
export PATH=$PATH:/tmp/go/bin
if ! go version >/dev/null 2>&1; then
    echo "⬇️  Go not found or invalid. Downloading Portable Go 1.21.5..."
    rm -rf /tmp/go
    
    ARCH=$(uname -m)
    if [ "$ARCH" = "aarch64" ]; then
        echo "Detected Architecture: ARM64"
        GO_URL="https://go.dev/dl/go1.21.5.linux-arm64.tar.gz"
    else
        echo "Detected Architecture: AMD64"
        GO_URL="https://go.dev/dl/go1.21.5.linux-amd64.tar.gz"
    fi

    wget -q "$GO_URL" -O /tmp/go.tar.gz
    
    echo "📦 Extracting Go..."
    tar -C /tmp -xzf /tmp/go.tar.gz
    rm /tmp/go.tar.gz
    echo "✅ Go installed to /tmp/go"
else
    echo "✅ Go already exists and works in /tmp/go"
fi

# 4. Verify
export PATH=/tmp/go/bin:$PATH
go version
echo "🎉 Environment Ready. You can now build in /tmp/astra_build/cloudslash"
