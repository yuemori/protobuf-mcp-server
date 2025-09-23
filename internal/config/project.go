package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ProjectConfig represents the configuration for a protobuf project
type ProjectConfig struct {
	RootDirectory   string                 `yaml:"root_directory"`
	IncludePaths    []string               `yaml:"include_paths"`
	ProtoPaths      []string               `yaml:"proto_paths"`
	CompilerOptions map[string]interface{} `yaml:"compiler_options"`
	IgnoredPatterns []string               `yaml:"ignored_patterns"`
	ShowLogs        bool                   `yaml:"show_logs"`
}

// DefaultProjectConfig returns a default configuration for a new project
func DefaultProjectConfig() *ProjectConfig {
	return &ProjectConfig{
		RootDirectory:   ".",
		IncludePaths:    []string{"."},
		ProtoPaths:      []string{"."},
		CompilerOptions: map[string]interface{}{},
		IgnoredPatterns: []string{
			"*_test.proto",
			"tmp/**",
		},
		ShowLogs: false,
	}
}

// LoadProjectConfig loads project configuration from the given directory
func LoadProjectConfig(projectRoot string) (*ProjectConfig, error) {
	configPath := filepath.Join(projectRoot, ".protobuf-mcp", "project.yml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read project config: %w", err)
	}

	var config ProjectConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse project config: %w", err)
	}

	return &config, nil
}

// SaveProjectConfig saves the project configuration to the given directory
func SaveProjectConfig(projectRoot string, config *ProjectConfig) error {
	configDir := filepath.Join(projectRoot, ".protobuf-mcp")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "project.yml")
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ProjectExists checks if a project is already initialized in the given directory
func ProjectExists(projectRoot string) bool {
	configPath := filepath.Join(projectRoot, ".protobuf-mcp", "project.yml")
	_, err := os.Stat(configPath)
	return err == nil
}
