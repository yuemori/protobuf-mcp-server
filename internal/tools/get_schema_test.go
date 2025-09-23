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
