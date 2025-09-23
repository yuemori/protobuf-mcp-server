package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

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
func CreateTestProject() (*compiler.ProtobufProject, error) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %v", err)
	}

	// Use existing test data from the compiler package
	// Go up from internal/tools to project root, then to testdata
	projectRoot := filepath.Join(cwd, "..", "..")
	rootDir := filepath.Join(projectRoot, "internal/compiler/testdata/simple")
	protoFiles := []string{
		"api.proto",
		"types.proto",
	}
	importPaths := []string{"."}

	// Compile the protos directly using the CompileProtos function
	ctx := context.Background()
	compiledProtos, err := compiler.CompileProtos(ctx, rootDir, protoFiles, importPaths)
	if err != nil {
		return nil, fmt.Errorf("compilation failed: %v", err)
	}

	// Create a basic config
	cfg := &config.ProjectConfig{
		ProtoFiles:  protoFiles,
		ImportPaths: importPaths,
	}

	// Create the project with compiled protos
	project := &compiler.ProtobufProject{
		ProjectRoot:    rootDir,
		Config:         cfg,
		CompiledProtos: compiledProtos,
	}

	return project, nil
}
