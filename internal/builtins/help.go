package builtins

import (
	"fmt"
	"io"
	"sort"
)

// HelpCommand represents the 'help' built-in command.
type HelpCommand struct{}

// Execute prints a list of all available built-in commands.
func (c *HelpCommand) Execute(args []string, out io.Writer, errOut io.Writer) error {
	if len(args) > 0 {
		return fmt.Errorf("help command does not accept arguments")
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
