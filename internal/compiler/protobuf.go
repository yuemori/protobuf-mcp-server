package compiler

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bufbuild/protocompile"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/yuemori/protobuf-mcp-server/internal/config"
)

// ProtobufProject represents a compiled protobuf project
type ProtobufProject struct {
	ProjectRoot    string
	Config         *config.ProjectConfig
	CompiledProtos *descriptorpb.FileDescriptorSet
	resolver       protocompile.Resolver
	protoFiles     []string
}

// NewProtobufProject creates a new ProtobufProject instance
func NewProtobufProject(projectRoot string, cfg *config.ProjectConfig) (*ProtobufProject, error) {
	// Resolve proto files using the new config system
	protoFiles, err := config.ResolveProtoFiles(cfg, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve proto files: %w", err)
	}

	// Convert absolute paths to relative paths for protocompile
	var relativeProtoFiles []string
	for _, file := range protoFiles {
		relPath, err := filepath.Rel(projectRoot, file)
		if err != nil {
			return nil, fmt.Errorf("failed to get relative path for %s: %w", file, err)
		}
		relativeProtoFiles = append(relativeProtoFiles, relPath)
	}

	// Create source resolver with project root as import path
	resolver := &protocompile.SourceResolver{
		ImportPaths: []string{projectRoot},
	}

	// Wrap with standard imports for well-known types
	resolverWithStdImports := protocompile.WithStandardImports(resolver)

	return &ProtobufProject{
		ProjectRoot:    projectRoot,
		Config:         cfg,
		CompiledProtos: nil,
		resolver:       resolverWithStdImports,
		protoFiles:     relativeProtoFiles,
	}, nil
}

// CompileProtos compiles all proto files in the project
func (p *ProtobufProject) CompileProtos(ctx context.Context) error {
	if len(p.protoFiles) == 0 {
		return fmt.Errorf("no proto files found in configured paths")
	}

	// Create compiler with our resolver
	compiler := protocompile.Compiler{
		Resolver:       p.resolver,
		MaxParallelism: 4,
		SourceInfoMode: protocompile.SourceInfoStandard,
	}

	// Compile all proto files with dependency resolution
	// Sort files to ensure dependencies are compiled first
	sortedFiles := p.sortProtoFilesByDependencies(p.protoFiles)

	// Compile all files together - protocompile will resolve dependencies automatically
	files, err := compiler.Compile(ctx, sortedFiles...)
	if err != nil {
		return fmt.Errorf("failed to compile proto files: %w", err)
	}

	// Convert compiled files to FileDescriptorSet
	filesList := make([]*descriptorpb.FileDescriptorProto, len(files))
	for i, file := range files {
		// Each file implements protoreflect.FileDescriptor
		// Convert to FileDescriptorProto to preserve all information
		filesList[i] = protodesc.ToFileDescriptorProto(file)
	}

	p.CompiledProtos = &descriptorpb.FileDescriptorSet{
		File: filesList,
	}

	return nil
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

// sortProtoFilesByDependencies sorts proto files by their dependencies
// Google API files should be compiled before files that depend on them
func (p *ProtobufProject) sortProtoFilesByDependencies(protoFiles []string) []string {
	// Simple dependency sorting: Google API files first, then others
	var googleFiles []string
	var otherFiles []string

	for _, file := range protoFiles {
		if strings.HasPrefix(file, "google/") {
			googleFiles = append(googleFiles, file)
		} else {
			otherFiles = append(otherFiles, file)
		}
	}

	// Return Google API files first, then others
	return append(googleFiles, otherFiles...)
}