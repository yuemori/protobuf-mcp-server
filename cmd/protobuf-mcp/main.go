package main

import (
	"fmt"
	"os"

	"github.com/yuemori/protobuf-mcp-server/internal/cli"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: protobuf-mcp <command>")
		fmt.Println("Commands:")
		fmt.Println("  init [project-path]  - Initialize project configuration")
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "init":
		if err := cli.InitCommand(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}