package builtins

import (
	"context"
	"fmt"
	"io"
	"os"

	"dush/internal/app"
)

// CDCommand implements the Command interface for the 'cd' builtin.
type CDCommand struct{}

// Execute changes the shell's current working directory.
func (c *CDCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) error {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Fprintln(out, "Usage: cd [directory]\nChange the current working directory.\nWith no arguments, changes to the home directory.")
		return nil
	}

	appInstance := app.GetApp()

	var targetPath string
	if len(args) == 0 {
		// No argument given, change to home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cd: could not get home directory: %w", err)
		}
		targetPath = homeDir
	} else {
		targetPath = args[0]
	}

	// Delegate path resolution to the OS
	if err := os.Chdir(targetPath); err != nil {
		return fmt.Errorf("cd: %w", err)
	}

	// Get the new absolute path to update our state
	newWD, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cd: failed to determine new working directory: %w", err)
	}

	// Update App state
	if err := appInstance.SetCurrentDir(newWD); err != nil {
		return fmt.Errorf("cd: %w", err)
	}

	return nil
}
func init() {
	RegisterBuiltin("cd", &CDCommand{})
}
