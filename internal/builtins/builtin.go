package builtins

import (
	"fmt"
	"io"
	"strings"
)

// Command is the interface that all built-in commands must implement.
type Command interface {
	Execute(args []string, out io.Writer, errOut io.Writer) error
}

var registeredCommands = make(map[string]Command)

// RegisterBuiltin registers a new built-in command.
func RegisterBuiltin(name string, cmd Command) {
	registeredCommands[name] = cmd
}

// RunBuiltin checks if the input is a registered built-in command and executes it.
// It returns true if a builtin was executed, false otherwise.
func RunBuiltin(input string, out io.Writer, errOut io.Writer) bool {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return false
	}

	cmdName := parts[0]
	if cmd, ok := registeredCommands[cmdName]; ok {
		err := cmd.Execute(parts[1:], out, errOut)
		if err != nil {
			fmt.Fprintf(errOut, "%s: %v\n", cmdName, err)
		}
		return true
	}
	return false
}
