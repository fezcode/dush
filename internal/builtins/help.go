package builtins

import (
	"context" // New import
	"fmt"
	"io"
	"sort"
)

// HelpCommand represents the 'help' built-in command.
type HelpCommand struct{}

// Execute prints a list of all available built-in commands.
func (c *HelpCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) error {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Fprintln(out, "Usage: help\nList all available built-in commands. Run any command with -h for details.")
		return nil
	}
	if len(args) > 0 {
		return fmt.Errorf("help: too many arguments")
	}

	commandNames := ListBuiltins()
	sort.Strings(commandNames) // Sort for consistent output

	fmt.Fprintln(out, "Available built-in commands:")
	for _, name := range commandNames {
		fmt.Fprintf(out, "  %s\n", name)
	}
	return nil
}

func init() {
	RegisterBuiltin("help", &HelpCommand{})
}
