package compiler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bufbuild/protocompile"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"

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
func CompileProtos(ctx context.Context, protoFiles []string, importPaths []string) (*descriptorpb.FileDescriptorSet, error) {
	if len(protoFiles) == 0 {
		return nil, fmt.Errorf("no proto files found in configured paths")
	}

	// Create source resolver with import paths
	var resolvedImportPaths []string
	for _, importPath := range importPaths {
		if filepath.IsAbs(importPath) {
			// Absolute path - use as is
			resolvedImportPaths = append(resolvedImportPaths, importPath)
		} else {
			// Relative path - resolve from current working directory
			wd, err := os.Getwd()
			if err != nil {
				return nil, fmt.Errorf("failed to get working directory: %w", err)
			}
			// Check if the path exists relative to current directory
			fullPath := filepath.Join(wd, importPath)
			if _, err := os.Stat(fullPath); err == nil {
				resolvedImportPaths = append(resolvedImportPaths, fullPath)
			} else {
				// If not found, try relative to project root (go up to find testdata)
				projectRoot := wd
				for {
					testPath := filepath.Join(projectRoot, importPath)
					if _, err := os.Stat(testPath); err == nil {
						resolvedImportPaths = append(resolvedImportPaths, testPath)
						break
					}
					parent := filepath.Dir(projectRoot)
					if parent == projectRoot {
						// Reached root, use original path
						resolvedImportPaths = append(resolvedImportPaths, fullPath)
						break
					}
					projectRoot = parent
				}
			}
		}
	}

	resolver := &protocompile.SourceResolver{
		ImportPaths: resolvedImportPaths,
	}

	// Wrap with standard imports for well-known types
	resolverWithStdImports := protocompile.WithStandardImports(resolver)

	// Convert proto files to relative paths based on import paths
	var relativeProtoFiles []string
	for _, protoFile := range protoFiles {
		if filepath.IsAbs(protoFile) {
			// Find which import path this file belongs to
			found := false
			for _, importPath := range resolvedImportPaths {
				if relPath, err := filepath.Rel(importPath, protoFile); err == nil && !strings.HasPrefix(relPath, "..") {
					relativeProtoFiles = append(relativeProtoFiles, relPath)
					found = true
					break
				}
			}
			if !found {
				// If not found in any import path, use the filename only
				relativeProtoFiles = append(relativeProtoFiles, filepath.Base(protoFile))
			}
		} else {
			// Already relative, use as is
			relativeProtoFiles = append(relativeProtoFiles, protoFile)
		}
	}

	// Change working directory to the first import path for protocompile
	if len(resolvedImportPaths) > 0 {
		originalWd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current working directory: %w", err)
		}
		defer os.Chdir(originalWd)

		if err := os.Chdir(resolvedImportPaths[0]); err != nil {
			return nil, fmt.Errorf("failed to change to import path directory: %w", err)
		}
	}

	// Create compiler with our resolver
	compiler := protocompile.Compiler{
		Resolver:       resolverWithStdImports,
		MaxParallelism: getMaxParallelism(),
		SourceInfoMode: protocompile.SourceInfoStandard,
	}

	// Compile all files together - protocompile will resolve dependencies automatically
	files, err := compiler.Compile(ctx, relativeProtoFiles...)
	if err != nil {
		return nil, fmt.Errorf("failed to compile proto files: %w", err)
	}

	// Convert compiled files to FileDescriptorSet
	filesList := make([]*descriptorpb.FileDescriptorProto, len(files))
	for i, file := range files {
		// Each file implements protoreflect.FileDescriptor
		// Convert to FileDescriptorProto to preserve all information
		filesList[i] = protodesc.ToFileDescriptorProto(file)
	}

	return &descriptorpb.FileDescriptorSet{
		File: filesList,
	}, nil
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
	compiledProtos, err := CompileProtos(ctx, p.protoFiles, importPaths)
	if err != nil {
		return err
	}

	p.CompiledProtos = compiledProtos
	return nil
}

// GetServices extracts all services from compiled protos
func (p *ProtobufProject) GetServices() ([]*descriptorpb.ServiceDescriptorProto, error) {
	if p.CompiledProtos == nil {
		return nil, fmt.Errorf("project not compiled yet, call CompileProtos first")
	}

	var services []*descriptorpb.ServiceDescriptorProto

	for _, file := range p.CompiledProtos.File {
		services = append(services, file.Service...)
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
