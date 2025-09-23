package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/yuemori/protobuf-mcp-server/internal/config"
)

func TestActivateProjectTool_Execute(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "protobuf-mcp-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test proto files
	testProtoDir := filepath.Join(tempDir, "proto")
	if err := os.MkdirAll(testProtoDir, 0755); err != nil {
		t.Fatalf("Failed to create proto dir: %v", err)
	}

	// Write a simple test proto file
	testProtoContent := `syntax = "proto3";

package test.v1;

option go_package = "github.com/test/v1";

// Test service
service TestService {
  // Test method
  rpc TestMethod(TestRequest) returns (TestResponse);
}

// Test request message
message TestRequest {
  string name = 1;
}

// Test response message
message TestResponse {
  string message = 1;
}
`

	testProtoFile := filepath.Join(testProtoDir, "test.proto")
	if err := os.WriteFile(testProtoFile, []byte(testProtoContent), 0644); err != nil {
		t.Fatalf("Failed to write test proto file: %v", err)
	}

	tool := NewActivateProjectTool()

	tests := []struct {
		name          string
		params        ActivateProjectParams
		expectSuccess bool
		expectError   bool
		setupProject  bool
	}{
		{
			name: "successful activation",
			params: ActivateProjectParams{
				ProjectPath: tempDir,
			},
			expectSuccess: true,
			expectError:   false,
			setupProject:  true,
		},
		{
			name: "project not initialized",
			params: ActivateProjectParams{
				ProjectPath: "/non/existent/project",
			},
			expectSuccess: false,
			expectError:   false,
			setupProject:  false,
		},
		{
			name: "empty project path",
			params: ActivateProjectParams{
				ProjectPath: "",
			},
			expectSuccess: false,
			expectError:   true,
			setupProject:  false,
		},
		{
			name: "non-existent project path",
			params: ActivateProjectParams{
				ProjectPath: "/non/existent/path",
			},
			expectSuccess: false,
			expectError:   false,
			setupProject:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup project if needed
			if tt.setupProject {
				// Initialize project configuration
				projectConfig := config.DefaultProjectConfig()
				projectConfig.ProtoPaths = []string{"proto"}
				if err := config.SaveProjectConfig(tempDir, projectConfig); err != nil {
					t.Fatalf("Failed to setup project: %v", err)
				}
			}

			// Marshal parameters
			paramsJSON, err := json.Marshal(tt.params)
			if err != nil {
				t.Fatalf("Failed to marshal params: %v", err)
			}

			// Execute tool
			result, err := tool.Execute(context.Background(), paramsJSON)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check result
			if !tt.expectError {
				response, ok := result.(*ActivateProjectResponse)
				if !ok {
					t.Fatalf("Expected ActivateProjectResponse, got %T", result)
				}

				if response.Success != tt.expectSuccess {
					t.Errorf("Expected success=%v, got %v", tt.expectSuccess, response.Success)
				}

				// For successful activation, check that we have some proto files
				if tt.expectSuccess && response.ProtoFiles == 0 {
					t.Errorf("Expected proto files to be found, got 0")
				}
			}
		})
	}
}

func TestActivateProjectTool_Name(t *testing.T) {
	tool := NewActivateProjectTool()
	expected := "activate_project"
	if tool.Name() != expected {
		t.Errorf("Expected name %s, got %s", expected, tool.Name())
	}
}

func TestActivateProjectTool_Description(t *testing.T) {
	tool := NewActivateProjectTool()
	description := tool.Description()
	if description == "" {
		t.Error("Expected non-empty description")
	}
}
