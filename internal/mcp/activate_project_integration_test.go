package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/yuemori/protobuf-mcp-server/internal/config"
	"github.com/yuemori/protobuf-mcp-server/internal/tools"
)

// TestActivateProjectTool_Integration tests the activate_project tool through MCP protocol
// TestActivateProjectTool_Integration tests the activate_project tool through MCP protocol
func TestActivateProjectTool_Integration(t *testing.T) {
	// Create a temporary directory with a test project
	tempDir, err := os.MkdirTemp("", "protobuf-activate-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create project config
	projectConfig := &config.ProjectConfig{
		ProtoFiles: []string{"**/*.proto"},
	}
	if err := config.SaveProjectConfig(tempDir, projectConfig); err != nil {
		t.Fatalf("Failed to save project config: %v", err)
	}

	// Create a test proto file
	protoContent := `syntax = "proto3";
package test;

service TestService {
  rpc TestMethod(TestRequest) returns (TestResponse);
}

message TestRequest {
  string name = 1;
}

message TestResponse {
  string message = 1;
}`
	protoFile := filepath.Join(tempDir, "test.proto")
	if err := os.WriteFile(protoFile, []byte(protoContent), 0644); err != nil {
		t.Fatalf("Failed to write proto file: %v", err)
	}

	// Create server and client
	server := NewMCPServer()
	mcpClient, err := client.NewInProcessClient(server.server)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Initialize
	_, err = mcpClient.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			ClientInfo: mcp.Implementation{
				Name:    "test-client",
				Version: "1.0.0",
			},
			Capabilities: mcp.ClientCapabilities{},
		},
	})
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Test successful activation
	t.Run("SuccessfulActivation", func(t *testing.T) {
		result, err := mcpClient.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "activate_project",
				Arguments: map[string]interface{}{
					"project_path": tempDir,
				},
			},
		})
		if err != nil {
			t.Fatalf("Failed to call activate_project: %v", err)
		}

		if result.IsError {
			if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
				t.Fatalf("activate_project returned error: %s", textContent.Text)
			} else {
				t.Fatalf("activate_project returned error")
			}
		}

		// Parse response
		var response tools.ActivateProjectResponse
		if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
			if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}
		} else {
			t.Fatalf("Expected text content in response")
		}

		if !response.Success {
			t.Fatalf("Expected success=true, got success=false: %s", response.Message)
		}

		if response.Services != 1 {
			t.Errorf("Expected 1 service, got %d", response.Services)
		}

		if response.Messages != 2 {
			t.Errorf("Expected 2 messages, got %d", response.Messages)
		}
	})

	// Test activation with non-existent path
	t.Run("NonExistentPath", func(t *testing.T) {
		result, err := mcpClient.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "activate_project",
				Arguments: map[string]interface{}{
					"project_path": "/non/existent/path",
				},
			},
		})
		if err != nil {
			t.Fatalf("Failed to call activate_project: %v", err)
		}

		var response tools.ActivateProjectResponse
		if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
			if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}
		}

		if response.Success {
			t.Error("Expected success=false for non-existent project")
		}
	})

	// Test activation with missing project_path parameter
	t.Run("MissingProjectPath", func(t *testing.T) {
		result, err := mcpClient.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "activate_project",
				Arguments: map[string]interface{}{},
			},
		})
		if err != nil {
			t.Fatalf("Failed to call activate_project: %v", err)
		}

		if !result.IsError {
			t.Error("Expected error for missing project_path parameter")
		}
	})
}
