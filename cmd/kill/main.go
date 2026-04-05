package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

var signalMap = map[string]syscall.Signal{
	"HUP":  syscall.SIGHUP,
	"INT":  syscall.SIGINT,
	"KILL": syscall.SIGKILL,
	"TERM": syscall.SIGTERM,
	"1":    syscall.SIGHUP,
	"2":    syscall.SIGINT,
	"9":    syscall.SIGKILL,
	"15":   syscall.SIGTERM,
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Fprintln(os.Stderr, `Usage: kill [-s signal] [-l] <pid>...

Send a signal to processes.

Options:
  -s <signal>   Signal to send (default: TERM)
  -l            List available signals
  -h, --help    Show this help

Signals: HUP(1), INT(2), KILL(9), TERM(15)

Examples:
  kill 1234           Send SIGTERM to PID 1234
  kill -s KILL 1234   Send SIGKILL to PID 1234
  kill -9 1234        Send SIGKILL to PID 1234`)
		os.Exit(0)
	}

	if args[0] == "-l" {
		fmt.Println("HUP(1)  INT(2)  KILL(9)  TERM(15)")
		os.Exit(0)
	}

	sig := syscall.SIGTERM
	var pids []int

	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "-s" && i+1 < len(args) {
			i++
			name := strings.ToUpper(args[i])
			if s, ok := signalMap[name]; ok {
				sig = s
			} else {
				fmt.Fprintf(os.Stderr, "kill: unknown signal: %s\n", args[i])
				os.Exit(1)
			}
		} else if strings.HasPrefix(a, "-") && len(a) > 1 {
			// Handle -9, -KILL style
			name := strings.ToUpper(a[1:])
			if s, ok := signalMap[name]; ok {
				sig = s
			} else {
				fmt.Fprintf(os.Stderr, "kill: unknown signal: %s\n", a)
				os.Exit(1)
			}
		} else {
			pid, err := strconv.Atoi(a)
			if err != nil {
				fmt.Fprintf(os.Stderr, "kill: invalid PID: %s\n", a)
				os.Exit(1)
			}
			pids = append(pids, pid)
		}
	}

	if len(pids) == 0 {
		fmt.Fprintln(os.Stderr, "kill: no PID specified")
		os.Exit(1)
	}

	exitCode := 0
	for _, pid := range pids {
		proc, err := os.FindProcess(pid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "kill: (%d) - no such process\n", pid)
			exitCode = 1
			continue
		}
		if err := proc.Signal(sig); err != nil {
			fmt.Fprintf(os.Stderr, "kill: (%d) - %v\n", pid, err)
			exitCode = 1
		}
	}
	os.Exit(exitCode)
}
