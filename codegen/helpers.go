package codegen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rightscale/rsc/gen/writers/text"
)

// CommandLine return the command used to run this process.
func CommandLine() string {
	cmd := fmt.Sprintf("$ %s %s", filepath.Base(os.Args[0]), strings.Join(os.Args[1:], " "))
	return strings.Replace(cmd, " --", "\n  --", -1)
}

// Comment produces line comments by concatenating the given strings and producing 80 characters
// long lines starting with "//"
func Comment(elems ...string) string {
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

// Tabs returns a string made of depth tab characters.
func Tabs(depth int) string {
	var tabs string
	for i := 0; i < depth; i++ {
		tabs += "\t"
	}
	//	return fmt.Sprintf("%d%s", depth, tabs)
	return tabs
}
