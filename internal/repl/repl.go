package repl

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

const PROMPT = ">> "

// Start starts the Read-Eval-Print Loop.
// It takes an io.Reader for input and an io.Writer for output.
func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Fprintf(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return // EOF or error
		}

		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == "exit" || trimmedLine == "quit" {
			fmt.Fprintf(out, "Exiting dush REPL.\n")
			return
		}

		// For now, just echo the input
		fmt.Fprintf(out, "Echo: %s\n", line)
	}
}
