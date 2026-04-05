package builtins

import (
	"context"
	"fmt"
	"io"

	"dush/internal/config"
)

// UnaliasCommand implements the `unalias` built-in command.
type UnaliasCommand struct {
}

// NewUnaliasCommand creates a new instance of UnaliasCommand.
func NewUnaliasCommand() *UnaliasCommand {
	return &UnaliasCommand{}
}

// Name returns the name of the built-in command.
func (c *UnaliasCommand) Name() string {
	return "unalias"
}

// Execute runs the unalias command.
func (c *UnaliasCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) error {
	cfg := config.GetConfig()
	aliases := cfg.Aliases

	var aliasName string

	// Parse flags
	filteredArgs := []string{}
	for _, arg := range args {
		switch arg {
		case "-h", "--help":
			fmt.Fprintln(out, "Usage: unalias <name>\nRemove an alias definition.")
			return nil
		case "-s", "--save":
			// ignored, aliases persist via ~/.dush/is
		default:
			filteredArgs = append(filteredArgs, arg)
		}
	}

	if len(filteredArgs) < 1 {
		fmt.Fprintln(errOut, "Usage: unalias [-s | --save] <name>")
		return fmt.Errorf("missing alias name for unalias")
	}

	aliasName = filteredArgs[0]

	if _, ok := aliases[aliasName]; ok {
		delete(aliases, aliasName)
		fmt.Fprintf(out, "Alias '%s' removed.\n", aliasName)
	} else {
		fmt.Fprintf(errOut, "Alias '%s' not found.\n", aliasName)
	}
	return nil
}

func init() {
	RegisterBuiltin("unalias", NewUnaliasCommand())
}
