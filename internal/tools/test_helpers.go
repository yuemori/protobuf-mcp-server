package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/yuemori/protobuf-mcp-server/internal/compiler"
	"github.com/yuemori/protobuf-mcp-server/internal/config"
)

// MockProjectManager is a mock implementation of ProjectManagerInterface
type MockProjectManager struct {
	project *compiler.ProtobufProject
}

// SetProject sets the current project
func (m *MockProjectManager) SetProject(project *compiler.ProtobufProject) {
	m.project = project
}

// GetProject gets the current project
func (m *MockProjectManager) GetProject() *compiler.ProtobufProject {
	return m.project
}

// CreateTestProject creates a real project using test data
func CreateTestProject(t *testing.T) (*compiler.ProtobufProject, error) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("Failed to restore working directory: %v", err)
		}
	})

	// Use existing test data
	rootDir := filepath.Join(cwd, "testdata")
	protoFiles := []string{
		"api.proto",
		"types.proto",
	}
	importPaths := []string{"."}

	// Create a basic config
	cfg := &config.ProjectConfig{
		ProtoFiles:  protoFiles,
		ImportPaths: importPaths,
	}

	// Create the project with compiled protos
	project := &compiler.ProtobufProject{
		ProjectRoot: rootDir,
		Config:      cfg,
	}

	return project, nil
}
