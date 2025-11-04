#!/bin/bash
# Production build script - builds and installs csync using Makefile

set -e  # Exit on error

echo "🔨 Building and installing csync for production use..."
echo ""

# Use the Makefile install target which builds and installs to ~/.local/bin
make install

echo ""
echo "✅ Done! You can now use 'csync' from anywhere."
echo ""

