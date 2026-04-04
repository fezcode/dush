package main

import (
	"fmt"
	"os"
	"path/filepath"

	"dush/internal/evaluator"
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
			fmt.Println("dush - A modern shell written in Go.")
			fmt.Println("Usage: dush [file] [flags]")
			fmt.Println("\nFlags:")
			fmt.Println("  -h, --help     Show help")
			fmt.Println("  -v, --version  Print version")
			fmt.Println("\nRun without arguments for interactive mode.")
			fmt.Println("Run with a file path to execute a script.")
			return
		}
	}

	// Bootstrap the application
	Bootstrap()

	// Set shell version for environment variables
	evaluator.ShellVersion = Version

	// Non-interactive mode: dush script.dush
	// Only sources ~/.dush/env (not ~/.dush/is)
	if len(os.Args) > 1 {
		env := evaluator.NewEnvironment()
		env.Stdout = os.Stdout
		env.Stderr = os.Stderr

		// Source ~/.dush/env (always loaded)
		if home, err := os.UserHomeDir(); err == nil {
			envPath := filepath.Join(home, ".dush", "env")
			if _, err := os.Stat(envPath); err == nil {
				evaluator.EvalSource(envPath, env)
			}
		}

		result := evaluator.EvalSource(os.Args[1], env)
		if result != nil && result.Inspect() == "false" {
			os.Exit(1)
		}
		return
	}

	// Interactive mode
	fmt.Printf("Welcome to dush v%s!\n", Version)
	fmt.Println("Type 'exit' or 'quit' to exit.")
	repl.Start(os.Stdin, os.Stdout, os.Stderr)
}
