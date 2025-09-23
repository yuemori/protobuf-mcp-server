package mcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/yuemori/protobuf-mcp-server/internal/tools"
)

func TestMCPServer_RegisterTool(t *testing.T) {
	server := NewMCPServer()
	tool := tools.NewActivateProjectTool()

	server.RegisterTool(tool)

	// Check if tool is registered
	if _, exists := server.tools[tool.Name()]; !exists {
		t.Errorf("Tool %s not registered", tool.Name())
	}
}

func TestMCPServer_HandleToolsList(t *testing.T) {
	server := NewMCPServer()
	tool := tools.NewActivateProjectTool()
	server.RegisterTool(tool)

	request := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
		Params:  json.RawMessage("{}"),
	}

	response := server.handleRequest(context.Background(), request)

	if response.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC 2.0, got %s", response.JSONRPC)
	}

	if response.ID != 1 {
		t.Errorf("Expected ID 1, got %v", response.ID)
	}

	if response.Error != nil {
		t.Errorf("Unexpected error: %v", response.Error)
	}

	if response.Result == nil {
		t.Error("Expected result, got nil")
	}

	// Check result structure
	result, ok := response.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result to be map, got %T", response.Result)
	}

	toolsList, ok := result["tools"].([]map[string]interface{})
	if !ok {
		t.Fatalf("Expected tools to be array, got %T", result["tools"])
	}

	if len(toolsList) != 3 {
		t.Errorf("Expected 3 tools, got %d", len(toolsList))
	}

	// Check that all tools are present
	toolNames := make(map[string]bool)
	for _, toolInfo := range toolsList {
		if name, ok := toolInfo["name"].(string); ok {
			toolNames[name] = true
		}
	}

	if !toolNames["activate_project"] {
		t.Errorf("Expected activate_project tool to be present")
	}
	if !toolNames["list_services"] {
		t.Errorf("Expected list_services tool to be present")
	}
	if !toolNames["get_schema"] {
		t.Errorf("Expected get_schema tool to be present")
	}
}

func TestMCPServer_HandleToolsCall_InvalidParams(t *testing.T) {
	server := NewMCPServer()
	tool := tools.NewActivateProjectTool()
	server.RegisterTool(tool)

	request := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  json.RawMessage("invalid json"),
	}

	response := server.handleRequest(context.Background(), request)

	if response.Error == nil {
		t.Error("Expected error for invalid params")
	}

	if response.Error.Code != -32602 {
		t.Errorf("Expected error code -32602, got %d", response.Error.Code)
	}
}

func TestMCPServer_HandleToolsCall_UnknownTool(t *testing.T) {
	server := NewMCPServer()

	params := map[string]interface{}{
		"name":      "unknown_tool",
		"arguments": json.RawMessage("{}"),
	}
	paramsJSON, _ := json.Marshal(params)

	request := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  paramsJSON,
	}

	response := server.handleRequest(context.Background(), request)

	if response.Error == nil {
		t.Error("Expected error for unknown tool")
	}

	if response.Error.Code != -32601 {
		t.Errorf("Expected error code -32601, got %d", response.Error.Code)
	}
}

func TestMCPServer_HandleRequest_UnknownMethod(t *testing.T) {
	server := NewMCPServer()

	request := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "unknown/method",
		Params:  json.RawMessage("{}"),
	}

	response := server.handleRequest(context.Background(), request)

	if response.Error == nil {
		t.Error("Expected error for unknown method")
	}

	if response.Error.Code != -32601 {
		t.Errorf("Expected error code -32601, got %d", response.Error.Code)
	}
}

func TestJSONRPCRequest_Unmarshal(t *testing.T) {
	jsonData := `{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/list",
		"params": {}
	}`

	var request JSONRPCRequest
	err := json.Unmarshal([]byte(jsonData), &request)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON-RPC request: %v", err)
	}

	if request.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC 2.0, got %s", request.JSONRPC)
	}

	if request.ID != float64(1) {
		t.Errorf("Expected ID 1, got %v", request.ID)
	}

	if request.Method != "tools/list" {
		t.Errorf("Expected method tools/list, got %s", request.Method)
	}
}

func TestJSONRPCResponse_Marshal(t *testing.T) {
	response := &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      1,
		Result: map[string]interface{}{
			"success": true,
		},
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal JSON-RPC response: %v", err)
	}

	// Check that it's valid JSON
	var parsed map[string]interface{}
	err = json.Unmarshal(jsonData, &parsed)
	if err != nil {
		t.Fatalf("Generated JSON is invalid: %v", err)
	}

	// Check structure
	if parsed["jsonrpc"] != "2.0" {
		t.Errorf("Expected JSONRPC 2.0, got %v", parsed["jsonrpc"])
	}

	if parsed["id"] != float64(1) {
		t.Errorf("Expected ID 1, got %v", parsed["id"])
	}

	if parsed["result"] == nil {
		t.Error("Expected result, got nil")
	}
}

func TestJSONRPCError_Marshal(t *testing.T) {
	error := &JSONRPCError{
		Code:    -32601,
		Message: "Method not found",
		Data:    "test_method",
	}

	jsonData, err := json.Marshal(error)
	if err != nil {
		t.Fatalf("Failed to marshal JSON-RPC error: %v", err)
	}

	// Check that it's valid JSON
	var parsed map[string]interface{}
	err = json.Unmarshal(jsonData, &parsed)
	if err != nil {
		t.Fatalf("Generated JSON is invalid: %v", err)
	}

	// Check structure
	if parsed["code"] != float64(-32601) {
		t.Errorf("Expected code -32601, got %v", parsed["code"])
	}

	if parsed["message"] != "Method not found" {
		t.Errorf("Expected message 'Method not found', got %v", parsed["message"])
	}

	if parsed["data"] != "test_method" {
		t.Errorf("Expected data 'test_method', got %v", parsed["data"])
	}
}
