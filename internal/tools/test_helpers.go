package tools

import (
	"github.com/yuemori/protobuf-mcp-server/internal/compiler"
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
