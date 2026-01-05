package utils

import (
	"strings"
)

// SplitArgs splits a command line string into arguments, respecting quotes.
// It handles single quotes, double quotes, and escaped characters.
func SplitArgs(line string) []string {
	var args []string
	var currentArg strings.Builder
	inQuote := rune(0) // 0 for no quote, ' or " for respective quote type
	escaped := false

	for _, r := range line {
		if escaped {
			currentArg.WriteRune(r)
			escaped = false
			continue
		}

		switch r {
		case '\\':
			escaped = true
		case '\'', '"':
			if inQuote == 0 {
				inQuote = r
			} else if inQuote == r {
				inQuote = 0
			} else {
				// Different quote type encountered within a quote, treat as literal
				currentArg.WriteRune(r)
			}
		case ' ', '\t':
			if inQuote == 0 {
				if currentArg.Len() > 0 {
					args = append(args, currentArg.String())
					currentArg.Reset()
				}
			} else {
				currentArg.WriteRune(r)
			}
		default:
			currentArg.WriteRune(r)
		}
	}

	if currentArg.Len() > 0 {
		args = append(args, currentArg.String())
	}

	return args
}
