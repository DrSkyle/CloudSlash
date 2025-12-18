#!/bin/bash
# ASTRA SYNC SCRIPT
# Run this in a separate terminal to keep your build folder up to date.

SRC="/home/astra/antigravity/Share/Portfolio/CloudSlash/"
DEST="/tmp/astra_build/cloudslash/"

echo "🔄 Watching for changes in $SRC..."
echo "📂 Syncing to $DEST..."

while true; do
  rsync -avq --delete \
    --exclude='node_modules' \
    --exclude='.git' \
    --exclude='cloudslash-out' \
    --exclude='builds' \
    --exclude='dist' \
    "$SRC" "$DEST"
  
  echo "✅ Synced at $(date +%T)"
  sleep 2
done
