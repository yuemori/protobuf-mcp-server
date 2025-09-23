package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ProjectConfig represents the configuration for a protobuf project
type ProjectConfig struct {
	ProtoFiles []string `yaml:"proto_files"`
}

// DefaultProjectConfig returns a default configuration for a new project
func DefaultProjectConfig() *ProjectConfig {
	return &ProjectConfig{
		ProtoFiles: []string{
			"proto/**/*.proto",
		},
	}
}

// LoadProjectConfig loads project configuration from the given directory
func LoadProjectConfig(projectRoot string) (*ProjectConfig, error) {
	configPath := filepath.Join(projectRoot, ".protobuf-mcp.yml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read project config: %w", err)
	}

	// Preprocess the YAML to handle unquoted glob patterns
	processedData := preprocessYAML(string(data))

	var config ProjectConfig
	if err := yaml.Unmarshal([]byte(processedData), &config); err != nil {
		return nil, fmt.Errorf("failed to parse project config: %w", err)
	}

	return &config, nil
}

// preprocessYAML handles unquoted glob patterns in YAML
func preprocessYAML(yamlContent string) string {
	lines := strings.Split(yamlContent, "\n")
	var processedLines []string

	for _, line := range lines {
		// Check if line contains unquoted glob patterns
		if strings.Contains(line, "- ") && (strings.Contains(line, "**") || strings.Contains(line, "*")) {
			// Find the pattern after "- "
			parts := strings.SplitN(line, "- ", 2)
			if len(parts) == 2 {
				pattern := strings.TrimSpace(parts[1])
				// Only quote if it's not already quoted and contains special characters
				if !strings.HasPrefix(pattern, "\"") && !strings.HasPrefix(pattern, "'") {
					if strings.Contains(pattern, "**") || strings.Contains(pattern, "*") {
						pattern = "\"" + pattern + "\""
					}
				}
				processedLines = append(processedLines, parts[0]+"- "+pattern)
				continue
			}
		}
		processedLines = append(processedLines, line)
	}

	return strings.Join(processedLines, "\n")
}

// SaveProjectConfig saves the project configuration to the given directory
func SaveProjectConfig(projectRoot string, config *ProjectConfig) error {
	configPath := filepath.Join(projectRoot, ".protobuf-mcp.yml")

	// Create a custom YAML structure to ensure proper quoting
	yamlData := struct {
		ProtoFiles []string `yaml:"proto_files"`
	}{
		ProtoFiles: config.ProtoFiles,
	}

	data, err := yaml.Marshal(yamlData)
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
	configPath := filepath.Join(projectRoot, ".protobuf-mcp.yml")
	_, err := os.Stat(configPath)
	return err == nil
}

// ResolveProtoFiles resolves proto file patterns to actual file paths
func ResolveProtoFiles(config *ProjectConfig, projectRoot string) ([]string, error) {
	var resolvedFiles []string

	for _, pattern := range config.ProtoFiles {
		if filepath.IsAbs(pattern) {
			// Absolute path: use as-is
			matches, err := filepath.Glob(pattern)
			if err != nil {
				return nil, fmt.Errorf("failed to glob absolute pattern %q: %w", pattern, err)
			}
			resolvedFiles = append(resolvedFiles, matches...)
		} else {
			// Relative path: resolve from project root
			fullPattern := filepath.Join(projectRoot, pattern)

			// Handle ** pattern by walking the directory tree
			if strings.Contains(pattern, "**") {
				matches, err := resolveRecursivePattern(pattern, projectRoot)
				if err != nil {
					return nil, fmt.Errorf("failed to resolve recursive pattern %q: %w", pattern, err)
				}
				resolvedFiles = append(resolvedFiles, matches...)
			} else {
				matches, err := filepath.Glob(fullPattern)
				if err != nil {
					return nil, fmt.Errorf("failed to glob relative pattern %q: %w", pattern, err)
				}
				resolvedFiles = append(resolvedFiles, matches...)
			}
		}
	}

	return resolvedFiles, nil
}

// resolveRecursivePattern handles ** patterns by walking the directory tree
// resolveRecursivePattern handles ** patterns by walking the directory tree
// resolveRecursivePattern handles ** patterns by walking the directory tree
// resolveRecursivePattern handles ** patterns by walking the directory tree
// resolveRecursivePattern handles ** patterns by walking the directory tree
// resolveRecursivePattern handles ** patterns by walking the directory tree
// resolveRecursivePattern handles ** patterns by walking the directory tree
// resolveRecursivePattern handles ** patterns by walking the directory tree
func resolveRecursivePattern(pattern, projectRoot string) ([]string, error) {
	var matches []string

	// Find the ** pattern in the string
	starStarIndex := strings.Index(pattern, "**")
	if starStarIndex == -1 {
		// Fallback to regular glob if not a ** pattern
		fullPattern := filepath.Join(projectRoot, pattern)
		return filepath.Glob(fullPattern)
	}

	baseDir := filepath.Join(projectRoot, strings.TrimSuffix(pattern[:starStarIndex], "/"))
	filePattern := strings.TrimPrefix(pattern[starStarIndex+2:], "/")

	// Check if baseDir exists
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		return matches, nil
	}

	// Walk the directory tree
	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			// For ** patterns, we need to check if the file matches the pattern
			// Since we're walking recursively, we can use the filename directly
			filename := filepath.Base(path)

			matched, err := filepath.Match(filePattern, filename)
			if err != nil {
				return err
			}

			if matched {
				matches = append(matches, path)
			}
		}

		return nil
	})

	return matches, err
}
