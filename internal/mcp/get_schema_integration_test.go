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

// TestGetSchemaTool_Integration tests the get_schema tool through MCP protocol
func TestGetSchemaTool_Integration(t *testing.T) {
	// Create a temporary directory with a test project
	tempDir, err := os.MkdirTemp("", "protobuf-get-schema-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create project config
	projectConfig := &config.ProjectConfig{
		RootDirectory: tempDir,
		ProtoPaths:    []string{"."},
		IncludePaths:  []string{"."},
	}
	if err := config.SaveProjectConfig(tempDir, projectConfig); err != nil {
		t.Fatalf("Failed to save project config: %v", err)
	}

	// Create a test proto file with various types
	protoContent := `syntax = "proto3";
package test;

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
}

message GetUserRequest {
  string user_id = 1;
}

message GetUserResponse {
  User user = 1;
}

message CreateUserRequest {
  string name = 1;
  string email = 2;
}

message CreateUserResponse {
  User user = 1;
}

message User {
  string id = 1;
  string name = 2;
  string email = 3;
  UserStatus status = 4;
}

enum UserStatus {
  UNKNOWN = 0;
  ACTIVE = 1;
  INACTIVE = 2;
  SUSPENDED = 3;
}

message Product {
  string id = 1;
  string name = 2;
  string description = 3;
  float price = 4;
  ProductCategory category = 5;
}

enum ProductCategory {
  UNKNOWN_CATEGORY = 0;
  ELECTRONICS = 1;
  CLOTHING = 2;
  BOOKS = 3;
}`
	protoFile := filepath.Join(tempDir, "schema.proto")
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

	// Activate project first
	activateResult, err := mcpClient.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "activate_project",
			Arguments: map[string]interface{}{
				"project_path": tempDir,
			},
		},
	})
	if err != nil {
		t.Fatalf("Failed to activate project: %v", err)
	}

	var activateResponse tools.ActivateProjectResponse
	if textContent, ok := mcp.AsTextContent(activateResult.Content[0]); ok {
		if err := json.Unmarshal([]byte(textContent.Text), &activateResponse); err != nil {
			t.Fatalf("Failed to unmarshal activate response: %v", err)
		}
	}
	if !activateResponse.Success {
		t.Fatalf("Project activation failed: %s", activateResponse.Message)
	}

	// Test calling get_schema without any filters
	t.Run("GetAllSchema", func(t *testing.T) {
		result, err := mcpClient.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "get_schema",
				Arguments: map[string]interface{}{},
			},
		})
		if err != nil {
			t.Fatalf("Failed to call get_schema: %v", err)
		}

		var response tools.GetSchemaResponse
		if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
			if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}
		}

		if !response.Success {
			t.Fatalf("Get schema failed: %s", response.Message)
		}

		if response.Schema == nil {
			t.Fatal("Expected schema to be present")
		}

		// Check that we have all expected types
		if len(response.Schema.Messages) < 6 {
			t.Errorf("Expected at least 6 messages, got %d", len(response.Schema.Messages))
		}

		if len(response.Schema.Services) != 1 {
			t.Errorf("Expected 1 service, got %d", len(response.Schema.Services))
		}

		if len(response.Schema.Enums) != 2 {
			t.Errorf("Expected 2 enums, got %d", len(response.Schema.Enums))
		}

		// Check stats
		if response.Schema.Stats.TotalMessages < 6 {
			t.Errorf("Expected at least 6 total messages, got %d", response.Schema.Stats.TotalMessages)
		}
		if response.Schema.Stats.TotalServices != 1 {
			t.Errorf("Expected 1 total service, got %d", response.Schema.Stats.TotalServices)
		}
		if response.Schema.Stats.TotalEnums != 2 {
			t.Errorf("Expected 2 total enums, got %d", response.Schema.Stats.TotalEnums)
		}
	})

	// Test get_schema with message type filter
	t.Run("FilterByMessageTypes", func(t *testing.T) {
		result, err := mcpClient.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_schema",
				Arguments: map[string]interface{}{
					"message_types": []interface{}{"User", "Product"},
				},
			},
		})
		if err != nil {
			t.Fatalf("Failed to call get_schema: %v", err)
		}

		var response tools.GetSchemaResponse
		if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
			if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}
		}

		if !response.Success {
			t.Fatalf("Get schema failed: %s", response.Message)
		}

		if len(response.Schema.Messages) != 2 {
			t.Errorf("Expected 2 messages, got %d", len(response.Schema.Messages))
		}

		// Check that we got the right messages
		messageNames := make(map[string]bool)
		for _, msg := range response.Schema.Messages {
			messageNames[msg.Name] = true
		}

		if !messageNames["User"] {
			t.Error("Expected User message to be present")
		}
		if !messageNames["Product"] {
			t.Error("Expected Product message to be present")
		}
	})

	// Test get_schema with enum type filter
	t.Run("FilterByEnumTypes", func(t *testing.T) {
		result, err := mcpClient.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_schema",
				Arguments: map[string]interface{}{
					"enum_types": []interface{}{"UserStatus"},
				},
			},
		})
		if err != nil {
			t.Fatalf("Failed to call get_schema: %v", err)
		}

		var response tools.GetSchemaResponse
		if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
			if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}
		}

		if !response.Success {
			t.Fatalf("Get schema failed: %s", response.Message)
		}

		if len(response.Schema.Enums) != 1 {
			t.Errorf("Expected 1 enum, got %d", len(response.Schema.Enums))
		}

		if response.Schema.Enums[0].Name != "UserStatus" {
			t.Errorf("Expected UserStatus enum, got %s", response.Schema.Enums[0].Name)
		}
	})

	// Test get_schema with include_file_info
	t.Run("IncludeFileInfo", func(t *testing.T) {
		result, err := mcpClient.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_schema",
				Arguments: map[string]interface{}{
					"include_file_info": true,
				},
			},
		})
		if err != nil {
			t.Fatalf("Failed to call get_schema: %v", err)
		}

		var response tools.GetSchemaResponse
		if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
			if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}
		}

		if !response.Success {
			t.Fatalf("Get schema failed: %s", response.Message)
		}

		// Check that file info is included
		if len(response.Schema.Files) == 0 {
			t.Error("Expected file information to be included")
		}

		// Check that we have the schema.proto file
		found := false
		for _, file := range response.Schema.Files {
			if file.Name == "schema.proto" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected schema.proto file to be present in file info")
		}
	})

	// Test get_schema without project activated
	t.Run("NoProjectActivated", func(t *testing.T) {
		// Create a new server without project
		newServer := NewMCPServer()
		newClient, err := client.NewInProcessClient(newServer.server)
		if err != nil {
			t.Fatalf("Failed to create new client: %v", err)
		}

		// Initialize
		_, err = newClient.Initialize(ctx, mcp.InitializeRequest{
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

		// Try to get schema without activating project
		result, err := newClient.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "get_schema",
				Arguments: map[string]interface{}{},
			},
		})
		if err != nil {
			t.Fatalf("Failed to call get_schema: %v", err)
		}

		var response tools.GetSchemaResponse
		if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
			if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}
		}

		if response.Success {
			t.Error("Expected success=false when no project is activated")
		}

		if response.Message != "No project activated. Use activate_project first." {
			t.Errorf("Unexpected error message: %s", response.Message)
		}
	})
}
