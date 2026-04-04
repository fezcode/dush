package config

import (
	"os"
	"path/filepath"
)

// Paths holds the resolved file paths for dush configuration.
// Override individual fields before calling Resolve() to use custom locations.
type Paths struct {
	Env     string // env file path (always loaded)
	Is      string // is file path (interactive only)
	History string // history file path
}

// DefaultPaths returns paths under ~/.dush/ with defaults.
// Any field already set (non-empty) is left as-is.
func (p *Paths) Resolve() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	dushDir := filepath.Join(home, ".dush")

	if p.Env == "" {
		p.Env = filepath.Join(dushDir, "env")
	}
	if p.Is == "" {
		p.Is = filepath.Join(dushDir, "is")
	}
	if p.History == "" {
		p.History = filepath.Join(dushDir, "history")
	}
}

// ShellPaths is the global resolved paths instance.
// Set fields before calling Resolve() to override defaults.
var ShellPaths = &Paths{}
