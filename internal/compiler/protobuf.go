package compiler

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/bufbuild/protocompile"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/yuemori/protobuf-mcp-server/internal/config"
)

// ProtobufProject represents a compiled protobuf project
type ProtobufProject struct {
	ProjectRoot    string
	Config         *config.ProjectConfig
	CompiledProtos *descriptorpb.FileDescriptorSet
	resolver       protocompile.Resolver
}

// NewProtobufProject creates a new ProtobufProject instance
func NewProtobufProject(projectRoot string, cfg *config.ProjectConfig) (*ProtobufProject, error) {
	// Create source resolver with import paths
	resolver := &protocompile.SourceResolver{
		ImportPaths: cfg.IncludePaths,
	}

	// Wrap with standard imports for well-known types
	resolverWithStdImports := protocompile.WithStandardImports(resolver)

	return &ProtobufProject{
		ProjectRoot:    projectRoot,
		Config:         cfg,
		CompiledProtos: nil,
		resolver:       resolverWithStdImports,
	}, nil
}

// CompileProtos compiles all proto files in the project
func (p *ProtobufProject) CompileProtos(ctx context.Context) error {
	// Find all proto files in the configured paths
	protoFiles, err := p.findProtoFiles()
	if err != nil {
		return fmt.Errorf("failed to find proto files: %w", err)
	}

	if len(protoFiles) == 0 {
		return fmt.Errorf("no proto files found in configured paths")
	}

	// Create compiler with our resolver
	compiler := protocompile.Compiler{
		Resolver:       p.resolver,
		MaxParallelism: 4, // Reasonable default
		SourceInfoMode: protocompile.SourceInfoStandard,
	}

	// Compile all proto files
	files, err := compiler.Compile(ctx, protoFiles...)
	if err != nil {
		return fmt.Errorf("failed to compile proto files: %w", err)
	}

	// Convert compiled files to FileDescriptorSet
	filesList := make([]*descriptorpb.FileDescriptorProto, len(files))
	for i, file := range files {
		// Each file implements protoreflect.FileDescriptor
		proto := &descriptorpb.FileDescriptorProto{}
		path := file.Path()
		proto.Name = &path
		
		// For now, just set the name. Full conversion would require
		// iterating through services, messages, enums, etc.
		// This is a simplified implementation for initial testing.
		filesList[i] = proto
	}
	
	p.CompiledProtos = &descriptorpb.FileDescriptorSet{
		File: filesList,
	}

	return nil
}

// findProtoFiles discovers all proto files in the configured proto paths
func (p *ProtobufProject) findProtoFiles() ([]string, error) {
	var protoFiles []string

	for _, protoPath := range p.Config.ProtoPaths {
		// Convert to absolute path relative to project root
		absolutePath := filepath.Join(p.ProjectRoot, protoPath)

		// Use glob to find all .proto files
		pattern := filepath.Join(absolutePath, "**", "*.proto")
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to glob proto files in %s: %w", protoPath, err)
		}

		// Convert back to relative paths for protocompile
		for _, match := range matches {
			// Skip ignored patterns
			if p.shouldIgnoreFile(match) {
				continue
			}

			// Convert to relative path from project root
			relPath, err := filepath.Rel(p.ProjectRoot, match)
			if err != nil {
				return nil, fmt.Errorf("failed to get relative path for %s: %w", match, err)
			}

			protoFiles = append(protoFiles, relPath)
		}

		// Also try direct pattern matching without ** (for simple cases)
		simplePattern := filepath.Join(absolutePath, "*.proto")
		simpleMatches, err := filepath.Glob(simplePattern)
		if err != nil {
			return nil, fmt.Errorf("failed to glob proto files in %s: %w", protoPath, err)
		}

		for _, match := range simpleMatches {
			if p.shouldIgnoreFile(match) {
				continue
			}

			relPath, err := filepath.Rel(p.ProjectRoot, match)
			if err != nil {
				return nil, fmt.Errorf("failed to get relative path for %s: %w", match, err)
			}

			// Avoid duplicates
			found := false
			for _, existing := range protoFiles {
				if existing == relPath {
					found = true
					break
				}
			}
			if !found {
				protoFiles = append(protoFiles, relPath)
			}
		}
	}

	return protoFiles, nil
}

// shouldIgnoreFile checks if a file should be ignored based on configured patterns
func (p *ProtobufProject) shouldIgnoreFile(filePath string) bool {
	for _, pattern := range p.Config.IgnoredPatterns {
		matched, err := filepath.Match(pattern, filepath.Base(filePath))
		if err != nil {
			continue // Skip invalid patterns
		}
		if matched {
			return true
		}
	}
	return false
}

// GetServices extracts all services from compiled protos
func (p *ProtobufProject) GetServices() ([]*descriptorpb.ServiceDescriptorProto, error) {
	if p.CompiledProtos == nil {
		return nil, fmt.Errorf("project not compiled yet, call CompileProtos first")
	}

	var services []*descriptorpb.ServiceDescriptorProto

	for _, file := range p.CompiledProtos.File {
		for _, service := range file.Service {
			services = append(services, service)
		}
	}

	return services, nil
}

// GetMessages extracts all messages from compiled protos
func (p *ProtobufProject) GetMessages() ([]*descriptorpb.DescriptorProto, error) {
	if p.CompiledProtos == nil {
		return nil, fmt.Errorf("project not compiled yet, call CompileProtos first")
	}

	var messages []*descriptorpb.DescriptorProto

	for _, file := range p.CompiledProtos.File {
		messages = append(messages, file.MessageType...)
	}

	return messages, nil
}

// GetEnums extracts all enums from compiled protos
func (p *ProtobufProject) GetEnums() ([]*descriptorpb.EnumDescriptorProto, error) {
	if p.CompiledProtos == nil {
		return nil, fmt.Errorf("project not compiled yet, call CompileProtos first")
	}

	var enums []*descriptorpb.EnumDescriptorProto

	for _, file := range p.CompiledProtos.File {
		enums = append(enums, file.EnumType...)
	}

	return enums, nil
}