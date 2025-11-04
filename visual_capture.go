package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	"github.com/creack/pty"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const (
	charWidth  = 7
	charHeight = 13
	cols       = 80
	rows       = 24
)

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func main() {
	outputDir := fmt.Sprintf("/tmp/cloud-sync-visual-%d", time.Now().Unix())
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		panic(err)
	}

	fmt.Printf("Visual Test Output: %s\n", outputDir)
	fmt.Println("Building application...")

	// Build the app
	buildCmd := exec.Command("go", "build", "-o", "cloud-sync-test", "./cmd/cloud-sync")
	if err := buildCmd.Run(); err != nil {
		panic(fmt.Errorf("build failed: %w", err))
	}

	fmt.Println("Starting application in PTY...")

	// Start the app in a PTY
	cmd := exec.Command("./cloud-sync-test")
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	ptmx, err := pty.Start(cmd)
	if err != nil {
		panic(err)
	}
	defer ptmx.Close()

	// Set PTY size
	if err := pty.Setsize(ptmx, &pty.Winsize{Rows: rows, Cols: cols}); err != nil {
		panic(err)
	}

	// Buffer to accumulate output
	var screenBuf bytes.Buffer
	captureOutput := make(chan []byte, 100)

	// Read from PTY
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := ptmx.Read(buf)
			if err != nil {
				if err != io.EOF {
					fmt.Printf("Read error: %v\n", err)
				}
				return
			}
			if n > 0 {
				data := make([]byte, n)
				copy(data, buf[:n])
				captureOutput <- data
			}
		}
	}()

	// Helper to capture current screen
	captureScreen := func(name string) {
		time.Sleep(100 * time.Millisecond) // Let rendering settle
		
		// Get current screen content
		screen := screenBuf.String()
		
		// Save raw output
		rawPath := filepath.Join(outputDir, fmt.Sprintf("%s_raw.txt", name))
		os.WriteFile(rawPath, []byte(screen), 0644)
		
		// Strip ANSI and save clean text
		clean := ansiRe.ReplaceAllString(screen, "")
		cleanPath := filepath.Join(outputDir, fmt.Sprintf("%s_clean.txt", name))
		os.WriteFile(cleanPath, []byte(clean), 0644)
		
		// Render as image
		img := renderToImage(clean)
		imgPath := filepath.Join(outputDir, fmt.Sprintf("%s.png", name))
		f, _ := os.Create(imgPath)
		png.Encode(f, img)
		f.Close()
		
		fmt.Printf("  📸 Captured: %s\n", name)
	}

	// Helper to send keys
	sendKey := func(key string) {
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
		time.Sleep(50 * time.Millisecond)
	}

	// Accumulate output
	go func() {
		for data := range captureOutput {
			screenBuf.Write(data)
		}
	}()

	// Wait for initial render
	time.Sleep(500 * time.Millisecond)
	captureScreen("00_initial")

	fmt.Println("\nTest Sequence:")
	fmt.Println("1. Pressing DOWN 12 times...")
	for i := 1; i <= 12; i++ {
		sendKey("down")
		captureScreen(fmt.Sprintf("%02d_down_%02d", i, i))
	}

	fmt.Println("2. Pressing UP 3 times...")
	for i := 1; i <= 3; i++ {
		sendKey("up")
		captureScreen(fmt.Sprintf("%02d_up_%02d", 12+i, i))
	}

	fmt.Println("3. Quitting...")
	sendKey("q")
	time.Sleep(200 * time.Millisecond)

	// Cleanup
	cmd.Process.Kill()
	cmd.Wait()

	fmt.Printf("\n✅ Test complete!\n")
	fmt.Printf("📁 Screenshots saved to: %s\n", outputDir)
	fmt.Println("\nCompare screenshots to see the issue:")
	fmt.Printf("  open %s\n", outputDir)
	fmt.Println("\nOr use ImageMagick to create a diff:")
	fmt.Printf("  compare %s/12_down_12.png %s/13_up_01.png %s/diff_down_vs_up.png\n",
		outputDir, outputDir, outputDir)
}

func renderToImage(text string) *image.RGBA {
	width := cols * charWidth
	height := rows * charHeight
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Black background
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.Black)
		}
	}

	// Draw text
	lines := splitLines(text, rows)
	y := charHeight
	for _, line := range lines {
		x := 0
		for _, ch := range line {
			if x >= width {
				break
			}
			drawChar(img, x, y, ch)
			x += charWidth
		}
		y += charHeight
		if y >= height {
			break
		}
	}

	return img
}

func splitLines(text string, maxLines int) []string {
	lines := []string{}
	current := ""
	
	for _, ch := range text {
		if ch == '\n' {
			lines = append(lines, current)
			current = ""
			if len(lines) >= maxLines {
				break
			}
		} else if ch >= 32 && ch < 127 {
			current += string(ch)
		}
	}
	
	if current != "" && len(lines) < maxLines {
		lines = append(lines, current)
	}
	
	// Pad to maxLines
	for len(lines) < maxLines {
		lines = append(lines, "")
	}
	
	return lines
}

func drawChar(img *image.RGBA, x, y int, ch rune) {
	point := fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(y * 64)}
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.White),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(string(ch))
}
