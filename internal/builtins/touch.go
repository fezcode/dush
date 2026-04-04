package builtins

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type TouchCommand struct{}

func (c *TouchCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) error {
	noCreate := false
	changeAccess := false
	changeMod := false
	var files []string

	for _, arg := range args {
		if strings.HasPrefix(arg, "-") && len(arg) > 1 && arg[1] != '-' {
			for _, ch := range arg[1:] {
				switch ch {
				case 'a':
					changeAccess = true
				case 'm':
					changeMod = true
				case 'c':
					noCreate = true
				case 'h':
					c.printHelp(out)
					return nil
				default:
					return fmt.Errorf("touch: invalid option -- '%c'\nTry 'touch --help' for more information.", ch)
				}
			}
		} else if strings.HasPrefix(arg, "--") {
			switch arg {
			case "--no-create":
				noCreate = true
			case "--help":
				c.printHelp(out)
				return nil
			default:
				return fmt.Errorf("touch: unrecognized option '%s'\nTry 'touch --help' for more information.", arg)
			}
		} else {
			files = append(files, arg)
		}
	}

	if len(files) == 0 {
		return fmt.Errorf("touch: missing file operand\nTry 'touch --help' for more information.")
	}

	// If neither -a nor -m is specified, change both
	if !changeAccess && !changeMod {
		changeAccess = true
		changeMod = true
	}

	now := time.Now()

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			if os.IsNotExist(err) {
				if noCreate {
					continue
				}
				// Create the file
				f, createErr := os.Create(file)
				if createErr != nil {
					fmt.Fprintf(errOut, "touch: cannot touch '%s': %v\n", file, createErr)
					continue
				}
				f.Close()
				continue
			} else {
				fmt.Fprintf(errOut, "touch: cannot stat '%s': %v\n", file, err)
				continue
			}
		}

		// Update times for an existing file
		// Note: Standard Go os package doesn't expose atime in os.FileInfo cleanly across platforms.
		// We use info.ModTime() as a fallback for atime if we are not updating it.
		atime := info.ModTime()
		mtime := info.ModTime()

		if changeAccess && changeMod {
			atime = now
			mtime = now
		} else if changeMod {
			// Change mtime, preserve atime (fallback to mtime)
			mtime = now
		} else if changeAccess {
			// Change atime, preserve mtime
			atime = now
		}

		err = os.Chtimes(file, atime, mtime)
		if err != nil {
			fmt.Fprintf(errOut, "touch: cannot touch '%s': %v\n", file, err)
		}
	}

	return nil
}

func (c *TouchCommand) printHelp(out io.Writer) {
	helpText := `Usage: touch [OPTION]... FILE...
Update the access and modification times of each FILE to the current time.

A FILE argument that does not exist is created empty, unless -c or --no-create is supplied.

Options:
  -a                    change only the access time
  -c, --no-create       do not create any files
  -m                    change only the modification time
  -h, --help            display this help and exit
`
	fmt.Fprint(out, helpText)
}

func init() {
	RegisterBuiltin("touch", &TouchCommand{})
}
