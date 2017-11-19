package codegen

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"unicode"

	"goa.design/goa/pkg"
)

// TemplateFuncs lists common template helper functions.
func TemplateFuncs() map[string]interface{} {
	return map[string]interface{}{
		"commandLine": CommandLine,
		"comment":     Comment,
	}
}

// CheckVersion returns an error if the ver is empty, contains an incorrect value or
// a version number that is not compatible with the version of this repo.
func CheckVersion(ver string) error {
	compat, err := pkg.Compatible(ver)
	if err != nil {
		return err
	}
	if !compat {
		return fmt.Errorf("version mismatch: using goa %s to generate code that compiles with goa %s",
			ver, pkg.Version())
	}
	return nil
}

// CommandLine return the command used to run this process.
func CommandLine() string {
	cmdl := "$ goa"
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "--cmd=") {
			cmdl = arg[6:]
			break
		}
	}
	return cmdl
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

	return Indent(WrapText(t, 77), "// ")
}

// Indent inserts prefix at the beginning of each non-empty line of s. The
// end-of-line marker is NL.
func Indent(s, prefix string) string {
	var (
		res []byte
		b   = []byte(s)
		p   = []byte(prefix)
		bol = true
	)
	for _, c := range b {
		if bol && c != '\n' {
			res = append(res, p...)
		}
		res = append(res, c)
		bol = c == '\n'
	}
	return string(res)
}

// Add adds two integers and returns the sum of the two.
func Add(a, b int) int { return a + b }

// Casing exceptions
var toLower = map[string]string{"OAuth": "oauth"}

// SnakeCase produces the snake_case version of the given CamelCase string.
func SnakeCase(name string) string {
	for u, l := range toLower {
		name = strings.Replace(name, u, l, -1)
	}
	var b bytes.Buffer
	var lastUnderscore bool
	ln := len(name)
	if ln == 0 {
		return ""
	}
	b.WriteRune(unicode.ToLower(rune(name[0])))
	for i := 1; i < ln; i++ {
		r := rune(name[i])
		nextIsLower := false
		if i < ln-1 {
			n := rune(name[i+1])
			nextIsLower = unicode.IsLower(n) && unicode.IsLetter(n)
		}
		if unicode.IsUpper(r) {
			if !lastUnderscore && nextIsLower {
				b.WriteRune('_')
				lastUnderscore = true
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
			lastUnderscore = false
		}
	}
	return b.String()
}

// KebabCase produces the kebab-case version of the given CamelCase string.
func KebabCase(name string) string {
	name = SnakeCase(name)
	ln := len(name)
	if name[ln-1] == '_' {
		name = name[:ln-1]
	}
	return strings.Replace(name, "_", "-", -1)
}

// WrapText produces lines with text capped at maxChars
// it will keep words intact and respects newlines.
func WrapText(text string, maxChars int) string {
	res := ""
	lines := strings.Split(text, "\n")
	for _, v := range lines {
		runes := []rune(strings.TrimSpace(v))
		for l := len(runes); l >= 0; l = len(runes) {
			if maxChars >= l {
				res = res + string(runes) + "\n"
				break
			}

			i := runeSpacePosRev(runes[:maxChars])
			if i == 0 {
				i = runeSpacePos(runes)
			}

			res = res + string(runes[:i]) + "\n"
			if l == i {
				break
			}
			runes = runes[i+1:]
		}
	}
	return res[:len(res)-1]
}

func runeSpacePosRev(r []rune) int {
	for i := len(r) - 1; i > 0; i-- {
		if unicode.IsSpace(r[i]) {
			return i
		}
	}
	return 0
}

func runeSpacePos(r []rune) int {
	for i := 0; i < len(r); i++ {
		if unicode.IsSpace(r[i]) {
			return i
		}
	}
	return len(r)
}
