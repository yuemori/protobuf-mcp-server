package compiler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/linker"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/yuemori/protobuf-mcp-server/internal/config"
)

// getMaxParallelism returns the maximum parallelism from environment variable or default
func getMaxParallelism() int {
	if val := os.Getenv("PROTOBUF_MCP_MAX_PARALLELISM"); val != "" {
		if parallelism, err := strconv.Atoi(val); err == nil && parallelism > 0 {
			return parallelism
		}
	}
	return 4 // default value
}

// ProtobufProject represents a compiled protobuf project
type ProtobufProject struct {
	ProjectRoot    string
	Config         *config.ProjectConfig
	CompiledProtos linker.Files
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

	// Create source resolver with configured import paths
	var importPaths []string
	for _, importPath := range cfg.ImportPaths {
		if filepath.IsAbs(importPath) {
			// Absolute path - use as is
			importPaths = append(importPaths, importPath)
		} else {
			// Relative path - resolve from project root
			importPaths = append(importPaths, filepath.Join(projectRoot, importPath))
		}
	}

	resolver := &protocompile.SourceResolver{
		ImportPaths: importPaths,
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

// CompileProtos compiles proto files with the given configuration
func CompileProtos(ctx context.Context, rootDir string, protoFiles []string, importPaths []string) (linker.Files, error) {
	if len(protoFiles) == 0 {
		return nil, fmt.Errorf("no proto files found in configured paths")
	}

	var absImportPaths []string
	for _, importPath := range importPaths {
		if filepath.IsAbs(importPath) {
			// Absolute path - use as is
			absImportPaths = append(absImportPaths, importPath)
		} else {
			// Relative path - resolve from rootDir
			absImportPaths = append(absImportPaths, filepath.Join(rootDir, importPath))
		}
	}

	resolver := &protocompile.SourceResolver{
		ImportPaths: absImportPaths,
	}

	// Wrap with standard imports for well-known types
	resolverWithStdImports := protocompile.WithStandardImports(resolver)

	// Create compiler with our resolver
	compiler := protocompile.Compiler{
		Resolver:       resolverWithStdImports,
		MaxParallelism: getMaxParallelism(),
		SourceInfoMode: protocompile.SourceInfoExtraOptionLocations,
	}

	var absProtoFiles []string

	for _, file := range protoFiles {
		absProtoFiles = append(absProtoFiles, filepath.Join(rootDir, file))
	}

	var relativeProtoFiles []string
	for _, file := range absProtoFiles {
		for _, importPath := range absImportPaths {
			relPath, err := filepath.Rel(importPath, file)
			if err == nil && !filepath.IsAbs(relPath) && relPath != file {
				relativeProtoFiles = append(relativeProtoFiles, relPath)
				break
			}
		}
	}

	// Compile all files together - protocompile will resolve dependencies automatically
	files, err := compiler.Compile(ctx, relativeProtoFiles...)
	if err != nil {
		return nil, fmt.Errorf("failed to compile proto files: %w", err)
	}

	return files, nil
}

// CompileProtos compiles all proto files in the project
func (p *ProtobufProject) CompileProtos(ctx context.Context) error {
	// Get import paths from config
	var importPaths []string
	for _, importPath := range p.Config.ImportPaths {
		if filepath.IsAbs(importPath) {
			// Absolute path - use as is
			importPaths = append(importPaths, importPath)
		} else {
			// Relative path - resolve from project root
			importPaths = append(importPaths, filepath.Join(p.ProjectRoot, importPath))
		}
	}

	// Use the standalone compile function with proto files (already relative)
	compiledProtos, err := CompileProtos(ctx, p.ProjectRoot, p.protoFiles, importPaths)
	if err != nil {
		return err
	}

	p.CompiledProtos = compiledProtos
	return nil
}

func (p *ProtobufProject) GetFileDescriptorSet() (linker.Files, error) {
	if p.CompiledProtos == nil {
		return nil, fmt.Errorf("project not compiled yet, call CompileProtos first")
	}
	return p.CompiledProtos, nil
}

// GetServices extracts all services from compiled protos
func (p *ProtobufProject) GetServices() ([]protoreflect.ServiceDescriptor, error) {
	if p.CompiledProtos == nil {
		return nil, fmt.Errorf("project not compiled yet, call CompileProtos first")
	}

	var services []protoreflect.ServiceDescriptor

	for _, file := range p.CompiledProtos {
		fileDesc := protoreflect.FileDescriptor(file)
		for i := 0; i < fileDesc.Services().Len(); i++ {
			serviceDesc := fileDesc.Services().Get(i)
			services = append(services, serviceDesc)
		}
	}

	return services, nil
}

// GetMessages extracts all messages from compiled protos
func (p *ProtobufProject) GetMessages() ([]protoreflect.MessageDescriptor, error) {
	if p.CompiledProtos == nil {
		return nil, fmt.Errorf("project not compiled yet, call CompileProtos first")
	}

	var messages []protoreflect.MessageDescriptor

	for _, file := range p.CompiledProtos {
		fileDesc := protoreflect.FileDescriptor(file)
		for i := 0; i < fileDesc.Messages().Len(); i++ {
			messageDesc := fileDesc.Messages().Get(i)
			messages = append(messages, messageDesc)
		}
	}

	return messages, nil
}

// GetEnums extracts all enums from compiled protos
func (p *ProtobufProject) GetEnums() ([]protoreflect.EnumDescriptor, error) {
	if p.CompiledProtos == nil {
		return nil, fmt.Errorf("project not compiled yet, call CompileProtos first")
	}

	var enums []protoreflect.EnumDescriptor

	for _, file := range p.CompiledProtos {
		fileDesc := protoreflect.FileDescriptor(file)
		for i := 0; i < fileDesc.Enums().Len(); i++ {
			enumDesc := fileDesc.Enums().Get(i)
			enums = append(enums, enumDesc)
		}
	}

	return enums, nil
}
