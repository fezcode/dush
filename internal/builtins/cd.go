package builtins

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"dush/internal/app" // New import
)

// CDCommand implements the Command interface for the 'cd' builtin.
type CDCommand struct{}

// Execute changes the shell's current working directory.
func (c *CDCommand) Execute(args []string, out io.Writer, errOut io.Writer) error {
	appInstance := app.GetApp() // Get the app singleton

	if len(args) == 0 {
		// No argument given, change to home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cd: could not get home directory: %w", err)
		}
		if err := appInstance.SetCurrentDir(homeDir); err != nil {
			return fmt.Errorf("cd: %w", err)
		}
		return nil
	}

	newPath := args[0]
	currentCWD := appInstance.GetCurrentDir() // Use the app singleton

	// Resolve the new path relative to currentCWD
	var absolutePath string
	if filepath.IsAbs(newPath) {
		absolutePath = filepath.Clean(newPath)
	} else {
		absolutePath = filepath.Clean(filepath.Join(currentCWD, newPath))
	}

	// Check if the path exists and is a directory
	fileInfo, err := os.Stat(absolutePath)
	if err != nil {
		return fmt.Errorf("cd: %s: %w", newPath, err)
	}
	if !fileInfo.IsDir() {
		return fmt.Errorf("cd: %s: Not a directory", newPath)
	}

	// Set the new current working directory
	if err := appInstance.SetCurrentDir(absolutePath); err != nil {
		return fmt.Errorf("cd: %w", err)
	}
	return nil
}

func init() {
	RegisterBuiltin("cd", &CDCommand{})
}
