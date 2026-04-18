package repl

import (
	"fmt"
	"strings"

	"dush/internal/utils"
)

// reverseSearch implements incremental backward history search (Ctrl+R).
//
// It takes over the terminal input for the duration of the search, reading
// bytes directly from the underlying reader and drawing a `(reverse-i-search)`
// prompt. On Enter, the matched line is submitted (a synthetic Enter is
// injected so the term package's ReadLine returns immediately). On Escape or
// Ctrl+G the original line is restored.
//
// Returns (newLine, newPos, ok) for the term package's AutoCompleteCallback.
// The caller (autoCompleteCb) redraws prompt+line; we only manage the search UI.
func (le *lineEditor) reverseSearch(originalLine string, originalPos int) (string, int, bool) {
	if le.io == nil || le.out == nil {
		return "", 0, false
	}

	history := utils.GetHistory()

	pattern := ""
	matchIdx := -1
	currentMatch := ""

	// findMatch scans history backwards from startIdx looking for the first
	// entry containing pat. Returns the index, or -1 if no match.
	findMatch := func(pat string, startIdx int) int {
		if pat == "" || startIdx < 0 {
			return -1
		}
		if startIdx >= len(history) {
			startIdx = len(history) - 1
		}
		for i := startIdx; i >= 0; i-- {
			if strings.Contains(history[i], pat) {
				return i
			}
		}
		return -1
	}

	// Save the starting cursor position: col 0 of the row Ctrl+R was pressed
	// on. Subsequent redraws restore to this point so wrapped matches don't
	// leave orphan content above the search prompt.
	fmt.Fprint(le.out, "\r\x1b[s\x1b[J")

	draw := func() {
		fmt.Fprint(le.out, "\x1b[u\x1b[J")
		fmt.Fprintf(le.out, "(reverse-i-search)`%s': %s", pattern, currentMatch)
	}

	// restorePromptLine repaints the prompt followed by lineContent starting
	// at the saved position. The term package's cursor tracking still points
	// to the end of originalLine; autoCompleteCb applies the wide-char
	// compensation based on the line it received, so we leave the real cursor
	// at the visual end of lineContent — that matches the term package's
	// tracking when lineContent == originalLine.
	restorePromptLine := func(lineContent string) {
		fmt.Fprint(le.out, "\x1b[u\x1b[J")
		fmt.Fprint(le.out, le.prompt)
		fmt.Fprint(le.out, lineContent)
	}

	// Commit: restore the display, inject Enter so the line auto-executes,
	// and return the match (or the original line if no match was found).
	commit := func(leftover []byte) (string, int, bool) {
		target, targetPos := currentMatch, len(currentMatch)
		if target == "" {
			target, targetPos = originalLine, originalPos
		}
		restorePromptLine(originalLine)
		if len(leftover) > 0 {
			le.io.Inject(leftover)
		}
		le.io.Inject([]byte{'\r'})
		return target, targetPos, true
	}

	cancel := func() (string, int, bool) {
		restorePromptLine(originalLine)
		return originalLine, originalPos, true
	}

	draw()

	buf := make([]byte, 32)
	for {
		n, err := le.io.Read(buf)
		if err != nil || n == 0 {
			return cancel()
		}

		for i := 0; i < n; i++ {
			b := buf[i]
			switch {
			case b == '\r' || b == '\n':
				return commit(buf[i+1 : n])

			case b == 0x1b: // Escape (may be a standalone ESC or the start of a CSI sequence)
				// Peek: if more bytes follow and start a CSI, treat as arrow etc. = cancel.
				// Either way, cancel and restore the original line.
				return cancel()

			case b == 0x03, b == 0x07: // Ctrl+C, Ctrl+G — cancel
				return cancel()

			case b == 0x12: // Ctrl+R — search older match with same pattern
				startFrom := len(history) - 1
				if matchIdx > 0 {
					startFrom = matchIdx - 1
				} else if matchIdx == 0 {
					// Already at oldest; stay put.
					break
				}
				if next := findMatch(pattern, startFrom); next >= 0 {
					matchIdx = next
					currentMatch = history[matchIdx]
				}

			case b == 0x7f, b == 0x08: // Backspace / Ctrl+H
				if len(pattern) > 0 {
					pattern = pattern[:len(pattern)-1]
					if pattern == "" {
						matchIdx = -1
						currentMatch = ""
					} else {
						matchIdx = findMatch(pattern, len(history)-1)
						if matchIdx >= 0 {
							currentMatch = history[matchIdx]
						} else {
							currentMatch = ""
						}
					}
				}

			case b >= 0x20 && b < 0x7f: // Printable ASCII — extend the pattern
				pattern += string(rune(b))
				matchIdx = findMatch(pattern, len(history)-1)
				if matchIdx >= 0 {
					currentMatch = history[matchIdx]
				} else {
					currentMatch = ""
				}

			default:
				// Ignore other control bytes for now.
			}
			draw()
		}
	}
}
