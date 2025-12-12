package builtins

import (
	"fmt"
	"io"

	"dush/internal/app"
)

// PWDCommand implements the Command interface for the 'pwd' builtin.
type PWDCommand struct{}

// Execute prints the current working directory to the output writer.
func (c *PWDCommand) Execute(args []string, out io.Writer, errOut io.Writer) error {
	if len(args) > 0 {
		return fmt.Errorf("pwd: too many arguments")
	}

	appInstance := app.GetApp() // Get the app singleton
	fmt.Fprintf(out, "%s\n", appInstance.GetCurrentDir())
	return nil
}

func init() {
	RegisterBuiltin("pwd", &PWDCommand{})
}
