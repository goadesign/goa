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

// Unmarshaler produces the go code that initializes a user type from its JSON representation.
// This include running any validation defined on the type.
func Unmarshaler(ds DataStructure) string {
	def := ds.Definition()
	switch actual := def.Type.(type) {
	}
}

type primitiveCoerceData struct {
	// Raw contains the name of the variable containing the raw (interface{}) value.
	Raw string

	// Type is the name of the type to coerce to.
	Type string

	// Target is the name of the target variable
	Target string

	// Error is the name of the error factory method. The method must take 4 arguments:
	// the name of the payload field or param being coerced, its value, the target type and
	// the conversion error.
	Error string

	// Name is the name of the payload field or parameter being coerced.
	Name string
}

type arrayCoerceData struct {
	// Raw contains the name of the variable containing the raw (interface{}) value.
	Raw string

	// Type is the name of the type to coerce to.
	Type string

	// ElemConversion is the source code used to convert a single array element.
	ElemConversion string
}

const (
	primitiveCoerce = `	if val, ok := {{.Raw}}.({{.Type}}); ok {
		{{.Target}} = val
	} else {
		err = goa.{{.Error}}("{{.Name}}", {{.Raw}}, "{{.Type}}", fmt.Printf("incompatible type"))
	}`

	arrayCoerce = `	if val, ok := {{.Raw}}.([]interface{}), ok {
		{{.Target}} = make([]{{.ElemType}}, len(val))
		for i, v := range val {
			var e {{.ElemType}}
			{{.ElemConversion}}	
		}`

	objectCoerve = `	if val, ok := {{.Raw}}.(map[string]interface{}), ok {
`
)

// TypeUnmarshaler produces the go code that initializes a data type from its JSON representation.
func TypeUnmarshaler(t DataType, target string) string {
	switch actual := def.Type.(type) {
	case Primitive:
		return fmt.Sprintf(primitivePayloadCoerce, GoTypeName(t), target, t.Kind().Name())
	case *Array:
		elemType := GoTypeName(actual.ElemType.Type)
		return fmt.Sprintf(arrayPayloadCoerce, target, elemType, elemType, Unmarshaler(actual.ElemType, "e"))
		for n, att := range actual {

		}
	}
}

// GoTypeDef returns the Go code that defines a Go type which matches the data structure
// definition (the part that comes after `type foo`).
func GoTypeDef(ds DataStructure) string {
	var buffer bytes.Buffer
	def := ds.Definition()
	t := def.Type
	switch actual := t.(type) {
	case Primitive:
		return GoTypeName(t)
	case *Array:
		return "[]" + GoTypeDef(actual.ElemType)
	case Object:
		buffer.WriteString("struct {\n")
		keys := make([]string, len(actual))
		i := 0
		for n := range actual {
			keys[i] = n
			i++
		}
		sort.Strings(keys)
		for _, name := range keys {
			typedef := GoTypeDef(actual[name])
			fname := Goify(name, true)
			var omit string
			if !def.IsRequired(name) {
				omit = ",omitempty"
			}
			field := fmt.Sprintf("\t%s %s `json:\"%s%s\"`\n", fname, typedef, name, omit)
			buffer.WriteString(field)
		}
		buffer.WriteString("}")
		return buffer.String()
	case *UserTypeDefinition, *MediaTypeDefinition:
		return "*" + GoTypeName(actual)
	default:
		panic("goa bug: unknown data structure type")
	}
}

// GoTypeRef returns the Go code that refers to the Go type which matches the given data type
// (the part that comes after `var foo`)
func GoTypeRef(t DataType) string {
	switch t.(type) {
	case *UserTypeDefinition, *MediaTypeDefinition:
		return "*" + GoTypeName(t)
	default:
		return GoTypeName(t)
	}
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
		return "[]" + GoTypeRef(actual.ElemType.Type)
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
