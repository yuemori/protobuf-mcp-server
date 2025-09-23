# Protobuf MCP Server

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

2. **Initialize your protobuf project**:

```bash
# Navigate to your protobuf project directory
cd /path/to/your/protobuf/project

# Initialize the project (using go run)
go run github.com/yuemori/protobuf-mcp-server/cmd/protobuf-mcp@latest init

# Or if installed globally
protobuf-mcp init
```

This creates a `.protobuf-mcp/project.yml` configuration file:

```yaml
root_directory: .
include_paths:
  - .
  - google # For Google API protos
proto_paths:
  - .
  - google
compiler_options: {}
ignored_patterns:
  - "*_test.proto"
  - "tmp/**"
show_logs: false
```

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

2. **Initialize your protobuf project** (same as Claude Code)

## Usage

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

## Configuration Options

### Project Configuration (`.protobuf-mcp/project.yml`)

```yaml
root_directory: . # Project root directory
include_paths: # Paths to search for imported protos
  - .
  - google
  - third_party
proto_paths: # Paths containing your proto files
  - .
  - api
  - internal
compiler_options: {} # Additional compiler options
ignored_patterns: # Patterns to ignore
  - "*_test.proto"
  - "tmp/**"
  - "generated/**"
show_logs: false # Enable debug logging
```

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
   - Ensure `.protobuf-mcp/project.yml` exists

2. **"No proto files found" error**:

   - Check your `proto_paths` configuration
   - Ensure `.proto` files exist in specified paths

3. **Import resolution errors**:

   - Add import paths to `include_paths` in configuration
   - Ensure imported files are accessible

4. **Compilation errors**:
   - Check proto file syntax
   - Verify all dependencies are available
   - Enable `show_logs: true` for detailed error information

### Debug Mode

Enable debug logging in your project configuration:

```yaml
show_logs: true
```

This will provide detailed information about:

- Proto file discovery
- Compilation process
- Import resolution
- Error details

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
