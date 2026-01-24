package builtins

import (
	"context"
	"fmt"
	"io"
	"strings"
)

type EchoCommand struct{}

func (c *EchoCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) error {
	// Simple echo implementation
	// TODO: Support flags like -n (no newline)
	fmt.Fprintln(out, strings.Join(args, " "))
	return nil
}

func init() {
	RegisterBuiltin("echo", &EchoCommand{})
}
