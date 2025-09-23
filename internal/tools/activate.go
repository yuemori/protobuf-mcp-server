package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/yuemori/protobuf-mcp-server/internal/compiler"
	"github.com/yuemori/protobuf-mcp-server/internal/config"
)

// ActivateProjectTool implements the activate_project MCP tool
type ActivateProjectTool struct{}

// NewActivateProjectTool creates a new ActivateProjectTool instance
func NewActivateProjectTool() *ActivateProjectTool {
	return &ActivateProjectTool{}
}

// Name returns the tool name
func (t *ActivateProjectTool) Name() string {
	return "activate_project"
}

// Description returns the tool description
func (t *ActivateProjectTool) Description() string {
	return "Activate a protobuf project by loading configuration and compiling proto files"
}

// ActivateProjectParams represents the parameters for activate_project
type ActivateProjectParams struct {
	ProjectPath string `json:"project_path"`
}

// ActivateProjectResponse represents the response from activate_project
type ActivateProjectResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	ProjectRoot string `json:"project_root,omitempty"`
	ProtoFiles  int    `json:"proto_files,omitempty"`
	Services    int    `json:"services,omitempty"`
	Messages    int    `json:"messages,omitempty"`
	Enums       int    `json:"enums,omitempty"`
}

// Execute executes the activate_project tool
func (t *ActivateProjectTool) Execute(ctx context.Context, params json.RawMessage) (interface{}, error) {
	// Parse parameters
	var activateParams ActivateProjectParams
	if err := json.Unmarshal(params, &activateParams); err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	// Validate project path
	if activateParams.ProjectPath == "" {
		return nil, fmt.Errorf("project_path parameter is required")
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(activateParams.ProjectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if project is initialized
	if !config.ProjectExists(absPath) {
		return &ActivateProjectResponse{
			Success: false,
			Message: "Project not initialized. Run 'go run cmd/protobuf-mcp/main.go init' first.",
		}, nil
	}

	// Load project configuration
	projectConfig, err := config.LoadProjectConfig(absPath)
	if err != nil {
		return &ActivateProjectResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to load project configuration: %v", err),
		}, nil
	}

	// Create protobuf project
	protobufProject, err := compiler.NewProtobufProject(absPath, projectConfig)
	if err != nil {
		return &ActivateProjectResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to create protobuf project: %v", err),
		}, nil
	}

	// Compile proto files
	if err := protobufProject.CompileProtos(ctx); err != nil {
		return &ActivateProjectResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to compile proto files: %v", err),
		}, nil
	}

	// Get statistics
	services, err := protobufProject.GetServices()
	if err != nil {
		return &ActivateProjectResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get services: %v", err),
		}, nil
	}

	messages, err := protobufProject.GetMessages()
	if err != nil {
		return &ActivateProjectResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get messages: %v", err),
		}, nil
	}

	enums, err := protobufProject.GetEnums()
	if err != nil {
		return &ActivateProjectResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get enums: %v", err),
		}, nil
	}

	// Count proto files
	protoFiles := 0
	if protobufProject.CompiledProtos != nil {
		protoFiles = len(protobufProject.CompiledProtos.File)
	}

	return &ActivateProjectResponse{
		Success:     true,
		Message:     "Project activated successfully",
		ProjectRoot: absPath,
		ProtoFiles:  protoFiles,
		Services:    len(services),
		Messages:    len(messages),
		Enums:       len(enums),
	}, nil
}
