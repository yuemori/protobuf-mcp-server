package main

import (
	"fmt"
	"os"

	"github.com/yuemori/protobuf-mcp-server/internal/mcp"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	if len(os.Args) < 2 {
		showHelp()
		os.Exit(1)
	}

	command := os.Args[1]

	// Handle help command
	if command == "help" || command == "--help" || command == "-h" {
		showHelp()
		os.Exit(0)
	}

	// Handle version command
	if command == "version" || command == "--version" || command == "-v" {
		fmt.Printf("protobuf-mcp version %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	switch command {
	case "server":
		if err := mcp.StartServer(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		showHelp()
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Println("Protobuf MCP Server - A Model Context Protocol server for Protocol Buffers")
	fmt.Println()
	fmt.Println("Usage: protobuf-mcp <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  init [project-path]  - Initialize project configuration")
	fmt.Println("  server               - Start MCP server")
	fmt.Println("  help                 - Show this help message")
	fmt.Println("  version              - Show version information")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  PROTOBUF_MCP_MAX_PARALLELISM  - Set maximum parallel compilation (default: 4)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  protobuf-mcp init                    # Initialize in current directory")
	fmt.Println("  protobuf-mcp init /path/to/project   # Initialize in specific directory")
	fmt.Println("  protobuf-mcp server                  # Start MCP server")
	fmt.Println("  protobuf-mcp help                    # Show this help")
}
