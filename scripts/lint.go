package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	fmt.Println("Running golangci-lint...")

	// Check if golangci-lint is installed
	_, err := exec.LookPath("golangci-lint")
	if err != nil {
		fmt.Println("Error: golangci-lint not found in PATH.")
		fmt.Println("Please install it: https://golangci-lint.run/usage/install/#local-installation")
		os.Exit(1)
	}

	// Run golangci-lint
	cmd := exec.Command("golangci-lint", "run", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = "." // Run from the project root

	err = cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "golangci-lint found issues or exited with an error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("golangci-lint completed with no issues found.")
}
