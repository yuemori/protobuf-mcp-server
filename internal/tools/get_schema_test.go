package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestGetSchemaTool_GetTool(t *testing.T) {
	mockProjectManager := &MockProjectManager{}
	tool := NewGetSchemaTool(mockProjectManager)

	mcpTool := tool.GetTool()
	if mcpTool.Name != "get_schema" {
		t.Fatalf("Expected tool name 'get_schema', got '%s'", mcpTool.Name)
	}

	if mcpTool.Description == "" {
		t.Fatalf("Expected non-empty description")
	}
}

func TestGetSchemaTool_Handle_NoProject(t *testing.T) {
	mockProjectManager := &MockProjectManager{}
	tool := NewGetSchemaTool(mockProjectManager)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "get_schema",
			Arguments: map[string]interface{}{},
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
	var response GetSchemaResponse
	if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
		if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
	} else {
		t.Fatalf("Expected text content in response")
	}

	if response.Success {
		t.Fatalf("Expected success=false when no project is activated, got success=true")
	}

	if !strings.Contains(response.Message, "No project activated") {
		t.Fatalf("Expected error message about no project activated, got: %s", response.Message)
	}
}

func TestGetSchemaTool_Handle_Success(t *testing.T) {
	// Create a test project with real protobuf data
	project, err := CreateTestProject()
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	mockProjectManager := &MockProjectManager{}
	mockProjectManager.SetProject(project)
	tool := NewGetSchemaTool(mockProjectManager)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "get_schema",
			Arguments: map[string]interface{}{},
		},
	}

	ctx := context.Background()
	result, err := tool.Handle(ctx, req)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}

	if result.IsError {
		t.Fatalf("Expected regular response, got error type")
	}

	// Parse response to check if it's successful
	var response GetSchemaResponse
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

	// Check that we have schema data
	if response.Schema == nil {
		t.Fatalf("Expected schema data, got nil")
	}

	// Check that count matches
	expectedCount := len(response.Schema.Messages) + len(response.Schema.Services) + len(response.Schema.Enums)
	if response.Count != expectedCount {
		t.Fatalf("Expected count=%d, got count=%d", expectedCount, response.Count)
	}

	// Check that we have services
	if len(response.Schema.Services) == 0 {
		t.Fatalf("Expected at least one service, got %d", len(response.Schema.Services))
	}

	// Check that we have messages
	if len(response.Schema.Messages) == 0 {
		t.Fatalf("Expected at least one message, got %d", len(response.Schema.Messages))
	}

	// Verify service structure
	for i, service := range response.Schema.Services {
		if service.Name == "" {
			t.Fatalf("Service %d has empty name", i)
		}
		if service.FullName == "" {
			t.Fatalf("Service %d has empty full_name", i)
		}
		if service.File == "" {
			t.Fatalf("Service %d has empty file", i)
		}
		if service.Package == "" {
			t.Fatalf("Service %d has empty package", i)
		}
	}

	// Verify message structure
	for i, message := range response.Schema.Messages {
		if message.Name == "" {
			t.Fatalf("Message %d has empty name", i)
		}
		if message.FullName == "" {
			t.Fatalf("Message %d has empty full_name", i)
		}
		if message.File == "" {
			t.Fatalf("Message %d has empty file", i)
		}
		if message.Package == "" {
			t.Fatalf("Message %d has empty package", i)
		}
	}

	// Check for expected services from test data
	foundGreetingService := false
	foundUserService := false
	for _, service := range response.Schema.Services {
		switch service.Name {
		case "GreetingService":
			foundGreetingService = true
		case "UserService":
			foundUserService = true
		}
	}

	if !foundGreetingService {
		t.Fatalf("Expected to find GreetingService")
	}
	if !foundUserService {
		t.Fatalf("Expected to find UserService")
	}
}

func TestGetSchemaTool_Handle_WithFilters(t *testing.T) {
	// Create a test project with real protobuf data
	project, err := CreateTestProject()
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	mockProjectManager := &MockProjectManager{}
	mockProjectManager.SetProject(project)
	tool := NewGetSchemaTool(mockProjectManager)

	// Test with name filter
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "get_schema",
			Arguments: map[string]interface{}{
				"name": "Greeting",
			},
		},
	}

	ctx := context.Background()
	result, err := tool.Handle(ctx, req)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}

	if result.IsError {
		t.Fatalf("Expected regular response, got error type")
	}

	// Parse response
	var response GetSchemaResponse
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

	// Should find GreetingService
	foundGreetingService := false
	for _, service := range response.Schema.Services {
		if service.Name == "GreetingService" {
			foundGreetingService = true
			break
		}
	}

	if !foundGreetingService {
		t.Fatalf("Expected to find GreetingService with name filter 'Greeting'")
	}

	// Test with type filter
	req = mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "get_schema",
			Arguments: map[string]interface{}{
				"type": "service",
			},
		},
	}

	result, err = tool.Handle(ctx, req)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}

	if result.IsError {
		t.Fatalf("Expected regular response, got error type")
	}

	// Parse response
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

	// Should only have services, no messages or enums
	if len(response.Schema.Services) == 0 {
		t.Fatalf("Expected at least one service with type filter 'service', got %d", len(response.Schema.Services))
	}

	// Should not have messages when filtering by service type
	if len(response.Schema.Messages) > 0 {
		t.Fatalf("Expected no messages with type filter 'service', got %d", len(response.Schema.Messages))
	}
}
