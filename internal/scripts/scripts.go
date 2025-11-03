package scripts

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed *.tmpl
var scriptTemplates embed.FS

// Generator handles script generation from templates
type Generator struct {
	templateFS embed.FS
}

// Config holds configuration for script generation
type Config struct {
	HomeDir      string
	Username     string
	RclonePath   string
	SourceRemote string
	SourceBucket string
	DestRemote   string
	DestBucket   string
	LogDir       string
	BinDir       string
}

// NewGenerator creates a new script generator
func NewGenerator() *Generator {
	return &Generator{
		templateFS: scriptTemplates,
	}
}

// NewGeneratorWithFS creates a generator with custom filesystem (for testing)
func NewGeneratorWithFS(fs embed.FS) *Generator {
	return &Generator{
		templateFS: fs,
	}
}

// CreateDirectories creates required directories for scripts and logs
func (g *Generator) CreateDirectories(config *Config) error {
	dirs := []string{
		config.BinDir,
		config.LogDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// GenerateEngineScript generates the run_rclone_sync.sh script
func (g *Generator) GenerateEngineScript(config *Config) error {
	return g.generateScript("run_rclone_sync.sh", config)
}

// GenerateMonthlyScript generates the monthly_backup.sh script
func (g *Generator) GenerateMonthlyScript(config *Config) error {
	return g.generateScript("monthly_backup.sh", config)
}

// GenerateManualScript generates the sync_now.sh script
func (g *Generator) GenerateManualScript(config *Config) error {
	return g.generateScript("sync_now.sh", config)
}

// GenerateShowTransfersScript generates the show_transfers.sh script
func (g *Generator) GenerateShowTransfersScript(config *Config) error {
	return g.generateScript("show_transfers.sh", config)
}

// GenerateAllScripts generates all scripts
func (g *Generator) GenerateAllScripts(config *Config) error {
	scripts := []string{
		"run_rclone_sync.sh",
		"monthly_backup.sh",
		"sync_now.sh",
		"show_transfers.sh",
	}

	for _, script := range scripts {
		if err := g.generateScript(script, config); err != nil {
			return fmt.Errorf("failed to generate %s: %w", script, err)
		}
	}

	return nil
}

// generateScript generates a script from template
func (g *Generator) generateScript(scriptName string, config *Config) error {
	// Read template
	templateName := scriptName + ".tmpl"
	tmplContent, err := g.templateFS.ReadFile(templateName)
	if err != nil {
		return fmt.Errorf("failed to read template %s: %w", templateName, err)
	}

	// Parse template
	tmpl, err := template.New(scriptName).Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", templateName, err)
	}

	// Create output file
	outputPath := filepath.Join(config.BinDir, scriptName)
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create script %s: %w", outputPath, err)
	}
	defer file.Close()

	// Execute template
	if err := tmpl.Execute(file, config); err != nil {
		return fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}

	// Make script executable
	if err := os.Chmod(outputPath, 0755); err != nil {
		return fmt.Errorf("failed to make script executable %s: %w", outputPath, err)
	}

	return nil
}

// MakeExecutable makes a script executable
func (g *Generator) MakeExecutable(filepath string) error {
	if err := os.Chmod(filepath, 0755); err != nil {
		return fmt.Errorf("failed to make %s executable: %w", filepath, err)
	}
	return nil
}

// ValidateConfig validates the script configuration
func ValidateConfig(config *Config) error {
	if config.HomeDir == "" {
		return fmt.Errorf("HomeDir is required")
	}
	if config.Username == "" {
		return fmt.Errorf("Username is required")
	}
	if config.RclonePath == "" {
		return fmt.Errorf("RclonePath is required")
	}
	if config.SourceRemote == "" {
		return fmt.Errorf("SourceRemote is required")
	}
	if config.SourceBucket == "" {
		return fmt.Errorf("SourceBucket is required")
	}
	if config.DestRemote == "" {
		return fmt.Errorf("DestRemote is required")
	}
	if config.DestBucket == "" {
		return fmt.Errorf("DestBucket is required")
	}
	if config.LogDir == "" {
		return fmt.Errorf("LogDir is required")
	}
	if config.BinDir == "" {
		return fmt.Errorf("BinDir is required")
	}
	return nil
}