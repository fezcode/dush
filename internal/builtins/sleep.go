package builtins

import (
	"context" // New import
	"fmt"
	"io"
	"strconv"
	"time"
)

type SleepCommand struct{}

func (c *SleepCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: sleep <duration>")
	}

	durationStr := args[0]
	var d time.Duration
	var err error

	// Try to parse as duration string (e.g., "1.5s", "10m")
	if d, err = time.ParseDuration(durationStr); err != nil {
		// If that fails, try to parse as a simple number of seconds
		seconds, convErr := strconv.ParseFloat(durationStr, 64)
		if convErr != nil {
			return fmt.Errorf("invalid duration: %s. Expected a number of seconds or a duration string (e.g., 1.5s, 10m)", durationStr)
		}
		d = time.Duration(seconds * float64(time.Second))
	}

	// Make sleep interruptible using the context
	select {
	case <-time.After(d):
		// Duration passed naturally
		return nil
	case <-ctx.Done():
		// Context was cancelled (e.g., by Ctrl+C)
		return ctx.Err() // This will be context.Canceled
	}
}

func init() {
	RegisterBuiltin("sleep", &SleepCommand{})
}
