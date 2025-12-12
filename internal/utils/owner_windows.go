//go:build windows

package utils

import (
	"io/fs"
	"syscall"
	"unsafe"
)

// 1. Load the necessary Windows DLLs
var (
	modadvapi32 = syscall.NewLazyDLL("advapi32.dll")
	modkernel32 = syscall.NewLazyDLL("kernel32.dll")

	procGetNamedSecurityInfoW = modadvapi32.NewProc("GetNamedSecurityInfoW")
	procLookupAccountSidW     = modadvapi32.NewProc("LookupAccountSidW")
	procLocalFree             = modkernel32.NewProc("LocalFree")
)

// 2. Define Windows Constants (Magic Numbers)
const (
	SE_FILE_OBJECT             = 1
	OWNER_SECURITY_INFORMATION = 0x00000001
	GROUP_SECURITY_INFORMATION = 0x00000002
)

type SID struct{} // Opaque structure

// getOwnerAndGroupNamesPlatform returns the owner and group names for a given file path on Windows.
func getOwnerAndGroupNamesPlatform(path string, info fs.FileInfo) (owner string, group string) {
	// path is the full path to the file
	var pSidOwner, pSidGroup *SID
	var pSecurityDescriptor uintptr

	// 3. Call GetNamedSecurityInfoW
	// This retrieves the Security Descriptor from the file system.
	// Arguments: Path, ObjectType, InfoFlags, OwnerSID, GroupSID, Dacl, Sacl, SecDesc
	ret, _, _ := procGetNamedSecurityInfoW.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(path))), // Use the provided 'path' argument
		uintptr(SE_FILE_OBJECT),
		uintptr(OWNER_SECURITY_INFORMATION|GROUP_SECURITY_INFORMATION),
		uintptr(unsafe.Pointer(&pSidOwner)),
		uintptr(unsafe.Pointer(&pSidGroup)),
		0, // DACL (not needed)
		0, // SACL (not needed)
		uintptr(unsafe.Pointer(&pSecurityDescriptor)),
	)

	if ret != 0 {
		return "Unknown", "Unknown" // Failed to get info
	}

	// Ensure we free the memory Windows allocated for the descriptor
	defer procLocalFree.Call(pSecurityDescriptor)

	// 4. Translate SIDs to Names
	ownerName := lookupSID(pSidOwner)
	groupName := lookupSID(pSidGroup)

	return ownerName, groupName
}

// Helper to call LookupAccountSidW
func lookupSID(sid *SID) string {
	var nameSize, domainSize uint32
	var peUse uint32

	// First Call: Ask Windows how much memory we need for the Name and Domain
	// This call is expected to fail with ERROR_INSUFFICIENT_BUFFER, but it populates the sizes.
	procLookupAccountSidW.Call(
		0, // System Name (NULL = local)
		uintptr(unsafe.Pointer(sid)),
		0,
		uintptr(unsafe.Pointer(&nameSize)),
		0,
		uintptr(unsafe.Pointer(&domainSize)),
		uintptr(unsafe.Pointer(&peUse)),
	)

	if nameSize == 0 {
		return "Unknown"
	}

	// Allocate buffers based on the sizes Windows gave us
	nameBuf := make([]uint16, nameSize)
	domainBuf := make([]uint16, domainSize)

	// Second Call: Actually get the names
	ret, _, _ := procLookupAccountSidW.Call(
		0,
		uintptr(unsafe.Pointer(sid)),
		uintptr(unsafe.Pointer(&nameBuf[0])),
		uintptr(unsafe.Pointer(&nameSize)),
		uintptr(unsafe.Pointer(&domainBuf[0])),
		uintptr(unsafe.Pointer(&domainSize)), // Corrected argument
		uintptr(unsafe.Pointer(&peUse)),
	)

	if ret == 0 {
		return "Unknown"
	}

	// Windows returns "Domain" and "Name". Usually we just want "Name" (e.g., "Administrator")
	// If you want "COMPUTER\User", you can join them here.
	return syscall.UTF16ToString(nameBuf)
}
