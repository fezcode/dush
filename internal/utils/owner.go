package utils

import (
	"io/fs"
)

// GetOwnerAndGroupNames returns the owner and group names for a given file path and FileInfo.
// This function uses build tags to call platform-specific implementations.
// The 'path' argument should be the absolute path to the file.
func GetOwnerAndGroupNames(path string, info fs.FileInfo) (owner string, group string) {
	// This function body is a placeholder. The actual implementation will be provided
	// by owner_windows.go and owner_unix.go using build tags.
	// If no platform-specific implementation is found (e.g., unexpected OS),
	// it will return generic placeholders.
	return getOwnerAndGroupNamesPlatform(path, info)
}
