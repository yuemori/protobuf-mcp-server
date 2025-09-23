package compiler

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/yuemori/protobuf-mcp-server/internal/config"
)

func TestNewProtobufProject(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "protobuf-compiler-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test proto file
	protoDir := filepath.Join(tempDir, "proto")
	err = os.MkdirAll(protoDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create proto directory: %v", err)
	}

	testProtoFile := filepath.Join(protoDir, "test.proto")
	err = os.WriteFile(testProtoFile, []byte("syntax = \"proto3\";\npackage test;\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test proto file: %v", err)
	}

	cfg := &config.ProjectConfig{
		ProtoFiles: []string{"proto/**/*.proto"},
	}

	project, err := NewProtobufProject(tempDir, cfg)
	if err != nil {
		t.Fatalf("Failed to create ProtobufProject: %v", err)
	}

	if project.ProjectRoot != tempDir {
		t.Errorf("Expected ProjectRoot to be %s, got %s", tempDir, project.ProjectRoot)
	}

	if project.Config != cfg {
		t.Error("Expected Config to be set correctly")
	}

	if project.resolver == nil {
		t.Error("Expected resolver to be initialized")
	}

	if project.CompiledProtos != nil {
		t.Error("Expected CompiledProtos to be nil initially")
	}

	if len(project.protoFiles) != 1 {
		t.Errorf("Expected 1 proto file, got %d", len(project.protoFiles))
	}
}

func TestCompileProtosIntegration(t *testing.T) {
	// Skip this test if we can't access the testdata files
	wd, err := os.Getwd()
	if err != nil {
		t.Skip("Cannot determine working directory")
	}

	projectRoot := filepath.Join(wd, "..", "..")
	testDataPath := filepath.Join(projectRoot, "internal", "compiler", "testdata", "simple")

	// Check if testdata exists
	if _, err := os.Stat(testDataPath); os.IsNotExist(err) {
		t.Skip("Testdata not available, skipping integration test")
	}

	// Use testdata/simple as the project root for compilation
	cfg := &config.ProjectConfig{
		ProtoFiles: []string{"**/*.proto"},
	}

	project, err := NewProtobufProject(testDataPath, cfg)
	if err != nil {
		t.Fatalf("Failed to create ProtobufProject: %v", err)
	}

	ctx := context.Background()
	err = project.CompileProtos(ctx)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	// Verify the results
	if project.CompiledProtos == nil {
		t.Error("Expected CompiledProtos to be set after successful compilation")
		return
	}

	if len(project.CompiledProtos.File) == 0 {
		t.Error("Expected to have compiled files")
		return
	}

	// Test service extraction
	services, err := project.GetServices()
	if err != nil {
		t.Errorf("Failed to get services: %v", err)
	} else {
		t.Logf("Found %d services", len(services))
		if len(services) != 2 {
			t.Errorf("Expected 2 services (GreetingService, UserService), got %d", len(services))
		}
	}

	// Test message extraction
	messages, err := project.GetMessages()
	if err != nil {
		t.Errorf("Failed to get messages: %v", err)
	} else {
		t.Logf("Found %d messages", len(messages))
		if len(messages) < 10 {
			t.Errorf("Expected at least 10 messages, got %d", len(messages))
		}
	}

	// Test enum extraction
	enums, err := project.GetEnums()
	if err != nil {
		t.Errorf("Failed to get enums: %v", err)
	} else {
		t.Logf("Found %d enums", len(enums))
		if len(enums) < 5 {
			t.Errorf("Expected at least 5 enums, got %d", len(enums))
		}
	}
}

func TestCompileComplexProtosIntegration(t *testing.T) {
	// Get the project root directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	
	// Find the project root (go up from internal/compiler to project root)
	projectRoot := filepath.Join(wd, "..", "..")
	
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "protobuf-mcp-test-complex-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Copy testdata/complex to temp directory
	srcDir := filepath.Join(projectRoot, "internal", "compiler", "testdata", "complex")
	destDir := filepath.Join(tempDir, "testdata", "complex")
	
	err = os.MkdirAll(destDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create dest dir: %v", err)
	}
	
	// Copy the complex test data
	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		
		destPath := filepath.Join(destDir, relPath)
		
		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		} else {
			// Copy file content
			srcFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer srcFile.Close()
			
			destFile, err := os.Create(destPath)
			if err != nil {
				return err
			}
			defer destFile.Close()
			
			_, err = io.Copy(destFile, srcFile)
			return err
		}
	})
	if err != nil {
		t.Fatalf("Failed to copy test data: %v", err)
	}

	// Use testdata/complex as the project root for compilation
	cfg := &config.ProjectConfig{
		ProtoFiles: []string{"**/*.proto"},
	}

	// Create protobuf project with testdata/complex as root
	project, err := NewProtobufProject(destDir, cfg)
	if err != nil {
		t.Fatalf("Failed to create protobuf project: %v", err)
	}

	// Compile protos
	err = project.CompileProtos(context.Background())
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	if project.CompiledProtos == nil {
		t.Fatal("Compiled result is nil")
	}

	// Verify we have the expected number of files
	expectedFiles := 2 // common.proto and user.proto
	if len(project.protoFiles) != expectedFiles {
		t.Errorf("Expected %d proto files, got %d", expectedFiles, len(project.protoFiles))
	}

	// Verify we can find services
	services, err := project.GetServices()
	if err != nil {
		t.Errorf("Failed to get services: %v", err)
	} else if len(services) == 0 {
		t.Error("No services found in complex test")
	} else {
		t.Logf("Found %d services in complex test", len(services))
	}

	// Verify we can find messages (should be many due to nested structures)
	messages, err := project.GetMessages()
	if err != nil {
		t.Errorf("Failed to get messages: %v", err)
	} else if len(messages) < 10 { // Adjusted expectation based on actual content
		t.Errorf("Expected at least 10 messages in complex test, got %d", len(messages))
	} else {
		t.Logf("Found %d messages in complex test", len(messages))
	}

	// Verify we can find enums (should be several due to nested enums)
	enums, err := project.GetEnums()
	if err != nil {
		t.Errorf("Failed to get enums: %v", err)
	} else if len(enums) < 2 { // Adjusted expectation based on actual content
		t.Errorf("Expected at least 2 enums in complex test, got %d", len(enums))
	} else {
		t.Logf("Found %d enums in complex test", len(enums))
	}
}

func TestGetServicesNotCompiled(t *testing.T) {
	cfg := &config.ProjectConfig{}
	project := &ProtobufProject{
		Config:         cfg,
		CompiledProtos: nil,
	}

	_, err := project.GetServices()
	if err == nil {
		t.Error("Expected error when getting services from non-compiled project")
	}

	expectedMessage := "project not compiled yet"
	if err.Error() != "project not compiled yet, call CompileProtos first" {
		t.Errorf("Expected error message to contain '%s', got: %s", expectedMessage, err.Error())
	}
}

func TestGetMessagesNotCompiled(t *testing.T) {
	cfg := &config.ProjectConfig{}
	project := &ProtobufProject{
		Config:         cfg,
		CompiledProtos: nil,
	}

	_, err := project.GetMessages()
	if err == nil {
		t.Error("Expected error when getting messages from non-compiled project")
	}
}

func TestGetEnumsNotCompiled(t *testing.T) {
	cfg := &config.ProjectConfig{}
	project := &ProtobufProject{
		Config:         cfg,
		CompiledProtos: nil,
	}

	_, err := project.GetEnums()
	if err == nil {
		t.Error("Expected error when getting enums from non-compiled project")
	}
}