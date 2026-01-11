package repl

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"dush/internal/app"
	"dush/internal/builtins"
	"dush/internal/config"
	"dush/internal/evaluator"
	"dush/internal/utils"

	"golang.org/x/term"
)

type lineEditor struct {
	prompt string
	line   []rune
	pos    int
}

type terminalIO struct {
	io.Reader
	io.Writer
}

func (le *lineEditor) readLine(stdin io.Reader, stdout io.Writer) (string, error) {
	t := term.NewTerminal(terminalIO{stdin, stdout}, le.prompt)

	// Set autocomplete callback
	t.AutoCompleteCallback = func(line string, pos int, key rune) (newLine string, newPos int, ok bool) {
		if key == '\t' {
			return le.autoComplete(line, pos)
		}
		return "", 0, false
	}

	return t.ReadLine()
}

func (le *lineEditor) autoComplete(line string, pos int) (string, int, bool) {
	before := line[:pos]
	after := line[pos:]

	fields := strings.Fields(before)

	// If the line is empty or we are completing the first word (command)
	if len(fields) == 0 || (len(fields) == 1 && !strings.HasSuffix(before, " ")) {
		prefix := ""
		if len(fields) > 0 {
			prefix = fields[0]
		}

		matches := []string{}
		// Builtins
		for _, name := range builtins.ListBuiltins() {
			if strings.HasPrefix(name, prefix) {
				matches = append(matches, name)
			}
		}
		// Aliases
		cfg := config.GetConfig()
		for name := range cfg.Aliases {
			if strings.HasPrefix(name, prefix) {
				matches = append(matches, name)
			}
		}

		if len(matches) == 0 {
			return "", 0, false
		}

		sort.Strings(matches)

		if len(matches) == 1 {
			return matches[0] + " " + after, len(matches[0]) + 1, true
		}

		// Multiple matches: find common prefix
		common := matches[0]
		for _, m := range matches[1:] {
			for i := 0; i < len(common) && i < len(m); i++ {
				if common[i] != m[i] {
					common = common[:i]
					break
				}
			}
			if len(common) == 0 {
				break
			}
		}
		return common + after, len(common), true
	}

	// File path completion
	lastField := ""
	if strings.HasSuffix(before, " ") {
		// New argument starting
		lastField = ""
	} else {
		lastField = fields[len(fields)-1]
	}

	dir := "."
	prefix := lastField
	if lastField != "" {
		dir = filepath.Dir(lastField)
		prefix = filepath.Base(lastField)
		if strings.HasSuffix(lastField, string(filepath.Separator)) || strings.HasSuffix(lastField, "/") {
			dir = lastField
			prefix = ""
		}
	}

	appInstance := app.GetApp()
	absDir := dir
	if !filepath.IsAbs(dir) {
		absDir = filepath.Join(appInstance.GetCurrentDir(), dir)
	}

	entries, err := os.ReadDir(absDir)
	if err != nil {
		return "", 0, false
	}

	matches := []string{}
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, prefix) {
			if entry.IsDir() {
				name += string(filepath.Separator)
			}
			matches = append(matches, name)
		}
	}

	if len(matches) == 0 {
		return "", 0, false
	}

	sort.Strings(matches)

	if len(matches) == 1 {
		completed := filepath.Join(dir, matches[0])
		// Reconstruct the line
		prev := before[:len(before)-len(lastField)]
		newLine := prev + completed + after
		return newLine, len(prev + completed), true
	}

	// Multiple matches: find common prefix
	common := matches[0]
	for _, m := range matches[1:] {
		for i := 0; i < len(common) && i < len(m); i++ {
			if common[i] != m[i] {
				common = common[:i]
				break
			}
		}
	}

	completed := filepath.Join(dir, common)
	prev := before[:len(before)-len(lastField)]
	return prev + completed + after, len(prev + completed), true
}

// Start starts the Read-Eval-Print Loop.
// It takes an io.Reader for input, an io.Writer for output, and an io.Writer for error output.
func Start(in io.Reader, out io.Writer, errOut io.Writer) {
	// Create a context for the entire REPL lifecycle, cancelled on SIGTERM/SIGHUP
	replCtx, replCancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGHUP)
	defer replCancel() // Ensure this context is cancelled when Start returns

	// Load history at the start of the REPL
	utils.LoadHistory()
	// Ensure history is saved when the REPL exits
	defer utils.SaveHistory()

	// Get the singleton App instance
	appInstance := app.GetApp()

	// Initialize currentCWD with the actual OS CWD at startup
	initialCWD, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(errOut, "Error getting initial working directory: %v. Defaulting to '/'.\n", err)
		initialCWD = "/" // Fallback if getting CWD fails
	}
	appInstance.SetCurrentDir(initialCWD) // Use the setter to initialize

	// Get the configuration once at the start of REPL
	cfg := config.GetConfig()

	// Check if stdin is a terminal
	isTerminal := term.IsTerminal(int(os.Stdin.Fd()))

	var oldState *term.State
	if isTerminal {
		oldState, err = term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			isTerminal = false
		} else {
			defer term.Restore(int(os.Stdin.Fd()), oldState)
		}
	}

	for {
		// Check if the main REPL context has been cancelled
		select {
		case <-replCtx.Done():
			if isTerminal {
				fmt.Fprintf(out, "\r\nExiting dush REPL gracefully...\n")
			} else {
				fmt.Fprintf(out, "\nExiting dush REPL gracefully...\n")
			}
			return
		default:
			// Continue
		}

		currentCWD := appInstance.GetCurrentDir()
		displayDirName := utils.GetDisplayDirName(currentCWD)

		// Construct the dynamic prompt using App's currentCWD
		promptLine := fmt.Sprintf("%s %s@%s%s ", cfg.PromptPrefix, cfg.UserName, displayDirName, cfg.PromptSuffix)

		var line string
		if isTerminal {
			le := &lineEditor{prompt: promptLine}
			line, err = le.readLine(in, out)
			if err != nil {
				if err == io.EOF {
					term.Restore(int(os.Stdin.Fd()), oldState)
					fmt.Fprintf(out, "\r\nExiting dush REPL.\n")
					return
				}
				// Other errors...
				continue
			}
		} else {
			fmt.Fprintf(out, promptLine)
			scanner := bufio.NewScanner(in)
			if !scanner.Scan() {
				fmt.Fprintf(out, "Exiting dush REPL.\n")
				return
			}
			line = scanner.Text()
		}

		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue // Skip empty lines
		}

		// Add command to history before processing it
		utils.AddCommand(trimmedLine)

		parts := strings.Fields(trimmedLine)
		cmdName := parts[0]
		args := parts[1:]

		// --- Start Alias Expansion ---
		if expandedValue, ok := cfg.Aliases[cmdName]; ok {
			// If the command name is an alias, expand it
			expandedParts := strings.Fields(expandedValue)
			if len(expandedParts) > 0 {
				cmdName = expandedParts[0]
				// Append original args to expanded alias args
				args = append(expandedParts[1:], args...)
			}
		}
		// --- End Alias Expansion ---

		if cmdName == "exit" || cmdName == "quit" {
			if isTerminal {
				fmt.Fprintf(out, "\r\nExiting dush REPL.\n")
			} else {
				fmt.Fprintf(out, "Exiting dush REPL.\n")
			}
			return
		}

		// Create a cancellable context for the current command
		cmdCtx, cmdCancel := context.WithCancel(replCtx)

		// Check and execute built-in commands
		if builtins.RunBuiltin(cmdCtx, cmdName, args, out, errOut) {
			// Builtin handled
		} else {
			// If not a built-in command, attempt to run as an external command
			if isTerminal {
				term.Restore(int(os.Stdin.Fd()), oldState)
			}

			err := evaluator.ExecuteExternal(cmdCtx, cmdName, args, out, errOut)
			if err != nil {
				if _, ok := err.(*exec.Error); ok {
					fmt.Fprintf(out, "Command not found: %s\n", cmdName)
				} else {
					// Other errors (like execution failure)
					fmt.Fprintf(out, "Error executing %s: %v\n", cmdName, err)
				}
			}

			if isTerminal {
				oldState, _ = term.MakeRaw(int(os.Stdin.Fd()))
			}
		}
		cmdCancel()
	}
}
