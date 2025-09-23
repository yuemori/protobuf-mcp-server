package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yuemori/protobuf-mcp-server/internal/config"
)

// InitCommand handles the initialization of a new protobuf project
// InitCommand handles the initialization of a new protobuf project
func InitCommand(args []string) error {
	var projectPath string

	// Determine project path
	if len(args) > 0 {
		projectPath = args[0]
	} else {
		// Use current directory if no path specified
		var err error
		projectPath, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if project already exists
	if config.ProjectExists(absPath) {
		return fmt.Errorf("project already initialized in %s", absPath)
	}

	// Create default configuration
	defaultConfig := config.DefaultProjectConfig()

	// Save configuration to the project directory
	if err := config.SaveProjectConfig(absPath, defaultConfig); err != nil {
		return fmt.Errorf("failed to initialize project: %w", err)
	}

	fmt.Printf("Successfully initialized protobuf MCP project in %s\n", absPath)
	fmt.Println("Configuration file created at: .protobuf-mcp.yml")
	fmt.Println("\nNext steps:")
	fmt.Println("1. Edit .protobuf-mcp.yml to configure your proto files")
	fmt.Println("2. Use MCP tools to analyze your protobuf files")

	return nil
}
