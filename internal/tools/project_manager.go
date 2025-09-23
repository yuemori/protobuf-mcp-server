package tools

import (
	"sync"

	"github.com/yuemori/protobuf-mcp-server/internal/compiler"
)

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

// SetCurrentProject sets the currently activated project
func (pm *ProjectManager) SetCurrentProject(project *compiler.ProtobufProject) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.currentProject = project
}

// GetCurrentProject returns the currently activated project
func (pm *ProjectManager) GetCurrentProject() *compiler.ProtobufProject {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.currentProject
}

// ClearCurrentProject clears the currently activated project
func (pm *ProjectManager) ClearCurrentProject() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.currentProject = nil
}
