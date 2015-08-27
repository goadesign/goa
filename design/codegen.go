package design

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"
	"unicode"
)

var (
	primitiveT *template.Template
	arrayT     *template.Template
	objectT    *template.Template
	userT      *template.Template
)

//  init instantiates the templates.
func init() {
	var err error
	fm := template.FuncMap{
		"unmarshalType":      typeUnmarshalerR,
		"unmarshalAttribute": attributeUnmarshalerR,
		"gotypename":         GoTypeName,
		"gotyperef":          GoTypeRef,
		"tabs":               tabs,
		"tabulate":           tabulate,
		"add":                func(a, b int) int { return a + b },
	}
	if primitiveT, err = template.New("primitive").Funcs(fm).Parse(primitiveTmpl); err != nil {
		panic(err)
	}
	if arrayT, err = template.New("array").Funcs(fm).Parse(arrayTmpl); err != nil {
		panic(err)
	}
	if objectT, err = template.New("object").Funcs(fm).Parse(objectTmpl); err != nil {
		panic(err)
	}
	if userT, err = template.New("user").Funcs(fm).Parse(userTmpl); err != nil {
		panic(err)
	}
}

// TypeUnmarshaler produces the go code that initializes a variable of the given type given
// a deserialized (interface{}) value.
// source is the name of the variable that contains the raw interface{} value and target the
// name of the variable to initialize.
// context is used to keep track of recursion to produce helpful error messages in case of type
// mismatch or validation error.
// The generated code assumes that there is a variable called "err" of type error that it can use
// to record errors.
func TypeUnmarshaler(t DataType, context, source, target string) string {
	return typeUnmarshalerR(t, context, source, target, 1)
}
func typeUnmarshalerR(t DataType, context, source, target string, depth int) string {
	switch actual := t.(type) {
	case Primitive:
		return primitiveUnmarshalerR(actual, context, source, target, depth)
	case *Array:
		return arrayUnmarshalerR(actual, context, source, target, depth)
	case Object:
		return objectUnmarshalerR(actual, context, source, target, depth)
	case NamedType:
		return namedTypeUnmarshalerR(actual, context, source, target, depth)
	default:
		panic("unknown type")
	}
}

// AttributeUnmarshaler produces the go code that initializes an attribute given a deserialized
// (interface{}) value.
// source is the name of the variable that contains the raw interface{} value and target the
// name of the variable to initialize.
// context is used to keep track of recursion to produce helpful error messages in case of type
// mismatch or validation error.
// The generated code assumes that there is a variable called "err" of type error that it can use
// to record errors.
func AttributeUnmarshaler(att *AttributeDefinition, context, source, target string) string {
	return attributeUnmarshalerR(att, context, source, target, 1)
}
func attributeUnmarshalerR(att *AttributeDefinition, context, source, target string, depth int) string {
	return typeUnmarshalerR(att.Type, context, source, target, depth) +
		validationCheckerR(att, context, target, depth)
}

// PrimitiveUnmarshaler produces the go code that initializes a primitive type from its JSON
// representation.
// source is the name of the variable that contains the raw interface{} value and target the
// name of the variable to initialize.
// The generated code assumes that there is a variable called "err" of type error that it can use
// to record errors.
func PrimitiveUnmarshaler(p Primitive, context, source, target string) string {
	return primitiveUnmarshalerR(p, context, source, target, 1)
}
func primitiveUnmarshalerR(p Primitive, context, source, target string, depth int) string {
	data := map[string]interface{}{
		"source":  source,
		"target":  target,
		"type":    p,
		"context": context,
		"depth":   depth,
	}
	var b bytes.Buffer
	err := primitiveT.Execute(&b, data)
	if err != nil {
		panic(err) // should never happen
	}
	return b.String()
}

// ArrayUnmarshaler produces the go code that initializes an array from its JSON representation.
// source is the name of the variable that contains the raw interface{} value and target the
// name of the variable to initialize.
// The generated code assumes that there is a variable called "err" of type error that it can use
// to record errors.
func ArrayUnmarshaler(a *Array, context, source, target string) string {
	return arrayUnmarshalerR(a, context, source, target, 1)
}
func arrayUnmarshalerR(a *Array, context, source, target string, depth int) string {
	data := map[string]interface{}{
		"source":   source,
		"target":   target,
		"elemType": a.ElemType,
		"context":  context,
		"depth":    depth,
	}
	var b bytes.Buffer
	err := arrayT.Execute(&b, data)
	if err != nil {
		panic(err) // should never happen
	}
	return b.String()
}

// ObjectUnmarshaler produces the go code that initializes an object type from its JSON
// representation.
// source is the name of the variable that contains the raw interface{} value and target the
// name of the variable to initialize.
// The generated code assumes that there is a variable called "err" of type error that it can use
// to record errors.
func ObjectUnmarshaler(o Object, context, source, target string) string {
	return objectUnmarshalerR(o, context, source, target, 1)
}
func objectUnmarshalerR(o Object, context, source, target string, depth int) string {
	data := map[string]interface{}{
		"source":  source,
		"target":  target,
		"type":    o,
		"context": context,
		"depth":   depth,
	}
	var b bytes.Buffer
	err := objectT.Execute(&b, data)
	if err != nil {
		panic(err)
	}
	return b.String()
}

// NamedTypeUnmarshaler produces the go code that initializes a named type from its JSON
// representation.
// This include running any validation defined on the type.
// source is the name of the variable that contains the raw interface{} value and target the
// name of the variable to initialize.
// context is used to keep track of recursion to produce helpful error messages in case of type
// mismatch or validation error.
// The generated code assumes that there is a variable called "err" of type error that it can use
// to record errors.
func NamedTypeUnmarshaler(t NamedType, context, source, target string) string {
	return namedTypeUnmarshalerR(t, context, source, target, 1)
}
func namedTypeUnmarshalerR(t NamedType, context, source, target string, depth int) string {
	data := map[string]interface{}{
		"source":  source,
		"target":  target,
		"context": context,
		"depth":   depth,
	}
	var b bytes.Buffer
	err := userT.Execute(&b, data)
	if err != nil {
		panic(err) // should never happen
	}
	return b.String()
}

// ValidationChecker produces Go code that runs the validation defined in the given attribute
// definition against the content of the variable named target recursively.
// context is used to keep track of recursion to produce helpful error messages in case of type
// validation error.
func ValidationChecker(att *AttributeDefinition, context, target string) string {
	return validationCheckerR(att, context, target, 1)
}
func validationCheckerR(att *AttributeDefinition, context, target string, depth int) string {
	var b bytes.Buffer
	//for _, v := range att.Validations {
	//switch actual := v.(type) {
	//case *EnumValidationDefinition:
	//case *FormatValidationDefinition:
	//case *MinimumValidationDefinition:
	//case *MaximumValidationDefinition:
	//case *MinLengthValidationDefinition:
	//case *MaxLengthValidationDefinition:
	//case *RequiredValidationDefinition:
	//}
	//}
	return b.String()
}

const (
	primitiveTmpl = `{{tabs .depth}}if val, ok := {{.source}}.({{gotyperef .type}}); ok {
{{tabs .depth}}	{{.target}} = val
{{tabs .depth}}} else {
{{tabs .depth}}	err = goa.IncompatibleTypeError(` + "`" + `{{.context}}` + "`" + `, {{.source}}, "{{gotypename .type}}")
{{tabs .depth}}}`

	arrayTmpl = `{{tabs .depth}}if val, ok := {{.source}}.([]interface{}); ok {
{{tabs .depth}}	{{.target}} = make([]{{gotyperef .elemType.Type}}, len(val))
{{tabs .depth}}	for i, v := range val {
{{unmarshalAttribute .elemType (printf "%s[*]" .context) "v" (printf "%s[i]" .target) (add .depth 2)}}
{{tabs .depth}}	}
{{tabs .depth}}} else {
{{tabs .depth}}	err = goa.IncompatibleTypeError(` + "`" + `{{.context}}` + "`" + `, {{.source}}, "[]{{gotyperef .elemType.Type}}")
{{tabs .depth}}}`

	objectTmpl = `{{tabs .depth}}if val, ok := {{.source}}.(map[string]interface{}); ok {
{{tabs .depth}}{{$context := .context}}{{$depth := .depth}}{{$target := .target}}	{{$target}} = make(map[string]interface{})
{{range $name, $att := .type}}{{tabs $depth}}	if v, ok := val["{{$name}}"]; ok {
{{unmarshalType $att.Type (printf "%s[\"%s\"]" $context $name) "v" (printf "%s[\"%s\"]" $target $name) (add $depth 2)}}
{{tabs $depth}}	}
{{end}}{{tabs $depth}}} else {
{{tabs .depth}}	err = goa.IncompatibleTypeError(` + "`" + `{{.context}}` + "`" + `, {{.source}}, ` + "`{{tabulate (gotypename .type) (add .depth 2)}}`)" + `
{{tabs .depth}}}`

	userTmpl = `{{tabs .depth}}if val, ok := {{.source}}.(map[string]interface{}); ok {
{{tabs .depth}}	{{.target}} = new({{.type.Name}})
{{range $name, $att := .type.Definition.Type}}{{tabs .depth}}	if v, ok := val["{{$name}}"]; ok {
{{unmarshalType $att.Type (printf "%s.%s" .context $name) "v" (printf "%s.%s" .target $name) (add .depth 2)}}
{{tabs .depth}}	}
{{end}}{{tabs .depth}}} else {
{{tabs .depth}}	err = goa.IncompatibleTypeError(` + "`" + `{{.context}}` + "`" + `, {{.source}}, {{gotypename .type}})
{{tabs .depth}}}`
)

// GoTypeDef returns the Go code that defines a Go type which matches the data structure
// definition (the part that comes after `type foo`).
func GoTypeDef(ds DataStructure, jsonTags bool) string {
	var buffer bytes.Buffer
	def := ds.Definition()
	t := def.Type
	switch actual := t.(type) {
	case Primitive:
		return GoTypeName(t)
	case *Array:
		return "[]" + GoTypeDef(actual.ElemType, jsonTags)
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
			typedef := GoTypeDef(actual[name], jsonTags)
			fname := Goify(name, true)
			var tags string
			if jsonTags {
				var omit string
				if !def.IsRequired(name) {
					omit = ",omitempty"
				}
				tags = fmt.Sprintf(" `json:\"%s%s\"`", name, omit)
			}
			field := fmt.Sprintf("\t%s %s%s\n", fname, typedef, tags)
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
	case Object, *UserTypeDefinition, *MediaTypeDefinition:
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
		return GoTypeDef(&AttributeDefinition{Type: actual}, false)
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

// tabs returns a string made of depth tab characters.
func tabs(depth int) string {
	var tabs string
	for i := 0; i < depth; i++ {
		tabs += "\t"
	}
	//	return fmt.Sprintf("%d%s", depth, tabs)
	return tabs
}

// tabulate prefixes each line of the given string with depth tab characters except the first one.
func tabulate(text string, depth int) string {
	lines := strings.Split(text, "\n")
	if len(lines) == 1 {
		return text
	}
	var b bytes.Buffer
	b.WriteString(lines[0])
	b.WriteByte('\n')
	for i := 1; i < len(lines); i++ {
		b.WriteString(tabs(depth))
		b.WriteString(lines[i])
		if i < len(lines)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}
