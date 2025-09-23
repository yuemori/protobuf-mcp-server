package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuemori/protobuf-mcp-server/internal/config"
)

func TestInitCommand(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "protobuf-mcp-init-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test init with explicit path
	testPath := filepath.Join(tempDir, "test-project")
	err = os.MkdirAll(testPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create test project directory: %v", err)
	}

	err = InitCommand([]string{testPath})
	if err != nil {
		t.Fatalf("InitCommand failed: %v", err)
	}

	// Check that project was initialized
	if !config.ProjectExists(testPath) {
		t.Error("Project was not initialized properly")
	}

	// Check that config file exists and is valid
	loadedConfig, err := config.LoadProjectConfig(testPath)
	if err != nil {
		t.Fatalf("Failed to load created project config: %v", err)
	}

	// Verify config contains expected default values
	expectedProtoFiles := []string{"proto/**/*.proto"}
	if len(loadedConfig.ProtoFiles) != len(expectedProtoFiles) {
		t.Errorf("Expected %d proto files, got %d", len(expectedProtoFiles), len(loadedConfig.ProtoFiles))
	}

	for i, expected := range expectedProtoFiles {
		if i >= len(loadedConfig.ProtoFiles) || loadedConfig.ProtoFiles[i] != expected {
			t.Errorf("Expected ProtoFiles[%d] to be %s, got %s", i, expected, loadedConfig.ProtoFiles[i])
		}
	}
}

func TestInitCommandWithoutArgs(t *testing.T) {
	// Save current directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Create temporary directory and change to it
	tempDir, err := os.MkdirTemp("", "protobuf-mcp-init-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		os.Chdir(originalWd) // Restore original directory
		os.RemoveAll(tempDir)
	}()

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Test init without arguments (should use current directory)
	err = InitCommand([]string{})
	if err != nil {
		t.Fatalf("InitCommand failed: %v", err)
	}

	// Check that project was initialized in current directory
	if !config.ProjectExists(tempDir) {
		t.Error("Project was not initialized in current directory")
	}
}

func TestInitCommandAlreadyInitialized(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "protobuf-mcp-init-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize project first time
	err = InitCommand([]string{tempDir})
	if err != nil {
		t.Fatalf("First InitCommand failed: %v", err)
	}

	// Try to initialize again - should fail
	err = InitCommand([]string{tempDir})
	if err == nil {
		t.Error("Expected error when initializing already initialized project")
	}

	// Check error message contains expected text
	expectedText := "already initialized"
	if err != nil && !contains(err.Error(), expectedText) {
		t.Errorf("Expected error message to contain '%s', got: %s", expectedText, err.Error())
	}
}

func TestInitCommandInvalidPath(t *testing.T) {
	// Test with a path that should cause issues (very long path that might not be valid)
	invalidPath := "/this/path/should/not/exist/and/should/cause/an/error/when/trying/to/create/directories"

	err := InitCommand([]string{invalidPath})
	if err == nil {
		t.Error("Expected error when using invalid path")
	}
}

func TestInitCommandRelativePath(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "protobuf-mcp-init-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Save current directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalWd)

	// Change to temp directory
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create subdirectory and use relative path
	subDir := "relative-project"
	err = os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Test init with relative path
	err = InitCommand([]string{subDir})
	if err != nil {
		t.Fatalf("InitCommand with relative path failed: %v", err)
	}

	// Check that project was initialized in absolute path
	absoluteSubDir := filepath.Join(tempDir, subDir)
	if !config.ProjectExists(absoluteSubDir) {
		t.Error("Project was not initialized with relative path")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}