package builtins

import (
	"context"
	"fmt"
	"io"
	"strings"
)

type EchoCommand struct{}

func (c *EchoCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) error {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Fprintln(out, "Usage: echo [-n] [args...]\nPrint arguments separated by spaces.\n\nOptions:\n  -n    Do not output a trailing newline\n  -h    Show this help")
		return nil
	}

	noNewline := false
	startIdx := 0

	if len(args) > 0 && args[0] == "-n" {
		noNewline = true
		startIdx = 1
	}

	output := strings.Join(args[startIdx:], " ")
	if noNewline {
		fmt.Fprint(out, output)
	} else {
		fmt.Fprintln(out, output)
	}
	return nil
}

func init() {
	RegisterBuiltin("echo", &EchoCommand{})
}
