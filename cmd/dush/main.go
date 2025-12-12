package main

import (
	"fmt"
	"os"

	"dush/internal/repl"
)

func main() {
	// Bootstrap the application
	Bootstrap() // Call the bootstrap function without arguments

	fmt.Println("Welcome to dush!")
	fmt.Println("Type 'exit' or 'quit' to exit.")
	repl.Start(os.Stdin, os.Stdout, os.Stderr)
}
