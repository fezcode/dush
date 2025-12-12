package utils

import (
	"path/filepath"
	"runtime"
)

// GetDisplayDirName returns a user-friendly name for the given directory path,
// handling special cases like Windows drive roots.
func GetDisplayDirName(path string) string {
	if runtime.GOOS == "windows" {
		// On Windows, if CWD is a drive root like "C:\", display "C:\" (or "C:") instead of "\"
		// Check for typical Windows drive root format (e.g., "C:\")
		if len(path) >= 3 && path[1] == ':' && (path[2] == '\\' || path[2] == '/') {
			// If it's just "C:\", display "C:\"
			if len(path) == 3 {
				return path
			} else {
				return filepath.Base(path)
			}
		} else if len(path) == 2 && path[1] == ':' { // "C:" without backslash
			return path
		} else {
			return filepath.Base(path)
		}
	} else {
		// For Unix-like systems, filepath.Base handles root "/" correctly
		return filepath.Base(path)
	}
}
