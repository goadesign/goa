package codegen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goadesign/goa/design"
)

// CommandLine return the command used to run this process.
func CommandLine() string {
	// We don't use the full path to the tool so that running goagen multiple times doesn't
	// end up creating different command line comments (because of the temporary directory it
	// runs in).
	var param string
	if len(os.Args) > 1 {
		args := make([]string, len(os.Args)-1)
		gopaths := filepath.SplitList(os.Getenv("GOPATH"))
		for i, a := range os.Args[1:] {
			for _, p := range gopaths {
				if strings.Contains(a, p) {
					args[i] = strings.Replace(a, p, "$(GOPATH)", -1)
					break
				}
			}
			if args[i] == "" {
				args[i] = a
			}
		}
		param = strings.Join(args, " ")
	}
	cmd := fmt.Sprintf("$ %s %s", filepath.Base(os.Args[0]), param)
	return strings.Replace(cmd, " --", "\n\t--", -1)
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

	return Indent(t, "// ")
}

// Indent inserts prefix at the beginning of each non-empty line of s. The
// end-of-line marker is NL.
func Indent(s, prefix string) string {
	return string(IndentBytes([]byte(s), []byte(prefix)))
}

// IndentBytes inserts prefix at the beginning of each non-empty line of b.
// The end-of-line marker is NL.
func IndentBytes(b, prefix []byte) []byte {
	var res []byte
	bol := true
	for _, c := range b {
		if bol && c != '\n' {
			res = append(res, prefix...)
		}
		res = append(res, c)
		bol = c == '\n'
	}
	return res
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

// Add adds two integers and returns the sum of the two.
func Add(a, b int) int { return a + b }

// CanonicalTemplate returns the resource URI template as a format string suitable for use in the
// fmt.Printf function family.
func CanonicalTemplate(r *design.ResourceDefinition) string {
	return design.WildcardRegex.ReplaceAllLiteralString(r.URITemplate(), "/%v")
}

// CanonicalParams returns the list of parameter names needed to build the canonical href to the
// resource. It returns nil if the resource does not have a canonical action.
func CanonicalParams(r *design.ResourceDefinition) []string {
	var params []string
	if ca := r.CanonicalAction(); ca != nil {
		if len(ca.Routes) > 0 {
			params = ca.Routes[0].Params()
		}
		for i, p := range params {
			params[i] = Goify(p, false)
		}
	}
	return params
}
