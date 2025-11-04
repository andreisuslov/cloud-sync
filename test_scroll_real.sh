#!/bin/bash
# Real terminal test with script recording

OUTPUT_DIR="/tmp/cloud-sync-test-$(date +%s)"
mkdir -p "$OUTPUT_DIR"

echo "Building app..."
go build -o cloud-sync-test ./cmd/cloud-sync

echo "Output: $OUTPUT_DIR"
echo ""
echo "This will run the app and you can manually test:"
echo "1. Press DOWN 12 times"
echo "2. Press UP 3 times"  
echo "3. Watch if the whole screen updates or just the top row"
echo "4. Press 'q' to quit"
echo ""
echo "The session will be recorded to: $OUTPUT_DIR/session.typescript"
echo ""
read -p "Press ENTER to start..."

# Record the session
TERM=xterm-256color script -q "$OUTPUT_DIR/session.typescript" ./cloud-sync-test

echo ""
echo "Session recorded!"
echo ""
echo "To replay:"
echo "  scriptreplay $OUTPUT_DIR/session.typescript"
echo ""
echo "To view raw:"
echo "  cat $OUTPUT_DIR/session.typescript | less -R"
