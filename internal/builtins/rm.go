package builtins

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
)

type RmCommand struct{}

func (c *RmCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) error {
	force := false
	recursive := false
	verbose := false
	var targets []string

	for _, arg := range args {
		if strings.HasPrefix(arg, "-") && len(arg) > 1 && arg[1] != '-' {
			for _, ch := range arg[1:] {
				switch ch {
				case 'f':
					force = true
				case 'r', 'R':
					recursive = true
				case 'v':
					verbose = true
				case 'h':
					c.printHelp(out)
					return nil
				default:
					return fmt.Errorf("rm: invalid option -- '%c'\nTry 'rm --help' for more information.", ch)
				}
			}
		} else if strings.HasPrefix(arg, "--") {
			switch arg {
			case "--force":
				force = true
			case "--recursive":
				recursive = true
			case "--verbose":
				verbose = true
			case "--help":
				c.printHelp(out)
				return nil
			default:
				return fmt.Errorf("rm: unrecognized option '%s'\nTry 'rm --help' for more information.", arg)
			}
		} else {
			targets = append(targets, arg)
		}
	}

	if len(targets) == 0 {
		if !force {
			return fmt.Errorf("rm: missing operand\nTry 'rm --help' for more information.")
		}
		return nil
	}

	for _, target := range targets {
		info, err := os.Lstat(target)
		if err != nil {
			if !force {
				fmt.Fprintf(errOut, "rm: cannot remove '%s': %v\n", target, err)
			}
			continue
		}

		if info.IsDir() {
			if !recursive {
				if !force {
					fmt.Fprintf(errOut, "rm: cannot remove '%s': Is a directory\n", target)
				}
				continue
			}
			err = os.RemoveAll(target)
		} else {
			err = os.Remove(target)
		}

		if err != nil {
			if !force {
				fmt.Fprintf(errOut, "rm: cannot remove '%s': %v\n", target, err)
			}
		} else if verbose {
			if info.IsDir() {
				fmt.Fprintf(out, "removed directory '%s'\n", target)
			} else {
				fmt.Fprintf(out, "removed '%s'\n", target)
			}
		}
	}
	return nil
}

func (c *RmCommand) printHelp(out io.Writer) {
	helpText := `Usage: rm [OPTION]... FILE...
Remove (unlink) the FILE(s).

Options:
  -f, --force           ignore nonexistent files and arguments, never prompt
  -r, -R, --recursive   remove directories and their contents recursively
  -v, --verbose         explain what is being done
  -h, --help            display this help and exit
`
	fmt.Fprint(out, helpText)
}

func init() {
	RegisterBuiltin("rm", &RmCommand{})
}
