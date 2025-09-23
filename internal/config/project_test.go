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

func TestYAMLParsing(t *testing.T) {
	tests := []struct {
		name          string
		yaml          string
		expectedFiles []string
		expectedPaths []string
	}{
		{
			name: "quoted glob patterns with import paths",
			yaml: `proto_files:
  - "**/*.proto"
  - "proto/**/*.proto"
import_paths:
  - "."
  - "proto"`,
			expectedFiles: []string{"**/*.proto", "proto/**/*.proto"},
			expectedPaths: []string{".", "proto"},
		},
		{
			name: "unquoted glob patterns with import paths",
			yaml: `proto_files:
  - **/*.proto
  - proto/**/*.proto
import_paths:
  - .
  - proto`,
			expectedFiles: []string{"**/*.proto", "proto/**/*.proto"},
			expectedPaths: []string{".", "proto"},
		},
		{
			name: "mixed quoted and unquoted patterns",
			yaml: `proto_files:
  - **/*.proto
  - "api/**/*.proto"
  - proto/**/*.proto
import_paths:
  - "."
  - api`,
			expectedFiles: []string{"**/*.proto", "api/**/*.proto", "proto/**/*.proto"},
			expectedPaths: []string{".", "api"},
		},
		{
			name: "simple patterns without glob",
			yaml: `proto_files:
  - api.proto
  - types.proto
import_paths:
  - .`,
			expectedFiles: []string{"api.proto", "types.proto"},
			expectedPaths: []string{"."},
		},
		{
			name: "single pattern with multiple import paths",
			yaml: `proto_files:
  - **/*.proto
import_paths:
  - .
  - proto
  - api`,
			expectedFiles: []string{"**/*.proto"},
			expectedPaths: []string{".", "proto", "api"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file with the test YAML
			tempDir, err := os.MkdirTemp("", "yaml-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			configPath := filepath.Join(tempDir, ".protobuf-mcp.yml")
			if err := os.WriteFile(configPath, []byte(tt.yaml), 0644); err != nil {
				t.Fatalf("Failed to write test config: %v", err)
			}

			// Load the config
			config, err := LoadProjectConfig(tempDir)
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			// Check the parsed proto files
			if len(config.ProtoFiles) != len(tt.expectedFiles) {
				t.Errorf("Expected %d proto files, got %d", len(tt.expectedFiles), len(config.ProtoFiles))
				return
			}

			for i, expected := range tt.expectedFiles {
				if i >= len(config.ProtoFiles) || config.ProtoFiles[i] != expected {
					t.Errorf("Expected ProtoFiles[%d] to be %s, got %s", i, expected, config.ProtoFiles[i])
				}
			}

			// Check the parsed import paths
			if len(config.ImportPaths) != len(tt.expectedPaths) {
				t.Errorf("Expected %d import paths, got %d", len(tt.expectedPaths), len(config.ImportPaths))
				return
			}

			for i, expected := range tt.expectedPaths {
				if i >= len(config.ImportPaths) || config.ImportPaths[i] != expected {
					t.Errorf("Expected ImportPaths[%d] to be %s, got %s", i, expected, config.ImportPaths[i])
				}
			}
		})
	}
}

func TestPreprocessYAML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "unquoted glob patterns",
			input: `proto_files:
  - **/*.proto
  - proto/**/*.proto`,
			expected: `proto_files:
  - "**/*.proto"
  - "proto/**/*.proto"`,
		},
		{
			name: "already quoted patterns",
			input: `proto_files:
  - "**/*.proto"
  - "proto/**/*.proto"`,
			expected: `proto_files:
  - "**/*.proto"
  - "proto/**/*.proto"`,
		},
		{
			name: "mixed patterns",
			input: `proto_files:
  - **/*.proto
  - "api/**/*.proto"
  - proto/**/*.proto`,
			expected: `proto_files:
  - "**/*.proto"
  - "api/**/*.proto"
  - "proto/**/*.proto"`,
		},
		{
			name: "simple patterns without glob",
			input: `proto_files:
  - api.proto
  - types.proto`,
			expected: `proto_files:
  - api.proto
  - types.proto`,
		},
		{
			name: "single asterisk patterns",
			input: `proto_files:
  - *.proto
  - api/*.proto`,
			expected: `proto_files:
  - "*.proto"
  - "api/*.proto"`,
		},
		{
			name: "unquoted import paths",
			input: `proto_files:
  - **/*.proto
import_paths:
  - .
  - proto`,
			expected: `proto_files:
  - "**/*.proto"
import_paths:
  - .
  - proto`,
		},
		{
			name: "mixed quoted and unquoted import paths",
			input: `proto_files:
  - **/*.proto
import_paths:
  - .
  - "proto"
  - api`,
			expected: `proto_files:
  - "**/*.proto"
import_paths:
  - .
  - "proto"
  - api`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := preprocessYAML(tt.input)
			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}
