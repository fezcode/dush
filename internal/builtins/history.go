package builtins

import (
	"context" // New import
	"fmt"
	"io"

	"dush/internal/utils" // Import the utils package
)

type HistoryCommand struct{}

func (c *HistoryCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) error {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Fprintln(out, "Usage: history\nDisplay the command history with line numbers.")
		return nil
	}

	history := utils.GetHistory()
	if len(history) == 0 {
		fmt.Fprintln(out, "No command history available.")
		return nil
	}

	for i, cmd := range history {
		fmt.Fprintf(out, "%5d  %s\n", i+1, cmd)
	}
	return nil
}

func init() {
	RegisterBuiltin("history", &HistoryCommand{})
}
