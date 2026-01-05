package repl

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall" // New import for specific signals

	"dush/internal/app"
	"dush/internal/builtins"
	"dush/internal/config"
	"dush/internal/utils"
)

// Start starts the Read-Eval-Print Loop.
// It takes an io.Reader for input, an io.Writer for output, and an io.Writer for error output.
func Start(in io.Reader, out io.Writer, errOut io.Writer) {
	scanner := bufio.NewScanner(in)

	// Create a context for the entire REPL lifecycle, cancelled on SIGTERM/SIGHUP
	replCtx, replCancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGHUP)
	defer replCancel() // Ensure this context is cancelled when Start returns

	// Load history at the start of the REPL
	utils.LoadHistory()
	// Ensure history is saved when the REPL exits
	defer utils.SaveHistory()

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
		// Check if the main REPL context has been cancelled
		select {
		case <-replCtx.Done():
			fmt.Fprintf(out, "\nExiting dush REPL gracefully...\n")
			return
		default:
			// Continue
		}

		currentCWD := appInstance.GetCurrentDir()
		displayDirName := utils.GetDisplayDirName(currentCWD)

		// Construct the dynamic prompt using App's currentCWD
		promptLine := fmt.Sprintf("%s %s@%s%s ", cfg.PromptPrefix, cfg.UserName, displayDirName, cfg.PromptSuffix)
		fmt.Fprintf(out, promptLine)

		// Create a cancellable context for the current command
		cmdCtx, cmdCancel := context.WithCancel(replCtx) // Child context of replCtx
		// Defer cmdCancel to ensure it's always called, but it might be called earlier by signal handler
		defer func() {
			// This defer will only run when the loop iteration finishes.
			// It's crucial to call cmdCancel() to release resources even if not explicitly cancelled by Ctrl+C.
			cmdCancel()
		}()

		// Set up signal handling for the current command
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt) // Capture Ctrl+C

		go func() {
			select {
			case <-sigChan: // Ctrl+C received
				fmt.Fprintln(errOut, "\nInterrupt received. Cancelling current command...")
				cmdCancel() // Cancel the context for the current command
			case <-cmdCtx.Done():
				// Command finished or was cancelled by other means, stop listening to sigChan
			}
		}()

		scanned := scanner.Scan()
		// After scanner.Scan() returns, stop listening for signals on sigChan for this command
		// and reset signal.Notify to its default or a new context for the next command.
		// However, signal.Stop is tricky with multiple listeners.
		// A simpler approach for cleanup is to rely on cmdCancel() and the goroutine
		// exiting when cmdCtx.Done() is closed. The signal.Notify will stay active.

		if !scanned {
			fmt.Fprintf(out, "Exiting dush REPL.\n") // Inform user on EOF
			return                                   // EOF or error
		}

		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue // Skip empty lines
		}

		// Add command to history before processing it
		utils.AddCommand(trimmedLine)

		parts := strings.Fields(trimmedLine)
		cmdName := parts[0]
		args := parts[1:]

		// --- Start Alias Expansion ---
		cfg := config.GetConfig() // Get current config to access aliases
		if expandedValue, ok := cfg.Aliases[cmdName]; ok {
			// If the command name is an alias, expand it
			expandedParts := strings.Fields(expandedValue)
			if len(expandedParts) > 0 {
				cmdName = expandedParts[0]
				// Append original args to expanded alias args
				args = append(expandedParts[1:], args...)
			}
		}
		// --- End Alias Expansion ---

		if cmdName == "exit" || cmdName == "quit" {
			fmt.Fprintf(out, "Exiting dush REPL.\n")
			return
		}

		// Check and execute built-in commands, passing the command context
		if builtins.RunBuiltin(cmdCtx, cmdName, args, out, errOut) {
			// Builtin handled, continue loop. cmdCancel will be called by defer.
		} else {
			// If not a built-in command, attempt to run as an external command
			// For now, this is a placeholder.
			fmt.Fprintf(out, "Command not found: %s\n", cmdName)
			// In a future step, we will implement logic to search PATH and execute external commands here.
		}

		// After command execution, ensure no lingering signal goroutine tries to cancel a context
		// that's about to be recreated. The current goroutine naturally exits when cmdCtx.Done() is closed by cmdCancel().
		// No explicit signal.Stop is needed if we're constantly notifying the same channel for all signals.
		// However, it's safer to ensure we're not stacking signal listeners.
		// signal.Reset(os.Interrupt) would remove all registrations for os.Interrupt.
		// A more precise approach is to re-create sigChan for each command, ensuring only one listener per context.
		// This is done by `sigChan := make(chan os.Signal, 1); signal.Notify(sigChan, os.Interrupt)` inside the loop.
	}
}
