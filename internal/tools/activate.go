package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/yuemori/protobuf-mcp-server/internal/compiler"
	"github.com/yuemori/protobuf-mcp-server/internal/config"
	"github.com/yuemori/protobuf-mcp-server/internal/templates"
)

// ActivateProjectTool implements the activate_project MCP tool using mcp-go
type ActivateProjectTool struct {
	projectManager ProjectManagerInterface
}

// NewActivateProjectTool creates a new ActivateProjectTool instance
func NewActivateProjectTool(projectManager ProjectManagerInterface) *ActivateProjectTool {
	return &ActivateProjectTool{
		projectManager: projectManager,
	}
}

// GetTool returns the MCP tool definition
func (t *ActivateProjectTool) GetTool() mcp.Tool {
	return mcp.NewTool(
		"activate_project",
		mcp.WithDescription("Activate a protobuf project by loading configuration and compiling proto files"),
		mcp.WithString("project_path",
			mcp.Required(),
			mcp.Description("Path to the protobuf project directory"),
		),
	)
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
}

// Handle handles the tool execution
func (t *ActivateProjectTool) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get project path from request
	projectPath := req.GetString("project_path", "")
	if projectPath == "" {
		return mcp.NewToolResultError("project_path parameter is required"), nil
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get absolute path: %v", err)), nil
	}

	// Check if project is initialized
	if !config.ProjectExists(absPath) {
		// Return onboarding prompt when config file doesn't exist
		onboardingPrompt, err := templates.GetOnboardingPrompt(absPath)
		if err != nil {
			response := &ActivateProjectResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to generate onboarding prompt: %v", err),
			}
			responseJSON, _ := json.Marshal(response)
			return mcp.NewToolResultText(string(responseJSON)), nil
		}

		response := &ActivateProjectResponse{
			Success: false,
			Message: onboardingPrompt,
		}
		responseJSON, _ := json.Marshal(response)
		return mcp.NewToolResultText(string(responseJSON)), nil
	}

	// Load project configuration
	projectConfig, err := config.LoadProjectConfig(absPath)
	if err != nil {
		response := &ActivateProjectResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to load project configuration: %v", err),
		}
		responseJSON, _ := json.Marshal(response)
		return mcp.NewToolResultText(string(responseJSON)), nil
	}

	// Create protobuf project
	protobufProject, err := compiler.NewProtobufProject(absPath, projectConfig)
	if err != nil {
		response := &ActivateProjectResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to create protobuf project: %v", err),
		}
		responseJSON, _ := json.Marshal(response)
		return mcp.NewToolResultText(string(responseJSON)), nil
	}

	// Set as current project
	t.projectManager.SetProject(protobufProject)
	response := &ActivateProjectResponse{
		Success:     true,
		Message:     "Project activated successfully",
		ProjectRoot: absPath,
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(responseJSON)), nil
}
