package repl

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"dush/internal/app"
	"dush/internal/builtins"
	"dush/internal/config"
	"dush/internal/evaluator"
	"dush/internal/evaluator/object"
	"dush/internal/parser/lexer"
	"dush/internal/parser/parser"
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

	// Initialize the Environment
	env := evaluator.NewEnvironment()
	env.Stdout = out
	env.Stderr = errOut

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
			fmt.Fprintf(out, "%s", promptLine)
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

		// Basic Alias handling (Pre-processing)
		parts := strings.Fields(trimmedLine)
		if len(parts) > 0 {
			if expandedValue, ok := cfg.Aliases[parts[0]]; ok {
				// Replace the first word with the alias value
				trimmedLine = expandedValue + strings.TrimPrefix(trimmedLine, parts[0])
			}
		}

		if trimmedLine == "exit" || trimmedLine == "quit" {
			if isTerminal {
				fmt.Fprintf(out, "\r\nExiting dush REPL.\n")
			} else {
				fmt.Fprintf(out, "Exiting dush REPL.\n")
			}
			return
		}

		if isTerminal {
			term.Restore(int(os.Stdin.Fd()), oldState)
		}

		// --- New Parser/Evaluator Logic ---
		l := lexer.New(trimmedLine)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			for _, msg := range p.Errors() {
				fmt.Fprintf(out, "Parse Error: %s\n", msg)
			}
		} else {
			evaluated := evaluator.Eval(program, env)
			if evaluated != nil {
				if evaluated.Type() != object.NULL_OBJ && evaluated.Type() != object.INTEGER_OBJ {
					fmt.Fprintf(out, "%s\n", evaluated.Inspect())
				}
				if evaluated.Type() == object.ERROR_OBJ {
					// Inspect handles printing "ERROR: msg"
				}
			}
		}

		if isTerminal {
			oldState, _ = term.MakeRaw(int(os.Stdin.Fd()))
		}
	}
}
