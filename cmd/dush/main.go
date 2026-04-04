package main

import (
	"fmt"
	"os"
	"strings"

	"dush/internal/config"
	"dush/internal/evaluator"
	"dush/internal/repl"
)

var Version = "dev"

func main() {
	// Parse flags and collect positional args
	var scriptFile string
	var envPath, isPath, historyPath string

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "-v" || arg == "--version":
			fmt.Printf("dush v%s\n", Version)
			return
		case arg == "-h" || arg == "--help":
			printHelp()
			return
		case arg == "--env" && i+1 < len(args):
			i++
			envPath = args[i]
		case strings.HasPrefix(arg, "--env="):
			envPath = strings.TrimPrefix(arg, "--env=")
		case arg == "--is" && i+1 < len(args):
			i++
			isPath = args[i]
		case strings.HasPrefix(arg, "--is="):
			isPath = strings.TrimPrefix(arg, "--is=")
		case arg == "--history" && i+1 < len(args):
			i++
			historyPath = args[i]
		case strings.HasPrefix(arg, "--history="):
			historyPath = strings.TrimPrefix(arg, "--history=")
		case !strings.HasPrefix(arg, "-"):
			if scriptFile == "" {
				scriptFile = arg
			}
		default:
			fmt.Fprintf(os.Stderr, "unknown flag: %s\n", arg)
			os.Exit(1)
		}
	}

	// Bootstrap the application
	Bootstrap()

	// Set shell version for environment variables
	evaluator.ShellVersion = Version

	// Apply flag overrides to paths, then resolve defaults for the rest
	if envPath != "" {
		config.ShellPaths.Env = envPath
	}
	if isPath != "" {
		config.ShellPaths.Is = isPath
	}
	if historyPath != "" {
		config.ShellPaths.History = historyPath
	}
	config.ShellPaths.Resolve()

	// Non-interactive mode: dush script.dush
	// Only sources env (not is)
	if scriptFile != "" {
		env := evaluator.NewEnvironment()
		env.Stdout = os.Stdout
		env.Stderr = os.Stderr

		// Source env file (always loaded)
		if ep := config.ShellPaths.Env; ep != "" {
			if _, err := os.Stat(ep); err == nil {
				evaluator.EvalSource(ep, env)
			}
		}

		result := evaluator.EvalSource(scriptFile, env)
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

func printHelp() {
	fmt.Println("dush - A modern shell written in Go.")
	fmt.Println("Usage: dush [file] [flags]")
	fmt.Println("\nFlags:")
	fmt.Println("  -h, --help            Show help")
	fmt.Println("  -v, --version         Print version")
	fmt.Println("  --env <path>          Override env file path (default: ~/.dush/env)")
	fmt.Println("  --is <path>           Override is file path (default: ~/.dush/is)")
	fmt.Println("  --history <path>      Override history file path (default: ~/.dush/history)")
	fmt.Println("\nRun without arguments for interactive mode.")
	fmt.Println("Run with a file path to execute a script.")
}
