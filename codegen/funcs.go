// This file contains helpers shared by Goa's code generators. The helpers
// receive names and types chosen while a generated file is planned and return
// the Go source inserted into that file.
package codegen

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"unicode"

	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/internal/codegenname"
)

type (
	// InitArgData contains the data needed to render code to initialize struct
	// fields with the given arguments.
	InitArgData struct {
		// Name is the argument name.
		Name string
		// Pointer if true indicates that the argument is a pointer.
		Pointer bool
		// Type is the argument type.
		Type expr.DataType
		// FieldName is the name of the field in the struct initialized by the
		// argument.
		FieldName string
		// FieldPointer if true indicates that the field in the struct is a
		// pointer.
		FieldPointer bool
		// FieldType is the type of the field in the struct.
		FieldType expr.DataType
		// FieldTypeRef is the field's final Go type name in the generated file.
		// Callers set it when converting an argument to a named primitive type.
		FieldTypeRef string
	}
)

// TemplateFuncs lists common template helper functions.
func TemplateFuncs() map[string]any {
	return map[string]any{
		"commandLine": CommandLine,
		"comment":     Comment,
	}
}

// CommandLine returns the command used to run this process.
func CommandLine() string {
	cmdl := "goa"
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "--cmd=") {
			cmdl = arg[6:]
			break
		}
	}
	return cmdl
}

// Comment produces line comments by concatenating the given strings and
// producing 80 characters long lines starting with "//".
func Comment(elems ...string) string {
	lineCount := 0
	for _, e := range elems {
		lineCount += strings.Count(e, "\n") + 1
	}
	lines := make([]string, 0, lineCount)
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
		b   = []byte(s)
		p   = []byte(prefix)
		res = make([]byte, 0, len(b)+len(b)/4*len(p)) // preallocate with estimated size
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

// Casing exceptions
var toLower = map[string]string{"OAuth": "oauth"}

// CamelCase produces the CamelCase version of the given string. It removes any
// non letter and non digit character.
//
// If firstUpper is true the first letter of the string is capitalized else
// the first letter is in lowercase.
//
// If acronym is true and a part of the string is a common acronym
// then it keeps the part capitalized (firstUpper = true)
// (e.g. APIVersion) or lowercase (firstUpper = false) (e.g. apiVersion).
func CamelCase(name string, firstUpper, acronym bool) string {
	if name == "" {
		return ""
	}

	// Use cache to avoid recomputing the same transformation
	key := cacheKey{
		input:      name,
		firstUpper: firstUpper,
		acronym:    acronym,
		operation:  "camel",
	}
	return globalStringCache.getCached(key, func() string {
		return codegenname.CamelCase(name, firstUpper, acronym)
	})
}

// SnakeCase produces the snake_case version of the given CamelCase string.
// News    => news
// OldNews => old_news
// CNNNews => cnn_news
func SnakeCase(name string) string {
	// Special handling for single "words" starting with multiple upper case letters
	for u, l := range toLower {
		name = strings.ReplaceAll(name, u, l)
	}

	// Remove leading and trailing blank spaces and replace any blank spaces in
	// between with a single underscore
	name = strings.Join(strings.Fields(name), "_")

	// Special handling for dashes and slashes to convert them into underscores
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, "/", "_")

	var b bytes.Buffer
	ln := len(name)
	if ln == 0 {
		return ""
	}
	n := rune(name[0])
	b.WriteRune(unicode.ToLower(n))
	var lastLower, isLower, lastUnder, isUnder bool
	for i := 1; i < ln; i++ {
		r := rune(name[i])
		isLower = unicode.IsLower(r) && unicode.IsLetter(r) || unicode.IsDigit(r)
		isUnder = r == '_'
		if !isLower && !isUnder {
			if lastLower && !lastUnder {
				b.WriteRune('_')
			} else if ln > i+1 {
				rn := rune(name[i+1])
				if unicode.IsLower(rn) && rn != '_' && !lastUnder {
					b.WriteRune('_')
				}
			}
		}
		b.WriteRune(unicode.ToLower(r))
		lastLower = isLower
		lastUnder = isUnder
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
	return strings.ReplaceAll(name, "_", "-")
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

// InitStructFields produces Go code to initialize a struct and its fields from
// the given init arguments.
func InitStructFields(args []*InitArgData, targetVar, sourcePkg, targetPkg string) (string, []*TransformFunctionData, error) {
	scope := NewNameScope()
	scope.Unique(targetVar)

	var (
		code    string
		helpers []*TransformFunctionData
	)
	for _, arg := range args {
		switch {
		case arg.FieldName == "" && arg.FieldType == nil:
		// do nothing
		case expr.Equal(unalias(arg.Type), arg.FieldType):
			// arg type and struct field type are the same. No need to call transform
			// to initialize the field
			deref := ""
			if !arg.Pointer && arg.FieldPointer && expr.IsPrimitive(arg.FieldType) {
				deref = "&"
			}
			code += fmt.Sprintf("%s.%s = %s%s\n", targetVar, arg.FieldName, deref, arg.Name)
		case expr.IsPrimitive(arg.FieldType):
			// aliased primitive type
			if arg.FieldTypeRef == "" {
				return "", helpers, fmt.Errorf("initialize field %q: missing final field type reference", arg.FieldName)
			}
			t := arg.FieldTypeRef
			cast := fmt.Sprintf("%s(%s)", t, arg.Name)
			if arg.Pointer {
				code += "if " + arg.Name + " != nil {\n"
				cast = fmt.Sprintf("%s(*%s)", t, arg.Name)
			}
			switch {
			case arg.FieldPointer:
				code += fmt.Sprintf("tmp%s := %s\n%s.%s = &tmp%s\n", arg.Name, cast, targetVar, arg.FieldName, arg.Name)
			case arg.FieldName != "":
				code += fmt.Sprintf("%s.%s = %s\n", targetVar, arg.FieldName, cast)
			default:
				code += fmt.Sprintf("%s := %s\n", targetVar, cast)
			}
			if arg.Pointer {
				code += "}\n"
			}
		default:
			srcctx := NewAttributeContext(arg.Pointer, false, true, sourcePkg, scope)
			tgtctx := NewAttributeContext(arg.FieldPointer, false, true, targetPkg, scope)
			c, h, err := GoTransform(
				&expr.AttributeExpr{Type: arg.Type}, &expr.AttributeExpr{Type: arg.FieldType},
				arg.Name, fmt.Sprintf("%s.%s", targetVar, arg.FieldName), srcctx, tgtctx, "", false)
			if err != nil {
				return "", helpers, err
			}
			code += c + "\n"
			helpers = AppendHelpers(helpers, h)
		}
	}
	return code, helpers, nil
}

// Get the underlying primitive type of a aliased type or return the type itself
// if not aliased.
func unalias(dt expr.DataType) expr.DataType {
	if ut, ok := dt.(expr.UserType); ok {
		if _, ok := ut.Attribute().Type.(expr.Primitive); ok {
			return ut.Attribute().Type
		}
		return unalias(ut.Attribute().Type)
	}
	return dt
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
	for i := range r {
		if unicode.IsSpace(r[i]) {
			return i
		}
	}
	return len(r)
}
