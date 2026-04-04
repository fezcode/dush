//go:build !windows

package utils

import (
	"io/fs"
	"time"
)

func getCreationTimePlatform(path string, info fs.FileInfo) time.Time {
	// Most Unix systems don't reliably expose birth time via Go's stdlib.
	// Fall back to modification time.
	return info.ModTime()
}
