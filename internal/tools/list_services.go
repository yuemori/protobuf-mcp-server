package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/protobuf/types/descriptorpb"
)

// ListServicesTool implements the list_services MCP tool using mcp-go
type ListServicesTool struct {
	projectManager ProjectManagerInterface
}

// NewListServicesTool creates a new ListServicesTool instance
func NewListServicesTool(projectManager ProjectManagerInterface) *ListServicesTool {
	return &ListServicesTool{
		projectManager: projectManager,
	}
}

// GetTool returns the MCP tool definition
func (t *ListServicesTool) GetTool() mcp.Tool {
	return mcp.NewTool(
		"list_services",
		mcp.WithDescription("List all services in the currently activated protobuf project"),
	)
}

// ListServicesResponse represents the response from list_services tool
type ListServicesResponse struct {
	Success  bool          `json:"success"`
	Message  string        `json:"message"`
	Services []ServiceInfo `json:"services,omitempty"`
	Count    int           `json:"count"`
}

// Handle handles the tool execution
func (t *ListServicesTool) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get current project
	project := t.projectManager.GetProject()
	if project == nil {
		response := &ListServicesResponse{
			Success: false,
			Message: "No project activated. Use activate_project first.",
		}
		responseJSON, _ := json.Marshal(response)
		return mcp.NewToolResultText(string(responseJSON)), nil
	}

	// Check if project is compiled
	if project.CompiledProtos == nil {
		response := &ListServicesResponse{
			Success: false,
			Message: "Project not compiled. Use activate_project first.",
		}
		responseJSON, _ := json.Marshal(response)
		return mcp.NewToolResultText(string(responseJSON)), nil
	}

	// Get services from compiled protos
	services, err := project.GetServices()
	if err != nil {
		response := &ListServicesResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get services: %v", err),
		}
		responseJSON, _ := json.Marshal(response)
		return mcp.NewToolResultText(string(responseJSON)), nil
	}

	// Convert to ServiceInfo
	serviceInfos := make([]ServiceInfo, 0, len(services))
	for _, service := range services {
		serviceInfo := t.convertServiceToInfo(service)
		serviceInfos = append(serviceInfos, serviceInfo)
	}

	response := &ListServicesResponse{
		Success:  true,
		Message:  fmt.Sprintf("Found %d services", len(serviceInfos)),
		Services: serviceInfos,
		Count:    len(serviceInfos),
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(responseJSON)), nil
}

// convertServiceToInfo converts a protobuf service to ServiceInfo
func (t *ListServicesTool) convertServiceToInfo(service *descriptorpb.ServiceDescriptorProto) ServiceInfo {
	// Convert methods
	methods := make([]MethodInfo, 0, len(service.Method))
	for _, method := range service.Method {
		methodInfo := MethodInfo{
			Name:            method.GetName(),
			InputType:       method.GetInputType(),
			OutputType:      method.GetOutputType(),
			ClientStreaming: method.GetClientStreaming(),
			ServerStreaming: method.GetServerStreaming(),
			Description:     "", // TODO: Extract description from method options
		}
		methods = append(methods, methodInfo)
	}

	return ServiceInfo{
		Name:        service.GetName(),
		FullName:    service.GetName(), // TODO: Add package prefix if needed
		Methods:     methods,
		File:        "", // TODO: Add file information if available
		Description: "", // TODO: Extract description from service options
	}
}
