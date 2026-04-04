package utils

import (
	"io/fs"
	"time"
)

// GetCreationTime returns the creation time (birth time) of a file.
// Falls back to modification time on platforms that don't support it.
func GetCreationTime(path string, info fs.FileInfo) time.Time {
	return getCreationTimePlatform(path, info)
}
