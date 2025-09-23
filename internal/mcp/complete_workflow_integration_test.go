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

// TestCompleteWorkflow_Integration tests the complete end-to-end workflow through MCP protocol
func TestCompleteWorkflow_Integration(t *testing.T) {
	// Create a temporary directory with a test project
	tempDir, err := os.MkdirTemp("", "protobuf-workflow-test-*")
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

	// Create a comprehensive test proto file
	protoContent := `syntax = "proto3";
package test;

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
  rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse);
  rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse);
}

service ProductService {
  rpc GetProduct(GetProductRequest) returns (GetProductResponse);
  rpc ListProducts(ListProductsRequest) returns (ListProductsResponse);
  rpc CreateProduct(CreateProductRequest) returns (CreateProductResponse);
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
  UserRole role = 3;
}

message CreateUserResponse {
  User user = 1;
}

message UpdateUserRequest {
  string user_id = 1;
  string name = 2;
  string email = 3;
  UserRole role = 4;
}

message UpdateUserResponse {
  User user = 1;
}

message DeleteUserRequest {
  string user_id = 1;
}

message DeleteUserResponse {
  bool success = 1;
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
  ProductCategory category = 3;
}

message ListProductsResponse {
  repeated Product products = 1;
  int32 total = 2;
  int32 page = 3;
}

message CreateProductRequest {
  string name = 1;
  string description = 2;
  float price = 3;
  ProductCategory category = 4;
}

message CreateProductResponse {
  Product product = 1;
}

message User {
  string id = 1;
  string name = 2;
  string email = 3;
  UserRole role = 4;
  UserStatus status = 5;
  int64 created_at = 6;
  int64 updated_at = 7;
}

message Product {
  string id = 1;
  string name = 2;
  string description = 3;
  float price = 4;
  ProductCategory category = 5;
  ProductStatus status = 6;
  int64 created_at = 7;
  int64 updated_at = 8;
}

enum UserRole {
  UNKNOWN_ROLE = 0;
  ADMIN = 1;
  USER = 2;
  MODERATOR = 3;
}

enum UserStatus {
  UNKNOWN_STATUS = 0;
  ACTIVE = 1;
  INACTIVE = 2;
  SUSPENDED = 3;
  PENDING = 4;
}

enum ProductCategory {
  UNKNOWN_CATEGORY = 0;
  ELECTRONICS = 1;
  CLOTHING = 2;
  BOOKS = 3;
  HOME = 4;
  SPORTS = 5;
}

enum ProductStatus {
  UNKNOWN_PRODUCT_STATUS = 0;
  AVAILABLE = 1;
  OUT_OF_STOCK = 2;
  DISCONTINUED = 3;
}`
	protoFile := filepath.Join(tempDir, "workflow.proto")
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

	// Step 1: Initialize
	t.Run("Step1_Initialize", func(t *testing.T) {
		_, err = mcpClient.Initialize(ctx, mcp.InitializeRequest{
			Params: mcp.InitializeParams{
				ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
				ClientInfo: mcp.Implementation{
					Name:    "workflow-test-client",
					Version: "1.0.0",
				},
				Capabilities: mcp.ClientCapabilities{},
			},
		})
		if err != nil {
			t.Fatalf("Failed to initialize: %v", err)
		}
	})

	// Step 2: List available tools
	t.Run("Step2_ListTools", func(t *testing.T) {
		toolsResult, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
		if err != nil {
			t.Fatalf("Failed to list tools: %v", err)
		}

		if len(toolsResult.Tools) != 3 {
			t.Errorf("Expected 3 tools, got %d", len(toolsResult.Tools))
		}
	})

	// Step 3: Activate project
	t.Run("Step3_ActivateProject", func(t *testing.T) {
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

		// Verify activation results
		if activateResponse.Services != 2 {
			t.Errorf("Expected 2 services, got %d", activateResponse.Services)
		}
		if activateResponse.Messages < 10 {
			t.Errorf("Expected at least 10 messages, got %d", activateResponse.Messages)
		}
	})

	// Step 4: List services
	t.Run("Step4_ListServices", func(t *testing.T) {
		listResult, err := mcpClient.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "list_services",
				Arguments: map[string]interface{}{},
			},
		})
		if err != nil {
			t.Fatalf("Failed to list services: %v", err)
		}

		var listResponse tools.ListServicesResponse
		if textContent, ok := mcp.AsTextContent(listResult.Content[0]); ok {
			if err := json.Unmarshal([]byte(textContent.Text), &listResponse); err != nil {
				t.Fatalf("Failed to unmarshal list response: %v", err)
			}
		}
		if !listResponse.Success {
			t.Fatalf("List services failed: %s", listResponse.Message)
		}

		if listResponse.Count != 2 {
			t.Errorf("Expected 2 services, got %d", listResponse.Count)
		}

		// Verify service names
		serviceNames := make(map[string]bool)
		for _, service := range listResponse.Services {
			serviceNames[service.Name] = true
		}

		if !serviceNames["UserService"] {
			t.Error("Expected UserService to be present")
		}
		if !serviceNames["ProductService"] {
			t.Error("Expected ProductService to be present")
		}
	})

	// Step 5: Get complete schema
	t.Run("Step5_GetCompleteSchema", func(t *testing.T) {
		schemaResult, err := mcpClient.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_schema",
				Arguments: map[string]interface{}{
					"include_file_info": true,
				},
			},
		})
		if err != nil {
			t.Fatalf("Failed to get schema: %v", err)
		}

		var schemaResponse tools.GetSchemaResponse
		if textContent, ok := mcp.AsTextContent(schemaResult.Content[0]); ok {
			if err := json.Unmarshal([]byte(textContent.Text), &schemaResponse); err != nil {
				t.Fatalf("Failed to unmarshal schema response: %v", err)
			}
		}
		if !schemaResponse.Success {
			t.Fatalf("Get schema failed: %s", schemaResponse.Message)
		}

		// Verify schema completeness
		if schemaResponse.Schema == nil {
			t.Fatal("Expected schema to be present")
		}

		// Check services
		if len(schemaResponse.Schema.Services) != 2 {
			t.Errorf("Expected 2 services, got %d", len(schemaResponse.Schema.Services))
		}

		// Check messages (should have all request/response messages)
		if len(schemaResponse.Schema.Messages) < 10 {
			t.Errorf("Expected at least 10 messages, got %d", len(schemaResponse.Schema.Messages))
		}

		// Check enums
		if len(schemaResponse.Schema.Enums) != 4 {
			t.Errorf("Expected 4 enums, got %d", len(schemaResponse.Schema.Enums))
		}
	})

	// Step 6: Get filtered schema
	t.Run("Step6_GetFilteredSchema", func(t *testing.T) {
		schemaResult, err := mcpClient.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_schema",
				Arguments: map[string]interface{}{
					"message_types": []interface{}{"User", "Product"},
					"enum_types":    []interface{}{"UserRole", "ProductCategory"},
				},
			},
		})
		if err != nil {
			t.Fatalf("Failed to get filtered schema: %v", err)
		}

		var schemaResponse tools.GetSchemaResponse
		if textContent, ok := mcp.AsTextContent(schemaResult.Content[0]); ok {
			if err := json.Unmarshal([]byte(textContent.Text), &schemaResponse); err != nil {
				t.Fatalf("Failed to unmarshal filtered schema response: %v", err)
			}
		}
		if !schemaResponse.Success {
			t.Fatalf("Get filtered schema failed: %s", schemaResponse.Message)
		}

		// Verify filtered results
		if len(schemaResponse.Schema.Messages) != 2 {
			t.Errorf("Expected 2 messages, got %d", len(schemaResponse.Schema.Messages))
		}

		if len(schemaResponse.Schema.Enums) != 2 {
			t.Errorf("Expected 2 enums, got %d", len(schemaResponse.Schema.Enums))
		}

		// Check specific types
		messageNames := make(map[string]bool)
		for _, msg := range schemaResponse.Schema.Messages {
			messageNames[msg.Name] = true
		}

		if !messageNames["User"] {
			t.Error("Expected User message to be present")
		}
		if !messageNames["Product"] {
			t.Error("Expected Product message to be present")
		}

		enumNames := make(map[string]bool)
		for _, enum := range schemaResponse.Schema.Enums {
			enumNames[enum.Name] = true
		}

		if !enumNames["UserRole"] {
			t.Error("Expected UserRole enum to be present")
		}
		if !enumNames["ProductCategory"] {
			t.Error("Expected ProductCategory enum to be present")
		}
	})

	// Step 7: Test error handling
	t.Run("Step7_ErrorHandling", func(t *testing.T) {
		// Test calling non-existent tool
		_, err := mcpClient.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "non_existent_tool",
				Arguments: map[string]interface{}{},
			},
		})
		if err == nil {
			t.Error("Expected error for non-existent tool")
		}

		// Test calling tool with invalid arguments
		result, err := mcpClient.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "activate_project",
				Arguments: map[string]interface{}{
					"invalid_param": "value",
				},
			},
		})
		if err != nil {
			t.Fatalf("Failed to call activate_project: %v", err)
		}

		if !result.IsError {
			t.Error("Expected error for invalid arguments")
		}
	})
}
