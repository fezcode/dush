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

	var (
		saveUnalias bool
		aliasName   string
	)

	// Parse flags
	filteredArgs := []string{}
	for _, arg := range args {
		switch arg {
		case "-s", "--save":
			saveUnalias = true
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
		if saveUnalias {
			if err := config.SaveAliases(); err != nil {
				fmt.Fprintf(errOut, "Error saving aliases: %v\n", err)
				return err
			}
			fmt.Fprintf(out, "Alias '%s' removed and saved.\n", aliasName)
		} else {
			fmt.Fprintf(out, "Alias '%s' removed (runtime only).\n", aliasName)
		}
	} else {
		fmt.Fprintf(errOut, "Alias '%s' not found.\n", aliasName)
	}
	return nil
}

func init() {
	RegisterBuiltin("unalias", NewUnaliasCommand())
}
