package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/yuemori/protobuf-mcp-server/internal/compiler"
)

// ProjectState represents the persistent project state
type ProjectState struct {
	ProjectRoot    string `json:"project_root"`
	CompiledProtos []byte `json:"compiled_protos,omitempty"` // Serialized FileDescriptorSet
}

// ProjectStateManager manages persistent project state
type ProjectStateManager struct {
	stateFile string
	mu        sync.RWMutex
}

// Global project state manager instance
var globalProjectStateManager = &ProjectStateManager{
	stateFile: ".protobuf-mcp/state.json",
}

// GetProjectStateManager returns the global project state manager instance
func GetProjectStateManager() *ProjectStateManager {
	return globalProjectStateManager
}

// SaveProjectState saves the current project state to disk
func (psm *ProjectStateManager) SaveProjectState(project *compiler.ProtobufProject) error {
	psm.mu.Lock()
	defer psm.mu.Unlock()

	if project == nil || project.CompiledProtos == nil {
		return nil // Nothing to save
	}

	// Serialize the compiled protos
	protosData, err := json.Marshal(project.CompiledProtos)
	if err != nil {
		return err
	}

	state := ProjectState{
		ProjectRoot:    project.ProjectRoot,
		CompiledProtos: protosData,
	}

	// Ensure directory exists
	stateDir := filepath.Dir(psm.stateFile)
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return err
	}

	// Write state to file
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	return os.WriteFile(psm.stateFile, data, 0644)
}

// LoadProjectState loads the project state from disk
func (psm *ProjectStateManager) LoadProjectState() (*compiler.ProtobufProject, error) {
	psm.mu.RLock()
	defer psm.mu.RUnlock()

	// Check if state file exists
	if _, err := os.Stat(psm.stateFile); os.IsNotExist(err) {
		return nil, nil // No state file
	}

	// Read state file
	data, err := os.ReadFile(psm.stateFile)
	if err != nil {
		return nil, err
	}

	var state ProjectState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	// Deserialize compiled protos
	var compiledProtos map[string]interface{}
	if err := json.Unmarshal(state.CompiledProtos, &compiledProtos); err != nil {
		return nil, err
	}

	// Create project instance
	project := &compiler.ProtobufProject{
		ProjectRoot: state.ProjectRoot,
		// Note: We can't easily deserialize FileDescriptorSet from JSON
		// This is a simplified approach for demonstration
		CompiledProtos: nil, // Will be recompiled when needed
	}

	return project, nil
}

// ClearProjectState clears the project state
func (psm *ProjectStateManager) ClearProjectState() error {
	psm.mu.Lock()
	defer psm.mu.Unlock()

	if _, err := os.Stat(psm.stateFile); os.IsNotExist(err) {
		return nil // Already cleared
	}

	return os.Remove(psm.stateFile)
}
