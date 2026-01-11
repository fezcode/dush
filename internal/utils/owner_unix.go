//go:build !windows

package utils

import (
	"io/fs"
	"os/user"
	"runtime"
	"strconv"
	"syscall"
)

func getOwnerAndGroupNamesPlatform(path string, info fs.FileInfo) (owner string, group string) {
	owner = "unknown"
	group = "unknown"

	// Only attempt this on Unix-like systems
	switch runtime.GOOS {
	case "linux", "darwin", "freebsd", "netbsd", "openbsd":
		if stat, ok := info.Sys().(*syscall.Stat_t); ok {
			// Get Owner
			u, err := user.LookupId(strconv.FormatUint(uint64(stat.Uid), 10))
			if err == nil {
				owner = u.Username
			} else {
				owner = strconv.FormatUint(uint64(stat.Uid), 10) // Fallback to numeric ID
			}

			// Get Group
			g, err := user.LookupGroupId(strconv.FormatUint(uint64(stat.Gid), 10))
			if err == nil {
				group = g.Name
			} else {
				group = strconv.FormatUint(uint64(stat.Gid), 10) // Fallback to numeric ID
			}
		}
	}
	return owner, group
}
