package builtins

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dush/internal/app"
	"dush/internal/utils" // Ensure utils is imported
)

type LsCommand struct{}

// lsOptions holds parsed options for the ls command.
type lsOptions struct {
	LongFormat bool
	Path       string
}

// parseLsArgs parses the arguments for the ls command.
func parseLsArgs(args []string) (lsOptions, error) {
	opts := lsOptions{
		Path: ".", // Default path
	}

	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			// Handle flags
			for _, flagChar := range arg[1:] {
				switch flagChar {
				case 'l':
					opts.LongFormat = true
				default:
					return opts, fmt.Errorf("ls: unknown option -- '%c'", flagChar)
				}
			}
		} else {
			// Assume it's a path if not a flag
			if opts.Path != "." {
				// Only one path argument supported for now
				return opts, fmt.Errorf("ls: too many arguments (only one path supported)")
			}
			opts.Path = arg
		}
	}
	return opts, nil
}

// colorizeFileName returns the file name wrapped in ANSI color codes based on its type.
// It requires the full path to accurately check for broken symlinks.
func colorizeFileName(fullPath string, info fs.FileInfo, name string) string {
	mode := info.Mode()
	if mode.IsDir() {
		return utils.Colorize(name, utils.ColorBlue)
	}
	if mode&fs.ModeSymlink != 0 {
		// Check if symlink is broken by Stat-ing the target of the symlink.
		// os.Stat on `fullPath` (which is the symlink itself) will resolve the link.
		// If the resolved target does not exist, Stat will return an error.
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return utils.Colorize(name, utils.ColorRed) // Broken symlink
		}
		return utils.Colorize(name, utils.ColorCyan)
	}
	if mode&0111 != 0 { // Check if any execute bit is set (user, group, or other)
		return utils.Colorize(name, utils.ColorGreen)
	}
	return name // Default: no special color for regular files
}

// formatLongListing formats file information into a long listing string.
// It requires the full path to accurately get owner/group and resolve symlink targets.
func formatLongListing(fullPath string, info fs.FileInfo) string {
	_ = time.Now() // Dummy usage to satisfy the compiler about "time" import
	// Permissions
	mode := info.Mode()
	perm := make([]byte, 10)
	perm[0] = '-'
	if mode.IsDir() {
		perm[0] = 'd'
	} else if mode&fs.ModeSymlink != 0 {
		perm[0] = 'l'
	}
	// Add more file types as needed (e.g., ModeSocket, ModeNamedPipe)

	if mode&0400 != 0 {
		perm[1] = 'r'
	} else {
		perm[1] = '-'
	}
	if mode&0200 != 0 {
		perm[2] = 'w'
	} else {
		perm[2] = '-'
	}
	if mode&0100 != 0 {
		perm[3] = 'x'
	} else {
		perm[3] = '-'
	}

	if mode&0040 != 0 {
		perm[4] = 'r'
	} else {
		perm[4] = '-'
	}
	if mode&0020 != 0 {
		perm[5] = 'w'
	} else {
		perm[5] = '-'
	}
	if mode&0010 != 0 {
		perm[6] = 'x'
	} else {
		perm[6] = '-'
	}

	if mode&0004 != 0 {
		perm[7] = 'r'
	} else {
		perm[7] = '-'
	}
	if mode&0002 != 0 {
		perm[8] = 'w'
	} else {
		perm[8] = '-'
	}
	if mode&0001 != 0 {
		perm[9] = 'x'
	} else {
		perm[9] = '-'
	}

	// Owner and Group
	owner, group := utils.GetOwnerAndGroupNames(fullPath, info)

	// Size
	sizeStr := fmt.Sprintf("%10d", info.Size()) // Right-aligned

	// Modification time
	modTimeStr := info.ModTime().Format("Jan _2 15:04") // e.g., "Jan  2 15:04"

	// Name
	displayFileName := info.Name()
	if mode&fs.ModeSymlink != 0 {
		// If it's a symlink, append " -> target"
		// os.Readlink takes the symlink path itself, not its target.
		symlinkTarget, err := os.Readlink(fullPath)
		if err == nil {
			displayFileName = fmt.Sprintf("%s -> %s", displayFileName, symlinkTarget)
		}
	}
	coloredName := colorizeFileName(fullPath, info, displayFileName)

	// Links (hard links, always 1 for simple stat for now)
	links := "1" // Placeholder

	return fmt.Sprintf("%s %s %s %s %s %s %s", string(perm), links, owner, group, sizeStr, modTimeStr, coloredName)
}

func (c *LsCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) error {
	opts, err := parseLsArgs(args)
	if err != nil {
		return fmt.Errorf("ls: %w", err)
	}

	// Get the app singleton
	appInstance := app.GetApp()

	// If no explicit path was provided, use the shell's current working directory
	if opts.Path == "." {
		opts.Path = appInstance.GetCurrentDir()
	}

	dirEntries, err := os.ReadDir(opts.Path)
	if err != nil {
		return fmt.Errorf("ls: cannot access '%s': %w", opts.Path, err)
	}

	for _, entry := range dirEntries {
		select {
		case <-ctx.Done():
			return ctx.Err() // Command interrupted
		default:
			// Construct the full path to the current entry
			fullEntryPath := filepath.Join(opts.Path, entry.Name())

			info, err := entry.Info() // Get FileInfo for coloring and long listing
			if err != nil {
				fmt.Fprintf(errOut, "ls: could not get info for '%s': %v\n", fullEntryPath, err)
				continue // Skip this entry
			}

			if opts.LongFormat {
				fmt.Fprintln(out, formatLongListing(fullEntryPath, info))
			} else {
				fmt.Fprintln(out, colorizeFileName(fullEntryPath, info, entry.Name()))
			}
		}
	}

	return nil
}

func init() {
	RegisterBuiltin("ls", &LsCommand{})
}
