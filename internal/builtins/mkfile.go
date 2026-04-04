package builtins

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

type MkfileCommand struct{}

func parseSize(sizeStr string) (int64, error) {
	if len(sizeStr) == 0 {
		return 0, fmt.Errorf("empty size")
	}

	multiplier := int64(1)
	lastChar := unicode.ToLower(rune(sizeStr[len(sizeStr)-1]))

	switch lastChar {
	case 'b':
		multiplier = 512 // blocks
		sizeStr = sizeStr[:len(sizeStr)-1]
	case 'k':
		multiplier = 1024
		sizeStr = sizeStr[:len(sizeStr)-1]
	case 'm':
		multiplier = 1024 * 1024
		sizeStr = sizeStr[:len(sizeStr)-1]
	case 'g':
		multiplier = 1024 * 1024 * 1024
		sizeStr = sizeStr[:len(sizeStr)-1]
	}

	// if no suffix or suffix was the only character
	if len(sizeStr) == 0 {
		return 0, fmt.Errorf("invalid size format")
	}

	val, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size number: %v", err)
	}

	return val * multiplier, nil
}

func (c *MkfileCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) error {
	createParents := false
	verbose := false
	var sizeStr string
	var files []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") && len(arg) > 1 && arg[1] != '-' {
			for j, ch := range arg[1:] {
				switch ch {
				case 'p':
					createParents = true
				case 'v':
					verbose = true
				case 'h':
					c.printHelp(out)
					return nil
				case 's':
					// size flag
					if len(arg) > j+2 { // e.g., -s10m
						sizeStr = arg[j+2:]
					} else if i+1 < len(args) { // e.g., -s 10m
						i++
						sizeStr = args[i]
					} else {
						return fmt.Errorf("mkfile: option requires an argument -- 's'")
					}
					goto NextArg // We handled the rest of this flag group
				default:
					return fmt.Errorf("mkfile: invalid option -- '%c'\nTry 'mkfile --help' for more information.", ch)
				}
			}
		} else if strings.HasPrefix(arg, "--") {
			if strings.HasPrefix(arg, "--size=") {
				sizeStr = strings.TrimPrefix(arg, "--size=")
			} else {
				switch arg {
				case "--parents":
					createParents = true
				case "--verbose":
					verbose = true
				case "--help":
					c.printHelp(out)
					return nil
				case "--size":
					if i+1 < len(args) {
						i++
						sizeStr = args[i]
					} else {
						return fmt.Errorf("mkfile: option '--size' requires an argument")
					}
				default:
					return fmt.Errorf("mkfile: unrecognized option '%s'\nTry 'mkfile --help' for more information.", arg)
				}
			}
		} else {
			files = append(files, arg)
		}
	NextArg:
	}

	if len(files) == 0 {
		return fmt.Errorf("mkfile: missing operand\nTry 'mkfile --help' for more information.")
	}

	var size int64 = 0
	if sizeStr != "" {
		parsedSize, err := parseSize(sizeStr)
		if err != nil {
			return fmt.Errorf("mkfile: invalid size '%s': %v", sizeStr, err)
		}
		size = parsedSize
	}

	for _, file := range files {
		if createParents {
			dir := filepath.Dir(file)
			if dir != "." && dir != "" {
				if err := os.MkdirAll(dir, 0755); err != nil {
					fmt.Fprintf(errOut, "mkfile: cannot create directory '%s': %v\n", dir, err)
					continue
				} else if verbose {
					fmt.Fprintf(out, "mkfile: created directory '%s'\n", dir)
				}
			}
		}

		f, err := os.Create(file)
		if err != nil {
			fmt.Fprintf(errOut, "mkfile: cannot create file '%s': %v\n", file, err)
		} else {
			if size > 0 {
				if err := f.Truncate(size); err != nil {
					fmt.Fprintf(errOut, "mkfile: cannot set size for '%s': %v\n", file, err)
				}
			}
			if verbose {
				if size > 0 {
					fmt.Fprintf(out, "mkfile: created file '%s' with size %d bytes\n", file, size)
				} else {
					fmt.Fprintf(out, "mkfile: created file '%s'\n", file)
				}
			}
			f.Close()
		}
	}
	return nil
}

func (c *MkfileCommand) printHelp(out io.Writer) {
	helpText := `Usage: mkfile [OPTION]... FILE...
Create the FILE(s), optionally setting their size.

Options:
  -p, --parents     make parent directories as needed
  -s, --size=SIZE   set file size (e.g., 10k, 5m, 2g, 512b)
  -v, --verbose     print a message for each created file and directory
  -h, --help        display this help and exit

SIZE is an integer and optional unit (example: 10K is 10*1024).
Units are K, M, G for kilobytes, megabytes, gigabytes. B is for blocks (512 bytes).
`
	fmt.Fprint(out, helpText)
}

func init() {
	RegisterBuiltin("mkfile", &MkfileCommand{})
}
