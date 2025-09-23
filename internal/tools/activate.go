package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yuemori/protobuf-mcp-server/internal/compiler"
	"github.com/yuemori/protobuf-mcp-server/internal/config"
)

// ActivateProjectParams represents the parameters for the activate_project tool
type ActivateProjectParams struct {
	ProjectPath string `json:"project_path"`
}

// ActivateProjectResponse represents the response from the activate_project tool
type ActivateProjectResponse struct {
	Success       bool              `json:"success"`
	ProjectRoot   string            `json:"project_root"`
	ConfigPath    string            `json:"config_path"`
	ProtoFiles    []string          `json:"proto_files"`
	ServicesCount int               `json:"services_count"`
	MessagesCount int               `json:"messages_count"`
	EnumsCount    int               `json:"enums_count"`
	Message       string            `json:"message"`
	Error         string            `json:"error,omitempty"`
}

// ActivateProjectTool implements the activate_project MCP tool
type ActivateProjectTool struct {
	currentProject *compiler.ProtobufProject
}

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
	return "Activates a protobuf project by loading configuration and compiling proto files"
}

// Execute executes the activate_project tool
func (t *ActivateProjectTool) Execute(ctx context.Context, params ActivateProjectParams) (*ActivateProjectResponse, error) {
	// Resolve absolute path for project
	projectPath := params.ProjectPath
	if projectPath == "" {
		// Default to current directory
		cwd, err := os.Getwd()
		if err != nil {
			return &ActivateProjectResponse{
				Success: false,
				Error:   fmt.Sprintf("Failed to get current directory: %v", err),
			}, nil
		}
		projectPath = cwd
	}

	absProjectPath, err := filepath.Abs(projectPath)
	if err != nil {
		return &ActivateProjectResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to resolve absolute path for '%s': %v", projectPath, err),
		}, nil
	}

	// Check if project is initialized
	configDir := filepath.Join(absProjectPath, ".protobuf-mcp")
	configPath := filepath.Join(configDir, "project.yml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &ActivateProjectResponse{
			Success: false,
			Error:   "Project not initialized. Run 'go run cmd/protobuf-mcp/main.go init' first.",
		}, nil
	}

	// Load project configuration
	cfg, err := config.LoadProjectConfig(configPath)
	if err != nil {
		return &ActivateProjectResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to load project configuration: %v", err),
		}, nil
	}

	// Create protobuf project instance
	project, err := compiler.NewProtobufProject(absProjectPath, cfg)
	if err != nil {
		return &ActivateProjectResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create protobuf project: %v", err),
		}, nil
	}

	// Compile proto files
	if err := project.CompileProtos(ctx); err != nil {
		return &ActivateProjectResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to compile proto files: %v", err),
		}, nil
	}

	// Store current project for subsequent tool calls
	t.currentProject = project

	// Get compiled statistics
	services, err := project.GetServices()
	if err != nil {
		return &ActivateProjectResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get services: %v", err),
		}, nil
	}

	messages, err := project.GetMessages()
	if err != nil {
		return &ActivateProjectResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get messages: %v", err),
		}, nil
	}

	enums, err := project.GetEnums()
	if err != nil {
		return &ActivateProjectResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get enums: %v", err),
		}, nil
	}

	// Extract proto file names from compiled files
	var protoFiles []string
	if project.CompiledProtos != nil {
		for _, file := range project.CompiledProtos.File {
			if file.Name != nil {
				protoFiles = append(protoFiles, *file.Name)
			}
		}
	}

	return &ActivateProjectResponse{
		Success:       true,
		ProjectRoot:   absProjectPath,
		ConfigPath:    configPath,
		ProtoFiles:    protoFiles,
		ServicesCount: len(services),
		MessagesCount: len(messages),
		EnumsCount:    len(enums),
		Message:       fmt.Sprintf("Project activated successfully with %d proto files", len(protoFiles)),
	}, nil
}

// GetCurrentProject returns the currently activated project
func (t *ActivateProjectTool) GetCurrentProject() *compiler.ProtobufProject {
	return t.currentProject
}

// IsProjectActivated returns true if a project is currently activated
func (t *ActivateProjectTool) IsProjectActivated() bool {
	return t.currentProject != nil
}