package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestListServicesTool_GetTool(t *testing.T) {
	mockProjectManager := &MockProjectManager{}
	tool := NewListServicesTool(mockProjectManager)

	mcpTool := tool.GetTool()
	if mcpTool.Name != "list_services" {
		t.Fatalf("Expected tool name 'list_services', got '%s'", mcpTool.Name)
	}

	if mcpTool.Description == "" {
		t.Fatalf("Expected non-empty description")
	}
}

func TestListServicesTool_Handle_NoProject(t *testing.T) {
	mockProjectManager := &MockProjectManager{}
	tool := NewListServicesTool(mockProjectManager)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "list_services",
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
	var response ListServicesResponse
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

func TestListServicesTool_Handle_Success(t *testing.T) {
	// Create a test project with real protobuf data
	project, err := CreateTestProject(t)
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	mockProjectManager := &MockProjectManager{}
	mockProjectManager.SetProject(project)
	tool := NewListServicesTool(mockProjectManager)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "list_services",
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
	var response ListServicesResponse
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

	// Check that we have services
	if len(response.Services) == 0 {
		t.Fatalf("Expected at least one service, got %d", len(response.Services))
	}

	// Check that count matches
	if response.Count != len(response.Services) {
		t.Fatalf("Expected count=%d, got count=%d", len(response.Services), response.Count)
	}

	// Verify service structure
	for i, service := range response.Services {
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
		// Methods should be present (at least for the test data)
		if len(service.Methods) == 0 {
			t.Fatalf("Service %d has no methods", i)
		}
	}

	// Check for expected services from test data
	foundGreetingService := false
	foundUserService := false
	for _, service := range response.Services {
		switch service.Name {
		case "GreetingService":
			foundGreetingService = true
			// Check methods
			if len(service.Methods) < 4 {
				t.Fatalf("GreetingService should have at least 4 methods, got %d", len(service.Methods))
			}
		case "UserService":
			foundUserService = true
			// Check methods
			if len(service.Methods) < 3 {
				t.Fatalf("UserService should have at least 3 methods, got %d", len(service.Methods))
			}
		}
	}

	if !foundGreetingService {
		t.Fatalf("Expected to find GreetingService")
	}
	if !foundUserService {
		t.Fatalf("Expected to find UserService")
	}
}
