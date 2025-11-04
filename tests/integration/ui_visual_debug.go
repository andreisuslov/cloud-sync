package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// VisualDebugSession runs the actual cloud-sync binary in a PTY and captures output
func VisualDebugSession(outputDir string) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}

	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "cloud-sync-debug", "./cmd/cloud-sync")
	buildCmd.Dir = "/Users/ansuslov/Documents/Development/cloud-sync"
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	// Use script command to capture terminal session with timing
	scriptFile := filepath.Join(outputDir, "session.typescript")
	timingFile := filepath.Join(outputDir, "session.timing")
	
	// Create a command file with the key sequence
	cmdFile := filepath.Join(outputDir, "commands.txt")
	commands := `# Wait for startup
sleep 0.5

# Send multiple downs (to get to ~50%)
for i in {1..12}; do
  # Send down arrow
  printf '\x1b[B'
  sleep 0.1
done

# Capture state after downs
sleep 0.3

# Send 3 ups
for i in {1..3}; do
  printf '\x1b[A'
  sleep 0.2
done

# Wait to see final state
sleep 0.5

# Quit
printf 'q'
sleep 0.2
`
	if err := os.WriteFile(cmdFile, []byte(commands), 0o644); err != nil {
		return err
	}

	// Run with script to capture
	cmd := exec.Command("script", "-q", "-t", timingFile, scriptFile, 
		"bash", "-c", 
		fmt.Sprintf("cd %s && ./cloud-sync-debug < %s", 
			"/Users/ansuslov/Documents/Development/cloud-sync", cmdFile))
	
	cmd.Env = append(os.Environ(), 
		"TERM=xterm-256color",
		"COLUMNS=80",
		"LINES=24",
	)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("script failed: %w\nOutput: %s", err, output)
	}

	// Also save raw output
	rawFile := filepath.Join(outputDir, "raw_output.txt")
	if err := os.WriteFile(rawFile, output, 0o644); err != nil {
		return err
	}

	fmt.Printf("Visual debug session captured to: %s\n", outputDir)
	fmt.Printf("- Typescript: %s\n", scriptFile)
	fmt.Printf("- Timing: %s\n", timingFile)
	fmt.Printf("- Raw: %s\n", rawFile)
	
	return nil
}

// ExtractFramesFromTypescript parses the typescript file and extracts visual frames
func ExtractFramesFromTypescript(typescriptPath, outputDir string) error {
	data, err := os.ReadFile(typescriptPath)
	if err != nil {
		return err
	}

	// Split on clear screen sequences
	frames := splitOnClearScreen(data)
	
	for i, frame := range frames {
		framePath := filepath.Join(outputDir, fmt.Sprintf("frame_%03d.txt", i))
		if err := os.WriteFile(framePath, frame, 0o644); err != nil {
			return err
		}
	}
	
	fmt.Printf("Extracted %d frames to %s/frame_*.txt\n", len(frames), outputDir)
	return nil
}

func splitOnClearScreen(data []byte) [][]byte {
	// Look for common clear sequences: ESC[2J ESC[H or ESC[H ESC[2J
	var frames [][]byte
	var current []byte
	
	i := 0
	for i < len(data) {
		// Check for ESC[2J (clear screen)
		if i+3 < len(data) && data[i] == 0x1b && data[i+1] == '[' && 
		   data[i+2] == '2' && data[i+3] == 'J' {
			if len(current) > 0 {
				frames = append(frames, append([]byte(nil), current...))
				current = nil
			}
			i += 4
			continue
		}
		// Check for ESC[H (home)
		if i+2 < len(data) && data[i] == 0x1b && data[i+1] == '[' && data[i+2] == 'H' {
			// This often comes with clear, so we might start a new frame
			i += 3
			continue
		}
		current = append(current, data[i])
		i++
	}
	
	if len(current) > 0 {
		frames = append(frames, current)
	}
	
	return frames
}
