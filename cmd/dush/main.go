package main

import (
	"fmt"
	"os"

	"dush/internal/repl"
)

var Version = "dev"

func main() {
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if arg == "-v" || arg == "--version" {
			fmt.Printf("dush v%s\n", Version)
			return
		}
		if arg == "-h" || arg == "--help" {
			fmt.Println("dush - A custom terminal shell written in Go.")
			fmt.Println("Usage: dush [flags]")
			fmt.Println("\nFlags:")
			fmt.Println("  -h, --help     Show context-sensitive help")
			fmt.Println("  -v, --version  Print version information and quit")
			fmt.Println("\nGitHub Description: A custom terminal shell written in Go.")
			fmt.Println("GitHub Tags: shell, cli, terminal, go, automation")
			return
		}
	}

	// Bootstrap the application
	Bootstrap() // Call the bootstrap function without arguments

	fmt.Printf("Welcome to dush v%s!\n", Version)
	fmt.Println("Type 'exit' or 'quit' to exit.")
	repl.Start(os.Stdin, os.Stdout, os.Stderr)
}
