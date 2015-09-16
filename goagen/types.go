package goagen

import (
	"bytes"
	"fmt"
	"sort"
	"text/template"
	"unicode"

	"github.com/raphael/goa/design"
)

var (
	// TempCount holds the value appended to variable names to make them unique.
	TempCount int

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
		"goify":              Goify,
		"tabs":               Tabs,
		"add":                func(a, b int) int { return a + b },
		"tempvar":            tempvar,
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
func TypeUnmarshaler(t design.DataType, context, source, target string) string {
	return typeUnmarshalerR(t, context, source, target, 1)
}
func typeUnmarshalerR(t design.DataType, context, source, target string, depth int) string {
	switch actual := t.(type) {
	case design.Primitive:
		return primitiveUnmarshalerR(actual, context, source, target, depth)
	case *design.Array:
		return arrayUnmarshalerR(actual, context, source, target, depth)
	case design.Object:
		return objectUnmarshalerR(actual, context, source, target, depth)
	case design.NamedType:
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
func AttributeUnmarshaler(att *design.AttributeDefinition, context, source, target string) string {
	return attributeUnmarshalerR(att, context, source, target, 1)
}
func attributeUnmarshalerR(att *design.AttributeDefinition, context, source, target string, depth int) string {
	return typeUnmarshalerR(att.Type, context, source, target, depth) +
		validationCheckerR(att, context, target, depth)
}

// PrimitiveUnmarshaler produces the go code that initializes a primitive type from its JSON
// representation.
// source is the name of the variable that contains the raw interface{} value and target the
// name of the variable to initialize.
// The generated code assumes that there is a variable called "err" of type error that it can use
// to record errors.
func PrimitiveUnmarshaler(p design.Primitive, context, source, target string) string {
	return primitiveUnmarshalerR(p, context, source, target, 1)
}
func primitiveUnmarshalerR(p design.Primitive, context, source, target string, depth int) string {
	data := map[string]interface{}{
		"source":  source,
		"target":  target,
		"type":    p,
		"context": context,
		"depth":   depth,
	}
	return runTemplate(primitiveT, data)
}

// ArrayUnmarshaler produces the go code that initializes an array from its JSON representation.
// source is the name of the variable that contains the raw interface{} value and target the
// name of the variable to initialize.
// The generated code assumes that there is a variable called "err" of type error that it can use
// to record errors.
func ArrayUnmarshaler(a *design.Array, context, source, target string) string {
	return arrayUnmarshalerR(a, context, source, target, 1)
}
func arrayUnmarshalerR(a *design.Array, context, source, target string, depth int) string {
	data := map[string]interface{}{
		"source":   source,
		"target":   target,
		"elemType": a.ElemType,
		"context":  context,
		"depth":    depth,
	}
	return runTemplate(arrayT, data)
}

// ObjectUnmarshaler produces the go code that initializes an object type from its JSON
// representation.
// source is the name of the variable that contains the raw interface{} value and target the
// name of the variable to initialize.
// The generated code assumes that there is a variable called "err" of type error that it can use
// to record errors.
func ObjectUnmarshaler(o design.Object, context, source, target string) string {
	return objectUnmarshalerR(o, context, source, target, 1)
}
func objectUnmarshalerR(o design.Object, context, source, target string, depth int) string {
	data := map[string]interface{}{
		"source":  source,
		"target":  target,
		"type":    o,
		"context": context,
		"depth":   depth,
	}
	return runTemplate(objectT, data)
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
func NamedTypeUnmarshaler(t design.NamedType, context, source, target string) string {
	return namedTypeUnmarshalerR(t, context, source, target, 1)
}
func namedTypeUnmarshalerR(t design.NamedType, context, source, target string, depth int) string {
	data := map[string]interface{}{
		"type":    t,
		"source":  source,
		"target":  target,
		"context": context,
		"depth":   depth,
	}
	return runTemplate(userT, data)
}

// GoTypeDef returns the Go code that defines a Go type which matches the data structure
// definition (the part that comes after `type foo`).
// tabs indicates the number of tab character(s) used to tabulate the definition however the first
// line is never indented.
// jsonTags controls whether to produce json tags.
// inner indicates whether to prefix the struct of an attribute of type object with *.
func GoTypeDef(ds design.DataStructure, tabs int, jsonTags, inner bool) string {
	var buffer bytes.Buffer
	def := ds.Definition()
	t := def.Type
	switch actual := t.(type) {
	case design.Primitive:
		return GoTypeName(t, tabs)
	case *design.Array:
		return "[]" + GoTypeDef(actual.ElemType, tabs, jsonTags, true)
	case design.Object:
		if inner {
			buffer.WriteByte('*')
		}
		buffer.WriteString("struct {\n")
		keys := make([]string, len(actual))
		i := 0
		for n := range actual {
			keys[i] = n
			i++
		}
		sort.Strings(keys)
		for _, name := range keys {
			WriteTabs(&buffer, tabs+1)
			typedef := GoTypeDef(actual[name], tabs+1, jsonTags, true)
			fname := Goify(name, true)
			var tags string
			if jsonTags {
				var omit string
				if !def.IsRequired(name) {
					omit = ",omitempty"
				}
				tags = fmt.Sprintf(" `json:\"%s%s\"`", name, omit)
			}
			buffer.WriteString(fmt.Sprintf("%s %s%s\n", fname, typedef, tags))
		}
		WriteTabs(&buffer, tabs)
		buffer.WriteString("}")
		return buffer.String()
	case *design.UserTypeDefinition, *design.MediaTypeDefinition:
		return "*" + GoTypeName(actual, tabs)
	default:
		panic("goa bug: unknown data structure type")
	}
}

// GoTypeRef returns the Go code that refers to the Go type which matches the given data type
// (the part that comes after `var foo`)
// tabs is used to properly tabulate the object struct fields and only applies to this case.
func GoTypeRef(t design.DataType, tabs int) string {
	switch t.(type) {
	case design.Object, *design.UserTypeDefinition, *design.MediaTypeDefinition:
		return "*" + GoTypeName(t, tabs)
	default:
		return GoTypeName(t, tabs)
	}
}

// GoTypeName returns the Go type name for a data type.
// tabs is used to properly tabulate the object struct fields and only applies to this case.
func GoTypeName(t design.DataType, tabs int) string {
	switch actual := t.(type) {
	case design.Primitive:
		switch actual.Kind() {
		case design.BooleanKind:
			return "bool"
		case design.IntegerKind:
			return "int"
		case design.NumberKind:
			return "float64"
		case design.StringKind:
			return "string"
		default:
			panic(fmt.Sprintf("goa bug: unknown primitive type %#v", actual))
		}
	case *design.Array:
		return "[]" + GoTypeRef(actual.ElemType.Type, tabs+1)
	case design.Object:
		return GoTypeDef(&design.AttributeDefinition{Type: actual}, tabs, false, false)
	case *design.UserTypeDefinition:
		return Goify(actual.Name, true)
	case *design.MediaTypeDefinition:
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

// WriteTabs is a helper function that writes count tabulation characters to buf.
func WriteTabs(buf *bytes.Buffer, count int) {
	for i := 0; i < count; i++ {
		buf.WriteByte('\t')
	}
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

// tempvar generates a unique temp var name.
func tempvar() string {
	TempCount++
	return fmt.Sprintf("tmp%d", TempCount)
}

// runTemplate executs the given template with the given input and returns
// the rendered string.
func runTemplate(tmpl *template.Template, data interface{}) string {
	var b bytes.Buffer
	err := tmpl.Execute(&b, data)
	if err != nil {
		panic(err) // should never happen, bug if it does.
	}
	return b.String()
}

const (
	primitiveTmpl = `{{tabs .depth}}if val, ok := {{.source}}.({{gotyperef .type (add .depth 1)}}); ok {
{{tabs .depth}}	{{.target}} = val
{{tabs .depth}}} else {
{{tabs .depth}}	err = goa.IncompatibleTypeError(` + "`" + `{{.context}}` + "`" + `, {{.source}}, "{{gotyperef .type (add .depth 1)}}")
{{tabs .depth}}}`

	arrayTmpl = `{{tabs .depth}}if val, ok := {{.source}}.([]interface{}); ok {
{{tabs .depth}}	{{.target}} = make([]{{gotyperef .elemType.Type (add .depth 2)}}, len(val))
{{tabs .depth}}	for i, v := range val {
{{tabs .depth}}		{{$temp := tempvar}}var {{$temp}} {{gotyperef .elemType.Type (add .depth 3)}}
{{unmarshalAttribute .elemType (printf "%s[*]" .context) "v" $temp (add .depth 2)}}
{{tabs .depth}}		{{printf "%s[i]" .target}} = {{$temp}}
{{tabs .depth}}	}
{{tabs .depth}}} else {
{{tabs .depth}}	err = goa.IncompatibleTypeError(` + "`" + `{{.context}}` + "`" + `, {{.source}}, "[]interface{}")
{{tabs .depth}}}`

	objectTmpl = `{{tabs .depth}}if val, ok := {{.source}}.(map[string]interface{}); ok {
{{tabs .depth}}{{$context := .context}}{{$depth := .depth}}{{$target := .target}}	{{$target}} = new({{gotypename .type (add .depth 1)}})
{{range $name, $att := .type}}{{tabs $depth}}	if v, ok := val["{{$name}}"]; ok {
{{tabs $depth}}		{{$temp := tempvar}}var {{$temp}} {{gotyperef $att.Type (add $depth 2)}}
{{unmarshalType $att.Type (printf "%s.%s" $context (goify $name true)) "v" $temp (add $depth 2)}}
{{tabs $depth}}		{{printf "%s.%s" $target (goify $name true)}} = {{$temp}}
{{tabs $depth}}	}
{{end}}{{tabs $depth}}} else {
{{tabs .depth}}	err = goa.IncompatibleTypeError(` + "`" + `{{.context}}` + "`" + `, {{.source}}, "map[string]interface{}")
{{tabs .depth}}}`

	userTmpl = `{{tabs .depth}}if val, ok := {{.source}}.(map[string]interface{}); ok {
{{tabs .depth}}{{$depth := .depth}}{{$context := .context}}{{$target := .target}}	{{.target}} = new({{.type.Name}})
{{range $name, $att := .type.Definition.Type.ToObject}}{{tabs $depth}}	if v, ok := val["{{$name}}"]; ok {
{{tabs $depth}}		{{$temp := tempvar}}var {{$temp}} {{gotyperef $att.Type (add $depth 2)}}
{{unmarshalType $att.Type (printf "%s.%s" $context (goify $name true)) "v" $temp (add $depth 2)}}
{{tabs $depth}}		{{printf "%s.%s" $target (goify $name true)}} = {{$temp}}
{{tabs $depth}}	}
{{end}}{{tabs .depth}}} else {
{{tabs .depth}}	err = goa.IncompatibleTypeError(` + "`" + `{{.context}}` + "`" + `, {{.source}}, "map[string]interface{}")
{{tabs .depth}}}`
)
