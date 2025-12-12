package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	fmt.Println("Running go fmt...")

	// Find all Go modules in the project
	// For simplicity, we assume the current directory is the root of the main module
	// and apply go fmt to all subdirectories.
	cmd := exec.Command("go", "fmt", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = "." // Run from the project root

	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running go fmt: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("go fmt completed.")

	// Check if any files were changed by go fmt
	// This helps in CI/CD or pre-commit hooks to ensure formatting is applied
	gitStatusCmd := exec.Command("git", "status", "--porcelain")
	output, err := gitStatusCmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking git status: %v\n", err)
		os.Exit(1)
	}

	if len(output) > 0 {
		fmt.Println("go fmt made changes to the following files:")
		// Split output by newline and print files that start with M (modified)
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, " M") || strings.HasPrefix(line, "M") {
				// Extract file path, typically " M path/to/file.go"
				fileName := strings.TrimSpace(line[2:])
				fmt.Printf("  %s\n", fileName)
			}
		}
		os.Exit(1) // Exit with error if formatting changes were made
	}

	fmt.Println("All Go files are correctly formatted.")
}
