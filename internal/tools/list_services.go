package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/protobuf/reflect/protoreflect"
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

	files, err := project.CompileProtos(ctx)
	if err != nil {
		response := &ListServicesResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to compile proto files: %v", err),
		}
		responseJSON, _ := json.Marshal(response)
		return mcp.NewToolResultText(string(responseJSON)), nil
	}

	// Get services from compiled protos
	services, err := project.GetServices(files)
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
func (t *ListServicesTool) convertServiceToInfo(service protoreflect.ServiceDescriptor) ServiceInfo {
	// Convert methods
	methods := make([]MethodInfo, 0, service.Methods().Len())

	for i := 0; i < service.Methods().Len(); i++ {
		method := service.Methods().Get(i)
		methodInfo := MethodInfo{
			Name:            string(method.Name()),
			InputType:       string(method.Input().Name()),
			OutputType:      string(method.Output().Name()),
			ClientStreaming: method.IsStreamingClient(),
			ServerStreaming: method.IsStreamingServer(),
			Description:     strings.TrimSpace(method.ParentFile().SourceLocations().ByDescriptor(method).LeadingComments),
		}
		methods = append(methods, methodInfo)
	}

	return ServiceInfo{
		Name:        string(service.Name()),
		FullName:    "." + string(service.FullName()),
		Methods:     methods,
		Package:     string(service.ParentFile().Package()),
		File:        string(service.ParentFile().Path()),
		Description: strings.TrimSpace(service.ParentFile().SourceLocations().ByDescriptor(service).LeadingComments),
	}
}
