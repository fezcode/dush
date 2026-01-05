package builtins

import (
	"context"
	"fmt"
	"io"
	"strings"

	"dush/internal/config"
)

// AliasCommand implements the `alias` built-in command.
type AliasCommand struct {
}

// NewAliasCommand creates a new instance of AliasCommand.
func NewAliasCommand() *AliasCommand {
	return &AliasCommand{}
}

// Name returns the name of the built-in command.
func (c *AliasCommand) Name() string {
	return "alias"
}

// Execute runs the alias command.
func (c *AliasCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) error {
	cfg := config.GetConfig()
	aliases := cfg.Aliases

	var (
		printAll  bool
		saveAlias bool
		setAlias  string // To hold "name=value" if setting an alias
		showAlias string // To hold "name" if showing an alias
	)

	// Parse flags
	filteredArgs := []string{}
	for _, arg := range args {
		switch arg {
		case "-p", "--print":
			printAll = true
		case "-s", "--save":
			saveAlias = true
		default:
			filteredArgs = append(filteredArgs, arg)
		}
	}

	if printAll {
		// If -p or --print is present, just print all aliases
		if len(aliases) == 0 {
			fmt.Fprintln(out, "No aliases defined.")
			return nil
		}
		for name, value := range aliases {
			fmt.Fprintf(out, "alias %s='%s'\n", name, value)
		}
		return nil
	}

	// Now process the filteredArgs
	if len(filteredArgs) == 0 {
		// If no other args and no -p, default to print all aliases
		if len(aliases) == 0 {
			fmt.Fprintln(out, "No aliases defined.")
			return nil
		}
		for name, value := range aliases {
			fmt.Fprintf(out, "alias %s='%s'\n", name, value)
		}
		return nil
	}

	fmt.Printf("=========\n%v\n============\n%d", filteredArgs, len(filteredArgs))
	fmt.Printf(">> %v\n", filteredArgs[0])
	fmt.Printf(">> %v\n", filteredArgs[1])

	if len(filteredArgs) == 1 {
		arg := filteredArgs[0]
		if strings.Contains(arg, "=") {
			setAlias = arg
		} else {
			showAlias = arg
		}
	} else {
		fmt.Fprintln(errOut, "Usage:")
		fmt.Fprintln(errOut, "  alias                               - List all aliases")
		fmt.Fprintln(errOut, "  alias -p | --print                  - Print all aliases")
		fmt.Fprintln(errOut, "  alias <name>                        - Show a specific alias")
		fmt.Fprintln(errOut, "  alias [-s | --save] <name>=<value>  - Set an alias (with or without saving)")
		return fmt.Errorf("invalid arguments for alias command")
	}

	if setAlias != "" {
		// `alias name=value`: Set an alias
		parts := strings.SplitN(setAlias, "=", 2)
		fmt.Printf("=========\n%v\n============\n", parts)

		name := parts[0]
		value := parts[1]

		// Remove quotes if present
		value = strings.Trim(value, "'\"")

		aliases[name] = value
		if saveAlias {
			if err := config.SaveAliases(); err != nil {
				fmt.Fprintf(errOut, "Error saving aliases: %v\n", err)
				return err
			}
			fmt.Fprintf(out, "Alias '%s' set to '%s' and saved.\n", name, value)
		} else {
			fmt.Fprintf(out, "Alias '%s' set to '%s' (runtime only).\n", name, value)
		}
		return nil
	}

	if showAlias != "" {
		// `alias name`: Show a specific alias
		name := showAlias
		if value, ok := aliases[name]; ok {
			fmt.Fprintf(out, "alias %s='%s'\n", name, value)
		} else {
			fmt.Fprintf(errOut, "Alias '%s' not found.\n", name)
		}
		return nil
	}

	return nil
}

func init() {
	RegisterBuiltin("alias", NewAliasCommand())
}
