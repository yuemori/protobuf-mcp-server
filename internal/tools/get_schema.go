package tools

import (
	"context"
	"encoding/json"
	"fmt"

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
		mcp.WithArray("message_types",
			mcp.Description("Filter by specific message types"),
			mcp.Items(map[string]any{"type": "string"}),
		),
		mcp.WithArray("service_types",
			mcp.Description("Filter by specific service types"),
			mcp.Items(map[string]any{"type": "string"}),
		),
		mcp.WithArray("enum_types",
			mcp.Description("Filter by specific enum types"),
			mcp.Items(map[string]any{"type": "string"}),
		),
		mcp.WithBoolean("include_file_info",
			mcp.DefaultBool(false),
			mcp.Description("Include file information in the response"),
		),
	)
}

// GetSchemaParams represents the parameters for get_schema tool
type GetSchemaParams struct {
	// Optional filters for specific types
	MessageTypes []string `json:"message_types,omitempty"`
	ServiceTypes []string `json:"service_types,omitempty"`
	EnumTypes    []string `json:"enum_types,omitempty"`
	// Include file information
	IncludeFileInfo bool `json:"include_file_info,omitempty"`
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
	Files    []FileInfo    `json:"files,omitempty"`
	Stats    StatsInfo     `json:"stats"`
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

// FileInfo represents information about a protobuf file
type FileInfo struct {
	Name         string       `json:"name"`
	Package      string       `json:"package"`
	Dependencies []string     `json:"dependencies"`
	Description  string       `json:"description"`
	Options      []OptionInfo `json:"options,omitempty"`
}

// OptionInfo is defined in types.go

// StatsInfo represents schema statistics
type StatsInfo struct {
	TotalMessages int `json:"total_messages"`
	TotalServices int `json:"total_services"`
	TotalEnums    int `json:"total_enums"`
	TotalFiles    int `json:"total_files"`
	TotalFields   int `json:"total_fields"`
	TotalMethods  int `json:"total_methods"`
	TotalValues   int `json:"total_values"`
}

// Handle handles the tool execution
func (t *GetSchemaTool) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Parse parameters
	var params GetSchemaParams
	params.MessageTypes = t.getStringArray(req, "message_types")
	params.ServiceTypes = t.getStringArray(req, "service_types")
	params.EnumTypes = t.getStringArray(req, "enum_types")
	params.IncludeFileInfo = req.GetBool("include_file_info", false)

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

	// Check if project is compiled
	if project.CompiledProtos == nil {
		response := &GetSchemaResponse{
			Success: false,
			Message: "Project not compiled. Use activate_project first.",
		}
		responseJSON, _ := json.Marshal(response)
		return mcp.NewToolResultText(string(responseJSON)), nil
	}

	// Get schema information
	schemaInfo, err := t.buildSchemaInfo(project, &params)
	if err != nil {
		response := &GetSchemaResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to build schema info: %v", err),
		}
		responseJSON, _ := json.Marshal(response)
		return mcp.NewToolResultText(string(responseJSON)), nil
	}

	// Calculate count
	count := schemaInfo.Stats.TotalMessages + schemaInfo.Stats.TotalServices + schemaInfo.Stats.TotalEnums

	response := &GetSchemaResponse{
		Success: true,
		Message: fmt.Sprintf("Retrieved schema information: %d messages, %d services, %d enums",
			schemaInfo.Stats.TotalMessages, schemaInfo.Stats.TotalServices, schemaInfo.Stats.TotalEnums),
		Schema: schemaInfo,
		Count:  count,
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(responseJSON)), nil
}

// getStringArray extracts a string array from the request
func (t *GetSchemaTool) getStringArray(req mcp.CallToolRequest, key string) []string {
	args := req.GetArguments()
	if args == nil {
		return []string{}
	}

	if arr, ok := args[key].([]interface{}); ok {
		result := make([]string, 0, len(arr))
		for _, item := range arr {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	}

	return []string{}
}

// buildSchemaInfo builds detailed schema information from the compiled project
func (t *GetSchemaTool) buildSchemaInfo(project *compiler.ProtobufProject, params *GetSchemaParams) (*SchemaInfo, error) {
	schemaInfo := &SchemaInfo{
		Messages: []MessageInfo{},
		Services: []ServiceInfo{},
		Enums:    []EnumInfo{},
		Files:    []FileInfo{},
		Stats:    StatsInfo{},
	}

	// Process each file in the compiled protos
	for _, file := range project.CompiledProtos {
		// Add file information if requested
		if params.IncludeFileInfo {
			fileInfo := t.convertFileToInfo(file)
			schemaInfo.Files = append(schemaInfo.Files, fileInfo)
			schemaInfo.Stats.TotalFiles++
		}

		// Process messages
		for i := 0; i < file.Messages().Len(); i++ {
			message := file.Messages().Get(i)
			if t.shouldIncludeMessage(message, params.MessageTypes) {
				messageInfo := t.convertMessageToInfo(message, file)
				schemaInfo.Messages = append(schemaInfo.Messages, messageInfo)
				schemaInfo.Stats.TotalMessages++
				schemaInfo.Stats.TotalFields += len(messageInfo.Fields)
			}
		}

		// Process services
		for i := 0; i < file.Services().Len(); i++ {
			service := file.Services().Get(i)
			if t.shouldIncludeService(service, params.ServiceTypes) {
				serviceInfo := t.convertServiceToInfo(service, file)
				schemaInfo.Services = append(schemaInfo.Services, serviceInfo)
				schemaInfo.Stats.TotalServices++
				schemaInfo.Stats.TotalMethods += len(serviceInfo.Methods)
			}
		}

		// Process enums
		for i := 0; i < file.Enums().Len(); i++ {
			enum := file.Enums().Get(i)
			if t.shouldIncludeEnum(enum, params.EnumTypes) {
				enumInfo := t.convertEnumToInfo(enum, file)
				schemaInfo.Enums = append(schemaInfo.Enums, enumInfo)
				schemaInfo.Stats.TotalEnums++
				schemaInfo.Stats.TotalValues += len(enumInfo.Values)
			}
		}
	}

	return schemaInfo, nil
}

// shouldIncludeMessage checks if a message should be included based on filters
func (t *GetSchemaTool) shouldIncludeMessage(message protoreflect.MessageDescriptor, filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	for _, filter := range filters {
		if string(message.Name()) == filter {
			return true
		}
	}
	return false
}

// shouldIncludeService checks if a service should be included based on filters
func (t *GetSchemaTool) shouldIncludeService(service protoreflect.ServiceDescriptor, filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	for _, filter := range filters {
		if string(service.Name()) == filter {
			return true
		}
	}
	return false
}

// shouldIncludeEnum checks if an enum should be included based on filters
func (t *GetSchemaTool) shouldIncludeEnum(enum protoreflect.EnumDescriptor, filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	for _, filter := range filters {
		if string(enum.Name()) == filter {
			return true
		}
	}
	return false
}

// convertMessageToInfo converts a protobuf message to MessageInfo
func (t *GetSchemaTool) convertMessageToInfo(message protoreflect.MessageDescriptor, file linker.File) MessageInfo {
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
			Description: "",             // TODO: Extract from field options
			Options:     []OptionInfo{}, // TODO: Extract field options
		}
		fields = append(fields, fieldInfo)
	}

	return MessageInfo{
		Name:        string(message.Name()),
		FullName:    string(message.FullName()),
		Fields:      fields,
		File:        string(file.Name()),
		Package:     string(file.Package().Name()),
		Description: "",             // TODO: Extract from message options
		Options:     []OptionInfo{}, // TODO: Extract message options
	}
}

// convertServiceToInfo converts a protobuf service to ServiceInfo
func (t *GetSchemaTool) convertServiceToInfo(service protoreflect.ServiceDescriptor, file linker.File) ServiceInfo {
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
			Description:     "",             // TODO: Extract from method options
			Options:         []OptionInfo{}, // TODO: Extract method options
		}
		methods = append(methods, methodInfo)
	}

	return ServiceInfo{
		Name:        string(service.Name()),
		FullName:    string(service.FullName()),
		Methods:     methods,
		File:        string(file.Name()),
		Package:     string(file.Package().Name()),
		Description: "",             // TODO: Extract from service options
		Options:     []OptionInfo{}, // TODO: Extract service options
	}
}

// convertEnumToInfo converts a protobuf enum to EnumInfo
func (t *GetSchemaTool) convertEnumToInfo(enum protoreflect.EnumDescriptor, file linker.File) EnumInfo {
	// Convert enum values
	values := make([]EnumValueInfo, 0, enum.Values().Len())
	for i := 0; i < enum.Values().Len(); i++ {
		value := enum.Values().Get(i)
		valueInfo := EnumValueInfo{
			Name:        string(value.Name()),
			Number:      int32(value.Number()),
			Description: "",             // TODO: Extract from value options
			Options:     []OptionInfo{}, // TODO: Extract value options
		}
		values = append(values, valueInfo)
	}

	return EnumInfo{
		Name:        string(enum.Name()),
		FullName:    string(enum.FullName()),
		Values:      values,
		File:        string(file.Name()),
		Package:     string(file.Package().Name()),
		Description: "",             // TODO: Extract from enum options
		Options:     []OptionInfo{}, // TODO: Extract enum options
	}
}

// convertFileToInfo converts a protobuf file to FileInfo
func (t *GetSchemaTool) convertFileToInfo(file linker.File) FileInfo {
	return FileInfo{
		Name:         string(file.Name()),
		Package:      string(file.Package()),
		Dependencies: []string{},     // TODO: Extract dependencies
		Description:  "",             // TODO: Extract from file options
		Options:      []OptionInfo{}, // TODO: Extract file options
	}
}
