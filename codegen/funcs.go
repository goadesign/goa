package codegen

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

	"goa.design/goa.v2/design"
	"goa.design/goa.v2/pkg"
)

var (
	// common words who need to keep their
	commonInitialisms = map[string]bool{
		"API":   true,
		"ASCII": true,
		"CPU":   true,
		"CSS":   true,
		"DNS":   true,
		"EOF":   true,
		"GUID":  true,
		"HTML":  true,
		"HTTP":  true,
		"HTTPS": true,
		"ID":    true,
		"IP":    true,
		"JMES":  true,
		"JSON":  true,
		"JWT":   true,
		"LHS":   true,
		"OK":    true,
		"QPS":   true,
		"RAM":   true,
		"RHS":   true,
		"RPC":   true,
		"SLA":   true,
		"SMTP":  true,
		"SQL":   true,
		"SSH":   true,
		"TCP":   true,
		"TLS":   true,
		"TTL":   true,
		"UDP":   true,
		"UI":    true,
		"UID":   true,
		"UUID":  true,
		"URI":   true,
		"URL":   true,
		"UTF8":  true,
		"VM":    true,
		"XML":   true,
		"XSRF":  true,
		"XSS":   true,
	}

	// reserved golang keywords and package names
	reserved = map[string]bool{
		"byte":       true,
		"complex128": true,
		"complex64":  true,
		"float32":    true,
		"float64":    true,
		"int":        true,
		"int16":      true,
		"int32":      true,
		"int64":      true,
		"int8":       true,
		"rune":       true,
		"string":     true,
		"uint16":     true,
		"uint32":     true,
		"uint64":     true,
		"uint8":      true,

		// reserved keywords
		"break":       true,
		"case":        true,
		"chan":        true,
		"const":       true,
		"continue":    true,
		"default":     true,
		"defer":       true,
		"else":        true,
		"fallthrough": true,
		"for":         true,
		"func":        true,
		"go":          true,
		"goto":        true,
		"if":          true,
		"import":      true,
		"interface":   true,
		"map":         true,
		"package":     true,
		"range":       true,
		"return":      true,
		"select":      true,
		"struct":      true,
		"switch":      true,
		"type":        true,
		"var":         true,

		// stdlib and goa packages used by generated code
		"fmt":  true,
		"http": true,
		"json": true,
		"os":   true,
		"url":  true,
		"time": true,
	}
)

// TemplateFuncs returns all the common helper functions.
func TemplateFuncs() template.FuncMap {
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
		return fmt.Errorf("version mismatch: using goagen %s to generate code that compiles with goa %s",
			ver, pkg.Version())
	}
	return nil
}

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
		param = " " + strings.Join(args, " ")
	}
	rawcmd := filepath.Base(os.Args[0])
	// Remove possible .exe suffix to not create different ouptut just because
	// you ran goagen on Windows.
	rawcmd = strings.TrimSuffix(rawcmd, ".exe")

	cmd := fmt.Sprintf("$ %s%s", rawcmd, param)
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
	return strings.Replace(name, "_", "-", -1)
}

// GoTypeRef returns the Go code that refers to the Go type which matches the given data type
func GoTypeRef(dt design.DataType) string {
	tname := GoTypeName(dt)
	if design.IsObject(dt) {
		return "*" + tname
	}
	return tname
}

// GoTypeName returns the Go type name for a data type.
// todo: TBD add support for maps, objects and usertypes
func GoTypeName(dt design.DataType) string {
	switch actual := dt.(type) {
	case design.Primitive:
		return GoNativeType(dt)
	case *design.Array:
		return "[]" + GoTypeRef(actual.ElemType.Type)
	default:
		panic(fmt.Sprintf("goa bug: unknown type %#v", actual))
	}
}

// GoNativeType returns the Go built-in type from which instances of provided datatype can be initialized.
// todo: TBD add support for maps, objects and usertypes
func GoNativeType(t design.DataType) string {
	switch actual := t.(type) {
	case design.Primitive:
		switch actual.Kind() {
		case design.BooleanKind:
			return "bool"
		case design.Int32Kind:
			return "int32"
		case design.Int64Kind:
			return "int64"
		case design.UInt32Kind:
			return "uint32"
		case design.UInt64Kind:
			return "uint64"
		case design.Float32Kind:
			return "float32"
		case design.Float64Kind:
			return "float64"
		case design.StringKind:
			return "string"
		case design.AnyKind:
			return "interface{}"
		default:
			panic(fmt.Sprintf("goa bug: unknown primitive type %#v", actual))
		}
	case *design.Array:
		return "[]" + GoNativeType(actual.ElemType.Type)
	case design.CompositeExpr:
		return GoNativeType(actual.Attribute().Type)
	default:
		panic(fmt.Sprintf("goa bug: unknown type %#v", actual))
	}
}

// Goify makes a valid Go identifier out of any string.
// It does that by removing any non letter and non digit character and by making sure the first
// character is a letter or "_".
// Goify produces a "CamelCase" version of the string, if firstUpper is true the first character
// of the identifier is uppercase otherwise it's lowercase.
func Goify(str string, firstUpper bool) string {
	runes := []rune(str)

	// remove trailing invalid identifiers (makes code below simpler)
	runes = removeTrailingInvalid(runes)

	w, i := 0, 0 // index of start of word, scan
	for i+1 <= len(runes) {
		eow := false // whether we hit the end of a word

		// remove leading invalid identifiers
		runes = removeInvalidAtIndex(i, runes)

		if i+1 == len(runes) {
			eow = true
		} else if !validIdentifier(runes[i]) {
			// get rid of it
			runes = append(runes[:i], runes[i+1:]...)
		} else if runes[i+1] == '_' {
			// underscore; shift the remainder forward over any run of underscores
			eow = true
			n := 1
			for i+n+1 < len(runes) && runes[i+n+1] == '_' {
				n++
			}
			copy(runes[i+1:], runes[i+n+1:])
			runes = runes[:len(runes)-n]
		} else if unicode.IsLower(runes[i]) && !unicode.IsLower(runes[i+1]) {
			// lower->non-lower
			eow = true
		}
		i++
		if !eow {
			continue
		}

		// [w,i] is a word.
		word := string(runes[w:i])
		// is it one of our initialisms?
		if u := strings.ToUpper(word); commonInitialisms[u] {
			if firstUpper {
				u = strings.ToUpper(u)
			} else if w == 0 {
				u = strings.ToLower(u)
			}

			// All the common initialisms are ASCII,
			// so we can replace the bytes exactly.
			copy(runes[w:], []rune(u))
		} else if w > 0 && strings.ToLower(word) == word {
			// already all lowercase, and not the first word, so uppercase the first character.
			runes[w] = unicode.ToUpper(runes[w])
		} else if w == 0 && strings.ToLower(word) == word && firstUpper {
			runes[w] = unicode.ToUpper(runes[w])
		}
		if w == 0 && !firstUpper {
			runes[w] = unicode.ToLower(runes[w])
		}
		//advance to next word
		w = i
	}

	return fixReserved(string(runes))
}

// validIdentifier returns true if the rune is a letter or number
func validIdentifier(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

// fixReserved appends an underscore on to Go reserved keywords.
func fixReserved(w string) string {
	if reserved[w] {
		w += "_"
	}
	return w
}

// removeTrailingInvalid removes trailing invalid identifiers from runes.
func removeTrailingInvalid(runes []rune) []rune {
	valid := len(runes) - 1
	for ; valid >= 0 && !validIdentifier(runes[valid]); valid-- {
	}

	return runes[0 : valid+1]
}

// removeInvalidAtIndex removes consecutive invalid identifiers from runes starting at index i.
func removeInvalidAtIndex(i int, runes []rune) []rune {
	valid := i
	for ; valid < len(runes) && !validIdentifier(runes[valid]); valid++ {
	}

	return append(runes[:i], runes[valid:]...)
}
