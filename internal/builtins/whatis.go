package builtins

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"dush/internal/config"
)

type WhatisCommand struct{}

func (c *WhatisCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: whatis <name>...")
	}

	for _, name := range args {
		if strings.HasPrefix(name, "-") {
			if name == "-h" || name == "--help" {
				fmt.Fprintln(out, "Usage: whatis <name>...\nShow how each name would be interpreted by the shell.")
				return nil
			}
			return fmt.Errorf("whatis: unknown option '%s'", name)
		}

		found := false

		// Check aliases first
		cfg := config.GetConfig()
		if val, ok := cfg.Aliases[name]; ok {
			fmt.Fprintf(out, "%s is aliased to '%s'\n", name, val)
			found = true
		}

		// Check builtins
		if _, ok := GetCommand(name); ok {
			fmt.Fprintf(out, "%s is a shell builtin\n", name)
			found = true
		}

		// Check special commands handled in evaluator
		switch name {
		case "source", ".", "read", "unset":
			if !found {
				fmt.Fprintf(out, "%s is a shell builtin\n", name)
				found = true
			}
		}

		// Check external commands on PATH
		if path, err := exec.LookPath(name); err == nil {
			fmt.Fprintf(out, "%s is %s\n", name, path)
			found = true
		}

		if !found {
			fmt.Fprintf(errOut, "%s: not found\n", name)
		}
	}
	return nil
}

func init() {
	RegisterBuiltin("whatis", &WhatisCommand{})
}
