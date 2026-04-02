package builtins

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
)

// ExportCommand implements the Command interface for the 'export' builtin.
type ExportCommand struct{}

// Execute sets an environment variable.
func (c *ExportCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) error {
	if len(args) == 0 {
		// No arguments, maybe print all exported variables like bash?
		// For simplicity, just return an error or print a help message.
		for _, env := range os.Environ() {
			fmt.Fprintln(out, env)
		}
		return nil
	}

	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]
			if err := os.Setenv(key, value); err != nil {
				return fmt.Errorf("export: %w", err)
			}
		} else {
			// If no '=', it might just export a variable already in the env
			// but in our shell, if it's just 'export VAR', we could either
			// look it up or do nothing. For now, let's do nothing or error.
			// Let's do nothing to match bash partially if VAR is not set.
		}
	}

	return nil
}

func init() {
	RegisterBuiltin("export", &ExportCommand{})
}
