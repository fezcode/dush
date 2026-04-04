package prompt

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Default prompt format
const DefaultPromptLine = "{fg:cyan}{user}{reset}@{fg:green}{dir}{reset} {fg:yellow}${reset} "
const DefaultContinuationPrompt = "{fg:yellow}...{reset} "

// Context holds dynamic values available to the prompt renderer.
type Context struct {
	UserName   string
	HostName   string
	CWD        string
	HomeDir    string
	LastStatus string
}

// Render expands a prompt format string, replacing tokens with values and ANSI colors.
func Render(format string, ctx *Context) string {
	if format == "" {
		format = DefaultPromptLine
	}

	var out strings.Builder
	i := 0
	for i < len(format) {
		if format[i] == '{' {
			end := strings.IndexByte(format[i:], '}')
			if end == -1 {
				out.WriteByte(format[i])
				i++
				continue
			}
			token := format[i+1 : i+end]
			out.WriteString(expandToken(token, ctx))
			i += end + 1
		} else {
			out.WriteByte(format[i])
			i++
		}
	}
	return out.String()
}

func expandToken(token string, ctx *Context) string {
	// Color tokens
	if strings.HasPrefix(token, "fg:") {
		color := token[3:]
		return fgColor(color)
	}
	if strings.HasPrefix(token, "bg:") {
		color := token[3:]
		return bgColor(color)
	}

	switch token {
	// Text styles
	case "reset":
		return "\033[0m"
	case "bold":
		return "\033[1m"
	case "dim":
		return "\033[2m"
	case "italic":
		return "\033[3m"
	case "underline":
		return "\033[4m"

	// Content tokens
	case "user":
		return ctx.UserName
	case "host":
		return ctx.HostName
	case "dir":
		return dirName(ctx.CWD)
	case "path":
		return ctx.CWD
	case "home_path":
		return homePath(ctx.CWD, ctx.HomeDir)
	case "time":
		return time.Now().Format("15:04:05")
	case "date":
		return time.Now().Format("2006-01-02")
	case "git":
		return gitBranch(ctx.CWD)
	case "last_status":
		return ctx.LastStatus
	case "newline":
		return "\n"
	case "os":
		return runtime.GOOS
	default:
		return "{" + token + "}" // Unknown token, pass through
	}
}

// dirName returns the basename of the path, handling OS-specific roots.
func dirName(path string) string {
	if runtime.GOOS == "windows" {
		if len(path) >= 2 && path[1] == ':' {
			if len(path) <= 3 {
				return path
			}
		}
	}
	if path == "/" {
		return "/"
	}
	return filepath.Base(path)
}

// homePath replaces the home directory prefix with ~.
func homePath(path, home string) string {
	if home == "" {
		return path
	}
	// Normalize separators for comparison
	normPath := filepath.ToSlash(path)
	normHome := filepath.ToSlash(home)
	if strings.HasPrefix(normPath, normHome) {
		rest := normPath[len(normHome):]
		if rest == "" {
			return "~"
		}
		if rest[0] == '/' {
			return "~" + rest
		}
	}
	return path
}

// gitBranch returns the current git branch name, or empty string.
func gitBranch(cwd string) string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = cwd
	cmd.Stderr = nil
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	branch := strings.TrimSpace(string(out))
	if branch == "" {
		return ""
	}
	return branch
}

// --- ANSI color mapping ---

var colorCodes = map[string]int{
	"black":          30,
	"red":            31,
	"green":          32,
	"yellow":         33,
	"blue":           34,
	"magenta":        35,
	"cyan":           36,
	"white":          37,
	"bright_black":   90,
	"bright_red":     91,
	"bright_green":   92,
	"bright_yellow":  93,
	"bright_blue":    94,
	"bright_magenta": 95,
	"bright_cyan":    96,
	"bright_white":   97,
}

func fgColor(name string) string {
	if strings.HasPrefix(name, "#") && len(name) == 7 {
		r, g, b := hexToRGB(name)
		return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
	}
	if code, ok := colorCodes[name]; ok {
		return fmt.Sprintf("\033[%dm", code)
	}
	return ""
}

func bgColor(name string) string {
	if strings.HasPrefix(name, "#") && len(name) == 7 {
		r, g, b := hexToRGB(name)
		return fmt.Sprintf("\033[48;2;%d;%d;%dm", r, g, b)
	}
	if code, ok := colorCodes[name]; ok {
		return fmt.Sprintf("\033[%dm", code+10) // bg = fg + 10
	}
	return ""
}

func hexToRGB(hex string) (int, int, int) {
	var r, g, b int
	fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
	return r, g, b
}

// BuildContext creates a prompt context from current system state.
func BuildContext(lastStatus string) *Context {
	userName := os.Getenv("USER")
	if userName == "" {
		userName = os.Getenv("USERNAME")
	}
	if userName == "" {
		userName = "user"
	}

	hostName, _ := os.Hostname()
	if hostName == "" {
		hostName = "localhost"
	}

	cwd, _ := os.Getwd()
	home, _ := os.UserHomeDir()

	return &Context{
		UserName:   userName,
		HostName:   hostName,
		CWD:        cwd,
		HomeDir:    home,
		LastStatus: lastStatus,
	}
}
