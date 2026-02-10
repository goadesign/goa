package expr_test

import "strings"

// stripValidationLocations removes leading "[file:line] " prefixes from error
// messages.
//
// Validation errors now include DSL declaration locations when available; most
// test cases assert error semantics and should remain stable when line numbers
// change due to unrelated edits.
func stripValidationLocations(message string) string {
	lines := strings.Split(message, "\n")
	for i, line := range lines {
		if !strings.HasPrefix(line, "[") {
			continue
		}
		end := strings.Index(line, "] ")
		if end == -1 {
			continue
		}
		lines[i] = line[end+2:]
	}
	return strings.Join(lines, "\n")
}
