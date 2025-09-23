package tools

import (
	"sync"

	"github.com/yuemori/protobuf-mcp-server/internal/compiler"
)

// ProjectManagerInterface defines the interface for project state management
type ProjectManagerInterface interface {
	SetProject(*compiler.ProtobufProject)
	GetProject() *compiler.ProtobufProject
}

// ProjectManager manages the global project state
type ProjectManager struct {
	currentProject *compiler.ProtobufProject
	mu             sync.RWMutex
}

// Global project manager instance
var globalProjectManager = &ProjectManager{}

// GetProjectManager returns the global project manager instance
func GetProjectManager() *ProjectManager {
	return globalProjectManager
}

// SetProject sets the currently activated project (implements ProjectManager interface)
func (pm *ProjectManager) SetProject(project *compiler.ProtobufProject) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.currentProject = project
}

// GetProject returns the currently activated project (implements ProjectManager interface)
func (pm *ProjectManager) GetProject() *compiler.ProtobufProject {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.currentProject
}

// SetCurrentProject sets the currently activated project (legacy method)
func (pm *ProjectManager) SetCurrentProject(project *compiler.ProtobufProject) {
	pm.SetProject(project)
}

// GetCurrentProject returns the currently activated project (legacy method)
func (pm *ProjectManager) GetCurrentProject() *compiler.ProtobufProject {
	return pm.GetProject()
}

// ClearCurrentProject clears the currently activated project
func (pm *ProjectManager) ClearCurrentProject() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.currentProject = nil
}
