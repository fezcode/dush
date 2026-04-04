package builtins

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
)

type MkdirCommand struct{}

func (c *MkdirCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) error {
	createParents := false
	verbose := false
	var dirs []string

	for _, arg := range args {
		if strings.HasPrefix(arg, "-") && len(arg) > 1 && arg[1] != '-' {
			for _, ch := range arg[1:] {
				switch ch {
				case 'p':
					createParents = true
				case 'v':
					verbose = true
				case 'h':
					c.printHelp(out)
					return nil
				default:
					return fmt.Errorf("mkdir: invalid option -- '%c'\nTry 'mkdir --help' for more information.", ch)
				}
			}
		} else if strings.HasPrefix(arg, "--") {
			switch arg {
			case "--parents":
				createParents = true
			case "--verbose":
				verbose = true
			case "--help":
				c.printHelp(out)
				return nil
			default:
				return fmt.Errorf("mkdir: unrecognized option '%s'\nTry 'mkdir --help' for more information.", arg)
			}
		} else {
			dirs = append(dirs, arg)
		}
	}

	if len(dirs) == 0 {
		return fmt.Errorf("mkdir: missing operand\nTry 'mkdir --help' for more information.")
	}

	for _, dir := range dirs {
		var err error
		if createParents {
			err = os.MkdirAll(dir, 0755)
		} else {
			err = os.Mkdir(dir, 0755)
		}

		if err != nil {
			fmt.Fprintf(errOut, "mkdir: cannot create directory '%s': %v\n", dir, err)
		} else if verbose {
			fmt.Fprintf(out, "mkdir: created directory '%s'\n", dir)
		}
	}
	return nil
}

func (c *MkdirCommand) printHelp(out io.Writer) {
	helpText := `Usage: mkdir [OPTION]... DIRECTORY...
Create the DIRECTORY(ies), if they do not already exist.

Options:
  -p, --parents     no error if existing, make parent directories as needed
  -v, --verbose     print a message for each created directory
  -h, --help        display this help and exit
`
	fmt.Fprint(out, helpText)
}

func init() {
	RegisterBuiltin("mkdir", &MkdirCommand{})
}
