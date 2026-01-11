//go:build darwin

package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"syscall"
	"unsafe"
)

const (
	SYS_PROC_INFO = 336

	PROC_INFO_CALL_LISTPIDS  = 1
	PROC_INFO_CALL_PIDFDINFO = 3

	PROC_PIDLISTFDS      = 1
	PROC_PIDFDSOCKETINFO = 3

	PROX_FDTYPE_SOCKET = 2

	SOCKINFO_TCP = 2
	SOCKINFO_UDP = 3
)

type proc_fdinfo struct {
	ProcFd     int32
	ProcFdType uint32
}

// Simplified socket info structures based on Darwin headers
type socket_fdinfo struct {
	Pfi proc_fileinfo
	Psi socket_info
}

type proc_fileinfo struct {
	Openflags uint32
	Status    uint32
	Offset    int64
	Receiver  int32
	Flag      uint32
}

type socket_info struct {
	SoiStat     [64]byte // vinfo_stat
	SoiSo       uint64   // uint64_t
	SoiPcb      uint64   // uint64_t
	SoiType     int32
	SoiProtocol int32
	SoiFamily   int32
	SoiOptions  short
	SoiLinger   short
	SoiState    short
	SoiQlimit   short
	SoiError    short
	SoiOobmark  uint32
	SoiRcv      [32]byte // sockbuf_info
	SoiSnd      [32]byte // sockbuf_info
	SoiKind     int32
	// ... followed by union of in_sockinfo, tcp_sockinfo, etc.
}

type short int16

func listPortsOS() ([]PortInfo, error) {
	// 1. List all PIDs
	pids, err := listPids()
	if err != nil {
		return nil, err
	}

	var allPorts []PortInfo

	for _, pid := range pids {
		// 2. List FDs for this PID
		fds, err := listFds(pid)
		if err != nil {
			continue
		}

		for _, fd := range fds {
			if fd.ProcFdType == PROX_FDTYPE_SOCKET {
				// 3. Get Socket Info
				info, err := getSocketInfo(pid, fd.ProcFd)
				if err != nil {
					continue
				}

				p := parseSocketInfo(pid, info)
				if p != nil {
					allPorts = append(allPorts, *p)
				}
			}
		}
	}

	return allPorts, nil
}

func listPids() ([]int32, error) {
	// First call to get size
	res, _, err := syscall.Syscall6(SYS_PROC_INFO, PROC_INFO_CALL_LISTPIDS, 1, 0, 0, 0, 0)
	if res <= 0 {
		return nil, err
	}

	size := int(res)
	pids := make([]int32, size/4)
	res, _, err = syscall.Syscall6(SYS_PROC_INFO, PROC_INFO_CALL_LISTPIDS, 1, 0, 0, uintptr(unsafe.Pointer(&pids[0])), uintptr(size))
	if res <= 0 {
		return nil, err
	}

	return pids[:int(res)/4], nil
}

func listFds(pid int32) ([]proc_fdinfo, error) {
	// Get buffer size needed
	res, _, err := syscall.Syscall6(SYS_PROC_INFO, PROC_INFO_CALL_PIDFDINFO, uintptr(pid), PROC_PIDLISTFDS, 0, 0, 0)
	if res <= 0 {
		return nil, err
	}

	size := int(res)
	fds := make([]proc_fdinfo, size/int(unsafe.Sizeof(proc_fdinfo{})))
	res, _, err = syscall.Syscall6(SYS_PROC_INFO, PROC_INFO_CALL_PIDFDINFO, uintptr(pid), PROC_PIDLISTFDS, 0, uintptr(unsafe.Pointer(&fds[0])), uintptr(size))
	if res <= 0 {
		return nil, err
	}

	return fds[:int(res)/int(unsafe.Sizeof(proc_fdinfo{}))], nil
}

// We use a large enough buffer to accommodate socket_fdinfo and its unions
func getSocketInfo(pid int32, fd int32) ([]byte, error) {
	const bufSize = 2048 // socket_fdinfo is roughly 1200-1500 bytes depending on version
	buf := make([]byte, bufSize)
	res, _, err := syscall.Syscall6(SYS_PROC_INFO, PROC_INFO_CALL_PIDFDINFO, uintptr(pid), PROC_PIDFDSOCKETINFO, uintptr(fd), uintptr(unsafe.Pointer(&buf[0])), uintptr(bufSize))
	if res <= 0 {
		return nil, err
	}
	return buf[:res], nil
}

func parseSocketInfo(pid int32, data []byte) *PortInfo {
	if len(data) < 144 { // Minimum size to reach kind and start of unions
		return nil
	}

	// Offset to soi_kind is around 140 (proc_fileinfo is 32, soi_stat is 64, etc.)
	// This is very fragile without CGO, but let's try a heuristic or offset-based approach.
	// Based on Darwin xnu headers:
	// proc_fileinfo: 32 bytes
	// socket_info:
	//   soi_stat: 64
	//   soi_so: 8
	//   soi_pcb: 8
	//   soi_type: 4
	//   soi_protocol: 4
	//   soi_family: 4
	//   soi_options: 2
	//   soi_linger: 2
	//   soi_state: 2
	//   soi_qlimit: 2
	//   soi_error: 2
	//   soi_oobmark: 4
	//   soi_rcv: 32
	//   soi_snd: 32
	//   soi_kind: 4  <-- Offset 32 + 64 + 8 + 8 + 4 + 4 + 4 + 2 + 2 + 2 + 2 + 2 + 4 + 32 + 32 = 198? No, padding.

	// Actually, let's use the known offsets for common versions.
	// A better way is to look for the protocol and family.

	// soi_family is at offset 32 + 64 + 8 + 8 + 4 + 4 = 120
	family := binary.LittleEndian.Uint32(data[120:124])
	if family != 2 && family != 30 { // AF_INET = 2, AF_INET6 = 30
		return nil
	}

	// soi_kind is at offset 192 (on 64-bit)
	kind := binary.LittleEndian.Uint32(data[192:196])

	proto := "???"
	state := "-"
	var local, remote string

	if kind == SOCKINFO_TCP {
		proto = "TCP"
		// tcp_sockinfo starts at 196 + 4 (padding) = 200?
		// Actually the union starts right after soi_kind.
		// tcp_sockinfo:
		//   ins_sockinfo: 104 bytes
		//   tcpsi_state: 4

		stateNum := binary.LittleEndian.Uint32(data[196+104 : 196+108])
		state = decodeTcpState(stateNum)
		local, remote = parseInSockInfo(data[196:196+104], family)
	} else if kind == SOCKINFO_UDP {
		proto = "UDP"
		local, remote = parseInSockInfo(data[196:196+104], family)
	} else {
		return nil
	}

	return &PortInfo{
		Proto:  proto,
		Local:  local,
		Remote: remote,
		State:  state,
		PID:    fmt.Sprintf("%d", pid),
	}
}

func parseInSockInfo(data []byte, family uint32) (string, string) {
	// in_sockinfo:
	//   ins_fport: 2
	//   ins_lport: 2
	//   ins_faddr: 16
	//   ins_laddr: 16
	//   ...
	fport := binary.BigEndian.Uint16(data[0:2])
	lport := binary.BigEndian.Uint16(data[2:4])

	var laddr, faddr string
	if family == 2 { // IPv4
		laddr = net.IP(data[4:8]).String()
		faddr = net.IP(data[20:24]).String()
	} else { // IPv6
		laddr = net.IP(data[4:20]).String()
		faddr = net.IP(data[20:36]).String()
	}

	return fmt.Sprintf("%s:%d", laddr, lport), fmt.Sprintf("%s:%d", faddr, fport)
}

func decodeTcpState(state uint32) string {
	states := []string{"CLOSED", "LISTEN", "SYN_SENT", "SYN_RCVD", "ESTABLISHED", "CLOSE_WAIT", "FIN_WAIT_1", "CLOSING", "LAST_ACK", "FIN_WAIT_2", "TIME_WAIT"}
	if int(state) < len(states) {
		return states[state]
	}
	return fmt.Sprintf("%d", state)
}
