//go:build windows

package utils

import (
	"io/fs"
	"syscall"
	"time"
)

func getCreationTimePlatform(path string, info fs.FileInfo) time.Time {
	if sys := info.Sys(); sys != nil {
		if attr, ok := sys.(*syscall.Win32FileAttributeData); ok {
			return time.Unix(0, attr.CreationTime.Nanoseconds())
		}
	}
	return info.ModTime()
}
