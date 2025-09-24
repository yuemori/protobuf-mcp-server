package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bufbuild/protocompile/linker"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/yuemori/protobuf-mcp-server/internal/compiler"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// GetSchemaTool implements the get_schema MCP tool using mcp-go
type GetSchemaTool struct {
	projectManager ProjectManagerInterface
}

// NewGetSchemaTool creates a new GetSchemaTool instance
func NewGetSchemaTool(projectManager ProjectManagerInterface) *GetSchemaTool {
	return &GetSchemaTool{
		projectManager: projectManager,
	}
}

// GetTool returns the MCP tool definition
func (t *GetSchemaTool) GetTool() mcp.Tool {
	return mcp.NewTool(
		"get_schema",
		mcp.WithDescription("Get detailed schema information from the activated protobuf project"),
		mcp.WithString("name",
			mcp.Description("Filter by name (searches both Name and FullName)"),
		),
		mcp.WithString("type",
			mcp.Description("Filter by type: 'service', 'enum', or 'message'"),
		),
	)
}

// GetSchemaParams represents the parameters for get_schema tool
type GetSchemaParams struct {
	// Name filter (searches both Name and FullName)
	Name string `json:"name,omitempty"`
	// Type filter: "service", "enum", or "message"
	Type string `json:"type,omitempty"`
}

// GetSchemaResponse represents the response for get_schema tool
type GetSchemaResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Schema  *SchemaInfo `json:"schema,omitempty"`
	Count   int         `json:"count"`
}

// SchemaInfo represents detailed schema information
type SchemaInfo struct {
	Messages []MessageInfo `json:"messages"`
	Services []ServiceInfo `json:"services"`
	Enums    []EnumInfo    `json:"enums"`
}

// MessageInfo represents detailed information about a protobuf message
type MessageInfo struct {
	Name        string       `json:"name"`
	FullName    string       `json:"full_name"`
	Fields      []FieldInfo  `json:"fields"`
	File        string       `json:"file"`
	Package     string       `json:"package"`
	Description string       `json:"description"`
	Options     []OptionInfo `json:"options,omitempty"`
}

// FieldInfo represents information about a message field
type FieldInfo struct {
	Name        string       `json:"name"`
	Number      int32        `json:"number"`
	Type        string       `json:"type"`
	Optional    bool         `json:"optional"`
	Repeated    bool         `json:"repeated"`
	Description string       `json:"description"`
	Options     []OptionInfo `json:"options,omitempty"`
}

// ServiceInfo and MethodInfo are defined in types.go

// EnumInfo represents detailed information about a protobuf enum
type EnumInfo struct {
	Name        string          `json:"name"`
	FullName    string          `json:"full_name"`
	Values      []EnumValueInfo `json:"values"`
	File        string          `json:"file"`
	Package     string          `json:"package"`
	Description string          `json:"description"`
	Options     []OptionInfo    `json:"options,omitempty"`
}

// EnumValueInfo represents information about an enum value
type EnumValueInfo struct {
	Name        string       `json:"name"`
	Number      int32        `json:"number"`
	Description string       `json:"description"`
	Options     []OptionInfo `json:"options,omitempty"`
}

// OptionInfo is defined in types.go

// Handle handles the tool execution
func (t *GetSchemaTool) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Parse parameters
	var params GetSchemaParams
	params.Name = req.GetString("name", "")
	params.Type = req.GetString("type", "")

	// Get current project
	project := t.projectManager.GetProject()
	if project == nil {
		response := &GetSchemaResponse{
			Success: false,
			Message: "No project activated. Use activate_project first.",
		}
		responseJSON, _ := json.Marshal(response)
		return mcp.NewToolResultText(string(responseJSON)), nil
	}

	files, err := project.CompileProtos(ctx)
	if err != nil {
		response := &GetSchemaResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to compile proto files: %v", err),
		}
		responseJSON, _ := json.Marshal(response)
		return mcp.NewToolResultText(string(responseJSON)), nil
	}

	// Get schema information
	schemaInfo, err := t.buildSchemaInfo(project, files, &params)
	if err != nil {
		response := &GetSchemaResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to build schema info: %v", err),
		}
		responseJSON, _ := json.Marshal(response)
		return mcp.NewToolResultText(string(responseJSON)), nil
	}

	// Calculate count
	count := len(schemaInfo.Messages) + len(schemaInfo.Services) + len(schemaInfo.Enums)

	response := &GetSchemaResponse{
		Success: true,
		Message: fmt.Sprintf("Retrieved schema information: %d messages, %d services, %d enums",
			len(schemaInfo.Messages), len(schemaInfo.Services), len(schemaInfo.Enums)),
		Schema: schemaInfo,
		Count:  count,
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(responseJSON)), nil
}

// matchesName checks if a name matches the search criteria
func (t *GetSchemaTool) matchesName(name, fullName, searchName string) bool {
	if searchName == "" {
		return true
	}
	searchName = strings.ToLower(searchName)
	return strings.Contains(strings.ToLower(name), searchName) ||
		strings.Contains(strings.ToLower(fullName), searchName)
}

// matchesType checks if a type matches the filter criteria
func (t *GetSchemaTool) matchesType(actualType, filterType string) bool {
	if filterType == "" {
		return true
	}
	return actualType == filterType
}

// buildSchemaInfo builds detailed schema information from the compiled project
func (t *GetSchemaTool) buildSchemaInfo(project *compiler.ProtobufProject, files linker.Files, params *GetSchemaParams) (*SchemaInfo, error) {
	schemaInfo := &SchemaInfo{
		Messages: []MessageInfo{},
		Services: []ServiceInfo{},
		Enums:    []EnumInfo{},
	}

	// Process each file in the compiled protos
	for _, file := range files {
		// Process messages
		if t.matchesType("message", params.Type) {
			for i := 0; i < file.Messages().Len(); i++ {
				message := file.Messages().Get(i)
				if t.matchesName(string(message.Name()), string(message.FullName()), params.Name) {
					messageInfo := t.convertMessageToInfo(message)
					schemaInfo.Messages = append(schemaInfo.Messages, messageInfo)
				}
			}
		}

		// Process services
		if t.matchesType("service", params.Type) {
			for i := 0; i < file.Services().Len(); i++ {
				service := file.Services().Get(i)
				if t.matchesName(string(service.Name()), string(service.FullName()), params.Name) {
					serviceInfo := t.convertServiceToInfo(service)
					schemaInfo.Services = append(schemaInfo.Services, serviceInfo)
				}
			}
		}

		// Process enums
		if t.matchesType("enum", params.Type) {
			for i := 0; i < file.Enums().Len(); i++ {
				enum := file.Enums().Get(i)
				if t.matchesName(string(enum.Name()), string(enum.FullName()), params.Name) {
					enumInfo := t.convertEnumToInfo(enum)
					schemaInfo.Enums = append(schemaInfo.Enums, enumInfo)
				}
			}
		}
	}

	return schemaInfo, nil
}

// convertMessageToInfo converts a protobuf message to MessageInfo
func (t *GetSchemaTool) convertMessageToInfo(message protoreflect.MessageDescriptor) MessageInfo {
	// Convert fields
	fields := make([]FieldInfo, 0, message.Fields().Len())
	for i := 0; i < message.Fields().Len(); i++ {
		field := message.Fields().Get(i)
		fieldInfo := FieldInfo{
			Name:        string(field.Name()),
			Number:      int32(field.Number()),
			Type:        field.Kind().String(),
			Optional:    field.HasPresence(),
			Repeated:    field.Cardinality() == protoreflect.Repeated,
			Description: strings.TrimSpace(message.ParentFile().SourceLocations().ByDescriptor(field).LeadingComments),
			Options:     []OptionInfo{}, // TODO: Extract field options
		}
		fields = append(fields, fieldInfo)
	}

	return MessageInfo{
		Name:        string(message.Name()),
		FullName:    string(message.FullName()),
		Fields:      fields,
		File:        string(message.ParentFile().Path()),
		Package:     string(message.ParentFile().Package()),
		Description: strings.TrimSpace(message.ParentFile().SourceLocations().ByDescriptor(message).LeadingComments),
		Options:     []OptionInfo{}, // TODO: Extract message options
	}
}

// convertServiceToInfo converts a protobuf service to ServiceInfo
func (t *GetSchemaTool) convertServiceToInfo(service protoreflect.ServiceDescriptor) ServiceInfo {
	// Convert methods
	methods := make([]MethodInfo, 0, service.Methods().Len())
	for i := 0; i < service.Methods().Len(); i++ {
		method := service.Methods().Get(i)
		methodInfo := MethodInfo{
			Name:            string(method.Name()),
			InputType:       string(method.Input().FullName()),
			OutputType:      string(method.Output().FullName()),
			ClientStreaming: method.IsStreamingClient(),
			ServerStreaming: method.IsStreamingServer(),
			Description:     strings.TrimSpace(service.ParentFile().SourceLocations().ByDescriptor(method).LeadingComments),
			Options:         []OptionInfo{}, // TODO: Extract method options
		}
		methods = append(methods, methodInfo)
	}

	return ServiceInfo{
		Name:        string(service.Name()),
		FullName:    string(service.FullName()),
		Methods:     methods,
		File:        string(service.ParentFile().Path()),
		Package:     string(service.ParentFile().Package()),
		Description: strings.TrimSpace(service.ParentFile().SourceLocations().ByDescriptor(service).LeadingComments),
		Options:     []OptionInfo{}, // TODO: Extract service options
	}
}

// convertEnumToInfo converts a protobuf enum to EnumInfo
func (t *GetSchemaTool) convertEnumToInfo(enum protoreflect.EnumDescriptor) EnumInfo {
	// Convert enum values
	values := make([]EnumValueInfo, 0, enum.Values().Len())
	for i := 0; i < enum.Values().Len(); i++ {
		value := enum.Values().Get(i)
		valueInfo := EnumValueInfo{
			Name:        string(value.Name()),
			Number:      int32(value.Number()),
			Description: strings.TrimSpace(enum.ParentFile().SourceLocations().ByDescriptor(value).LeadingComments),
			Options:     []OptionInfo{}, // TODO: Extract value options
		}
		values = append(values, valueInfo)
	}

	return EnumInfo{
		Name:        string(enum.Name()),
		FullName:    string(enum.FullName()),
		Values:      values,
		File:        string(enum.ParentFile().Path()),
		Package:     string(enum.ParentFile().Package()),
		Description: strings.TrimSpace(enum.ParentFile().SourceLocations().ByDescriptor(enum).LeadingComments),
		Options:     []OptionInfo{}, // TODO: Extract enum options
	}
}
