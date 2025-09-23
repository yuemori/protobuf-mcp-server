package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/yuemori/protobuf-mcp-server/internal/config"
)

func TestActivateProjectTool_Handle(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "protobuf-mcp-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize the project
	projectConfig := &config.ProjectConfig{
		ProtoFiles: []string{"**/*.proto"},
	}
	// Save the config using the proper YAML format
	if err := config.SaveProjectConfig(tempDir, projectConfig); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify project exists
	if !config.ProjectExists(tempDir) {
		t.Fatalf("Project should exist after saving config")
	}

	// Create a test proto file
	protoContent := `syntax = "proto3";

package test;

message TestMessage {
  string name = 1;
  int32 value = 2;
}

service TestService {
  rpc GetTest(TestMessage) returns (TestMessage);
}`
	protoPath := filepath.Join(tempDir, "test.proto")
	if err := os.WriteFile(protoPath, []byte(protoContent), 0644); err != nil {
		t.Fatalf("Failed to write proto file: %v", err)
	}

	// Create tool with mock project manager
	mockProjectManager := &MockProjectManager{}
	tool := NewActivateProjectTool(mockProjectManager)

	// Test successful activation
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "activate_project",
			Arguments: map[string]interface{}{
				"project_path": tempDir,
			},
		},
	}

	ctx := context.Background()
	result, err := tool.Handle(ctx, req)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}

	if result.IsError {
		if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
			t.Fatalf("Expected success, got error: %s", textContent.Text)
		} else {
			t.Fatalf("Expected success, got error")
		}
	}

	// Parse response
	var response ActivateProjectResponse
	if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
		if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
	} else {
		t.Fatalf("Expected text content")
	}

	if !response.Success {
		t.Fatalf("Expected success=true, got success=false: %s", response.Message)
	}

	if response.ProjectRoot != tempDir {
		t.Fatalf("Expected ProjectRoot=%s, got %s", tempDir, response.ProjectRoot)
	}

	if response.ProtoFiles != 1 {
		t.Fatalf("Expected ProtoFiles=1, got %d", response.ProtoFiles)
	}

	if response.Services != 1 {
		t.Fatalf("Expected Services=1, got %d", response.Services)
	}

	if response.Messages != 1 {
		t.Fatalf("Expected Messages=1, got %d", response.Messages)
	}
}

func TestActivateProjectTool_Handle_InvalidPath(t *testing.T) {
	// Create tool with mock project manager
	mockProjectManager := &MockProjectManager{}
	tool := NewActivateProjectTool(mockProjectManager)

	// Test with invalid path
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "activate_project",
			Arguments: map[string]interface{}{
				"project_path": "/nonexistent/path",
			},
		},
	}

	ctx := context.Background()
	result, err := tool.Handle(ctx, req)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}

	// The result is not an error type, but contains a failure response
	if result.IsError {
		t.Fatalf("Expected regular response, got error type")
	}

	// Parse response to check if it's a failure
	var response ActivateProjectResponse
	if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
		if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
	} else {
		t.Fatalf("Expected text content in response")
	}

	if response.Success {
		t.Fatalf("Expected success=false for invalid path, got success=true")
	}

	if !strings.Contains(response.Message, "Project not initialized") {
		t.Fatalf("Expected error message about project not initialized, got: %s", response.Message)
	}
}

func TestActivateProjectTool_Handle_MissingPath(t *testing.T) {
	// Create tool with mock project manager
	mockProjectManager := &MockProjectManager{}
	tool := NewActivateProjectTool(mockProjectManager)

	// Test with missing path
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "activate_project",
			Arguments: map[string]interface{}{},
		},
	}

	ctx := context.Background()
	result, err := tool.Handle(ctx, req)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}

	if !result.IsError {
		t.Fatalf("Expected error, got success")
	}

	// Check error message
	if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
		if !strings.Contains(textContent.Text, "project_path parameter is required") {
			t.Fatalf("Expected error message about missing project_path, got: %s", textContent.Text)
		}
	} else {
		t.Fatalf("Expected text content in error response")
	}
}
