package builtins

import (
	"context"
	"errors"
	"fmt"
	"io"
)

// ListBuiltins returns a slice of strings containing the names of all registered built-in commands.
func ListBuiltins() []string {
	names := make([]string, 0, len(registeredCommands))
	for name := range registeredCommands {
		names = append(names, name)
	}
	return names
}

// Command is the interface that all built-in commands must implement.
type Command interface {
	Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) error
}

var registeredCommands = make(map[string]Command)

// RegisterBuiltin registers a new built-in command.
func RegisterBuiltin(name string, cmd Command) {
	registeredCommands[name] = cmd
}

// RunBuiltin checks if the given command name is a registered built-in command and executes it.
// It returns true if a builtin was executed, false otherwise.
// The context should be passed from the REPL to allow for cancellation.
func RunBuiltin(ctx context.Context, cmdName string, args []string, out io.Writer, errOut io.Writer) bool {
	if cmd, ok := registeredCommands[cmdName]; ok {
		err := cmd.Execute(ctx, args, out, errOut)
		if err != nil {
			// Do not print error if context was cancelled, as it's an expected interruption
			if errors.Is(err, context.Canceled) {
				fmt.Fprintln(errOut, "Command interrupted.")
			} else {
				fmt.Fprintf(errOut, "%s: %v\n", cmdName, err)
			}
		}
		return true
	}
	return false
}
