# Protobuf MCP Server

[![CI](https://github.com/yuemori/protobuf-mcp-server/workflows/CI/badge.svg)](https://github.com/yuemori/protobuf-mcp-server/actions)
[![CodeQL](https://github.com/yuemori/protobuf-mcp-server/workflows/CodeQL/badge.svg)](https://github.com/yuemori/protobuf-mcp-server/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/yuemori/protobuf-mcp-server)](https://goreportcard.com/report/github.com/yuemori/protobuf-mcp-server)
[![Go Version](https://img.shields.io/badge/go-1.23-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

A Model Context Protocol (MCP) server that provides semantic information about Protocol Buffers files, enabling AI assistants to work with protobuf schemas without RAG search, reducing token consumption and improving context efficiency.

## Features

- **Project Management**: Initialize and manage protobuf projects with configuration
- **Service Discovery**: List all services with detailed method information
- **Schema Analysis**: Get comprehensive schema information including messages, enums, and services
- **Pattern Search**: Search for specific patterns across protobuf files (coming soon)
- **Streaming Support**: Detect and handle streaming methods
- **Filtering**: Filter results by message, service, or enum types

## Installation

### Prerequisites

- Go 1.23 or later
- Protocol Buffers project with `.proto` files

### Quick Start (Recommended)

#### Option 1: Direct Run (No Installation)

```bash
# Run directly from source (requires Go)
go run github.com/yuemori/protobuf-mcp-server/cmd/protobuf-mcp@latest server
```

#### Option 2: Install Globally

```bash
# Install the latest version globally
go install github.com/yuemori/protobuf-mcp-server/cmd/protobuf-mcp@latest

# Verify installation
protobuf-mcp --help
```

#### Option 3: Build from Source

```bash
# Clone the repository
git clone https://github.com/yuemori/protobuf-mcp-server.git
cd protobuf-mcp-server

# Install dependencies
go mod tidy

# Build the server
go build -o protobuf-mcp ./cmd/protobuf-mcp
```

### Initialize Your Project

After installation, initialize your protobuf project:

```bash
# Navigate to your protobuf project directory
cd /path/to/your/protobuf/project

# Initialize the project
protobuf-mcp init

# Or using go run (if not installed globally)
go run github.com/yuemori/protobuf-mcp-server/cmd/protobuf-mcp@latest init
```

This creates a `.protobuf-mcp.yml` configuration file with default settings:

```yaml
proto_files:
  - "proto/**/*.proto"
import_paths:
  - "."
```

## Configuration

### Claude Code Setup

1. **Add to your Claude Code configuration**:

```json
{
  "mcpServers": {
    "protobuf-mcp": {
      "command": "go",
      "args": [
        "run",
        "github.com/yuemori/protobuf-mcp-server/cmd/protobuf-mcp@latest",
        "server"
      ],
      "env": {}
    }
  }
}
```

**Or if you installed globally**:

```json
{
  "mcpServers": {
    "protobuf-mcp": {
      "command": "protobuf-mcp",
      "args": ["server"],
      "env": {}
    }
  }
}
```

2. **Initialize your protobuf project** (see Installation section above)

### Cursor Setup

1. **Add to your Cursor settings**:

```json
{
  "mcp": {
    "servers": {
      "protobuf-mcp": {
        "command": "go",
        "args": [
          "run",
          "github.com/yuemori/protobuf-mcp-server/cmd/protobuf-mcp@latest",
          "server"
        ]
      }
    }
  }
}
```

**Or if you installed globally**:

```json
{
  "mcp": {
    "servers": {
      "protobuf-mcp": {
        "command": "protobuf-mcp",
        "args": ["server"]
      }
    }
  }
}
```

2. **Initialize your protobuf project** (see Installation section above)

## Usage

### Configuration File

The `.protobuf-mcp.yml` configuration file supports the following options:

```yaml
# Glob patterns to match proto files
proto_files:
  - "proto/**/*.proto"
  - "api/**/*.proto"

# Import paths for protobuf compiler
import_paths:
  - "."
  - "proto"
  - "third_party"
```

#### Configuration Options

- **proto_files**: Glob patterns to match your `.proto` files

  - Supports `**` for recursive directory matching
  - Can specify multiple patterns
  - Relative paths are resolved from the config file location

- **import_paths**: Directories where the protobuf compiler should look for imported files
  - Used when resolving `import` statements in proto files
  - Defaults to `["."]` if not specified
  - Supports both relative and absolute paths

#### Re-initialize Project

To update the configuration or re-initialize:

```bash
# Re-initialize with updated settings
go run github.com/yuemori/protobuf-mcp-server/cmd/protobuf-mcp@latest init

# Or if installed globally
protobuf-mcp init
```

### Available Tools

The MCP server provides the following tools:

#### 1. `activate_project`

Activates a protobuf project by loading configuration and compiling proto files.

**Parameters:**

- `project_path` (string): Path to the protobuf project directory

**Example:**

```json
{
  "name": "activate_project",
  "arguments": {
    "project_path": "/path/to/your/protobuf/project"
  }
}
```

#### 2. `list_services`

Lists all services in the activated protobuf project.

**Parameters:** None

**Example:**

```json
{
  "name": "list_services",
  "arguments": {}
}
```

#### 3. `get_schema`

Get detailed schema information from the activated protobuf project.

**Parameters:**

- `message_types` (array, optional): Filter by specific message types
- `service_types` (array, optional): Filter by specific service types
- `enum_types` (array, optional): Filter by specific enum types
- `include_file_info` (boolean, optional): Include file information

**Example:**

```json
{
  "name": "get_schema",
  "arguments": {
    "message_types": ["User", "Product"],
    "include_file_info": true
  }
}
```

### Example Workflow

1. **Initialize a project**:

```bash
cd /path/to/your/protobuf/project

# Using go run (recommended)
go run github.com/yuemori/protobuf-mcp-server/cmd/protobuf-mcp@latest init

# Or if installed globally
protobuf-mcp init
```

2. **Activate the project** (via MCP):

```json
{
  "name": "activate_project",
  "arguments": {
    "project_path": "/path/to/your/protobuf/project"
  }
}
```

3. **List services**:

```json
{
  "name": "list_services",
  "arguments": {}
}
```

4. **Get detailed schema**:

```json
{
  "name": "get_schema",
  "arguments": {
    "include_file_info": true
  }
}
```

## Advanced Configuration

### Environment Variables

- `PROTOBUF_MCP_LOG_LEVEL`: Set log level (debug, info, warn, error)
- `PROTOBUF_MCP_MAX_PARALLELISM`: Set maximum parallel compilation (default: 4)

## Supported Features

### Protocol Buffers Support

- **Messages**: Complete field information with types, labels, and options
- **Services**: Method details including input/output types and streaming
- **Enums**: Value information with numbers and options
- **Imports**: Dependency resolution and cross-file references
- **Options**: Custom options and well-known types

### Streaming Support

- **Client Streaming**: Detect client-side streaming methods
- **Server Streaming**: Detect server-side streaming methods
- **Bidirectional Streaming**: Detect bidirectional streaming methods

### Filtering and Search

- **Type Filtering**: Filter by message, service, or enum types
- **Pattern Search**: Search for specific patterns (coming soon)
- **Context Display**: Show relevant context around matches

## Troubleshooting

### Common Issues

1. **"Project not initialized" error**:

   - Run `go run github.com/yuemori/protobuf-mcp-server/cmd/protobuf-mcp@latest init` in your project directory
   - Or if installed globally: `protobuf-mcp init`
   - Ensure `.protobuf-mcp.yml` exists

2. **"No proto files found" error**:

   - Check your `proto_files` configuration patterns
   - Ensure `.proto` files exist in specified paths
   - Verify glob patterns match your file structure

3. **Import resolution errors**:

   - Add import paths to `import_paths` in configuration
   - Ensure imported files are accessible from specified import paths

4. **Compilation errors**:
   - Check proto file syntax
   - Verify all dependencies are available
   - Check import paths are correctly configured

### Debug Mode

Enable debug logging using environment variables:

```bash
export PROTOBUF_MCP_LOG_LEVEL=debug
protobuf-mcp server
```

This will provide detailed information about:

- Proto file discovery and pattern matching
- Compilation process
- Import path resolution
- Error details

## CLI Commands

The protobuf-mcp server provides several CLI commands for project management:

### `init` - Initialize Project

Initialize a new protobuf project with configuration.

```bash
# Initialize project
go run github.com/yuemori/protobuf-mcp-server/cmd/protobuf-mcp@latest init

# Or if installed globally
protobuf-mcp init
```

**Options:**

- Creates `.protobuf-mcp.yml` configuration file
- Detects existing configuration and prompts for overwrite
- Sets up default paths and patterns

**Example Output:**

```
Initializing protobuf project...
Created .protobuf-mcp.yml
Project initialized successfully!
```

### `server` - Start MCP Server

Start the MCP server for use with Claude Code or Cursor.

```bash
# Start server
go run github.com/yuemori/protobuf-mcp-server/cmd/protobuf-mcp@latest server

# Or if installed globally
protobuf-mcp server
```

**Features:**

- JSON-RPC over stdio communication
- Automatic project detection
- Real-time protobuf compilation
- Error handling and logging

### `help` - Show Help

Display help information and available commands.

```bash
# Show help
go run github.com/yuemori/protobuf-mcp-server/cmd/protobuf-mcp@latest --help

# Or if installed globally
protobuf-mcp --help
```

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/tools
go test ./internal/mcp

# Run with verbose output
go test ./... -v
```

### Building

```bash
# Build for current platform
go build -o protobuf-mcp ./cmd/protobuf-mcp

# Build for specific platforms
GOOS=linux GOARCH=amd64 go build -o protobuf-mcp-linux ./cmd/protobuf-mcp
GOOS=windows GOARCH=amd64 go build -o protobuf-mcp.exe ./cmd/protobuf-mcp

# Or install globally for development
go install ./cmd/protobuf-mcp
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite
6. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [protocompile](https://github.com/bufbuild/protocompile) for Protocol Buffers compilation
- Follows the [Model Context Protocol](https://modelcontextprotocol.io/) specification
- Inspired by the need for efficient protobuf schema analysis in AI workflows
