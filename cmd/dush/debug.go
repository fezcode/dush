package main

import (
	"dush/cmd/dush/buildinfo"
	"fmt"
)

// DebugPrint prints a formatted string to stdout only if buildinfo.IsTestBuild() is true.
func DebugPrint(format string, a ...interface{}) {
	if buildinfo.IsTestBuild() {
		fmt.Printf("[DEBUG] "+format+"\n", a...)
	}
}
