package builtins

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"dush/internal/app"
	"dush/internal/utils"
)

type LsCommand struct{}

type lsOptions struct {
	LongFormat bool
	All        bool // show hidden (dot) files
	Header     bool // show table header
	Grid       bool // grid layout (default for non-long)
	OnePerLine bool // force one-per-line
	SortBy     string
	Reverse    bool
	NoIcons    bool
	ShowHelp   bool
	Path       string
}

func parseLsArgs(args []string) (lsOptions, error) {
	opts := lsOptions{
		Path:   ".",
		Header: true,
		Grid:   true,
		SortBy: "name",
	}

	for _, arg := range args {
		if strings.HasPrefix(arg, "-") && len(arg) > 1 && arg[1] != '-' {
			for _, ch := range arg[1:] {
				switch ch {
				case 'l':
					opts.LongFormat = true
					opts.Grid = false
				case 'a':
					opts.All = true
				case '1':
					opts.OnePerLine = true
					opts.Grid = false
				case 'r':
					opts.Reverse = true
				case 'S':
					opts.SortBy = "size"
				case 't':
					opts.SortBy = "time"
				case 'X':
					opts.SortBy = "ext"
				case 'h':
					opts.ShowHelp = true
				default:
					return opts, fmt.Errorf("unknown option -- '%c'", ch)
				}
			}
		} else if strings.HasPrefix(arg, "--") {
			switch arg {
			case "--help":
				opts.ShowHelp = true
			case "--all":
				opts.All = true
			case "--long":
				opts.LongFormat = true
				opts.Grid = false
			case "--grid":
				opts.Grid = true
			case "--oneline":
				opts.OnePerLine = true
				opts.Grid = false
			case "--sort=name":
				opts.SortBy = "name"
			case "--sort=size":
				opts.SortBy = "size"
			case "--sort=time":
				opts.SortBy = "time"
			case "--sort=ext":
				opts.SortBy = "ext"
			case "--reverse":
				opts.Reverse = true
			case "--no-header":
				opts.Header = false
			case "--no-icons":
				opts.NoIcons = true
			default:
				return opts, fmt.Errorf("unknown option: %s", arg)
			}
		} else {
			if opts.Path != "." {
				return opts, fmt.Errorf("too many arguments")
			}
			opts.Path = arg
		}
	}
	return opts, nil
}

type fileEntry struct {
	name    string
	info    fs.FileInfo
	path    string
	isDir   bool
	ext     string
	symlink string
}

func colorIcon(ch string, bg string, fg string) string {
	return bg + fg + utils.StyleBold + " " + ch + " " + utils.ColorReset
}

func fileIcon(e fileEntry) string {
	if e.isDir {
		return colorIcon(">", utils.BgBrightBlue, utils.ColorWhite)
	}
	switch strings.ToLower(e.ext) {
	case ".go":
		return colorIcon("G", utils.BgCyan, utils.ColorWhite)
	case ".js", ".ts", ".jsx", ".tsx":
		return colorIcon("J", utils.BgYellow, utils.ColorBlack)
	case ".py":
		return colorIcon("P", utils.BgGreen, utils.ColorWhite)
	case ".rs":
		return colorIcon("R", utils.BgRed, utils.ColorWhite)
	case ".rb":
		return colorIcon("r", utils.BgRed, utils.ColorWhite)
	case ".java", ".class", ".jar":
		return colorIcon("j", utils.BgBrightRed, utils.ColorWhite)
	case ".c", ".h":
		return colorIcon("C", utils.BgBlue, utils.ColorWhite)
	case ".cpp", ".cc", ".cxx", ".hpp":
		return colorIcon("+", utils.BgBlue, utils.ColorWhite)
	case ".md", ".txt", ".doc", ".docx":
		return colorIcon("d", utils.BgBrightBlack, utils.ColorWhite)
	case ".json", ".yaml", ".yml", ".toml", ".xml":
		return colorIcon("~", utils.BgYellow, utils.ColorBlack)
	case ".sh", ".bash", ".zsh", ".fish", ".dush":
		return colorIcon("$", utils.BgGreen, utils.ColorWhite)
	case ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".bmp", ".webp":
		return colorIcon("i", utils.BgMagenta, utils.ColorWhite)
	case ".mp3", ".wav", ".flac", ".ogg", ".m4a":
		return colorIcon("m", utils.BgMagenta, utils.ColorWhite)
	case ".mp4", ".mkv", ".avi", ".mov", ".webm":
		return colorIcon("v", utils.BgBrightMagenta, utils.ColorWhite)
	case ".zip", ".tar", ".gz", ".bz2", ".xz", ".7z", ".rar":
		return colorIcon("z", utils.BgRed, utils.ColorWhite)
	case ".exe", ".msi", ".dll", ".so", ".dylib":
		return colorIcon("*", utils.BgGreen, utils.ColorBlack)
	case ".pdf":
		return colorIcon("p", utils.BgRed, utils.ColorWhite)
	case ".html", ".htm", ".css":
		return colorIcon("w", utils.BgCyan, utils.ColorWhite)
	case ".sql", ".db", ".sqlite":
		return colorIcon("D", utils.BgBlue, utils.ColorWhite)
	case ".lock":
		return colorIcon("!", utils.BgBrightBlack, utils.ColorBrightWhite)
	case ".env":
		return colorIcon(".", utils.BgBrightBlack, utils.ColorBrightYellow)
	case ".log":
		return colorIcon("l", utils.BgBrightBlack, utils.ColorBrightWhite)
	case ".git", ".gitignore":
		return colorIcon("g", utils.BgBrightRed, utils.ColorWhite)
	default:
		if e.info.Mode()&0111 != 0 {
			return colorIcon("*", utils.BgGreen, utils.ColorBlack)
		}
		return colorIcon("-", utils.BgBrightBlack, utils.ColorBrightWhite)
	}
}

func colorName(e fileEntry) string {
	name := e.name
	if e.isDir {
		return utils.Colorize(name, utils.ColorBrightBlue+utils.StyleBold)
	}
	if e.info.Mode()&fs.ModeSymlink != 0 {
		target := name
		if e.symlink != "" {
			target = name + " → " + e.symlink
		}
		if _, err := os.Stat(e.path); os.IsNotExist(err) {
			return utils.Colorize(target, utils.ColorRed+utils.StyleCrossedOut)
		}
		return utils.Colorize(target, utils.ColorCyan+utils.StyleItalic)
	}
	if e.info.Mode()&0111 != 0 {
		return utils.Colorize(name, utils.ColorGreen+utils.StyleBold)
	}
	switch strings.ToLower(e.ext) {
	case ".go":
		return utils.Colorize(name, utils.ColorCyan)
	case ".js", ".ts", ".jsx", ".tsx":
		return utils.Colorize(name, utils.ColorYellow)
	case ".py":
		return utils.Colorize(name, utils.ColorBrightGreen)
	case ".rs":
		return utils.Colorize(name, utils.ColorBrightRed)
	case ".md", ".txt":
		return utils.Colorize(name, utils.ColorWhite)
	case ".json", ".yaml", ".yml", ".toml":
		return utils.Colorize(name, utils.ColorBrightYellow)
	case ".png", ".jpg", ".jpeg", ".gif", ".svg", ".bmp":
		return utils.Colorize(name, utils.ColorMagenta)
	case ".zip", ".tar", ".gz", ".7z", ".rar":
		return utils.Colorize(name, utils.ColorRed)
	case ".log":
		return utils.Colorize(name, utils.ColorBrightBlack)
	}
	return name
}

func humanSize(bytes int64) string {
	const (
		_  = iota
		KB = 1 << (10 * iota)
		MB
		GB
		TB
	)
	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.1fT", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.1fG", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1fM", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1fK", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

func formatPermissions(mode fs.FileMode) string {
	buf := make([]byte, 10)
	switch {
	case mode.IsDir():
		buf[0] = 'd'
	case mode&fs.ModeSymlink != 0:
		buf[0] = 'l'
	case mode&fs.ModeNamedPipe != 0:
		buf[0] = 'p'
	default:
		buf[0] = '.'
	}
	perms := [9]struct{ bit fs.FileMode; ch byte }{
		{0400, 'r'}, {0200, 'w'}, {0100, 'x'},
		{0040, 'r'}, {0020, 'w'}, {0010, 'x'},
		{0004, 'r'}, {0002, 'w'}, {0001, 'x'},
	}
	for i, p := range perms {
		if mode&p.bit != 0 {
			buf[i+1] = p.ch
		} else {
			buf[i+1] = '-'
		}
	}
	return string(buf)
}

func colorPerms(s string) string {
	var out strings.Builder
	for _, ch := range s {
		switch ch {
		case 'r':
			out.WriteString(utils.ColorBrightYellow)
			out.WriteRune(ch)
			out.WriteString(utils.ColorReset)
		case 'w':
			out.WriteString(utils.ColorBrightRed)
			out.WriteRune(ch)
			out.WriteString(utils.ColorReset)
		case 'x':
			out.WriteString(utils.ColorBrightGreen)
			out.WriteRune(ch)
			out.WriteString(utils.ColorReset)
		case 'd', 'l', 'p':
			out.WriteString(utils.ColorBrightBlue)
			out.WriteRune(ch)
			out.WriteString(utils.ColorReset)
		default:
			out.WriteString(utils.ColorBrightBlack)
			out.WriteRune(ch)
			out.WriteString(utils.ColorReset)
		}
	}
	return out.String()
}

func formatTime(t time.Time) string {
	now := time.Now()
	if t.Year() == now.Year() {
		return t.Format("Jan _2 15:04")
	}
	return t.Format("Jan _2  2006")
}

func sortEntries(entries []fileEntry, sortBy string, reverse bool) {
	sort.SliceStable(entries, func(i, j int) bool {
		// Directories first
		if entries[i].isDir != entries[j].isDir {
			return entries[i].isDir
		}

		var less bool
		switch sortBy {
		case "size":
			less = entries[i].info.Size() < entries[j].info.Size()
		case "time":
			less = entries[i].info.ModTime().Before(entries[j].info.ModTime())
		case "ext":
			if entries[i].ext == entries[j].ext {
				less = strings.ToLower(entries[i].name) < strings.ToLower(entries[j].name)
			} else {
				less = entries[i].ext < entries[j].ext
			}
		default: // name
			less = strings.ToLower(entries[i].name) < strings.ToLower(entries[j].name)
		}

		if reverse {
			return !less
		}
		return less
	})
}

func getTerminalWidth() int {
	// Default width; a more robust approach would query the terminal
	return 80
}

func printGrid(out io.Writer, entries []fileEntry, noIcons bool) {
	if len(entries) == 0 {
		return
	}

	// Calculate max display name width
	maxWidth := 0
	names := make([]string, len(entries))
	for i, e := range entries {
		icon := ""
		if !noIcons {
			icon = fileIcon(e) + " "
		}
		names[i] = icon + colorName(e)
		// Use raw name length for column calculation (no ANSI codes)
		rawLen := len(e.name)
		if !noIcons {
			rawLen += 4 // " X " icon (3) + trailing space (1)
		}
		if rawLen > maxWidth {
			maxWidth = rawLen
		}
	}

	termWidth := getTerminalWidth()
	colWidth := maxWidth + 2 // padding
	if colWidth < 4 {
		colWidth = 4
	}
	cols := termWidth / colWidth
	if cols < 1 {
		cols = 1
	}

	for i, n := range names {
		if i > 0 && i%cols == 0 {
			fmt.Fprintln(out)
		}
		// Pad with spaces to fill the column
		padding := colWidth - len(entries[i].name)
		if !noIcons {
			padding -= 4 // " X " icon (3) + trailing space (1)
		}
		if padding < 1 {
			padding = 1
		}
		fmt.Fprintf(out, "%s%s", n, strings.Repeat(" ", padding))
	}
	fmt.Fprintln(out)
}

func printLsHelp(out io.Writer) {
	help := `Usage: ls [OPTIONS] [PATH]

List directory contents with colors and icons.

Options:
  -l, --long        Long format with permissions, size, owner, created/modified times
  -a, --all         Show hidden (dot) files
  -1, --oneline     One entry per line
  -r, --reverse     Reverse sort order
  -S, --sort=size   Sort by file size
  -t, --sort=time   Sort by modification time
  -X, --sort=ext    Sort by file extension
      --no-header   Hide table header in long format
      --no-icons    Hide file type icons
  -h, --help        Show this help message

Examples:
  ls                List current directory
  ls -la            Long format with hidden files
  ls -lS            Long format sorted by size
  ls -lt            Long format sorted by time
  ls -r             Reverse sort order
  ls /path/to/dir   List a specific directory`
	fmt.Fprintln(out, help)
}

func (c *LsCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) error {
	opts, err := parseLsArgs(args)
	if err != nil {
		return fmt.Errorf("ls: %w", err)
	}

	if opts.ShowHelp {
		printLsHelp(out)
		return nil
	}

	appInstance := app.GetApp()
	if opts.Path == "." {
		opts.Path = appInstance.GetCurrentDir()
	}

	dirEntries, err := os.ReadDir(opts.Path)
	if err != nil {
		return fmt.Errorf("ls: cannot access '%s': %w", opts.Path, err)
	}

	// Build file entries
	var entries []fileEntry
	for _, de := range dirEntries {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		name := de.Name()
		if !opts.All && strings.HasPrefix(name, ".") {
			continue
		}

		fullPath := filepath.Join(opts.Path, name)
		info, err := de.Info()
		if err != nil {
			fmt.Fprintf(errOut, "ls: could not stat '%s': %v\n", fullPath, err)
			continue
		}

		e := fileEntry{
			name:  name,
			info:  info,
			path:  fullPath,
			isDir: info.IsDir(),
			ext:   filepath.Ext(name),
		}

		if info.Mode()&fs.ModeSymlink != 0 {
			if target, err := os.Readlink(fullPath); err == nil {
				e.symlink = target
			}
		}

		entries = append(entries, e)
	}

	sortEntries(entries, opts.SortBy, opts.Reverse)

	if len(entries) == 0 {
		return nil
	}

	if opts.LongFormat {
		// Table header
		if opts.Header {
			header := fmt.Sprintf(
				"%s  %s  %s  %s  %s  %s  %s  %s",
				utils.Colorize("Permissions", utils.StyleBold+utils.StyleUnderline),
				utils.Colorize("Size", utils.StyleBold+utils.StyleUnderline),
				utils.Colorize("User", utils.StyleBold+utils.StyleUnderline),
				utils.Colorize("Group", utils.StyleBold+utils.StyleUnderline),
				utils.Colorize("Created", utils.StyleBold+utils.StyleUnderline),
				utils.Colorize("Modified", utils.StyleBold+utils.StyleUnderline),
				utils.Colorize("", ""), // icon placeholder
				utils.Colorize("Name", utils.StyleBold+utils.StyleUnderline),
			)
			fmt.Fprintln(out, header)
		}

		// Calculate max widths for alignment
		maxSize := 0
		maxUser := 0
		maxGroup := 0
		for _, e := range entries {
			sz := len(humanSize(e.info.Size()))
			if sz > maxSize {
				maxSize = sz
			}
			owner, group := utils.GetOwnerAndGroupNames(e.path, e.info)
			if len(owner) > maxUser {
				maxUser = len(owner)
			}
			if len(group) > maxGroup {
				maxGroup = len(group)
			}
		}

		for _, e := range entries {
			perms := colorPerms(formatPermissions(e.info.Mode()))
			size := humanSize(e.info.Size())
			if e.isDir {
				size = utils.Colorize(fmt.Sprintf("%*s", maxSize, "-"), utils.ColorBrightBlack)
			} else {
				size = utils.Colorize(fmt.Sprintf("%*s", maxSize, size), utils.ColorBrightGreen)
			}
			owner, group := utils.GetOwnerAndGroupNames(e.path, e.info)
			owner = utils.Colorize(fmt.Sprintf("%-*s", maxUser, owner), utils.ColorBrightYellow)
			group = utils.Colorize(fmt.Sprintf("%-*s", maxGroup, group), utils.ColorBrightYellow)
			ctime := utils.GetCreationTime(e.path, e.info)
			createdTime := utils.Colorize(formatTime(ctime), utils.ColorBrightCyan)
			modTime := utils.Colorize(formatTime(e.info.ModTime()), utils.ColorBrightBlue)

			icon := ""
			if !opts.NoIcons {
				icon = fileIcon(e)
			}
			name := colorName(e)

			fmt.Fprintf(out, "%s  %s  %s  %s  %s  %s  %s %s\n",
				perms, size, owner, group, createdTime, modTime, icon, name)
		}
	} else if opts.OnePerLine {
		for _, e := range entries {
			icon := ""
			if !opts.NoIcons {
				icon = fileIcon(e) + " "
			}
			fmt.Fprintf(out, "%s%s\n", icon, colorName(e))
		}
	} else {
		// Grid mode
		printGrid(out, entries, opts.NoIcons)
	}

	return nil
}

func init() {
	RegisterBuiltin("ls", &LsCommand{})
}
