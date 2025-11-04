#!/bin/bash
# Comprehensive visual debugging script

set -e

echo "========================================="
echo "Cloud Sync Visual Scroll Debug"
echo "========================================="
echo ""

# Build debug version
echo "1. Building debug version..."
go build -o cloud-sync-debug ./cmd/cloud-sync-debug
echo "   ✓ Built cloud-sync-debug"
echo ""

# Create output directory
OUTPUT_DIR="/tmp/cloud-sync-visual-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$OUTPUT_DIR"
echo "2. Output directory: $OUTPUT_DIR"
echo ""

# Run the debug version
echo "3. Running interactive test..."
echo "   Instructions:"
echo "   - The app will start"
echo "   - Press DOWN 12 times to scroll to ~50%"
echo "   - Press UP 3 times"
echo "   - Watch if only the top row updates or the whole screen"
echo "   - Press 'q' to quit"
echo ""
echo "   Debug log will be written to: /tmp/cloud-sync-debug.log"
echo ""
read -p "Press ENTER to start the test..." 

# Run it
./cloud-sync-debug

echo ""
echo "========================================="
echo "Test Complete!"
echo "========================================="
echo ""
echo "Debug log saved to: /tmp/cloud-sync-debug.log"
echo ""
echo "To review the log:"
echo "  cat /tmp/cloud-sync-debug.log"
echo ""
echo "To see just the key presses:"
echo "  grep 'KeyMsg' /tmp/cloud-sync-debug.log"
echo ""
echo "To see list index changes:"
echo "  grep 'List Index' /tmp/cloud-sync-debug.log"
echo ""

# Analyze the log
if [ -f "/tmp/cloud-sync-debug.log" ]; then
    echo "Quick Analysis:"
    echo "---------------"
    echo "Total updates: $(grep -c 'Update #' /tmp/cloud-sync-debug.log || echo 0)"
    echo "Key presses: $(grep -c 'KeyMsg' /tmp/cloud-sync-debug.log || echo 0)"
    echo "Down keys: $(grep -c 'KeyMsg: \"down\"' /tmp/cloud-sync-debug.log || echo 0)"
    echo "Up keys: $(grep -c 'KeyMsg: \"up\"' /tmp/cloud-sync-debug.log || echo 0)"
    echo ""
    
    echo "List index progression:"
    grep 'List Index' /tmp/cloud-sync-debug.log | tail -20 || true
fi
