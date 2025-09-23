package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/yuemori/protobuf-mcp-server/internal/config"
)

func TestNewActivateProjectTool(t *testing.T) {
	tool := NewActivateProjectTool()
	if tool == nil {
		t.Fatal("NewActivateProjectTool() returned nil")
	}
	
	if tool.Name() != "activate_project" {
		t.Errorf("Expected tool name 'activate_project', got '%s'", tool.Name())
	}
	
	if tool.Description() == "" {
		t.Error("Tool description should not be empty")
	}
	
	if tool.IsProjectActivated() {
		t.Error("Project should not be activated initially")
	}
}

func TestActivateProjectTool_Execute_ProjectNotInitialized(t *testing.T) {
	// Create temporary directory without .protobuf-mcp configuration
	tempDir, err := os.MkdirTemp("", "protobuf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	tool := NewActivateProjectTool()
	params := ActivateProjectParams{
		ProjectPath: tempDir,
	}
	
	ctx := context.Background()
	response, err := tool.Execute(ctx, params)
	
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}
	
	if response.Success {
		t.Error("Expected success=false for uninitialized project")
	}
	
	expectedError := "Project not initialized. Run 'go run cmd/protobuf-mcp/main.go init' first."
	if response.Error != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, response.Error)
	}
	
	if tool.IsProjectActivated() {
		t.Error("Project should not be activated after failed execution")
	}
}

func TestActivateProjectTool_Execute_Success(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "protobuf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create .protobuf-mcp directory and config
	configDir := filepath.Join(tempDir, ".protobuf-mcp")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	
	// Create project.yml config file
	configPath := filepath.Join(configDir, "project.yml")
	cfg := &config.ProjectConfig{
		RootDirectory: ".",
		IncludePaths:  []string{"."},
		ProtoPaths:    []string{"."},
		CompilerOptions: make(map[string]interface{}),
		IgnoredPatterns: []string{"*_test.proto"},
		ShowLogs:        false,
	}
	
	if err := config.SaveProjectConfig(configPath, cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}
	
	// Create a simple proto file for testing
	protoContent := `syntax = "proto3";

package test;

service TestService {
    rpc GetUser(GetUserRequest) returns (GetUserResponse);
}

message GetUserRequest {
    string user_id = 1;
}

message GetUserResponse {
    string name = 1;
    string email = 2;
}

enum Status {
    UNKNOWN = 0;
    ACTIVE = 1;
    INACTIVE = 2;
}
`
	protoPath := filepath.Join(tempDir, "test.proto")
	if err := os.WriteFile(protoPath, []byte(protoContent), 0644); err != nil {
		t.Fatalf("Failed to create proto file: %v", err)
	}
	
	tool := NewActivateProjectTool()
	params := ActivateProjectParams{
		ProjectPath: tempDir,
	}
	
	ctx := context.Background()
	response, err := tool.Execute(ctx, params)
	
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}
	
	if !response.Success {
		t.Errorf("Expected success=true, got false. Error: %s", response.Error)
	}
	
	if response.ProjectRoot != tempDir {
		t.Errorf("Expected project root '%s', got '%s'", tempDir, response.ProjectRoot)
	}
	
	if response.ConfigPath != configPath {
		t.Errorf("Expected config path '%s', got '%s'", configPath, response.ConfigPath)
	}
	
	if len(response.ProtoFiles) == 0 {
		t.Error("Expected at least one proto file")
	}
	
	if response.Message == "" {
		t.Error("Expected success message")
	}
	
	if !tool.IsProjectActivated() {
		t.Error("Project should be activated after successful execution")
	}
	
	// Verify we can get the current project
	project := tool.GetCurrentProject()
	if project == nil {
		t.Error("GetCurrentProject() should return non-nil after activation")
	}
}

func TestActivateProjectTool_Execute_InvalidPath(t *testing.T) {
	tool := NewActivateProjectTool()
	params := ActivateProjectParams{
		ProjectPath: "/nonexistent/path/that/should/not/exist",
	}
	
	ctx := context.Background()
	response, err := tool.Execute(ctx, params)
	
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}
	
	if response.Success {
		t.Error("Expected success=false for invalid path")
	}
	
	if response.Error == "" {
		t.Error("Expected error message for invalid path")
	}
}

func TestActivateProjectTool_Execute_EmptyProjectPath(t *testing.T) {
	tool := NewActivateProjectTool()
	params := ActivateProjectParams{
		ProjectPath: "", // Empty path should default to current directory
	}
	
	ctx := context.Background()
	response, err := tool.Execute(ctx, params)
	
	// This test depends on the current directory not having .protobuf-mcp
	// In a real scenario, this would likely fail with "not initialized" error
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}
	
	// Should not crash and should provide a meaningful response
	if response == nil {
		t.Error("Response should not be nil")
	}
}

func TestActivateProjectTool_Execute_ConfigLoadError(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "protobuf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create .protobuf-mcp directory
	configDir := filepath.Join(tempDir, ".protobuf-mcp")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	
	// Create invalid config file (corrupted YAML)
	configPath := filepath.Join(configDir, "project.yml")
	invalidConfig := "invalid: yaml: content: ["
	if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("Failed to create invalid config: %v", err)
	}
	
	tool := NewActivateProjectTool()
	params := ActivateProjectParams{
		ProjectPath: tempDir,
	}
	
	ctx := context.Background()
	response, err := tool.Execute(ctx, params)
	
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}
	
	if response.Success {
		t.Error("Expected success=false for invalid config")
	}
	
	if response.Error == "" {
		t.Error("Expected error message for invalid config")
	}
	
	if tool.IsProjectActivated() {
		t.Error("Project should not be activated after config error")
	}
}