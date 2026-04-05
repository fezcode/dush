package builtins

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
)

// TrapHandler is called by the REPL/evaluator when a signal is received.
// The handler function receives the signal name and should execute the command string.
type TrapHandler func(command string)

var trapStore = struct {
	mu    sync.Mutex
	traps map[string]string // signal name -> command string
}{
	traps: make(map[string]string),
}

var trapHandler TrapHandler

// SetTrapHandler sets the function that will execute trap commands.
func SetTrapHandler(h TrapHandler) {
	trapHandler = h
}

// GetTrap returns the command string registered for a signal, if any.
func GetTrap(signal string) (string, bool) {
	trapStore.mu.Lock()
	defer trapStore.mu.Unlock()
	cmd, ok := trapStore.traps[strings.ToUpper(signal)]
	return cmd, ok
}

// ListTraps returns all registered traps.
func ListTraps() map[string]string {
	trapStore.mu.Lock()
	defer trapStore.mu.Unlock()
	result := make(map[string]string, len(trapStore.traps))
	for k, v := range trapStore.traps {
		result[k] = v
	}
	return result
}

var validSignals = map[string]bool{
	"INT": true, "TERM": true, "HUP": true,
	"EXIT": true, "ERR": true,
}

type TrapCommand struct{}

func (c *TrapCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) error {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			fmt.Fprint(out, `Usage: trap <command> <signal>...
       trap -l
       trap '' <signal>     (remove trap)

Set a command to run when the shell receives a signal.

Signals: INT, TERM, HUP, EXIT, ERR

Examples:
  trap 'echo bye' EXIT       Run "echo bye" on shell exit
  trap '' INT                 Remove the INT trap
  trap -l                    List all active traps
`)
			return nil
		}
	}

	if len(args) == 0 || (len(args) == 1 && args[0] == "-l") {
		// List traps
		traps := ListTraps()
		if len(traps) == 0 {
			fmt.Fprintln(out, "no traps set")
			return nil
		}
		for sig, cmd := range traps {
			fmt.Fprintf(out, "trap '%s' %s\n", cmd, sig)
		}
		return nil
	}

	if len(args) < 2 {
		return fmt.Errorf("trap: usage: trap <command> <signal>...")
	}

	command := args[0]
	signals := args[1:]

	trapStore.mu.Lock()
	defer trapStore.mu.Unlock()

	for _, sig := range signals {
		sig = strings.ToUpper(sig)
		if !validSignals[sig] {
			fmt.Fprintf(errOut, "trap: unknown signal: %s\n", sig)
			continue
		}
		if command == "" {
			delete(trapStore.traps, sig)
		} else {
			trapStore.traps[sig] = command
		}
	}

	return nil
}

func init() {
	RegisterBuiltin("trap", &TrapCommand{})
}
