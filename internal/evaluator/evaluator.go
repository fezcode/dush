package evaluator

import (
	"context"
	"dush/internal/app"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ExecuteExternal runs an external command.
func ExecuteExternal(ctx context.Context, cmdName string, args []string, out io.Writer, errOut io.Writer) error {
	appInstance := app.GetApp()
	currentDir := appInstance.GetCurrentDir()

	fullPath := cmdName
	// If the command doesn't have a path separator, check the current directory
	if !strings.ContainsAny(cmdName, "/\\") {
		localPath := filepath.Join(currentDir, cmdName)
		if info, err := os.Stat(localPath); err == nil && !info.IsDir() {
			fullPath = localPath
		}
	}

	cmd := exec.CommandContext(ctx, fullPath, args...)
	cmd.Dir = currentDir
	cmd.Stdout = out
	cmd.Stderr = errOut
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
