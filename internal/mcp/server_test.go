package mcp

import (
	"testing"

	"github.com/yuemori/protobuf-mcp-server/internal/compiler"
	"github.com/yuemori/protobuf-mcp-server/internal/tools"
)

// MockProjectManager for testing
type MockProjectManager struct {
	project *compiler.ProtobufProject
}

func (m *MockProjectManager) SetProject(project *compiler.ProtobufProject) {
	m.project = project
}

func (m *MockProjectManager) GetProject() *compiler.ProtobufProject {
	return m.project
}

func TestNewMCPServer(t *testing.T) {
	server := NewMCPServer()
	if server == nil {
		t.Fatal("Expected server to be created")
	}

	if server.server == nil {
		t.Fatal("Expected internal mcp-go server to be initialized")
	}
}

func TestMCPServer_ToolRegistration(t *testing.T) {
	server := NewMCPServer()

	// The tools should be registered during server creation
	// We can't directly inspect the registered tools in mcp-go,
	// but we can verify the server was created successfully
	if server.server == nil {
		t.Fatal("Expected server to have tools registered")
	}
}

func TestMCPProjectManager_SetAndGetProject(t *testing.T) {
	manager := &MCPProjectManager{}

	// Test nil project
	if manager.GetProject() != nil {
		t.Fatal("Expected nil project initially")
	}

	// Create a mock project
	project := &compiler.ProtobufProject{
		ProjectRoot: "/test/path",
	}

	// Set project
	manager.SetProject(project)

	// Get project
	retrieved := manager.GetProject()
	if retrieved != project {
		t.Fatal("Expected to retrieve the same project")
	}

	if retrieved.ProjectRoot != "/test/path" {
		t.Fatalf("Expected project root to be /test/path, got %s", retrieved.ProjectRoot)
	}
}

func TestActivateProjectTool_Creation(t *testing.T) {
	// This is a simple test to verify tool creation
	manager := &MockProjectManager{}
	tool := tools.NewActivateProjectTool(manager)

	if tool == nil {
		t.Fatal("Expected tool to be created")
	}

	mcpTool := tool.GetTool()
	if mcpTool.Name != "activate_project" {
		t.Fatalf("Expected tool name to be 'activate_project', got '%s'", mcpTool.Name)
	}
}

func TestListServicesTool_Creation(t *testing.T) {
	// This is a simple test to verify tool creation
	manager := &MockProjectManager{}
	tool := tools.NewListServicesTool(manager)

	if tool == nil {
		t.Fatal("Expected tool to be created")
	}

	mcpTool := tool.GetTool()
	if mcpTool.Name != "list_services" {
		t.Fatalf("Expected tool name to be 'list_services', got '%s'", mcpTool.Name)
	}
}

func TestGetSchemaTool_Creation(t *testing.T) {
	// This is a simple test to verify tool creation
	manager := &MockProjectManager{}
	tool := tools.NewGetSchemaTool(manager)

	if tool == nil {
		t.Fatal("Expected tool to be created")
	}

	mcpTool := tool.GetTool()
	if mcpTool.Name != "get_schema" {
		t.Fatalf("Expected tool name to be 'get_schema', got '%s'", mcpTool.Name)
	}
}

// Test that tools can handle requests through the server's project manager
func TestToolsWithServerProjectManager(t *testing.T) {
	// Create server which internally sets up project manager
	server := NewMCPServer()

	if server == nil {
		t.Fatal("Expected server to be created")
	}

	// The server should have tools registered and ready to handle requests
	// In actual usage, the tools will access the project manager through their own reference
	if server.server == nil {
		t.Fatal("Expected mcp-go server to be initialized")
	}
}
