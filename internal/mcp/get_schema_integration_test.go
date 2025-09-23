package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestMCPServerWithGetSchemaTool(t *testing.T) {
	// Change to project root directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Change to project root (go up from internal/mcp)
	if err := os.Chdir("../.."); err != nil {
		t.Fatalf("Failed to change to project root: %v", err)
	}

	// Build the MCP server
	buildCmd := exec.Command("go", "build", "-o", "protobuf-mcp", "./cmd/protobuf-mcp")
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build MCP server: %v\nOutput: %s", err, string(buildOutput))
	}
	defer os.Remove("protobuf-mcp")

	t.Run("Get Schema Integration", func(t *testing.T) {
		// Start the MCP server
		cmd := exec.Command("./protobuf-mcp", "server")

		stdin, err := cmd.StdinPipe()
		if err != nil {
			t.Fatalf("Failed to create stdin pipe: %v", err)
		}
		defer stdin.Close()

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			t.Fatalf("Failed to create stdout pipe: %v", err)
		}
		defer stdout.Close()

		if err := cmd.Start(); err != nil {
			t.Fatalf("Failed to start MCP server: %v", err)
		}
		defer cmd.Process.Kill()

		// Wait for server to start
		time.Sleep(100 * time.Millisecond)

		// Test 1: Initialize
		initRequest := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      1,
			Method:  "initialize",
			Params:  json.RawMessage(`{"protocolVersion":"2024-11-05","capabilities":{"tools":{}},"clientInfo":{"name":"test-client","version":"1.0.0"}}`),
		}

		if err := sendRequest(stdin, initRequest); err != nil {
			t.Fatalf("Failed to send initialize request: %v", err)
		}

		initResponse, err := readResponse(stdout)
		if err != nil {
			t.Fatalf("Failed to read initialize response: %v", err)
		}

		if initResponse.Error != nil {
			t.Fatalf("Initialize failed: %v", initResponse.Error)
		}

		// Test 2: List tools
		listRequest := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      2,
			Method:  "tools/list",
			Params:  json.RawMessage(`{}`),
		}

		if err := sendRequest(stdin, listRequest); err != nil {
			t.Fatalf("Failed to send list tools request: %v", err)
		}

		listResponse, err := readResponse(stdout)
		if err != nil {
			t.Fatalf("Failed to read list tools response: %v", err)
		}

		if listResponse.Error != nil {
			t.Fatalf("List tools failed: %v", listResponse.Error)
		}

		// Verify get_schema tool is present
		result, ok := listResponse.Result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected result to be map, got %T", listResponse.Result)
		}

		tools, ok := result["tools"].([]interface{})
		if !ok {
			t.Fatalf("Expected tools to be array, got %T", result["tools"])
		}

		if len(tools) != 3 {
			t.Errorf("Expected 3 tools, got %d", len(tools))
		}

		// Check that get_schema tool is present
		toolNames := make(map[string]bool)
		for _, tool := range tools {
			if toolMap, ok := tool.(map[string]interface{}); ok {
				if name, ok := toolMap["name"].(string); ok {
					toolNames[name] = true
				}
			}
		}

		if !toolNames["get_schema"] {
			t.Error("Expected get_schema tool to be present")
		}

		// Test 3: Activate project
		activateRequest := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      3,
			Method:  "tools/call",
			Params:  json.RawMessage(`{"name":"activate_project","arguments":{"project_path":"test-project"}}`),
		}

		if err := sendRequest(stdin, activateRequest); err != nil {
			t.Fatalf("Failed to send activate project request: %v", err)
		}

		activateResponse, err := readResponse(stdout)
		if err != nil {
			t.Fatalf("Failed to read activate project response: %v", err)
		}

		if activateResponse.Error != nil {
			t.Fatalf("Activate project failed: %v", activateResponse.Error)
		}

		// Test 4: Get schema
		getSchemaRequest := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      4,
			Method:  "tools/call",
			Params:  json.RawMessage(`{"name":"get_schema","arguments":{}}`),
		}

		if err := sendRequest(stdin, getSchemaRequest); err != nil {
			t.Fatalf("Failed to send get schema request: %v", err)
		}

		getSchemaResponse, err := readResponse(stdout)
		if err != nil {
			t.Fatalf("Failed to read get schema response: %v", err)
		}

		if getSchemaResponse.Error != nil {
			t.Fatalf("Get schema failed: %v", getSchemaResponse.Error)
		}

		// Verify schema response
		schemaResult, ok := getSchemaResponse.Result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected schema result to be map, got %T", getSchemaResponse.Result)
		}

		success, ok := schemaResult["success"].(bool)
		if !ok || !success {
			t.Errorf("Expected success=true, got %v", schemaResult["success"])
		}

		// Check that schema information is present
		if schema, ok := schemaResult["schema"].(map[string]interface{}); ok {
			// Check messages
			if messages, ok := schema["messages"].([]interface{}); ok {
				if len(messages) == 0 {
					t.Error("Expected messages to be present in schema")
				}
			}

			// Check services
			if services, ok := schema["services"].([]interface{}); ok {
				if len(services) == 0 {
					t.Error("Expected services to be present in schema")
				}
			}

			// Check enums
			if enums, ok := schema["enums"].([]interface{}); ok {
				if len(enums) == 0 {
					t.Error("Expected enums to be present in schema")
				}
			}

			// Check stats
			if stats, ok := schema["stats"].(map[string]interface{}); ok {
				if totalMessages, ok := stats["total_messages"].(float64); ok {
					if totalMessages == 0 {
						t.Error("Expected total_messages to be > 0")
					}
				}
			}
		} else {
			t.Error("Expected schema information to be present")
		}

		// Test 5: Get schema with filters
		getSchemaFilteredRequest := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      5,
			Method:  "tools/call",
			Params:  json.RawMessage(`{"name":"get_schema","arguments":{"message_types":["User"],"include_file_info":true}}`),
		}

		if err := sendRequest(stdin, getSchemaFilteredRequest); err != nil {
			t.Fatalf("Failed to send get schema filtered request: %v", err)
		}

		getSchemaFilteredResponse, err := readResponse(stdout)
		if err != nil {
			t.Fatalf("Failed to read get schema filtered response: %v", err)
		}

		if getSchemaFilteredResponse.Error != nil {
			t.Fatalf("Get schema filtered failed: %v", getSchemaFilteredResponse.Error)
		}

		// Verify filtered response
		filteredResult, ok := getSchemaFilteredResponse.Result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected filtered result to be map, got %T", getSchemaFilteredResponse.Result)
		}

		success, ok = filteredResult["success"].(bool)
		if !ok || !success {
			t.Errorf("Expected success=true for filtered request, got %v", filteredResult["success"])
		}
	})
}

// Helper functions for JSON-RPC communication
func sendRequest(stdin io.WriteCloser, request JSONRPCRequest) error {
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdin, "%s\n", string(requestBytes))
	return err
}

func readResponse(stdout io.ReadCloser) (*JSONRPCResponse, error) {
	scanner := bufio.NewScanner(stdout)
	if !scanner.Scan() {
		return nil, fmt.Errorf("no response received")
	}

	line := scanner.Text()
	var response JSONRPCResponse
	if err := json.Unmarshal([]byte(line), &response); err != nil {
		return nil, err
	}

	return &response, nil
}
