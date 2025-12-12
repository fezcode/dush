package builtins

import (
	"fmt"
	"io"
)

// ClearCommand implements the Command interface for the 'clear' builtin.
type ClearCommand struct{}

// Execute clears the terminal screen using ANSI escape codes.
func (c *ClearCommand) Execute(args []string, out io.Writer, errOut io.Writer) error {
	if len(args) > 0 {
		return fmt.Errorf("clear: too many arguments")
	}

	// ANSI escape code for clearing screen and moving cursor to top-left
	fmt.Fprint(out, "\x1b[2J\x1b[H")
	return nil
}

func init() {
	RegisterBuiltin("clear", &ClearCommand{})
}
