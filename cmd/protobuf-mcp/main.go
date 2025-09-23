package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: protobuf-mcp <command>")
		fmt.Println("Commands:")
		fmt.Println("  init [project-path]  - Initialize project configuration")
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "init":
		// TODO: Implement init command
		fmt.Println("Init command not implemented yet")
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}