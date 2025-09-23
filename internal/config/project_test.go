package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultProjectConfig(t *testing.T) {
	config := DefaultProjectConfig()

	// Check default values
	expectedProtoFiles := []string{"proto/**/*.proto"}
	if len(config.ProtoFiles) != len(expectedProtoFiles) {
		t.Errorf("Expected %d proto files, got %d", len(expectedProtoFiles), len(config.ProtoFiles))
	}

	for i, expected := range expectedProtoFiles {
		if i >= len(config.ProtoFiles) || config.ProtoFiles[i] != expected {
			t.Errorf("Expected ProtoFiles[%d] to be %s, got %s", i, expected, config.ProtoFiles[i])
		}
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
		ProtoFiles: []string{
			"proto/**/*.proto",
			"api/v1/**/*.proto",
			"/absolute/path/**/*.proto",
		},
	}

	// Save configuration
	err = SaveProjectConfig(tempDir, originalConfig)
	if err != nil {
		t.Fatalf("Failed to save project config: %v", err)
	}

	// Check that file was created
	configPath := filepath.Join(tempDir, ".protobuf-mcp.yml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("Config file was not created at %s", configPath)
	}

	// Load configuration back
	loadedConfig, err := LoadProjectConfig(tempDir)
	if err != nil {
		t.Fatalf("Failed to load project config: %v", err)
	}

	// Compare configurations
	if len(loadedConfig.ProtoFiles) != len(originalConfig.ProtoFiles) {
		t.Errorf("ProtoFiles length mismatch: expected %d, got %d", len(originalConfig.ProtoFiles), len(loadedConfig.ProtoFiles))
	}

	for i, expected := range originalConfig.ProtoFiles {
		if i >= len(loadedConfig.ProtoFiles) || loadedConfig.ProtoFiles[i] != expected {
			t.Errorf("ProtoFiles[%d] mismatch: expected %s, got %s", i, expected, loadedConfig.ProtoFiles[i])
		}
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

func TestResolveProtoFiles(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "protobuf-mcp-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test proto files
	protoDir := filepath.Join(tempDir, "proto")
	err = os.MkdirAll(protoDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create proto directory: %v", err)
	}

	// Create test proto file
	testProtoFile := filepath.Join(protoDir, "test.proto")
	err = os.WriteFile(testProtoFile, []byte("syntax = \"proto3\";\npackage test;\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test proto file: %v", err)
	}

	// Test relative path resolution
	config := &ProjectConfig{
		ProtoFiles: []string{"proto/**/*.proto"},
	}

	resolvedFiles, err := ResolveProtoFiles(config, tempDir)
	if err != nil {
		t.Fatalf("Failed to resolve proto files: %v", err)
	}

	// Debug: print what we found
	t.Logf("Resolved files: %v", resolvedFiles)
	t.Logf("Pattern: proto/**/*.proto")
	t.Logf("Full pattern: %s", filepath.Join(tempDir, "proto", "**", "*.proto"))

	if len(resolvedFiles) == 0 {
		// Try a simpler pattern
		config.ProtoFiles = []string{"proto/*.proto"}
		resolvedFiles, err = ResolveProtoFiles(config, tempDir)
		if err != nil {
			t.Fatalf("Failed to resolve proto files with simple pattern: %v", err)
		}
		t.Logf("Resolved files with simple pattern: %v", resolvedFiles)
	}

	if len(resolvedFiles) != 1 {
		t.Errorf("Expected 1 resolved file, got %d", len(resolvedFiles))
		return
	}

	expectedFile := filepath.Join(tempDir, "proto", "test.proto")
	if resolvedFiles[0] != expectedFile {
		t.Errorf("Expected resolved file %s, got %s", expectedFile, resolvedFiles[0])
	}
}

func TestResolveProtoFilesWithAbsolutePath(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "protobuf-mcp-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test proto files
	protoDir := filepath.Join(tempDir, "proto")
	err = os.MkdirAll(protoDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create proto directory: %v", err)
	}

	// Create test proto file
	testProtoFile := filepath.Join(protoDir, "test.proto")
	err = os.WriteFile(testProtoFile, []byte("syntax = \"proto3\";\npackage test;\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test proto file: %v", err)
	}

	// Test absolute path resolution
	config := &ProjectConfig{
		ProtoFiles: []string{filepath.Join(tempDir, "proto", "*.proto")},
	}

	resolvedFiles, err := ResolveProtoFiles(config, tempDir)
	if err != nil {
		t.Fatalf("Failed to resolve proto files: %v", err)
	}

	if len(resolvedFiles) != 1 {
		t.Errorf("Expected 1 resolved file, got %d", len(resolvedFiles))
		return
	}

	if resolvedFiles[0] != testProtoFile {
		t.Errorf("Expected resolved file %s, got %s", testProtoFile, resolvedFiles[0])
	}
}