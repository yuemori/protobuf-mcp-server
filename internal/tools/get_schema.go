package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/yuemori/protobuf-mcp-server/internal/compiler"
	"google.golang.org/protobuf/types/descriptorpb"
)

// GetSchemaTool implements the get_schema MCP tool
type GetSchemaTool struct {
	project *compiler.ProtobufProject
}

// SetProject sets the current project
func (t *GetSchemaTool) SetProject(project *compiler.ProtobufProject) {
	t.project = project
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
	Label       string       `json:"label"`
	Description string       `json:"description"`
	Default     string       `json:"default,omitempty"`
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

// Name returns the tool name
func (t *GetSchemaTool) Name() string {
	return "get_schema"
}

// Description returns the tool description
func (t *GetSchemaTool) Description() string {
	return "Get detailed schema information from the activated protobuf project"
}

// Execute gets detailed schema information from the activated protobuf project
func (t *GetSchemaTool) Execute(ctx context.Context, params json.RawMessage) (interface{}, error) {
	// Parse parameters
	var getSchemaParams GetSchemaParams
	if len(params) > 0 {
		if err := json.Unmarshal(params, &getSchemaParams); err != nil {
			return &GetSchemaResponse{
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
		return &GetSchemaResponse{
			Success: false,
			Message: "No project activated. Use activate_project first.",
		}, nil
	}

	// Check if project is compiled
	if project.CompiledProtos == nil {
		return &GetSchemaResponse{
			Success: false,
			Message: "Project not compiled. Use activate_project first.",
		}, nil
	}

	// Get schema information
	schemaInfo, err := t.buildSchemaInfo(project, &getSchemaParams)
	if err != nil {
		return &GetSchemaResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to build schema info: %v", err),
		}, nil
	}

	// Calculate count
	count := schemaInfo.Stats.TotalMessages + schemaInfo.Stats.TotalServices + schemaInfo.Stats.TotalEnums

	return &GetSchemaResponse{
		Success: true,
		Message: fmt.Sprintf("Retrieved schema information: %d messages, %d services, %d enums",
			schemaInfo.Stats.TotalMessages, schemaInfo.Stats.TotalServices, schemaInfo.Stats.TotalEnums),
		Schema: schemaInfo,
		Count:  count,
	}, nil
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
	for _, file := range project.CompiledProtos.File {
		// Add file information if requested
		if params.IncludeFileInfo {
			fileInfo := t.convertFileToInfo(file)
			schemaInfo.Files = append(schemaInfo.Files, fileInfo)
			schemaInfo.Stats.TotalFiles++
		}

		// Process messages
		for _, message := range file.MessageType {
			if t.shouldIncludeMessage(message, params.MessageTypes) {
				messageInfo := t.convertMessageToInfo(message, file)
				schemaInfo.Messages = append(schemaInfo.Messages, messageInfo)
				schemaInfo.Stats.TotalMessages++
				schemaInfo.Stats.TotalFields += len(messageInfo.Fields)
			}
		}

		// Process services
		for _, service := range file.Service {
			if t.shouldIncludeService(service, params.ServiceTypes) {
				serviceInfo := t.convertServiceToInfo(service, file)
				schemaInfo.Services = append(schemaInfo.Services, serviceInfo)
				schemaInfo.Stats.TotalServices++
				schemaInfo.Stats.TotalMethods += len(serviceInfo.Methods)
			}
		}

		// Process enums
		for _, enum := range file.EnumType {
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
func (t *GetSchemaTool) shouldIncludeMessage(message *descriptorpb.DescriptorProto, filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	for _, filter := range filters {
		if message.GetName() == filter {
			return true
		}
	}
	return false
}

// shouldIncludeService checks if a service should be included based on filters
func (t *GetSchemaTool) shouldIncludeService(service *descriptorpb.ServiceDescriptorProto, filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	for _, filter := range filters {
		if service.GetName() == filter {
			return true
		}
	}
	return false
}

// shouldIncludeEnum checks if an enum should be included based on filters
func (t *GetSchemaTool) shouldIncludeEnum(enum *descriptorpb.EnumDescriptorProto, filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	for _, filter := range filters {
		if enum.GetName() == filter {
			return true
		}
	}
	return false
}

// convertMessageToInfo converts a protobuf message to MessageInfo
func (t *GetSchemaTool) convertMessageToInfo(message *descriptorpb.DescriptorProto, file *descriptorpb.FileDescriptorProto) MessageInfo {
	// Convert fields
	fields := make([]FieldInfo, 0, len(message.Field))
	for _, field := range message.Field {
		fieldInfo := FieldInfo{
			Name:        field.GetName(),
			Number:      field.GetNumber(),
			Type:        field.GetTypeName(),
			Label:       field.GetLabel().String(),
			Description: "", // TODO: Extract from field options
			Default:     field.GetDefaultValue(),
			Options:     []OptionInfo{}, // TODO: Extract field options
		}
		fields = append(fields, fieldInfo)
	}

	return MessageInfo{
		Name:        message.GetName(),
		FullName:    message.GetName(), // TODO: Add package prefix
		Fields:      fields,
		File:        file.GetName(),
		Package:     file.GetPackage(),
		Description: "",             // TODO: Extract from message options
		Options:     []OptionInfo{}, // TODO: Extract message options
	}
}

// convertServiceToInfo converts a protobuf service to ServiceInfo
func (t *GetSchemaTool) convertServiceToInfo(service *descriptorpb.ServiceDescriptorProto, file *descriptorpb.FileDescriptorProto) ServiceInfo {
	// Convert methods
	methods := make([]MethodInfo, 0, len(service.Method))
	for _, method := range service.Method {
		methodInfo := MethodInfo{
			Name:            method.GetName(),
			InputType:       method.GetInputType(),
			OutputType:      method.GetOutputType(),
			ClientStreaming: method.GetClientStreaming(),
			ServerStreaming: method.GetServerStreaming(),
			Description:     "",             // TODO: Extract from method options
			Options:         []OptionInfo{}, // TODO: Extract method options
		}
		methods = append(methods, methodInfo)
	}

	return ServiceInfo{
		Name:        service.GetName(),
		FullName:    service.GetName(), // TODO: Add package prefix
		Methods:     methods,
		File:        file.GetName(),
		Package:     file.GetPackage(),
		Description: "",             // TODO: Extract from service options
		Options:     []OptionInfo{}, // TODO: Extract service options
	}
}

// convertEnumToInfo converts a protobuf enum to EnumInfo
func (t *GetSchemaTool) convertEnumToInfo(enum *descriptorpb.EnumDescriptorProto, file *descriptorpb.FileDescriptorProto) EnumInfo {
	// Convert enum values
	values := make([]EnumValueInfo, 0, len(enum.Value))
	for _, value := range enum.Value {
		valueInfo := EnumValueInfo{
			Name:        value.GetName(),
			Number:      value.GetNumber(),
			Description: "",             // TODO: Extract from value options
			Options:     []OptionInfo{}, // TODO: Extract value options
		}
		values = append(values, valueInfo)
	}

	return EnumInfo{
		Name:        enum.GetName(),
		FullName:    enum.GetName(), // TODO: Add package prefix
		Values:      values,
		File:        file.GetName(),
		Package:     file.GetPackage(),
		Description: "",             // TODO: Extract from enum options
		Options:     []OptionInfo{}, // TODO: Extract enum options
	}
}

// convertFileToInfo converts a protobuf file to FileInfo
func (t *GetSchemaTool) convertFileToInfo(file *descriptorpb.FileDescriptorProto) FileInfo {
	return FileInfo{
		Name:         file.GetName(),
		Package:      file.GetPackage(),
		Dependencies: file.GetDependency(),
		Description:  "",             // TODO: Extract from file options
		Options:      []OptionInfo{}, // TODO: Extract file options
	}
}
