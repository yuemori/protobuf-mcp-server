package mcp

import (
	"context"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"github.com/yuemori/protobuf-mcp-server/internal/compiler"
	"github.com/yuemori/protobuf-mcp-server/internal/tools"
)

// MCPServer represents the MCP JSON-RPC server using mcp-go
type MCPServer struct {
	server  *server.MCPServer
	project *compiler.ProtobufProject
}

// MCPProjectManager manages the current project state for MCP server
type MCPProjectManager struct {
	server  *server.MCPServer
	project *compiler.ProtobufProject
}

// SetProject sets the current project
func (pm *MCPProjectManager) SetProject(project *compiler.ProtobufProject) {
	pm.project = project
}

// GetProject returns the current project
func (pm *MCPProjectManager) GetProject() *compiler.ProtobufProject {
	return pm.project
}

// NewMCPServer creates a new MCP server instance using mcp-go
func NewMCPServer() *MCPServer {
	// Create server with tool capabilities
	s := server.NewMCPServer(
		"protobuf-mcp-server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Create project manager for state management
	projectManager := &MCPProjectManager{
		server: s,
	}

	// Create tools
	activateTool := tools.NewActivateProjectTool(projectManager)
	listServicesTool := tools.NewListServicesTool(projectManager)
	getSchemaTool := tools.NewGetSchemaTool(projectManager)

	// Register tools with the server
	s.AddTool(activateTool.GetTool(), activateTool.Handle)
	s.AddTool(listServicesTool.GetTool(), listServicesTool.Handle)
	s.AddTool(getSchemaTool.GetTool(), getSchemaTool.Handle)

	return &MCPServer{
		server: s,
	}
}

// SetProject sets the current project
func (s *MCPServer) SetProject(project *compiler.ProtobufProject) {
	s.project = project
}

// GetProject returns the current project
func (s *MCPServer) GetProject() *compiler.ProtobufProject {
	return s.project
}

// Run starts the MCP server and processes requests from stdin
func (s *MCPServer) Run(ctx context.Context) error {
	log.SetOutput(os.Stderr) // Log to stderr to avoid interfering with JSON-RPC
	log.Println("Starting Protobuf MCP Server (using mcp-go)...")

	// Run the server with stdio transport using mcp-go
	return server.ServeStdio(s.server)
}

// StartServer starts the MCP server (convenience function)
func StartServer() error {
	server := NewMCPServer()
	ctx := context.Background()
	return server.Run(ctx)
}
