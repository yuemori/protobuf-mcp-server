package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/yuemori/protobuf-mcp-server/internal/config"
	"github.com/yuemori/protobuf-mcp-server/internal/templates"
)

// OnboardingTool implements the onboarding MCP tool using mcp-go
type OnboardingTool struct {
	projectManager ProjectManagerInterface
}

// NewOnboardingTool creates a new OnboardingTool instance
func NewOnboardingTool(projectManager ProjectManagerInterface) *OnboardingTool {
	return &OnboardingTool{
		projectManager: projectManager,
	}
}

// GetTool returns the MCP tool definition
func (t *OnboardingTool) GetTool() mcp.Tool {
	return mcp.NewTool(
		"onboarding",
		mcp.WithDescription("Initialize protobuf MCP server configuration and provide setup guidance"),
		mcp.WithString("project_path",
			mcp.Required(),
			mcp.Description("Path to the protobuf project directory"),
		),
	)
}

// OnboardingParams represents the parameters for onboarding
type OnboardingParams struct {
	ProjectPath string `json:"project_path"`
}

// OnboardingResponse represents the response from onboarding
type OnboardingResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	ProjectRoot string `json:"project_root,omitempty"`
	ConfigFile  string `json:"config_file,omitempty"`
}

// Handle handles the tool execution
func (t *OnboardingTool) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	// Check if project is already initialized
	if config.ProjectExists(absPath) {
		response := &OnboardingResponse{
			Success:     false,
			Message:     "Project already initialized. Configuration file already exists.",
			ProjectRoot: absPath,
		}
		responseJSON, _ := json.Marshal(response)
		return mcp.NewToolResultText(string(responseJSON)), nil
	}

	// Create the configuration file with help comments
	defaultConfig := config.DefaultProjectConfig()
	if err := config.SaveProjectConfigWithComments(absPath, defaultConfig); err != nil {
		response := &OnboardingResponse{
			Success:     false,
			Message:     fmt.Sprintf("Failed to create configuration file: %v", err),
			ProjectRoot: absPath,
		}
		responseJSON, _ := json.Marshal(response)
		return mcp.NewToolResultText(string(responseJSON)), nil
	}

	// Generate onboarding prompt using template
	onboardingPrompt, err := templates.GetOnboardingPrompt(absPath)
	if err != nil {
		response := &OnboardingResponse{
			Success:     false,
			Message:     fmt.Sprintf("Failed to generate onboarding prompt: %v", err),
			ProjectRoot: absPath,
		}
		responseJSON, _ := json.Marshal(response)
		return mcp.NewToolResultText(string(responseJSON)), nil
	}

	response := &OnboardingResponse{
		Success:     true,
		Message:     onboardingPrompt,
		ProjectRoot: absPath,
		ConfigFile:  ".protobuf-mcp.yml",
	}
	responseJSON, _ := json.Marshal(response)
	return mcp.NewToolResultText(string(responseJSON)), nil
}
