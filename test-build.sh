#!/bin/bash
# Test build script - builds and runs cloud-sync for testing

set -e  # Exit on error

echo "🔨 Building cloud-sync for testing..."
go build -o cloud-sync ./cmd/cloud-sync

echo "✓ Build successful!"
echo ""
echo "🚀 Running cloud-sync..."
echo "----------------------------------------"
./cloud-sync
