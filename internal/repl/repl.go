package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"            // os is needed for os.Getwd()
	"path/filepath" // Keep for filepath.Base
	"strings"

	"dush/internal/app"
	"dush/internal/builtins"
	"dush/internal/config"
)

// Start starts the Read-Eval-Print Loop.
// It takes an io.Reader for input, an io.Writer for output, and an io.Writer for error output.
func Start(in io.Reader, out io.Writer, errOut io.Writer) {
	scanner := bufio.NewScanner(in)

	// Get the singleton App instance
	appInstance := app.GetApp()

	// Initialize currentCWD with the actual OS CWD at startup
	initialCWD, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(errOut, "Error getting initial working directory: %v. Defaulting to '/'.\n", err)
		initialCWD = "/" // Fallback if getting CWD fails
	}
	appInstance.SetCurrentDir(initialCWD) // Use the setter to initialize

	// Get the configuration once at the start of REPL
	cfg := config.GetConfig()

	for {
		// Construct the dynamic prompt using App's currentCWD
		promptLine := fmt.Sprintf("%s %s@%s%s ", cfg.PromptPrefix, cfg.UserName, filepath.Base(appInstance.GetCurrentDir()), cfg.PromptSuffix)
		fmt.Fprintf(out, promptLine)

		scanned := scanner.Scan()
		if !scanned {
			return // EOF or error
		}

		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue // Skip empty lines
		}

		parts := strings.Fields(trimmedLine)
		cmdName := parts[0]
		args := parts[1:]

		if cmdName == "exit" || cmdName == "quit" {
			fmt.Fprintf(out, "Exiting dush REPL.\n")
			return
		}

		// Check and execute built-in commands
		if builtins.RunBuiltin(cmdName, args, out, errOut) {
			continue // If a builtin was executed, skip further processing
		} else {
			// If not a built-in command, attempt to run as an external command
			// For now, this is a placeholder.
			fmt.Fprintf(out, "Command not found: %s\n", cmdName)
			// In a future step, we will implement logic to search PATH and execute external commands here.
		}
	}
}
