package code

import (
	"fmt"
	"os"
	"strings"

	"github.com/rightscale/rsc/gen/writers/text"
)

// commandLine return the command used to run this process.
func commandLine() string {
	return fmt.Sprintf("$ %s %s", os.Args[0], strings.Join(os.Args[1:], " "))
}

// comment produces line comments by concatenating the given strings and producing 80 characters
// long lines starting with "//"
func comment(elems ...string) string {
	var lines []string
	for _, e := range elems {
		lines = append(lines, strings.Split(e, "\n")...)
	}
	var trimmed = make([]string, len(lines))
	for i, l := range lines {
		trimmed[i] = strings.TrimLeft(l, " \t")
	}
	t := strings.Join(trimmed, "\n")
	return text.Indent(t, "// ")
}
