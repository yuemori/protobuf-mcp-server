package mcp

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestServerInitialization_Integration tests server initialization and basic capabilities
func TestServerInitialization_Integration(t *testing.T) {
	// Create server
	server := NewMCPServer()

	// Create in-process client
	mcpClient, err := client.NewInProcessClient(server.server)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test server initialization
	t.Run("Initialize", func(t *testing.T) {
		initResult, err := mcpClient.Initialize(ctx, mcp.InitializeRequest{
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

		// Verify server info
		if initResult.ServerInfo.Name != "protobuf-mcp-server" {
			t.Errorf("Expected server name 'protobuf-mcp-server', got '%s'", initResult.ServerInfo.Name)
		}

		if initResult.ServerInfo.Version != "1.0.0" {
			t.Errorf("Expected server version '1.0.0', got '%s'", initResult.ServerInfo.Version)
		}

		// Verify capabilities
		if initResult.Capabilities.Tools == nil {
			t.Error("Expected tools capability to be enabled")
		}
	})

	// Test listing tools
	t.Run("ListTools", func(t *testing.T) {
		// Initialize first
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

		// List tools
		toolsResult, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
		if err != nil {
			t.Fatalf("Failed to list tools: %v", err)
		}

		// Verify we have the expected tools
		expectedTools := map[string]bool{
			"activate_project": false,
			"list_services":    false,
			"get_schema":       false,
		}

		for _, tool := range toolsResult.Tools {
			if _, expected := expectedTools[tool.Name]; expected {
				expectedTools[tool.Name] = true
			}
		}

		for name, found := range expectedTools {
			if !found {
				t.Errorf("Expected tool '%s' not found", name)
			}
		}

		// Verify tool descriptions
		for _, tool := range toolsResult.Tools {
			switch tool.Name {
			case "activate_project":
				if tool.Description == "" {
					t.Error("Expected activate_project tool to have description")
				}
			case "list_services":
				if tool.Description == "" {
					t.Error("Expected list_services tool to have description")
				}
			case "get_schema":
				if tool.Description == "" {
					t.Error("Expected get_schema tool to have description")
				}
			}
		}
	})

	// Test server capabilities
	t.Run("ServerCapabilities", func(t *testing.T) {
		// Initialize first
		initResult, err := mcpClient.Initialize(ctx, mcp.InitializeRequest{
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

		// Verify server info from initialization result
		if initResult.ServerInfo.Name != "protobuf-mcp-server" {
			t.Errorf("Expected server name 'protobuf-mcp-server', got '%s'", initResult.ServerInfo.Name)
		}

		if initResult.ServerInfo.Version != "1.0.0" {
			t.Errorf("Expected server version '1.0.0', got '%s'", initResult.ServerInfo.Version)
		}

		// Verify capabilities
		if initResult.Capabilities.Tools == nil {
			t.Error("Expected tools capability to be enabled")
		}
	})
}
