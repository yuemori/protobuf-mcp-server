# MCP Server Tests

## Overview

This directory contains tests for the MCP (Model Context Protocol) server implementation, organized by tool and functionality.

## Test Structure

### Unit Tests (`server_test.go`)

- Tests for server creation and initialization
- Tool registration tests
- Project manager state management tests
- Basic tool creation tests

### Integration Tests (Tool-Specific)

#### `activate_project_integration_test.go`

- **SuccessfulActivation**: Tests successful project activation with proto files
- **NonExistentPath**: Tests error handling for invalid project paths
- **MissingProjectPath**: Tests error handling for missing required parameters

#### `list_services_integration_test.go`

- **NoProjectActivated**: Tests error handling when no project is activated
- **SuccessfulListServices**: Tests successful service listing after project activation

#### `get_schema_integration_test.go`

- **GetAllSchema**: Tests retrieving complete schema information
- **FilterByMessageTypes**: Tests filtering by specific message types
- **FilterByEnumTypes**: Tests filtering by specific enum types
- **IncludeFileInfo**: Tests including file information in schema response
- **NoProjectActivated**: Tests error handling when no project is activated

#### `server_initialization_integration_test.go`

- **Initialize**: Tests server initialization and basic capabilities
- **ListTools**: Tests listing available tools
- **ServerCapabilities**: Tests server capabilities and info

#### `complete_workflow_integration_test.go`

- **Step1_Initialize**: Server initialization
- **Step2_ListTools**: List available tools
- **Step3_ActivateProject**: Activate a protobuf project
- **Step4_ListServices**: List services in the activated project
- **Step5_GetCompleteSchema**: Get complete schema information
- **Step6_GetFilteredSchema**: Get filtered schema with specific types
- **Step7_ErrorHandling**: Test error handling scenarios

## Key Features

### In-Process Client Testing

All integration tests use `client.NewInProcessClient` from mcp-go, which provides:

1. **Faster Execution**: No process spawning or binary building
2. **Better Reliability**: No IPC issues or timing problems
3. **Real Protocol Testing**: Still tests actual MCP JSON-RPC protocol
4. **Easier Debugging**: Everything runs in the same process

### Comprehensive Coverage

- **Tool-Specific Tests**: Each MCP tool has dedicated integration tests
- **Error Handling**: Tests cover various error scenarios
- **Complete Workflows**: End-to-end testing of typical usage patterns
- **Filtering**: Tests schema filtering capabilities
- **State Management**: Tests project activation and state persistence

## Running Tests

```bash
# Run all MCP tests
go test ./internal/mcp -v

# Run specific tool tests
go test ./internal/mcp -run TestActivateProjectTool_Integration -v
go test ./internal/mcp -run TestListServicesTool_Integration -v
go test ./internal/mcp -run TestGetSchemaTool_Integration -v

# Run workflow tests
go test ./internal/mcp -run TestCompleteWorkflow_Integration -v

# Run initialization tests
go test ./internal/mcp -run TestServerInitialization_Integration -v

# Run all tests
go test ./...
```

## Test Organization Benefits

1. **Maintainability**: Each tool has its own test file for easier maintenance
2. **Clarity**: Test names clearly indicate what functionality is being tested
3. **Isolation**: Tests can be run independently for specific tools
4. **Comprehensive**: Both individual tool tests and complete workflow tests
5. **Performance**: Fast execution with in-process client approach
