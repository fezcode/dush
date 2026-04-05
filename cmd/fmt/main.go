package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		help := "Usage: fmt <format> [args...]\n" +
			"\nPrintf-style formatting.\n" +
			"\nFormat specifiers:\n" +
			"  %%s    string\n" +
			"  %%d    integer\n" +
			"  %%f    float\n" +
			"  %%x    hex\n" +
			"  %%o    octal\n" +
			"  %%b    binary\n" +
			"  %%q    quoted string\n" +
			"  %%v    default format\n" +
			"  %%%%    literal percent\n" +
			"  \\n    newline\n" +
			"  \\t    tab\n" +
			"\nExamples:\n" +
			"  fmt \"%%s is %%d years old\\n\" Alice 30\n" +
			"  fmt \"hex: %%x\\n\" 255\n" +
			"  fmt \"pi is %%.2f\\n\" 3.14159\n"
		fmt.Fprint(os.Stderr, help)
		os.Exit(0)
	}

	format := args[0]
	// Process escape sequences in format string
	format = strings.ReplaceAll(format, `\n`, "\n")
	format = strings.ReplaceAll(format, `\t`, "\t")
	format = strings.ReplaceAll(format, `\r`, "\r")
	format = strings.ReplaceAll(format, `\\`, "\x00") // placeholder
	format = strings.ReplaceAll(format, "\x00", `\`)

	fmtArgs := make([]any, 0, len(args)-1)
	for _, a := range args[1:] {
		// Try integer
		if n, err := strconv.ParseInt(a, 10, 64); err == nil {
			fmtArgs = append(fmtArgs, n)
			continue
		}
		// Try float
		if f, err := strconv.ParseFloat(a, 64); err == nil {
			fmtArgs = append(fmtArgs, f)
			continue
		}
		// Default: string
		fmtArgs = append(fmtArgs, a)
	}

	fmt.Printf(format, fmtArgs...)
}
