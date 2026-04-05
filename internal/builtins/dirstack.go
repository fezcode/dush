package builtins

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"dush/internal/app"
)

var dirStack = struct {
	mu    sync.Mutex
	stack []string
}{}

type PushdirCommand struct{}

func (c *PushdirCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) error {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			fmt.Fprintln(out, "Usage: pushdir <directory>\nPush the current directory onto the stack and change to <directory>.")
			return nil
		}
	}

	if len(args) == 0 {
		return fmt.Errorf("pushdir: no directory specified")
	}

	appInstance := app.GetApp()
	cwd := appInstance.GetCurrentDir()

	target := args[0]
	if err := os.Chdir(target); err != nil {
		return fmt.Errorf("pushdir: %w", err)
	}

	newWD, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("pushdir: %w", err)
	}

	dirStack.mu.Lock()
	dirStack.stack = append(dirStack.stack, cwd)
	dirStack.mu.Unlock()

	if err := appInstance.SetCurrentDir(newWD); err != nil {
		return fmt.Errorf("pushdir: %w", err)
	}

	fmt.Fprintln(out, newWD)
	return nil
}

type PopdirCommand struct{}

func (c *PopdirCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) error {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			fmt.Fprintln(out, "Usage: popdir\nPop the top directory from the stack and change to it.")
			return nil
		}
	}

	dirStack.mu.Lock()
	if len(dirStack.stack) == 0 {
		dirStack.mu.Unlock()
		return fmt.Errorf("popdir: directory stack is empty")
	}
	target := dirStack.stack[len(dirStack.stack)-1]
	dirStack.stack = dirStack.stack[:len(dirStack.stack)-1]
	dirStack.mu.Unlock()

	if err := os.Chdir(target); err != nil {
		return fmt.Errorf("popdir: %w", err)
	}

	newWD, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("popdir: %w", err)
	}

	appInstance := app.GetApp()
	if err := appInstance.SetCurrentDir(newWD); err != nil {
		return fmt.Errorf("popdir: %w", err)
	}

	fmt.Fprintln(out, newWD)
	return nil
}

type DirsCommand struct{}

func (c *DirsCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) error {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			fmt.Fprintln(out, "Usage: dirs [-c]\nShow the directory stack. Use -c to clear it.")
			return nil
		}
	}

	clear := false
	for _, arg := range args {
		if arg == "-c" || arg == "--clear" {
			clear = true
		}
	}

	dirStack.mu.Lock()
	defer dirStack.mu.Unlock()

	if clear {
		dirStack.stack = dirStack.stack[:0]
		return nil
	}

	appInstance := app.GetApp()
	cwd := appInstance.GetCurrentDir()

	// Print current dir first, then stack top-to-bottom
	var parts []string
	parts = append(parts, cwd)
	for i := len(dirStack.stack) - 1; i >= 0; i-- {
		parts = append(parts, dirStack.stack[i])
	}
	fmt.Fprintln(out, strings.Join(parts, "  "))
	return nil
}

func init() {
	RegisterBuiltin("pushdir", &PushdirCommand{})
	RegisterBuiltin("popdir", &PopdirCommand{})
	RegisterBuiltin("dirs", &DirsCommand{})
}
