#!/bin/bash
# Real visual test with actual screenshots

set -e

OUTPUT_DIR="/tmp/cloud-sync-visual-$(date +%s)"
mkdir -p "$OUTPUT_DIR"

echo "========================================="
echo "Cloud Sync Visual Screenshot Test"
echo "========================================="
echo ""
echo "Output directory: $OUTPUT_DIR"
echo ""

# Build the app
echo "Building..."
go build -o cloud-sync-test ./cmd/cloud-sync

# Create a script to automate the key presses
cat > "$OUTPUT_DIR/input.sh" << 'EOF'
#!/usr/bin/expect -f
set timeout 2

spawn ./cloud-sync-test
expect {
    timeout { exit 1 }
    eof { exit 0 }
    "*" { }
}

# Wait for initial render
sleep 0.5

# Take initial screenshot
exec screencapture -x "$OUTPUT_DIR/00_initial.png"

# Press down 12 times
for {set i 1} {$i <= 12} {incr i} {
    send "\033\[B"
    sleep 0.1
    exec screencapture -x [format "$OUTPUT_DIR/%02d_down_%02d.png" $i $i]
}

# Press up 3 times
for {set i 1} {$i <= 3} {incr i} {
    send "\033\[A"
    sleep 0.1
    set num [expr {12 + $i}]
    exec screencapture -x [format "$OUTPUT_DIR/%02d_up_%02d.png" $num $i]
}

# Quit
send "q"
sleep 0.2

expect eof
EOF

chmod +x "$OUTPUT_DIR/input.sh"

echo "Running visual test..."
echo "This will capture actual screenshots of your terminal."
echo ""

# Check if expect is installed
if ! command -v expect &> /dev/null; then
    echo "❌ 'expect' is not installed. Install it with:"
    echo "   brew install expect"
    exit 1
fi

# Run the test
cd "$(pwd)"
"$OUTPUT_DIR/input.sh"

echo ""
echo "✅ Test complete!"
echo ""
echo "📁 Screenshots saved to: $OUTPUT_DIR"
echo ""
echo "To view screenshots:"
echo "  open $OUTPUT_DIR"
echo ""
echo "To compare specific frames:"
echo "  open $OUTPUT_DIR/12_down_12.png $OUTPUT_DIR/13_up_01.png"
echo ""
echo "The screenshots show the ACTUAL terminal output."
echo "Compare them to see if only the top row changes or the whole screen updates."
