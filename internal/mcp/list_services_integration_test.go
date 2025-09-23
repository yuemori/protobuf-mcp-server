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

// TestListServicesTool_Integration tests the list_services tool through MCP protocol
func TestListServicesTool_Integration(t *testing.T) {
	// Create a temporary directory with a test project
	tempDir, err := os.MkdirTemp("", "protobuf-list-services-test-*")
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

	// Create a test proto file with multiple services
	protoContent := `syntax = "proto3";
package test;

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
}

service ProductService {
  rpc GetProduct(GetProductRequest) returns (GetProductResponse);
  rpc ListProducts(ListProductsRequest) returns (ListProductsResponse);
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

message GetProductRequest {
  string product_id = 1;
}

message GetProductResponse {
  Product product = 1;
}

message ListProductsRequest {
  int32 page = 1;
  int32 limit = 2;
}

message ListProductsResponse {
  repeated Product products = 1;
  int32 total = 2;
}

message User {
  string id = 1;
  string name = 2;
  string email = 3;
}

message Product {
  string id = 1;
  string name = 2;
  string description = 3;
  float price = 4;
}`
	protoFile := filepath.Join(tempDir, "services.proto")
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

	// Test calling list_services without activating a project
	t.Run("NoProjectActivated", func(t *testing.T) {
		result, err := mcpClient.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "list_services",
				Arguments: map[string]interface{}{},
			},
		})
		if err != nil {
			t.Fatalf("Failed to call list_services: %v", err)
		}

		var response tools.ListServicesResponse
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

	// Test successful list_services after project activation
	t.Run("SuccessfulListServices", func(t *testing.T) {
		// First activate the project
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

		// Now list services
		result, err := mcpClient.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "list_services",
				Arguments: map[string]interface{}{},
			},
		})
		if err != nil {
			t.Fatalf("Failed to call list_services: %v", err)
		}

		var response tools.ListServicesResponse
		if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
			if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}
		}

		if !response.Success {
			t.Fatalf("List services failed: %s", response.Message)
		}

		if response.Count != 2 {
			t.Errorf("Expected 2 services, got %d", response.Count)
		}

		if len(response.Services) != 2 {
			t.Errorf("Expected 2 services in response, got %d", len(response.Services))
		}

		// Check service names
		serviceNames := make(map[string]bool)
		for _, service := range response.Services {
			serviceNames[service.Name] = true
		}

		if !serviceNames["UserService"] {
			t.Error("Expected UserService to be present")
		}
		if !serviceNames["ProductService"] {
			t.Error("Expected ProductService to be present")
		}
	})
}
