//go:build !windows

package utils

func init() {
	// ANSI colors are natively supported on Unix terminals.
}

// EnableVirtualTerminalProcessing is a no-op on non-Windows systems.
func EnableVirtualTerminalProcessing() {}
