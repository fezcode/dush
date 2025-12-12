package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"dush/internal/config" // New import
)

// Start starts the Read-Eval-Print Loop.
// It takes an io.Reader for input and an io.Writer for output.
func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	// Get the configuration once at the start of REPL
	// The config is already bootstrapped in main, so GetConfig will return the existing instance.
	cfg := config.GetConfig()

	for {
		// Get current working directory
		cwd, err := os.Getwd()
		if err != nil {
			// If we can't get CWD, fallback to a generic prompt
			fmt.Fprintf(out, "%s %s%s", cfg.PromptPrefix, cfg.UserName, cfg.PromptSuffix)
		} else {
			// Construct the dynamic prompt
			// Example: $ user@base_dir >>
			promptLine := fmt.Sprintf("%s %s@%s%s ", cfg.PromptPrefix, cfg.UserName, filepath.Base(cwd), cfg.PromptSuffix)
			fmt.Fprintf(out, promptLine)
		}

		scanned := scanner.Scan()
		if !scanned {
			return // EOF or error
		}

		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == "exit" || trimmedLine == "quit" {
			fmt.Fprintf(out, "Exiting dush REPL.\n")
			return
		}

		// For now, just echo the input
		fmt.Fprintf(out, "Echo: %s\n", line)
	}
}
