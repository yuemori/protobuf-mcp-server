package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/yuemori/protobuf-mcp-server/internal/compiler"
	"github.com/yuemori/protobuf-mcp-server/internal/tools"
)

// MCPServer represents the MCP JSON-RPC server
type MCPServer struct {
	tools   map[string]Tool
	project *compiler.ProtobufProject
}

// Tool represents an MCP tool interface
type Tool interface {
	Name() string
	Description() string
	Execute(ctx context.Context, params json.RawMessage) (interface{}, error)
}

// ProjectAwareTool represents a tool that can set the current project
type ProjectAwareTool interface {
	Tool
	SetProject(project *compiler.ProtobufProject)
}

// JSONRPCRequest represents a JSON-RPC request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

// JSONRPCResponse represents a JSON-RPC response
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer() *MCPServer {
	server := &MCPServer{
		tools: make(map[string]Tool),
	}

	// Create and register tools
	activateTool := tools.NewActivateProjectTool()
	listServicesTool := &tools.ListServicesTool{}

	// Set server reference in activate tool
	activateTool.SetServer(server)

	// Register tools
	server.RegisterTool(activateTool)
	server.RegisterTool(listServicesTool)

	return server
}

// RegisterTool registers a tool with the server
func (s *MCPServer) RegisterTool(tool Tool) {
	s.tools[tool.Name()] = tool

	// If tool is project-aware, set the current project
	if projectAwareTool, ok := tool.(ProjectAwareTool); ok {
		projectAwareTool.SetProject(s.project)
	}
}

// SetProject sets the current project
func (s *MCPServer) SetProject(project *compiler.ProtobufProject) {
	s.project = project

	// Update all project-aware tools
	for _, tool := range s.tools {
		if projectAwareTool, ok := tool.(ProjectAwareTool); ok {
			projectAwareTool.SetProject(project)
		}
	}
}

// GetProject returns the current project
func (s *MCPServer) GetProject() *compiler.ProtobufProject {
	return s.project
}

// Run starts the MCP server and processes JSON-RPC requests from stdin
func (s *MCPServer) Run(ctx context.Context) error {
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			var request JSONRPCRequest
			if err := decoder.Decode(&request); err != nil {
				if err == io.EOF {
					return nil
				}
				return fmt.Errorf("failed to decode JSON-RPC request: %w", err)
			}

			response := s.handleRequest(ctx, &request)
			if err := encoder.Encode(response); err != nil {
				return fmt.Errorf("failed to encode JSON-RPC response: %w", err)
			}
		}
	}
}

// handleRequest processes a JSON-RPC request
func (s *MCPServer) handleRequest(ctx context.Context, request *JSONRPCRequest) *JSONRPCResponse {
	// Handle different methods
	switch request.Method {
	case "initialize":
		return s.handleInitialize(request)
	case "tools/list":
		return s.handleToolsList(request)
	case "tools/call":
		return s.handleToolsCall(ctx, request)
	default:
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &JSONRPCError{
				Code:    -32601,
				Message: "Method not found",
			},
		}
	}
}

// handleInitialize handles the initialize method
func (s *MCPServer) handleInitialize(request *JSONRPCRequest) *JSONRPCResponse {
	// Parse the initialize parameters
	var params struct {
		ProtocolVersion string `json:"protocolVersion"`
		Capabilities    struct {
			Tools map[string]interface{} `json:"tools"`
		} `json:"capabilities"`
		ClientInfo struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"clientInfo"`
	}

	if err := json.Unmarshal(request.Params, &params); err != nil {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &JSONRPCError{
				Code:    -32602,
				Message: "Invalid params",
				Data:    err.Error(),
			},
		}
	}

	// Return server capabilities
	result := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    "protobuf-mcp-server",
			"version": "1.0.0",
		},
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}
}

// handleToolsList returns the list of available tools
func (s *MCPServer) handleToolsList(request *JSONRPCRequest) *JSONRPCResponse {
	var tools []map[string]interface{}

	for _, tool := range s.tools {
		tools = append(tools, map[string]interface{}{
			"name":        tool.Name(),
			"description": tool.Description(),
		})
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result: map[string]interface{}{
			"tools": tools,
		},
	}
}

// handleToolsCall executes a tool
func (s *MCPServer) handleToolsCall(ctx context.Context, request *JSONRPCRequest) *JSONRPCResponse {
	// Parse tool call parameters
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}

	if err := json.Unmarshal(request.Params, &params); err != nil {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &JSONRPCError{
				Code:    -32602,
				Message: "Invalid params",
				Data:    err.Error(),
			},
		}
	}

	// Find the tool
	tool, exists := s.tools[params.Name]
	if !exists {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &JSONRPCError{
				Code:    -32601,
				Message: "Tool not found",
				Data:    params.Name,
			},
		}
	}

	// Execute the tool
	result, err := tool.Execute(ctx, params.Arguments)
	if err != nil {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &JSONRPCError{
				Code:    -32603,
				Message: "Internal error",
				Data:    err.Error(),
			},
		}
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}
}

// StartServer starts the MCP server (convenience function)
func StartServer() error {
	server := NewMCPServer()
	ctx := context.Background()

	log.SetOutput(os.Stderr) // Log to stderr to avoid interfering with JSON-RPC
	log.Println("Starting Protobuf MCP Server...")

	return server.Run(ctx)
}
