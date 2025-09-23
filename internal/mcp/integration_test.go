package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestMCPServerIntegration tests the full MCP server with real JSON-RPC communication
func TestMCPServerIntegration(t *testing.T) {
	// Get the project root directory
	projectRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Navigate to project root (assuming we're in internal/mcp)
	projectRoot = filepath.Join(projectRoot, "..", "..")

	// Build the protobuf-mcp binary
	binaryPath := filepath.Join(projectRoot, "protobuf-mcp")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "cmd/protobuf-mcp/main.go")
	buildCmd.Dir = projectRoot
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build protobuf-mcp: %v", err)
	}
	defer os.Remove(binaryPath) // Clean up binary after test

	// Test cases
	tests := []struct {
		name     string
		request  JSONRPCRequest
		validate func(t *testing.T, response JSONRPCResponse)
	}{
		{
			name: "Initialize request",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      1,
				Method:  "initialize",
				Params:  json.RawMessage(`{"protocolVersion":"2024-11-05","capabilities":{"tools":{}},"clientInfo":{"name":"test-client","version":"1.0.0"}}`),
			},
			validate: func(t *testing.T, response JSONRPCResponse) {
				if response.Error != nil {
					t.Errorf("Expected success, got error: %v", response.Error)
				}
				if response.Result == nil {
					t.Error("Expected result, got nil")
				}
			},
		},
		{
			name: "List tools",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      2,
				Method:  "tools/list",
				Params:  json.RawMessage(`{}`),
			},
			validate: func(t *testing.T, response JSONRPCResponse) {
				if response.Error != nil {
					t.Errorf("Expected success, got error: %v", response.Error)
				}
				if response.Result == nil {
					t.Error("Expected result, got nil")
				}

				// Check if tools list contains activate_project
				result, ok := response.Result.(map[string]interface{})
				if !ok {
					t.Error("Expected result to be a map")
					return
				}

				tools, ok := result["tools"].([]interface{})
				if !ok {
					t.Error("Expected tools to be an array")
					return
				}

				if len(tools) == 0 {
					t.Error("Expected at least one tool")
					return
				}

				// Check if activate_project tool is present
				found := false
				for _, tool := range tools {
					if toolMap, ok := tool.(map[string]interface{}); ok {
						if name, ok := toolMap["name"].(string); ok && name == "activate_project" {
							found = true
							break
						}
					}
				}
				if !found {
					t.Error("Expected activate_project tool to be present")
				}
			},
		},
		{
			name: "Call activate_project tool",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      3,
				Method:  "tools/call",
				Params:  json.RawMessage(`{"name":"activate_project","arguments":{"project_path":"test-project"}}`),
			},
			validate: func(t *testing.T, response JSONRPCResponse) {
				if response.Error != nil {
					t.Errorf("Expected success, got error: %v", response.Error)
				}
				if response.Result == nil {
					t.Error("Expected result, got nil")
				}

				// Check if result contains expected fields
				result, ok := response.Result.(map[string]interface{})
				if !ok {
					t.Error("Expected result to be a map")
					return
				}

				success, ok := result["success"].(bool)
				if !ok {
					t.Error("Expected success field to be boolean")
					return
				}
				if !success {
					t.Errorf("Expected success to be true, got false. Message: %v", result["message"])
				}

				// Check if project_root is set correctly
				projectRoot, ok := result["project_root"].(string)
				if !ok {
					t.Error("Expected project_root field to be string")
					return
				}
				if !strings.Contains(projectRoot, "test-project") {
					t.Errorf("Expected project_root to contain 'test-project', got: %s", projectRoot)
				}
			},
		},
	}

	// Run each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Start the MCP server
			cmd := exec.Command(binaryPath, "server")
			cmd.Dir = projectRoot

			stdin, err := cmd.StdinPipe()
			if err != nil {
				t.Fatalf("Failed to create stdin pipe: %v", err)
			}

			stdout, err := cmd.StdoutPipe()
			if err != nil {
				t.Fatalf("Failed to create stdout pipe: %v", err)
			}

			// Start the server
			if err := cmd.Start(); err != nil {
				t.Fatalf("Failed to start MCP server: %v", err)
			}

			// Create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			// Send request
			requestBytes, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			// Send the request
			go func() {
				defer stdin.Close()
				fmt.Fprintf(stdin, "%s\n", string(requestBytes))
			}()

			// Read response
			responseChan := make(chan JSONRPCResponse, 1)
			errorChan := make(chan error, 1)

			go func() {
				defer stdout.Close()
				responseBytes, err := io.ReadAll(stdout)
				if err != nil {
					errorChan <- err
					return
				}

				var response JSONRPCResponse
				if err := json.Unmarshal(responseBytes, &response); err != nil {
					errorChan <- err
					return
				}

				responseChan <- response
			}()

			// Wait for response or timeout
			select {
			case response := <-responseChan:
				// Validate the response
				tt.validate(t, response)
			case err := <-errorChan:
				t.Fatalf("Failed to read response: %v", err)
			case <-ctx.Done():
				t.Fatalf("Test timed out")
			}

			// Clean up
			cmd.Process.Kill()
			cmd.Wait()
		})
	}
}

// TestMCPServerWithTestProject tests the MCP server specifically with test-project
func TestMCPServerWithTestProject(t *testing.T) {
	// Get the project root directory
	projectRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Navigate to project root
	projectRoot = filepath.Join(projectRoot, "..", "..")

	// Check if test-project exists
	testProjectPath := filepath.Join(projectRoot, "test-project")
	if _, err := os.Stat(testProjectPath); os.IsNotExist(err) {
		t.Skip("test-project directory not found, skipping integration test")
	}

	// Build the protobuf-mcp binary
	binaryPath := filepath.Join(projectRoot, "protobuf-mcp")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "cmd/protobuf-mcp/main.go")
	buildCmd.Dir = projectRoot
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build protobuf-mcp: %v", err)
	}
	defer os.Remove(binaryPath)

	// Test the full workflow
	t.Run("Full MCP Workflow", func(t *testing.T) {
		// Start the MCP server
		cmd := exec.Command(binaryPath, "server")
		cmd.Dir = projectRoot

		stdin, err := cmd.StdinPipe()
		if err != nil {
			t.Fatalf("Failed to create stdin pipe: %v", err)
		}

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			t.Fatalf("Failed to create stdout pipe: %v", err)
		}

		// Start the server
		if err := cmd.Start(); err != nil {
			t.Fatalf("Failed to start MCP server: %v", err)
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test sequence: initialize -> tools/list -> tools/call
		testSequence := []JSONRPCRequest{
			{
				JSONRPC: "2.0",
				ID:      1,
				Method:  "initialize",
				Params:  json.RawMessage(`{"protocolVersion":"2024-11-05","capabilities":{"tools":{}},"clientInfo":{"name":"integration-test","version":"1.0.0"}}`),
			},
			{
				JSONRPC: "2.0",
				ID:      2,
				Method:  "tools/list",
				Params:  json.RawMessage(`{}`),
			},
			{
				JSONRPC: "2.0",
				ID:      3,
				Method:  "tools/call",
				Params:  json.RawMessage(`{"name":"activate_project","arguments":{"project_path":"test-project"}}`),
			},
		}

		// Send all requests
		go func() {
			defer stdin.Close()
			for _, request := range testSequence {
				requestBytes, err := json.Marshal(request)
				if err != nil {
					t.Errorf("Failed to marshal request: %v", err)
					return
				}
				fmt.Fprintf(stdin, "%s\n", string(requestBytes))
				time.Sleep(100 * time.Millisecond) // Small delay between requests
			}
		}()

		// Read all responses
		responseCount := 0
		scanner := func() {
			defer stdout.Close()
			responseBytes, err := io.ReadAll(stdout)
			if err != nil {
				t.Errorf("Failed to read response: %v", err)
				return
			}

			// Split by newlines and process each response
			lines := strings.Split(string(responseBytes), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}

				var response JSONRPCResponse
				if err := json.Unmarshal([]byte(line), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
					continue
				}

				responseCount++

				// Validate each response
				switch response.ID {
				case 1: // initialize
					if response.Error != nil {
						t.Errorf("Initialize failed: %v", response.Error)
					}
				case 2: // tools/list
					if response.Error != nil {
						t.Errorf("Tools list failed: %v", response.Error)
					}
				case 3: // tools/call
					if response.Error != nil {
						t.Errorf("Activate project failed: %v", response.Error)
					}
					if response.Result != nil {
						if result, ok := response.Result.(map[string]interface{}); ok {
							if success, ok := result["success"].(bool); ok && success {
								t.Logf("Successfully activated project: %v", result["project_root"])
							}
						}
					}
				}
			}
		}

		// Run scanner in goroutine
		go scanner()

		// Wait for completion or timeout
		select {
		case <-ctx.Done():
			t.Fatalf("Test timed out")
		default:
			// Wait a bit for responses
			time.Sleep(2 * time.Second)
		}

		// Clean up
		cmd.Process.Kill()
		cmd.Wait()

		// Verify we got responses
		if responseCount == 0 {
			t.Error("Expected at least one response, got none")
		}
	})
}
