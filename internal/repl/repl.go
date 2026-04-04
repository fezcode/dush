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
	"dush/internal/prompt"
	"dush/internal/utils"

	"golang.org/x/term"
)

type lineEditor struct {
	prompt string
	line   []rune
	pos    int
	out    io.Writer

	// Tab completion state — tracks consecutive tabs for list/cycle behavior
	lastTabLine string   // line content on previous tab press
	lastTabPos  int      // cursor position on previous tab press
	tabMatches  []string // cached matches from first tab
	tabIndex    int      // cycle index (-1 = common prefix shown, 0+ = cycling)
	tabPrev     string   // prefix before the completed arg
	tabAfter    string   // suffix after the cursor
	tabQuoted   bool     // whether the arg was quoted
}

type terminalIO struct {
	io.Reader
	io.Writer
}

func (le *lineEditor) resetTabState() {
	le.lastTabLine = ""
	le.lastTabPos = 0
	le.tabMatches = nil
	le.tabIndex = -1
}

func (le *lineEditor) autoComplete(line string, pos int) (string, int, bool) {
	// Check if this is a consecutive tab press (same line as last tab result)
	if le.lastTabLine == line && le.lastTabPos == pos && len(le.tabMatches) > 1 {
		return le.cycleCompletion()
	}

	// Fresh completion — reset state
	le.resetTabState()

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

		// Multiple: complete to common prefix, cache for cycling
		common := commonPrefix(matches)

		// Cache cycle state (command matches are plain strings, no quoting needed)
		le.tabPrev = ""
		le.tabAfter = after
		le.tabQuoted = false
		le.tabIndex = -1
		// Store full completions for cycling
		le.tabMatches = make([]string, len(matches))
		for i, m := range matches {
			le.tabMatches[i] = m + " "
		}

		result := common + after
		le.lastTabLine = result
		le.lastTabPos = len(common)
		return result, len(common), true
	}

	// File path completion — extract last argument respecting quotes
	rawArg, argStart, quoted := extractLastArg(before)

	dir := "."
	prefix := rawArg
	if rawArg != "" {
		dir = filepath.Dir(rawArg)
		prefix = filepath.Base(rawArg)
		if strings.HasSuffix(rawArg, string(filepath.Separator)) || strings.HasSuffix(rawArg, "/") {
			dir = rawArg
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

	prev := before[:argStart]

	if len(matches) == 1 {
		completed := filepath.Join(dir, matches[0])
		formatted := quoteIfNeeded(completed, quoted)
		newLine := prev + formatted + after
		return newLine, len(prev+formatted), true
	}

	// Multiple matches: complete to common prefix, cache for cycling
	common := commonPrefix(matches)
	completed := filepath.Join(dir, common)
	formatted := quoteIfNeeded(completed, quoted)

	// Cache cycle state
	le.tabPrev = prev
	le.tabAfter = after
	le.tabQuoted = quoted
	le.tabIndex = -1
	le.tabMatches = make([]string, len(matches))
	for i, m := range matches {
		le.tabMatches[i] = quoteIfNeeded(filepath.Join(dir, m), quoted)
	}

	result := prev + formatted + after
	le.lastTabLine = result
	le.lastTabPos = len(prev + formatted)
	return result, len(prev + formatted), true
}

func (le *lineEditor) cycleCompletion() (string, int, bool) {
	le.tabIndex++
	if le.tabIndex >= len(le.tabMatches) {
		le.tabIndex = 0
	}

	choice := le.tabMatches[le.tabIndex]
	result := le.tabPrev + choice + le.tabAfter
	newPos := len(le.tabPrev + choice)

	le.lastTabLine = result
	le.lastTabPos = newPos
	return result, newPos, true
}

func commonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	common := strs[0]
	for _, s := range strs[1:] {
		for i := 0; i < len(common) && i < len(s); i++ {
			if common[i] != s[i] {
				common = common[:i]
				break
			}
		}
		if len(s) < len(common) {
			common = common[:len(s)]
		}
		if len(common) == 0 {
			break
		}
	}
	return common
}

// extractLastArg extracts the last argument from the line, respecting quotes.
// Returns the raw (unquoted) argument value, its start position in the line,
// and whether it was inside quotes.
func extractLastArg(before string) (arg string, start int, quoted bool) {
	// Scan backwards to find where the last argument starts
	i := len(before) - 1

	// If trailing space with no open quote, it's a new empty argument
	if i >= 0 && before[i] == ' ' {
		return "", len(before), false
	}

	// Check if we're inside an open quote
	inQuote := false
	quoteStart := -1
	for j := 0; j < len(before); j++ {
		if before[j] == '"' {
			if inQuote {
				inQuote = false
			} else {
				inQuote = true
				quoteStart = j
			}
		}
	}

	if inQuote && quoteStart >= 0 {
		// We're inside an unclosed quote — the arg is everything after the quote char
		return before[quoteStart+1:], quoteStart, true
	}

	// Not in a quote — scan backwards to the last unquoted space
	for i >= 0 {
		if before[i] == ' ' {
			break
		}
		// If we hit a closing quote, find its opening quote
		if before[i] == '"' {
			j := i - 1
			for j >= 0 && before[j] != '"' {
				j--
			}
			if j >= 0 {
				// Found a complete quoted arg: strip the quotes, return inner value
				return before[j+1 : i], j, true
			}
		}
		i--
	}

	raw := before[i+1:]
	return raw, i + 1, false
}

// quoteIfNeeded wraps a completed path in double quotes if it contains spaces.
// If the arg was already quoted (open quote), it keeps the quote style and closes it
// only for single-match completions (trailing entries handled by caller).
func quoteIfNeeded(completed string, alreadyQuoted bool) string {
	if alreadyQuoted {
		// Caller started with a quote — keep content unquoted (the open quote is in prev)
		// Add closing quote only if this is a final completion (not a dir prefix)
		if strings.HasSuffix(completed, string(filepath.Separator)) || strings.HasSuffix(completed, "/") {
			return "\"" + completed
		}
		return "\"" + completed + "\""
	}
	if strings.Contains(completed, " ") {
		if strings.HasSuffix(completed, string(filepath.Separator)) || strings.HasSuffix(completed, "/") {
			return "\"" + completed
		}
		return "\"" + completed + "\""
	}
	return completed
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

	// Source env file (always loaded first)
	if envPath := config.ShellPaths.Env; envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			evaluator.EvalSource(envPath, env)
		}
	}

	// Source is file (interactive sessions only)
	if isPath := config.ShellPaths.Is; isPath != "" {
		if _, err := os.Stat(isPath); err == nil {
			evaluator.EvalSource(isPath, env)
		}
	}

	// Check if stdin is a terminal
	isTerminal := term.IsTerminal(int(os.Stdin.Fd()))

	var oldState *term.State
	var t *term.Terminal
	var le *lineEditor

	if isTerminal {
		oldState, err = term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			isTerminal = false
		} else {
			defer term.Restore(int(os.Stdin.Fd()), oldState)
			t = term.NewTerminal(terminalIO{in, out}, "")
			le = &lineEditor{}
			t.AutoCompleteCallback = func(line string, pos int, key rune) (newLine string, newPos int, ok bool) {
				if key == '\t' {
					return le.autoComplete(line, pos)
				}
				return "", 0, false
			}

			// Pre-fill history for arrow keys navigation
			// The term package's default History records the last 100 lines.
			for _, h := range utils.GetHistory() {
				if t.History != nil {
					t.History.Add(h)
				}
			}
		}
	}

	var commandBuffer string

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

		// Build prompt context
		lastStatus := "0"
		if val, ok := env.Get("LAST_STATUS"); ok {
			lastStatus = val.Inspect()
		}
		pctx := prompt.BuildContext(lastStatus)
		pctx.CWD = currentCWD

		// Construct the dynamic prompt
		var promptLine string
		if commandBuffer == "" {
			promptFmt := envStringOr(env, "PROMPT_LINE", prompt.DefaultPromptLine)
			promptLine = prompt.Render(promptFmt, pctx)
		} else {
			contFmt := envStringOr(env, "CONTINUATION_PROMPT", prompt.DefaultContinuationPrompt)
			promptLine = prompt.Render(contFmt, pctx)
		}

		var line string
		if isTerminal {
			t.SetPrompt(promptLine)
			line, err = t.ReadLine()
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

		// If line is empty, continue logic depends on buffer
		if trimmedLine == "" {
			if commandBuffer != "" {
				commandBuffer += "\n" // Add newline to buffer
				continue
			}
			continue // Skip empty lines if buffer is empty
		}

		// Append to buffer
		if commandBuffer == "" {
			commandBuffer = trimmedLine
		} else {
			commandBuffer += "\n" + trimmedLine
		}

		// Attempt to parse
		l := lexer.New(commandBuffer)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			// Check for unexpected EOF
			isIncomplete := false
			for _, msg := range p.Errors() {
				if strings.Contains(msg, "unexpected EOF") ||
					strings.Contains(msg, "expected next token to be }, got EOF instead") ||
					strings.Contains(msg, "expected next token to be ), got EOF instead") {
					isIncomplete = true
					break
				}
			}

			if isIncomplete {
				continue // Read more input
			}

			// Real Error
			for _, msg := range p.Errors() {
				fmt.Fprintf(out, "Parse Error: %s\n", msg)
			}
			// Reset buffer on error
			commandBuffer = ""
			utils.AddCommand(trimmedLine) // Add failed command to history? Or buffer?
			// Maybe add the whole buffer? For now, simplistic history.
		} else {
			// Success parsing

			// Add command to history before processing it
			utils.AddCommand(commandBuffer)

			// Pre-processing aliases?
			// Aliases usually apply to the first word of the command.
			// Only apply if it's a single line or check first word of buffer?
			// Applying to buffer:
			parts := strings.Fields(commandBuffer)
			if len(parts) > 0 {
				if expandedValue, ok := cfg.Aliases[parts[0]]; ok {
					commandBuffer = expandedValue + strings.TrimPrefix(commandBuffer, parts[0])
				}
			}

			if commandBuffer == "exit" || commandBuffer == "quit" {
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

			evaluated := evaluator.Eval(program, env)
			if evaluated != nil {
				if evaluated.Type() == object.ERROR_OBJ {
					fmt.Fprintf(out, "%s\n", evaluated.Inspect())
				}
			}

			// Reset buffer after execution
			commandBuffer = ""

			if isTerminal {
				oldState, _ = term.MakeRaw(int(os.Stdin.Fd()))
			}
		}
	}
}

// envStringOr reads a variable from the dush environment, returning fallback if not set.
func envStringOr(env *evaluator.Environment, name string, fallback string) string {
	if val, ok := env.Get(name); ok {
		return val.Inspect()
	}
	return fallback
}
