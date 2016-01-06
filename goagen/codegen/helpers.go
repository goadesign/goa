package codegen

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/raphael/goa/design"
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
				args[i] = strings.Replace(a, p, "$(GOPATH)", -1)
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

var (
	majorRegex       = regexp.MustCompile(`([0-9]+)\.`)
	digitPrefixRegex = regexp.MustCompile(`^[0-9]`)
)

// VersionPackage computes a given version package name.
// v1 => v1, V1 => v1, 1 => v1, 1.0 => v1 if unique - v1dot0 otherwise.
func VersionPackage(version string) string {
	var others []string
	design.Design.IterateVersions(func(v *design.APIVersionDefinition) error {
		others = append(others, v.Version)
		return nil
	})
	idx := strings.Index(version, ".")
	if idx == 0 {
		// weird but OK
		version = strings.Replace(version, ".", "dot", -1)
	} else if idx > 0 {
		uniqueMajor := true
		match := majorRegex.FindStringSubmatch(version)
		if len(match) > 1 {
			major := match[1]
			for _, o := range others {
				match = majorRegex.FindStringSubmatch(o)
				if len(match) > 1 && major != match[1] {
					uniqueMajor = false
					break
				}
			}
		}
		if uniqueMajor {
			version = version[:idx]
		} else {
			strings.Replace(version, ".", "dot", -1)
		}
	}
	if digitPrefixRegex.MatchString(version) {
		version = "v" + version
	}
	return Goify(version, false)
}
