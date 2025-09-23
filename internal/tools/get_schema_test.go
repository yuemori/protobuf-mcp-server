package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/yuemori/protobuf-mcp-server/internal/compiler"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestGetSchemaTool_NameAndDescription(t *testing.T) {
	tool := &GetSchemaTool{}

	if tool.Name() != "get_schema" {
		t.Errorf("Expected name 'get_schema', got '%s'", tool.Name())
	}

	expectedDesc := "Get detailed schema information from the activated protobuf project"
	if tool.Description() != expectedDesc {
		t.Errorf("Expected description '%s', got '%s'", expectedDesc, tool.Description())
	}
}

func TestGetSchemaTool_Execute_NoProject(t *testing.T) {
	tool := &GetSchemaTool{}

	// Clear any existing project state
	GetProjectManager().ClearCurrentProject()

	// Test with no project activated
	params := json.RawMessage(`{}`)
	result, err := tool.Execute(context.Background(), params)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	response, ok := result.(*GetSchemaResponse)
	if !ok {
		t.Fatalf("Expected GetSchemaResponse, got %T", result)
	}

	if response.Success {
		t.Error("Expected failure response when no project is activated")
	}

	expectedMessage := "No project activated. Use activate_project first."
	if response.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, response.Message)
	}
}

func TestGetSchemaTool_Execute_ProjectNotCompiled(t *testing.T) {
	tool := &GetSchemaTool{}

	// Create a project but don't compile it
	project := &compiler.ProtobufProject{
		CompiledProtos: nil, // Not compiled
	}
	tool.SetProject(project)

	params := json.RawMessage(`{}`)
	result, err := tool.Execute(context.Background(), params)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	response, ok := result.(*GetSchemaResponse)
	if !ok {
		t.Fatalf("Expected GetSchemaResponse, got %T", result)
	}

	if response.Success {
		t.Error("Expected failure response when project is not compiled")
	}

	expectedMessage := "Project not compiled. Use activate_project first."
	if response.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, response.Message)
	}
}

func TestGetSchemaTool_Execute_Success(t *testing.T) {
	tool := &GetSchemaTool{}

	// Create a mock compiled project
	project := &compiler.ProtobufProject{
		CompiledProtos: &descriptorpb.FileDescriptorSet{
			File: []*descriptorpb.FileDescriptorProto{
				{
					Name:    stringPtr("test.proto"),
					Package: stringPtr("test.v1"),
					MessageType: []*descriptorpb.DescriptorProto{
						{
							Name: stringPtr("TestMessage"),
							Field: []*descriptorpb.FieldDescriptorProto{
								{
									Name:   stringPtr("id"),
									Number: int32Ptr(1),
									Type:   descriptorpb.FieldDescriptorProto_TYPE_INT32.Enum(),
									Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
								},
							},
						},
					},
					Service: []*descriptorpb.ServiceDescriptorProto{
						{
							Name: stringPtr("TestService"),
							Method: []*descriptorpb.MethodDescriptorProto{
								{
									Name:       stringPtr("TestMethod"),
									InputType:  stringPtr(".test.v1.TestRequest"),
									OutputType: stringPtr(".test.v1.TestResponse"),
								},
							},
						},
					},
					EnumType: []*descriptorpb.EnumDescriptorProto{
						{
							Name: stringPtr("TestEnum"),
							Value: []*descriptorpb.EnumValueDescriptorProto{
								{
									Name:   stringPtr("UNKNOWN"),
									Number: int32Ptr(0),
								},
							},
						},
					},
				},
			},
		},
	}
	tool.SetProject(project)

	params := json.RawMessage(`{}`)
	result, err := tool.Execute(context.Background(), params)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	response, ok := result.(*GetSchemaResponse)
	if !ok {
		t.Fatalf("Expected GetSchemaResponse, got %T", result)
	}

	if !response.Success {
		t.Errorf("Expected success response, got failure: %s", response.Message)
	}

	if response.Schema == nil {
		t.Fatal("Expected schema information, got nil")
	}

	// Check message count
	if len(response.Schema.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(response.Schema.Messages))
	}

	// Check service count
	if len(response.Schema.Services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(response.Schema.Services))
	}

	// Check enum count
	if len(response.Schema.Enums) != 1 {
		t.Errorf("Expected 1 enum, got %d", len(response.Schema.Enums))
	}

	// Check stats
	stats := response.Schema.Stats
	if stats.TotalMessages != 1 {
		t.Errorf("Expected 1 total message, got %d", stats.TotalMessages)
	}
	if stats.TotalServices != 1 {
		t.Errorf("Expected 1 total service, got %d", stats.TotalServices)
	}
	if stats.TotalEnums != 1 {
		t.Errorf("Expected 1 total enum, got %d", stats.TotalEnums)
	}
	if stats.TotalFields != 1 {
		t.Errorf("Expected 1 total field, got %d", stats.TotalFields)
	}
	if stats.TotalMethods != 1 {
		t.Errorf("Expected 1 total method, got %d", stats.TotalMethods)
	}
	if stats.TotalValues != 1 {
		t.Errorf("Expected 1 total value, got %d", stats.TotalValues)
	}
}

func TestGetSchemaTool_Execute_WithFilters(t *testing.T) {
	tool := &GetSchemaTool{}

	// Create a mock compiled project
	project := &compiler.ProtobufProject{
		CompiledProtos: &descriptorpb.FileDescriptorSet{
			File: []*descriptorpb.FileDescriptorProto{
				{
					Name:    stringPtr("test.proto"),
					Package: stringPtr("test.v1"),
					MessageType: []*descriptorpb.DescriptorProto{
						{
							Name: stringPtr("TestMessage"),
						},
						{
							Name: stringPtr("AnotherMessage"),
						},
					},
					Service: []*descriptorpb.ServiceDescriptorProto{
						{
							Name: stringPtr("TestService"),
						},
						{
							Name: stringPtr("AnotherService"),
						},
					},
				},
			},
		},
	}
	tool.SetProject(project)

	// Test with filters
	params := json.RawMessage(`{
		"message_types": ["TestMessage"],
		"service_types": ["TestService"],
		"include_file_info": true
	}`)
	result, err := tool.Execute(context.Background(), params)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	response, ok := result.(*GetSchemaResponse)
	if !ok {
		t.Fatalf("Expected GetSchemaResponse, got %T", result)
	}

	if !response.Success {
		t.Errorf("Expected success response, got failure: %s", response.Message)
	}

	// Check filtered results
	if len(response.Schema.Messages) != 1 {
		t.Errorf("Expected 1 filtered message, got %d", len(response.Schema.Messages))
	}

	if len(response.Schema.Services) != 1 {
		t.Errorf("Expected 1 filtered service, got %d", len(response.Schema.Services))
	}

	// Check file info is included
	if len(response.Schema.Files) != 1 {
		t.Errorf("Expected 1 file info, got %d", len(response.Schema.Files))
	}
}

func TestGetSchemaTool_Execute_InvalidParams(t *testing.T) {
	tool := &GetSchemaTool{}

	// Test with invalid JSON parameters
	params := json.RawMessage(`{"invalid": json}`)
	result, err := tool.Execute(context.Background(), params)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	response, ok := result.(*GetSchemaResponse)
	if !ok {
		t.Fatalf("Expected GetSchemaResponse, got %T", result)
	}

	if response.Success {
		t.Error("Expected failure response for invalid parameters")
	}

	if !contains(response.Message, "Invalid parameters") {
		t.Errorf("Expected 'Invalid parameters' in message, got '%s'", response.Message)
	}
}

// Helper functions are defined in types.go
