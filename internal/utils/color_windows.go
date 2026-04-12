//go:build windows

package utils

import (
	"os"
	"syscall"
	"unsafe"
)

func init() {
	EnableVirtualTerminalProcessing()
}

func EnableVirtualTerminalProcessing() {
	const ENABLE_VIRTUAL_TERMINAL_PROCESSING = 0x0004

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getConsoleMode := kernel32.NewProc("GetConsoleMode")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")

	// Enable on stdout
	h := syscall.Handle(os.Stdout.Fd())
	var mode uint32
	r, _, _ := getConsoleMode.Call(uintptr(h), uintptr(unsafe.Pointer(&mode)))
	if r != 0 {
		setConsoleMode.Call(uintptr(h), uintptr(mode|ENABLE_VIRTUAL_TERMINAL_PROCESSING))
	}

	// Enable on stderr too
	h = syscall.Handle(os.Stderr.Fd())
	r, _, _ = getConsoleMode.Call(uintptr(h), uintptr(unsafe.Pointer(&mode)))
	if r != 0 {
		setConsoleMode.Call(uintptr(h), uintptr(mode|ENABLE_VIRTUAL_TERMINAL_PROCESSING))
	}
}
