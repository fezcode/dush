package builtins

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
)

type AppsCommand struct{}

func (c *AppsCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) error {
	buildDir := "build"
	entries, err := os.ReadDir(buildDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(out, "No build directory found. Use build scripts first.")
			return nil
		}
		return err
	}

	fmt.Fprintln(out, "Available applications in build/:")
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		isExec := false
		if runtime.GOOS == "windows" {
			if strings.HasSuffix(strings.ToLower(name), ".exe") {
				isExec = true
			}
		} else {
			info, err := entry.Info()
			if err == nil && info.Mode()&0111 != 0 {
				isExec = true
			}
		}

		if isExec {
			fmt.Fprintf(out, "  - %s\n", name)
		}
	}
	return nil
}

func init() {
	RegisterBuiltin("apps", &AppsCommand{})
}
