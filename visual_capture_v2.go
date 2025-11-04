package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/creack/pty"
)

func main() {
	outputDir := fmt.Sprintf("/tmp/cloud-sync-visual-%d", time.Now().Unix())
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		panic(err)
	}

	fmt.Printf("📁 Visual Test Output: %s\n\n", outputDir)
	fmt.Println("🔨 Building application...")

	// Build with alt screen mode for better capture
	buildCmd := exec.Command("go", "build", "-o", "cloud-sync-test", "./cmd/cloud-sync")
	if err := buildCmd.Run(); err != nil {
		panic(fmt.Errorf("build failed: %w", err))
	}

	fmt.Println("🚀 Starting application in PTY...\n")

	// Start the app in a PTY
	cmd := exec.Command("./cloud-sync-test")
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	ptmx, err := pty.Start(cmd)
	if err != nil {
		panic(err)
	}
	defer ptmx.Close()

	// Set PTY size
	if err := pty.Setsize(ptmx, &pty.Winsize{Rows: 24, Cols: 80}); err != nil {
		panic(err)
	}

	// Capture all output to file
	logFile, _ := os.Create(filepath.Join(outputDir, "full_session.log"))
	defer logFile.Close()

	// Tee output to both file and discard
	go io.Copy(io.MultiWriter(logFile, io.Discard), ptmx)

	// Helper to send keys and capture
	sendKey := func(key string, name string) {
		var seq []byte
		switch key {
		case "down":
			seq = []byte("\x1b[B")
		case "up":
			seq = []byte("\x1b[A")
		case "enter":
			seq = []byte("\r")
		case "q":
			seq = []byte("q")
		default:
			seq = []byte(key)
		}
		ptmx.Write(seq)
		time.Sleep(150 * time.Millisecond) // Give time to render
		
		// Use script to capture the current terminal state
		captureCmd := exec.Command("bash", "-c", 
			fmt.Sprintf("tput cup 0 0 && cat /dev/tty"))
		captureCmd.Stdin = ptmx
		captureCmd.Stdout, _ = os.Create(filepath.Join(outputDir, fmt.Sprintf("%s.txt", name)))
		captureCmd.Run()
		
		fmt.Printf("  📸 %s\n", name)
	}

	// Wait for initial render
	time.Sleep(500 * time.Millisecond)
	fmt.Println("Test Sequence:")
	fmt.Println("1. Pressing DOWN 12 times...")
	
	for i := 1; i <= 12; i++ {
		sendKey("down", fmt.Sprintf("%02d_down_%02d", i, i))
	}

	fmt.Println("\n2. Pressing UP 3 times...")
	for i := 1; i <= 3; i++ {
		sendKey("up", fmt.Sprintf("%02d_up_%02d", 12+i, i))
	}

	fmt.Println("\n3. Quitting...")
	sendKey("q", "99_quit")
	time.Sleep(200 * time.Millisecond)

	// Cleanup
	cmd.Process.Kill()
	cmd.Wait()

	fmt.Printf("\n✅ Test complete!\n")
	fmt.Printf("📁 Output: %s\n", outputDir)
	fmt.Printf("\n💡 To view the session log:\n")
	fmt.Printf("   cat %s/full_session.log | less -R\n", outputDir)
	fmt.Printf("\n💡 To see what changed between frames:\n")
	fmt.Printf("   diff %s/12_down_12.txt %s/13_up_01.txt\n", outputDir, outputDir)
}
