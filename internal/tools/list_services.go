package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/yuemori/protobuf-mcp-server/internal/compiler"
	"google.golang.org/protobuf/types/descriptorpb"
)

// ListServicesTool implements the list_services MCP tool
type ListServicesTool struct {
	project *compiler.ProtobufProject
}

// SetProject sets the current project
func (t *ListServicesTool) SetProject(project *compiler.ProtobufProject) {
	t.project = project
}

// ListServicesParams represents the parameters for list_services tool
type ListServicesParams struct {
	// No parameters needed - uses currently activated project
}

// ServiceInfo and MethodInfo are defined in types.go

// ListServicesResponse represents the response from list_services tool
type ListServicesResponse struct {
	Success  bool          `json:"success"`
	Message  string        `json:"message"`
	Services []ServiceInfo `json:"services,omitempty"`
	Count    int           `json:"count"`
}

// Name returns the tool name
func (t *ListServicesTool) Name() string {
	return "list_services"
}

// Description returns the tool description
func (t *ListServicesTool) Description() string {
	return "List all services in the currently activated protobuf project"
}

// Execute executes the list_services tool
func (t *ListServicesTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	// Parse parameters
	var params ListServicesParams
	if len(args) > 0 {
		if err := json.Unmarshal(args, &params); err != nil {
			return &ListServicesResponse{
				Success: false,
				Message: fmt.Sprintf("Invalid parameters: %v", err),
			}, nil
		}
	}

	// Get current project from tool or global manager
	var project *compiler.ProtobufProject
	if t.project != nil {
		project = t.project
	} else {
		project = GetProjectManager().GetCurrentProject()
	}

	if project == nil {
		return &ListServicesResponse{
			Success: false,
			Message: "No project activated. Use activate_project first.",
		}, nil
	}

	// Check if project is compiled
	if project.CompiledProtos == nil {
		return &ListServicesResponse{
			Success: false,
			Message: "Project not compiled. Use activate_project first.",
		}, nil
	}

	// Get services from compiled protos
	services, err := project.GetServices()
	if err != nil {
		return &ListServicesResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get services: %v", err),
		}, nil
	}

	// Debug: Log the number of services found
	fmt.Fprintf(os.Stderr, "Debug: Found %d services in project\n", len(services))

	// Convert to ServiceInfo
	serviceInfos := make([]ServiceInfo, 0, len(services))
	for _, service := range services {
		serviceInfo := t.convertServiceToInfo(service)
		serviceInfos = append(serviceInfos, serviceInfo)
	}

	return &ListServicesResponse{
		Success:  true,
		Message:  fmt.Sprintf("Found %d services", len(serviceInfos)),
		Services: serviceInfos,
		Count:    len(serviceInfos),
	}, nil
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
