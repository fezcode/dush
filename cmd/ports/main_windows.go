//go:build windows

package main

import (
	"fmt"
	"net"
	"syscall"
	"unsafe"
)

var (
	modiphlpapi             = syscall.NewLazyDLL("iphlpapi.dll")
	procGetExtendedTcpTable = modiphlpapi.NewProc("GetExtendedTcpTable")
	procGetExtendedUdpTable = modiphlpapi.NewProc("GetExtendedUdpTable")
)

const (
	TCP_TABLE_OWNER_PID_ALL = 5
	UDP_TABLE_OWNER_PID     = 1
	AF_INET                 = 2
	AF_INET6                = 23
)

type MIB_TCPROW_OWNER_PID struct {
	State      uint32
	LocalAddr  uint32
	LocalPort  uint32
	RemoteAddr uint32
	RemotePort uint32
	OwningPid  uint32
}

type MIB_UDPROW_OWNER_PID struct {
	LocalAddr uint32
	LocalPort uint32
	OwningPid uint32
}

func listPortsOS() ([]PortInfo, error) {
	var ports []PortInfo

	// TCP IPv4
	if tcp4, err := getTcpTable(AF_INET); err == nil {
		ports = append(ports, tcp4...)
	}
	// TCP IPv6
	if tcp6, err := getTcp6Table(AF_INET6); err == nil {
		ports = append(ports, tcp6...)
	}

	// UDP IPv4
	if udp4, err := getUdpTable(AF_INET); err == nil {
		ports = append(ports, udp4...)
	}
	// UDP IPv6
	if udp6, err := getUdp6Table(AF_INET6); err == nil {
		ports = append(ports, udp6...)
	}

	return ports, nil
}

func getTcpTable(family uint32) ([]PortInfo, error) {
	var size uint32
	procGetExtendedTcpTable.Call(0, uintptr(unsafe.Pointer(&size)), 0, uintptr(family), TCP_TABLE_OWNER_PID_ALL, 0)

	buf := make([]byte, size)
	ret, _, _ := procGetExtendedTcpTable.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)), 0, uintptr(family), TCP_TABLE_OWNER_PID_ALL, 0)
	if ret != 0 {
		return nil, fmt.Errorf("GetExtendedTcpTable failed: %d", ret)
	}

	numEntries := *(*uint32)(unsafe.Pointer(&buf[0]))
	table := (*[1 << 20]MIB_TCPROW_OWNER_PID)(unsafe.Pointer(&buf[4]))[:numEntries:numEntries]

	var ports []PortInfo
	for _, row := range table {
		ports = append(ports, PortInfo{
			Proto:  "TCP",
			Local:  fmt.Sprintf("%s:%d", parseIPv4(row.LocalAddr), ntohs(uint16(row.LocalPort))),
			Remote: fmt.Sprintf("%s:%d", parseIPv4(row.RemoteAddr), ntohs(uint16(row.RemotePort))),
			State:  decodeTcpState(row.State),
			PID:    fmt.Sprintf("%d", row.OwningPid),
		})
	}
	return ports, nil
}

type MIB_TCP6ROW_OWNER_PID struct {
	LocalAddr     [16]byte
	LocalScopeId  uint32
	LocalPort     uint32
	RemoteAddr    [16]byte
	RemoteScopeId uint32
	RemotePort    uint32
	State         uint32
	OwningPid     uint32
}

func getTcp6Table(family uint32) ([]PortInfo, error) {
	var size uint32
	procGetExtendedTcpTable.Call(0, uintptr(unsafe.Pointer(&size)), 0, uintptr(family), TCP_TABLE_OWNER_PID_ALL, 0)

	buf := make([]byte, size)
	ret, _, _ := procGetExtendedTcpTable.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)), 0, uintptr(family), TCP_TABLE_OWNER_PID_ALL, 0)
	if ret != 0 {
		return nil, fmt.Errorf("GetExtendedTcpTable failed: %d", ret)
	}

	numEntries := *(*uint32)(unsafe.Pointer(&buf[0]))
	// The table structure for IPv6 is slightly different due to alignment
	tablePtr := uintptr(unsafe.Pointer(&buf[4]))

	var ports []PortInfo
	for i := uint32(0); i < numEntries; i++ {
		row := (*MIB_TCP6ROW_OWNER_PID)(unsafe.Pointer(tablePtr + uintptr(i)*unsafe.Sizeof(MIB_TCP6ROW_OWNER_PID{})))
		ports = append(ports, PortInfo{
			Proto:  "TCP6",
			Local:  fmt.Sprintf("[%s]:%d", net.IP(row.LocalAddr[:]).String(), ntohs(uint16(row.LocalPort))),
			Remote: fmt.Sprintf("[%s]:%d", net.IP(row.RemoteAddr[:]).String(), ntohs(uint16(row.RemotePort))),
			State:  decodeTcpState(row.State),
			PID:    fmt.Sprintf("%d", row.OwningPid),
		})
	}
	return ports, nil
}

func getUdpTable(family uint32) ([]PortInfo, error) {
	var size uint32
	procGetExtendedUdpTable.Call(0, uintptr(unsafe.Pointer(&size)), 0, uintptr(family), UDP_TABLE_OWNER_PID, 0)

	buf := make([]byte, size)
	ret, _, _ := procGetExtendedUdpTable.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)), 0, uintptr(family), UDP_TABLE_OWNER_PID, 0)
	if ret != 0 {
		return nil, fmt.Errorf("GetExtendedUdpTable failed: %d", ret)
	}

	numEntries := *(*uint32)(unsafe.Pointer(&buf[0]))
	table := (*[1 << 20]MIB_UDPROW_OWNER_PID)(unsafe.Pointer(&buf[4]))[:numEntries:numEntries]

	var ports []PortInfo
	for _, row := range table {
		ports = append(ports, PortInfo{
			Proto:  "UDP",
			Local:  fmt.Sprintf("%s:%d", parseIPv4(row.LocalAddr), ntohs(uint16(row.LocalPort))),
			Remote: "*:*",
			State:  "-",
			PID:    fmt.Sprintf("%d", row.OwningPid),
		})
	}
	return ports, nil
}

type MIB_UDP6ROW_OWNER_PID struct {
	LocalAddr    [16]byte
	LocalScopeId uint32
	LocalPort    uint32
	OwningPid    uint32
}

func getUdp6Table(family uint32) ([]PortInfo, error) {
	var size uint32
	procGetExtendedUdpTable.Call(0, uintptr(unsafe.Pointer(&size)), 0, uintptr(family), UDP_TABLE_OWNER_PID, 0)

	buf := make([]byte, size)
	ret, _, _ := procGetExtendedUdpTable.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)), 0, uintptr(family), UDP_TABLE_OWNER_PID, 0)
	if ret != 0 {
		return nil, fmt.Errorf("GetExtendedUdpTable failed: %d", ret)
	}

	numEntries := *(*uint32)(unsafe.Pointer(&buf[0]))
	tablePtr := uintptr(unsafe.Pointer(&buf[4]))

	var ports []PortInfo
	for i := uint32(0); i < numEntries; i++ {
		row := (*MIB_UDP6ROW_OWNER_PID)(unsafe.Pointer(tablePtr + uintptr(i)*unsafe.Sizeof(MIB_UDP6ROW_OWNER_PID{})))
		ports = append(ports, PortInfo{
			Proto:  "UDP6",
			Local:  fmt.Sprintf("[%s]:%d", net.IP(row.LocalAddr[:]).String(), ntohs(uint16(row.LocalPort))),
			Remote: "*:*",
			State:  "-",
			PID:    fmt.Sprintf("%d", row.OwningPid),
		})
	}
	return ports, nil
}

func parseIPv4(addr uint32) string {
	return net.IPv4(byte(addr), byte(addr>>8), byte(addr>>16), byte(addr>>24)).String()
}

func ntohs(i uint16) uint16 {
	return (i<<8)&0xff00 | (i>>8)&0x00ff
}

func decodeTcpState(state uint32) string {
	switch state {
	case 1:
		return "CLOSED"
	case 2:
		return "LISTEN"
	case 3:
		return "SYN_SENT"
	case 4:
		return "SYN_RCVD"
	case 5:
		return "ESTAB"
	case 6:
		return "FIN_WAIT1"
	case 7:
		return "FIN_WAIT2"
	case 8:
		return "CLOSE_WAIT"
	case 9:
		return "CLOSING"
	case 10:
		return "LAST_ACK"
	case 11:
		return "TIME_WAIT"
	case 12:
		return "DELETE_TCB"
	default:
		return fmt.Sprintf("%d", state)
	}
}
