package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/yuemori/protobuf-mcp-server/internal/compiler"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestListServicesTool_Execute(t *testing.T) {
	tool := &ListServicesTool{}

	tests := []struct {
		name            string
		args            json.RawMessage
		setupProject    func()
		expectedSuccess bool
		expectedCount   int
	}{
		{
			name:            "No project activated",
			args:            json.RawMessage(`{}`),
			setupProject:    func() { GetProjectManager().ClearCurrentProject() },
			expectedSuccess: false,
		},
		{
			name: "Project activated with services",
			args: json.RawMessage(`{}`),
			setupProject: func() {
				// Create a mock project with services
				project := &compiler.ProtobufProject{
					CompiledProtos: &descriptorpb.FileDescriptorSet{
						File: []*descriptorpb.FileDescriptorProto{
							{
								Name:    stringPtr("test.proto"),
								Package: stringPtr("test.v1"),
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
							},
						},
					},
				}
				GetProjectManager().SetCurrentProject(project)
			},
			expectedSuccess: true,
			expectedCount:   1,
		},
		{
			name: "Project activated but not compiled",
			args: json.RawMessage(`{}`),
			setupProject: func() {
				// Create a mock project without compiled protos
				project := &compiler.ProtobufProject{
					CompiledProtos: nil,
				}
				GetProjectManager().SetCurrentProject(project)
			},
			expectedSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tt.setupProject()

			// Execute
			result, err := tool.Execute(context.Background(), tt.args)

			// Verify no error
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			// Verify result type
			response, ok := result.(*ListServicesResponse)
			if !ok {
				t.Fatalf("Execute() result = %T, want *ListServicesResponse", result)
			}

			// Verify success
			if response.Success != tt.expectedSuccess {
				t.Errorf("Execute() success = %v, want %v", response.Success, tt.expectedSuccess)
			}

			// Verify count if expected
			if tt.expectedCount >= 0 && response.Count != tt.expectedCount {
				t.Errorf("Execute() count = %v, want %v", response.Count, tt.expectedCount)
			}
		})
	}
}

func TestListServicesTool_ConvertServiceToInfo(t *testing.T) {
	tool := &ListServicesTool{}

	// Create a mock service
	service := &descriptorpb.ServiceDescriptorProto{
		Name: stringPtr("TestService"),
		Method: []*descriptorpb.MethodDescriptorProto{
			{
				Name:            stringPtr("TestMethod"),
				InputType:       stringPtr(".test.v1.TestRequest"),
				OutputType:      stringPtr(".test.v1.TestResponse"),
				ClientStreaming: boolPtr(false),
				ServerStreaming: boolPtr(false),
			},
			{
				Name:            stringPtr("StreamMethod"),
				InputType:       stringPtr(".test.v1.StreamRequest"),
				OutputType:      stringPtr(".test.v1.StreamResponse"),
				ClientStreaming: boolPtr(true),
				ServerStreaming: boolPtr(true),
			},
		},
	}

	// Convert to ServiceInfo
	serviceInfo := tool.convertServiceToInfo(service)

	// Verify service info
	if serviceInfo.Name != "TestService" {
		t.Errorf("convertServiceToInfo() name = %v, want TestService", serviceInfo.Name)
	}

	if len(serviceInfo.Methods) != 2 {
		t.Errorf("convertServiceToInfo() methods count = %v, want 2", len(serviceInfo.Methods))
	}

	// Verify first method
	method1 := serviceInfo.Methods[0]
	if method1.Name != "TestMethod" {
		t.Errorf("convertServiceToInfo() method1 name = %v, want TestMethod", method1.Name)
	}
	if method1.InputType != ".test.v1.TestRequest" {
		t.Errorf("convertServiceToInfo() method1 input type = %v, want .test.v1.TestRequest", method1.InputType)
	}
	if method1.OutputType != ".test.v1.TestResponse" {
		t.Errorf("convertServiceToInfo() method1 output type = %v, want .test.v1.TestResponse", method1.OutputType)
	}
	if method1.ClientStreaming {
		t.Errorf("convertServiceToInfo() method1 client streaming = %v, want false", method1.ClientStreaming)
	}
	if method1.ServerStreaming {
		t.Errorf("convertServiceToInfo() method1 server streaming = %v, want false", method1.ServerStreaming)
	}

	// Verify second method (streaming)
	method2 := serviceInfo.Methods[1]
	if method2.Name != "StreamMethod" {
		t.Errorf("convertServiceToInfo() method2 name = %v, want StreamMethod", method2.Name)
	}
	if !method2.ClientStreaming {
		t.Errorf("convertServiceToInfo() method2 client streaming = %v, want true", method2.ClientStreaming)
	}
	if !method2.ServerStreaming {
		t.Errorf("convertServiceToInfo() method2 server streaming = %v, want true", method2.ServerStreaming)
	}
}

func TestListServicesTool_NameAndDescription(t *testing.T) {
	tool := &ListServicesTool{}

	if tool.Name() != "list_services" {
		t.Errorf("Name() = %v, want list_services", tool.Name())
	}

	if tool.Description() == "" {
		t.Errorf("Description() is empty")
	}
}

// Helper functions are defined in types.go
