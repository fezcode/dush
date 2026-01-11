//go:build linux

package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func listPortsOS() ([]PortInfo, error) {
	var allPorts []PortInfo

	files := []struct {
		path  string
		proto string
	}{
		{"/proc/net/tcp", "TCP"},
		{"/proc/net/tcp6", "TCP6"},
		{"/proc/net/udp", "UDP"},
		{"/proc/net/udp6", "UDP6"},
	}

	inodeToPid := make(map[string]string)
	entries, _ := os.ReadDir("/proc")
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid := entry.Name()
		if _, err := strconv.Atoi(pid); err != nil {
			continue
		}

		fdPath := filepath.Join("/proc", pid, "fd")
		fds, err := os.ReadDir(fdPath)
		if err != nil {
			continue
		}

		comm, _ := os.ReadFile(filepath.Join("/proc", pid, "comm"))
		progName := strings.TrimSpace(string(comm))

		for _, fd := range fds {
			link, err := os.Readlink(filepath.Join(fdPath, fd.Name()))
			if err != nil {
				continue
			}
			if strings.HasPrefix(link, "socket:[") {
				inode := link[8 : len(link)-1]
				inodeToPid[inode] = fmt.Sprintf("%s/%s", pid, progName)
			}
		}
	}

	for _, f := range files {
		ports, err := parseProcNet(f.path, f.proto)
		if err == nil {
			for i := range ports {
				if pid, ok := inodeToPid[ports[i].Inode]; ok {
					ports[i].PID = pid
				}
				allPorts = append(allPorts, ports[i])
			}
		}
	}

	return allPorts, nil
}

func parseProcNet(path string, proto string) ([]PortInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var ports []PortInfo
	scanner := bufio.NewScanner(file)
	scanner.Scan() // skip header

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 10 {
			continue
		}

		local := parseHexAddr(fields[1])
		remote := parseHexAddr(fields[2])
		state := decodeState(fields[3], proto)
		inode := fields[9]

		ports = append(ports, PortInfo{
			Proto:  proto,
			Local:  local,
			Remote: remote,
			State:  state,
			Inode:  inode,
			PID:    "-",
		})
	}
	return ports, nil
}

func parseHexAddr(hexStr string) string {
	parts := strings.Split(hexStr, ":")
	if len(parts) != 2 {
		return hexStr
	}

	addrHex, _ := hex.DecodeString(parts[0])
	port, _ := strconv.ParseUint(parts[1], 16, 16)

	if len(addrHex) == 4 {
		// IPv4: /proc/net/tcp stores it in little-endian (for some reason)
		return fmt.Sprintf("%d.%d.%d.%d:%d", addrHex[3], addrHex[2], addrHex[1], addrHex[0], port)
	} else if len(addrHex) == 16 {
		// IPv6: /proc/net/tcp6 stores it in 4 32-bit words, each in host-endian
		// We need to swap each 4-byte block
		for i := 0; i < 16; i += 4 {
			addrHex[i], addrHex[i+1], addrHex[i+2], addrHex[i+3] = addrHex[i+3], addrHex[i+2], addrHex[i+1], addrHex[i]
		}
		ip := net.IP(addrHex)
		return fmt.Sprintf("[%s]:%d", ip.String(), port)
	}

	return hexStr
}

func decodeState(stateHex string, proto string) string {
	if strings.HasPrefix(proto, "UDP") {
		return "-"
	}
	switch stateHex {
	case "01":
		return "LISTEN"
	case "02":
		return "SYN_SENT"
	case "03":
		return "SYN_RECV"
	case "04":
		return "ESTAB"
	case "05":
		return "FIN_WAIT1"
	case "06":
		return "FIN_WAIT2"
	case "07":
		return "CLOSE_WAIT"
	case "08":
		return "LAST_ACK"
	case "09":
		return "CLOSING"
	case "0A":
		return "TIME_WAIT"
	case "0B":
		return "CLOSE"
	default:
		return stateHex
	}
}
