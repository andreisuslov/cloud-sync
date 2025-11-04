package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVisualDebug_CaptureScrollBehavior(t *testing.T) {
	if os.Getenv("VISUAL_DEBUG") == "" {
		t.Skip("Set VISUAL_DEBUG=1 to run visual debugging tests")
	}

	req := require.New(t)
	
	outputDir := filepath.Join(os.TempDir(), "cloud-sync-visual-debug")
	req.NoError(os.RemoveAll(outputDir))
	req.NoError(os.MkdirAll(outputDir, 0o755))
	
	t.Logf("Capturing visual session to: %s", outputDir)
	
	err := VisualDebugSession(outputDir)
	req.NoError(err)
	
	// Extract frames
	typescriptPath := filepath.Join(outputDir, "session.typescript")
	framesDir := filepath.Join(outputDir, "frames")
	req.NoError(os.MkdirAll(framesDir, 0o755))
	
	err = ExtractFramesFromTypescript(typescriptPath, framesDir)
	req.NoError(err)
	
	t.Logf("Visual debug complete. Review frames in: %s", framesDir)
	t.Logf("To replay: scriptreplay -t %s/session.timing %s/session.typescript", 
		outputDir, outputDir)
}
