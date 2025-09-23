package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultProjectConfig(t *testing.T) {
	config := DefaultProjectConfig()

	// Check default values
	if config.RootDirectory != "." {
		t.Errorf("Expected RootDirectory to be '.', got %s", config.RootDirectory)
	}

	if len(config.IncludePaths) != 1 || config.IncludePaths[0] != "." {
		t.Errorf("Expected IncludePaths to be ['.'], got %v", config.IncludePaths)
	}

	if len(config.ProtoPaths) != 1 || config.ProtoPaths[0] != "." {
		t.Errorf("Expected ProtoPaths to be ['.'], got %v", config.ProtoPaths)
	}

	if config.CompilerOptions == nil {
		t.Error("Expected CompilerOptions to be initialized")
	}

	expectedIgnoredPatterns := []string{"*_test.proto", "tmp/**"}
	if len(config.IgnoredPatterns) != len(expectedIgnoredPatterns) {
		t.Errorf("Expected %d ignored patterns, got %d", len(expectedIgnoredPatterns), len(config.IgnoredPatterns))
	}

	for i, pattern := range expectedIgnoredPatterns {
		if i >= len(config.IgnoredPatterns) || config.IgnoredPatterns[i] != pattern {
			t.Errorf("Expected ignored pattern at index %d to be %s, got %s", i, pattern, config.IgnoredPatterns[i])
		}
	}

	if config.ShowLogs != false {
		t.Errorf("Expected ShowLogs to be false, got %v", config.ShowLogs)
	}
}

func TestSaveAndLoadProjectConfig(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "protobuf-mcp-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test configuration
	originalConfig := &ProjectConfig{
		RootDirectory: "./test",
		IncludePaths:  []string{"./proto", "./vendor"},
		ProtoPaths:    []string{"./api", "./internal"},
		CompilerOptions: map[string]interface{}{
			"experimental_allow_proto3_optional": true,
		},
		IgnoredPatterns: []string{"*_test.proto", "tmp/**", "*.bak"},
		ShowLogs:        true,
	}

	// Save configuration
	err = SaveProjectConfig(tempDir, originalConfig)
	if err != nil {
		t.Fatalf("Failed to save project config: %v", err)
	}

	// Check that directory and file were created
	configPath := filepath.Join(tempDir, ".protobuf-mcp", "project.yml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("Config file was not created at %s", configPath)
	}

	// Load configuration back
	loadedConfig, err := LoadProjectConfig(tempDir)
	if err != nil {
		t.Fatalf("Failed to load project config: %v", err)
	}

	// Compare configurations
	if loadedConfig.RootDirectory != originalConfig.RootDirectory {
		t.Errorf("RootDirectory mismatch: expected %s, got %s", originalConfig.RootDirectory, loadedConfig.RootDirectory)
	}

	if len(loadedConfig.IncludePaths) != len(originalConfig.IncludePaths) {
		t.Errorf("IncludePaths length mismatch: expected %d, got %d", len(originalConfig.IncludePaths), len(loadedConfig.IncludePaths))
	}

	for i, path := range originalConfig.IncludePaths {
		if i >= len(loadedConfig.IncludePaths) || loadedConfig.IncludePaths[i] != path {
			t.Errorf("IncludePaths[%d] mismatch: expected %s, got %s", i, path, loadedConfig.IncludePaths[i])
		}
	}

	if len(loadedConfig.ProtoPaths) != len(originalConfig.ProtoPaths) {
		t.Errorf("ProtoPaths length mismatch: expected %d, got %d", len(originalConfig.ProtoPaths), len(loadedConfig.ProtoPaths))
	}

	if loadedConfig.ShowLogs != originalConfig.ShowLogs {
		t.Errorf("ShowLogs mismatch: expected %v, got %v", originalConfig.ShowLogs, loadedConfig.ShowLogs)
	}
}

func TestProjectExists(t *testing.T) {
	// Test with non-existent project
	tempDir, err := os.MkdirTemp("", "protobuf-mcp-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Should return false for non-existent project
	if ProjectExists(tempDir) {
		t.Error("ProjectExists should return false for non-existent project")
	}

	// Create a project
	config := DefaultProjectConfig()
	err = SaveProjectConfig(tempDir, config)
	if err != nil {
		t.Fatalf("Failed to save project config: %v", err)
	}

	// Should return true for existing project
	if !ProjectExists(tempDir) {
		t.Error("ProjectExists should return true for existing project")
	}
}

func TestLoadProjectConfigWithNonExistentFile(t *testing.T) {
	// Test loading from non-existent directory
	tempDir, err := os.MkdirTemp("", "protobuf-mcp-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	_, err = LoadProjectConfig(tempDir)
	if err == nil {
		t.Error("Expected error when loading non-existent config file")
	}
}

func TestSaveProjectConfigWithInvalidPath(t *testing.T) {
	// Try to save to a non-writable location (this may vary by system)
	config := DefaultProjectConfig()
	err := SaveProjectConfig("/invalid/path/that/should/not/exist", config)
	if err == nil {
		t.Error("Expected error when saving to invalid path")
	}
}
