package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"dush/internal/utils"
)

type PortInfo struct {
	Proto  string
	Local  string
	Remote string
	State  string
	PID    string
	Inode  string
}

func main() {
	showTCP := flag.Bool("tcp", true, "Show TCP ports")
	showUDP := flag.Bool("udp", true, "Show UDP ports")
	showListen := flag.Bool("listen", false, "Show only listening ports")
	showPID := flag.Bool("pid", true, "Show PID (if available)")
	flag.Parse()

	var ports []PortInfo
	var err error

	ports, err = listPorts()

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	printPorts(ports, *showTCP, *showUDP, *showListen, *showPID)
}

func printPorts(ports []PortInfo, tcp, udp, listenOnly, pid bool) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	header := "PROTO\tLOCAL ADDRESS\tFOREIGN ADDRESS\tSTATE"
	if pid {
		header += "\tPID/Program"
	}
	fmt.Fprintln(w, utils.Colorize(header, utils.StyleBold))

	for _, p := range ports {
		isTCP := strings.HasPrefix(strings.ToLower(p.Proto), "tcp")
		isUDP := strings.HasPrefix(strings.ToLower(p.Proto), "udp")

		if isTCP && !tcp {
			continue
		}
		if isUDP && !udp {
			continue
		}

		stateUpper := strings.ToUpper(p.State)
		if listenOnly && stateUpper != "LISTEN" && stateUpper != "LISTENING" {
			continue
		}

		row := fmt.Sprintf("%s\t%s\t%s\t%s", p.Proto, p.Local, p.Remote, p.State)
		if pid {
			row += fmt.Sprintf("\t%s", p.PID)
		}
		fmt.Fprintln(w, row)
	}
	w.Flush()
}

func listPorts() ([]PortInfo, error) {
	return listPortsOS()
}
