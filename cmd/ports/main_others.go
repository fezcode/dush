//go:build !windows && !linux && !darwin

package main

import "fmt"

func listPortsOS() ([]PortInfo, error) {
	return nil, fmt.Errorf("port listing is not implemented for this operating system")
}
