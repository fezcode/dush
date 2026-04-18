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

var previousDir string

// Execute changes the shell's current working directory.
func (c *CDCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) error {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Fprintln(out, "Usage: cd [directory]\nChange the current working directory.\n\nWith no arguments, changes to the home directory.\nUse 'cd -' to return to the previous directory.")
		return nil
	}
	if len(args) > 1 {
		return fmt.Errorf("cd: too many arguments")
	}

	appInstance := app.GetApp()
	currentDir := appInstance.GetCurrentDir()

	var targetPath string
	if len(args) == 0 {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("could not get home directory: %v", err)
		}
		targetPath = homeDir
	} else if args[0] == "-" {
		if previousDir == "" {
			return fmt.Errorf("no previous directory")
		}
		targetPath = previousDir
		fmt.Fprintln(out, targetPath)
	} else {
		targetPath = args[0]
	}

	if err := os.Chdir(targetPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("'%s': no such directory", targetPath)
		}
		return fmt.Errorf("'%s': %v", targetPath, err)
	}

	newWD, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to determine new working directory: %v", err)
	}

	previousDir = currentDir

	if err := appInstance.SetCurrentDir(newWD); err != nil {
		return fmt.Errorf("%v", err)
	}

	return nil
}
func init() {
	RegisterBuiltin("cd", &CDCommand{})
}
