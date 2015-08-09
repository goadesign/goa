package design

import (
	"bytes"
	"fmt"
	"sort"
	"unicode"

	"bitbucket.org/pkg/inflect"
)

// ContextName computes the name of the context data structure that corresponds to this action.
func (a *ActionDefinition) ContextName() string {
	return inflect.Camelize(a.Name) + inflect.Camelize(a.Resource.Name) + "Context"
}

// SourceCode returns the Go code that defines a Go type which matches the data structure
// definition.
func SourceCode(d DataStructure) string {
	var buffer bytes.Buffer
	buffer.WriteString("struct {\n")
	o := d.Obj()
	keys := make([]string, len(o))
	i := 0
	for n := range o {
		keys[i] = n
		i++
	}
	sort.Strings(keys)
	for _, name := range keys {
		typedef := TypeDef(o[name].Type)
		fname := Goify(name, true)
		field := fmt.Sprintf("\t%s %s `json:\"%s\"`\n", fname, typedef, name)
		buffer.WriteString(field)
	}
	buffer.WriteString("}")
	return buffer.String()
}

// TypeDef returns the Go type definition (the part that comes after 'type foo') for the given data
// type.
func TypeDef(t DataType) string {
	var typedef string
	switch actual := t.(type) {
	case *UserTypeDefinition, *MediaTypeDefinition:
		typedef = "*" + GoTypeName(actual)
	case Object:
		typedef = SourceCode(actual)
	case *Array:
		typedef = "[]" + TypeDef(actual.ElemType.Type)
	case Primitive:
		typedef = GoTypeName(actual)
	}
	return typedef
}

// GoTypeName returns the Go type name for a data type.
func GoTypeName(t DataType) string {
	switch actual := t.(type) {
	case Primitive:
		switch actual.Kind() {
		case BooleanKind:
			return "bool"
		case IntegerKind:
			return "int"
		case NumberKind:
			return "float64"
		case StringKind:
			return "string"
		default:
			panic(fmt.Sprintf("goa bug: unknown primitive type %#v", actual))
		}
	case *Array:
		return "[]" + GoTypeName(actual.ElemType.Type)
	case Object:
		return "map[string]interface{}"
	case *UserTypeDefinition:
		return Goify(actual.Name, true)
	case *MediaTypeDefinition:
		return Goify(actual.Name, true)
	default:
		panic(fmt.Sprintf("goa bug: unknown type %#v", actual))
	}
}

// Goify makes a valid go identifier out of any string.
// It does that by removing any non letter and non digit character and by making sure the first
// character is a letter or "_".
// Goify produces a "CamelCase" version of the string, if firstUpper is true the first character
// of the identifier is uppercase otherwise it's lowercase.
func Goify(str string, firstUpper bool) string {
	var b bytes.Buffer
	var firstWritten, nextUpper bool
	for i := 0; i < len(str); i++ {
		r := rune(str[i])
		if r == '_' {
			nextUpper = true
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if !firstWritten {
				if firstUpper {
					r = unicode.ToUpper(r)
				} else {
					r = unicode.ToLower(r)
				}
				firstWritten = true
				nextUpper = false
			} else if nextUpper {
				r = unicode.ToUpper(r)
				nextUpper = false
			}
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "_v" // you have a better idea?
	}
	res := b.String()
	if _, ok := reserved[res]; ok {
		res += "_"
	}
	return res
}

// reserved golang keywords
var reserved = map[string]bool{
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
}
