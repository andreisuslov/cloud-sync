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
echo "🚀 Opening new terminal for testing..."

# Open a new terminal window with PATH set correctly
osascript -e 'tell application "Terminal" to do script "export PATH=\"$HOME/.local/bin:$PATH\"; echo \"Terminal ready for testing csync\"; echo \"Try: csync --help or just: csync\"; echo \"\""'
