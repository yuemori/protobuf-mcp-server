package compiler

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/yuemori/protobuf-mcp-server/internal/config"
)

func TestNewProtobufProject(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "protobuf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test proto file
	testProtoFile := filepath.Join(tempDir, "test.proto")
	err = os.WriteFile(testProtoFile, []byte(`syntax = "proto3";
package test;
message TestMessage {
  string name = 1;
}`), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test proto file: %v", err)
	}

	// Create a config
	cfg := &config.ProjectConfig{
		ProtoFiles:  []string{"test.proto"},
		ImportPaths: []string{"."},
	}

	// Create project
	project, err := NewProtobufProject(tempDir, cfg)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Verify project structure
	if project.ProjectRoot != tempDir {
		t.Errorf("Expected ProjectRoot %s, got %s", tempDir, project.ProjectRoot)
	}
}

func TestCompileProtosIntegration(t *testing.T) {
	// Skip this test - replaced by TestCompileProtosWithImportPaths
	t.Skip("Replaced by TestCompileProtosWithImportPaths")
}

func TestCompileComplexProtosIntegration(t *testing.T) {
	// Skip this test - replaced by TestCompileProtosWithNestedImportPaths
	t.Skip("Replaced by TestCompileProtosWithNestedImportPaths")
}

func TestCompileProtosWithImportPaths(t *testing.T) {
	// Test compilation with import paths using the standalone function
	ctx := context.Background()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("Failed to restore working directory: %v", err)
		}
	})

	// Test with simple proto files - use relative paths
	rootDir := filepath.Join(cwd, "testdata/simple")
	protoFiles := []string{
		"api.proto",
		"types.proto",
	}
	importPaths := []string{"."}

	compiledProtos, err := CompileProtos(ctx, rootDir, protoFiles, importPaths)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	// Verify compilation results
	if compiledProtos == nil {
		t.Fatal("Expected compiled protos, got nil")
	}

	if len(compiledProtos) == 0 {
		t.Fatal("Expected compiled files, got empty")
	}
}

func TestCompileProtosWithNestedImportPaths(t *testing.T) {
	// Test compilation with nested import paths
	ctx := context.Background()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("Failed to restore working directory: %v", err)
		}
	})

	// Test with nested proto files - use relative paths
	protoFiles := []string{
		"proto/my-service/api/v1/hoge.proto",
		"proto/my-service/api/v1/foo.proto",
	}
	importPaths := []string{"proto"}

	rootDir := filepath.Join(cwd, "testdata/nested")

	compiledProtos, err := CompileProtos(ctx, rootDir, protoFiles, importPaths)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	// Verify compilation results
	if compiledProtos == nil {
		t.Fatal("Expected compiled protos, got nil")
	}

	if len(compiledProtos) == 0 {
		t.Fatal("Expected compiled files, got empty")
	}

	// Check that we have the expected files
	fileNames := make(map[string]bool)
	for _, file := range compiledProtos {
		fileNames[string(file.Name())] = true
	}

	// Verify that the import was resolved correctly
	// hoge.proto imports foo.proto, so both should be compiled
	if len(compiledProtos) < 2 {
		t.Errorf("Expected at least 2 files (hoge.proto and foo.proto), got %d", len(compiledProtos))
	}
}
