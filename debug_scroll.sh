#!/bin/bash
# Debug script to capture scrolling behavior visually

set -e

echo "Building cloud-sync..."
go build -o cloud-sync-debug ./cmd/cloud-sync

OUTPUT_DIR="/tmp/cloud-sync-debug-$(date +%s)"
mkdir -p "$OUTPUT_DIR"

echo "Output directory: $OUTPUT_DIR"

# Capture with script command
echo "Running visual capture session..."
echo "This will:"
echo "  1. Start cloud-sync"
echo "  2. Press Down 12 times (to ~50%)"
echo "  3. Press Up 3 times"
echo "  4. Quit"
echo ""

# Create input sequence
cat > "$OUTPUT_DIR/input.txt" << 'EOF'


EOF

# Add 12 down arrows
for i in {1..12}; do
    echo -ne '\x1b[B' >> "$OUTPUT_DIR/input.txt"
    echo "" >> "$OUTPUT_DIR/input.txt"
done

# Add 3 up arrows
for i in {1..3}; do
    echo -ne '\x1b[A' >> "$OUTPUT_DIR/input.txt"
    echo "" >> "$OUTPUT_DIR/input.txt"
done

# Add quit
echo "q" >> "$OUTPUT_DIR/input.txt"

# Run with typescript capture
TERM=xterm-256color COLUMNS=80 LINES=24 \
    script -q "$OUTPUT_DIR/session.typescript" \
    bash -c "./cloud-sync-debug" < "$OUTPUT_DIR/input.txt" || true

echo ""
echo "Session captured to: $OUTPUT_DIR/session.typescript"
echo ""
echo "To replay the session:"
echo "  cat $OUTPUT_DIR/session.typescript"
echo ""
echo "To see raw ANSI codes:"
echo "  cat -v $OUTPUT_DIR/session.typescript | less"
echo ""

# Try to extract frames
echo "Extracting frames..."
python3 - << 'PYTHON' "$OUTPUT_DIR"
import sys
import os
import re

output_dir = sys.argv[1]
typescript_path = os.path.join(output_dir, "session.typescript")

with open(typescript_path, 'rb') as f:
    data = f.read()

# Split on clear screen + home cursor
clear_pattern = b'\x1b[2J\x1b[H'
frames = data.split(clear_pattern)

frames_dir = os.path.join(output_dir, "frames")
os.makedirs(frames_dir, exist_ok=True)

for i, frame in enumerate(frames):
    if not frame.strip():
        continue
    frame_path = os.path.join(frames_dir, f"frame_{i:03d}.txt")
    with open(frame_path, 'wb') as f:
        f.write(frame)
    # Also create a stripped version
    stripped = re.sub(b'\x1b\[[0-9;]*[a-zA-Z]', b'', frame)
    stripped_path = os.path.join(frames_dir, f"frame_{i:03d}_clean.txt")
    with open(stripped_path, 'wb') as f:
        f.write(stripped)

print(f"Extracted {len([f for f in frames if f.strip()])} frames to {frames_dir}")
print(f"\nTo compare frames:")
print(f"  diff {frames_dir}/frame_012_clean.txt {frames_dir}/frame_013_clean.txt")
print(f"  diff {frames_dir}/frame_013_clean.txt {frames_dir}/frame_014_clean.txt")
PYTHON

echo ""
echo "Debug session complete!"
echo "Review the frames in: $OUTPUT_DIR/frames/"
