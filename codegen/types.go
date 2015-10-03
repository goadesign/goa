package codegen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/template"
	"unicode"

	"github.com/raphael/goa/design"
)

var (
	// TempCount holds the value appended to variable names to make them unique.
	TempCount int

	mArrayT          *template.Template
	mObjectT         *template.Template
	mLinkT           *template.Template
	unmPrimitiveT    *template.Template
	unmArrayT        *template.Template
	unmObjectOrUserT *template.Template
)

//  init instantiates the templates.
func init() {
	var err error
	fm := template.FuncMap{
		"marshalType":        typeMarshalerR,
		"marshalAttribute":   attributeMarshalerR,
		"marshalMediaType":   mediaTypeMarshalerR,
		"unmarshalType":      typeUnmarshalerR,
		"unmarshalAttribute": attributeUnmarshalerR,
		"validate":           validationCheckerR,
		"gotypename":         GoTypeName,
		"gotyperef":          GoTypeRef,
		"goify":              Goify,
		"tabs":               Tabs,
		"add":                func(a, b int) int { return a + b },
		"tempvar":            tempvar,
		"has":                has,
	}
	if mArrayT, err = template.New("array marshaler").Funcs(fm).Parse(mArrayTmpl); err != nil {
		panic(err)
	}
	if mObjectT, err = template.New("object marshaler").Funcs(fm).Parse(mObjectTmpl); err != nil {
		panic(err)
	}
	if mLinkT, err = template.New("links marshaler").Funcs(fm).Parse(mLinkTmpl); err != nil {
		panic(err)
	}
	if unmPrimitiveT, err = template.New("primitive unmarshaler").Funcs(fm).Parse(unmPrimitiveTmpl); err != nil {
		panic(err)
	}
	if unmArrayT, err = template.New("array unmarshaler").Funcs(fm).Parse(unmArrayTmpl); err != nil {
		panic(err)
	}
	if unmObjectOrUserT, err = template.New("object unmarshaler").Funcs(fm).Parse(unmObjectTmpl); err != nil {
		panic(err)
	}
}

// TypeMarshaler produces the Go code that initiliazes the variable named target which is an
// interface{} with the content of the variable named source which contains an instance of the type
// data structure.
// The code takes care of rendering media types according to the view defined on the attribute
// if any. It also renders media type links. Finally it validates the results using any type
// validation that is defined on the type attributes (if the type contains attributes).
// The generated code assumes that there is a variable called "err" of type error that it can use
// to record errors.
func TypeMarshaler(t design.DataType, context, source, target string) string {
	return typeMarshalerR(t, context, source, target, 1)
}
func typeMarshalerR(t design.DataType, context, source, target string, depth int) string {
	switch actual := t.(type) {
	case design.Primitive:
		return fmt.Sprintf("%s%s = %s", Tabs(depth), target, source)
	case *design.Array:
		return arrayMarshalerR(actual, context, source, target, depth)
	case design.Object, *design.UserTypeDefinition:
		return objectMarshalerR(actual.ToObject(), nil, context, source, target, depth)
	default:
		// this should never get called with a MediaType, MediaTypeMarshaler should be
		// called instead so the view is properly taken into account.
		panic(actual)
	}
}

// AttributeMarshaler produces the Go code that initiliazes the variable named target which holds an
// interface{} with the content of the variable named source which contains an
// instance of the attribute type data structure. The attribute view is used to render child
// attributes if there are any. As with TypeMarshaler the code renders media type links and runs any
// validation defined on the type definition.
//
// The generated code assumes that there is a variable called "err" of type error that it can use
// to record errors.
func AttributeMarshaler(att *design.AttributeDefinition, context, source, target string) string {
	return attributeMarshalerR(att, context, source, target, 1)
}
func attributeMarshalerR(att *design.AttributeDefinition, context, source, target string, depth int) string {
	var marshaler string
	switch actual := att.Type.(type) {
	case *design.MediaTypeDefinition:
		marshaler = mediaTypeMarshalerR(actual, context, source, target, att.View, depth)
	case design.Object:
		marshaler = objectMarshalerR(actual, att.AllRequired(), context, source, target, depth)
	default:
		marshaler = typeMarshalerR(att.Type, context, source, target, depth)
	}
	return marshaler + validationCheckerR(att, context, target, depth)
}

// ArrayMarshaler produces the Go code that marshals an array for rendering.
// source is the name of the variable that contains the array value and target the name of the
// variable to initialize.
// The generated code assumes that there is a variable called "err" of type error that it can use
// to record errors.
func ArrayMarshaler(a *design.Array, context, source, target string) string {
	return arrayMarshalerR(a, context, source, target, 1)
}
func arrayMarshalerR(a *design.Array, context, source, target string, depth int) string {
	data := map[string]interface{}{
		"source":   source,
		"target":   target,
		"elemType": a.ElemType,
		"context":  context,
		"depth":    depth,
	}
	return runTemplate(mArrayT, data)
}

// ObjectMarshaler produces the Go code that initializes the variable named target which holds a
// map of string to interface{} with the content of the variable named source which contains an
// instance of the object data structure. The code runs any validation defined on the object
// attribute definitions.
// The generated code assumes that there is a variable called "err" of type error that it can use
// to record errors.
func ObjectMarshaler(o design.Object, context, source, target string) string {
	return objectMarshalerR(o, nil, context, source, target, 1)
}
func objectMarshalerR(o design.Object, required []string, context, source, target string, depth int) string {
	data := map[string]interface{}{
		"type":     o,
		"required": required,
		"context":  context,
		"source":   source,
		"target":   target,
		"depth":    depth,
	}
	return runTemplate(mObjectT, data)
}

// MediaTypeMarshaler produces the Go code that initializes the variable named target which holds a
// map of string to interface{} with the content of the variable named source which contains an
// instance of the media type data structure. The code runs any validation defined on the media
// type definition. Also view is used to know which fields to copy and which ones to omit and for
// fields that are media types which view to use to render it. The rendering also takes care of
// following links.
// The generated code assumes that there is a variable called "err" of type error that it can use
// to record errors.
func MediaTypeMarshaler(mt *design.MediaTypeDefinition, context, source, target, view string) string {
	return mediaTypeMarshalerR(mt, context, source, target, view, 1)
}
func mediaTypeMarshalerR(mt *design.MediaTypeDefinition, context, source, target, view string, depth int) string {
	rendered := mt.AttributeDefinition
	if view == "" {
		view = "default"
	}
	if v, ok := mt.Views[view]; ok {
		rendered = &design.AttributeDefinition{
			Type:        v.Type.ToObject(),
			Validations: mt.Validations,
		}
	}
	links := mt.Links
	for n, l := range links {
		if l.MediaType == nil {
			// DSL validation makes sure this won't blow up.
			l.MediaType = mt.ToObject()[n].Type.(*design.MediaTypeDefinition)
		}
	}
	data := map[string]interface{}{
		"links":   links,
		"context": context,
		"source":  source,
		"target":  target,
		"view":    view,
		"depth":   depth,
	}
	linkMarshaler := runTemplate(mLinkT, data)
	return attributeMarshalerR(rendered, context, source, target, depth) + linkMarshaler
}

// TypeUnmarshaler produces the Go code that initializes a variable of the given type given
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
		return objectUnmarshalerR(actual, nil, context, source, target, depth)
	case *design.UserTypeDefinition, *design.MediaTypeDefinition:
		return namedTypeUnmarshalerR(actual, context, source, target, depth)
	default:
		panic(actual)
	}
}

// AttributeUnmarshaler produces the Go code that initializes an attribute given a deserialized
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
	var unmarshaler string
	if o, ok := att.Type.(design.Object); ok {
		unmarshaler = objectUnmarshalerR(o, att.AllRequired(), context, source, target, depth)
	} else {
		unmarshaler = typeUnmarshalerR(att.Type, context, source, target, depth)
	}
	return unmarshaler
}

// PrimitiveUnmarshaler produces the Go code that initializes a primitive type from its JSON
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
	return runTemplate(unmPrimitiveT, data)
}

// ArrayUnmarshaler produces the Go code that initializes an array from its JSON representation.
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
	return runTemplate(unmArrayT, data)
}

// ObjectUnmarshaler produces the Go code that initializes an object type from its JSON
// representation.
// source is the name of the variable that contains the raw interface{} value and target the
// name of the variable to initialize.
// The generated code assumes that there is a variable called "err" of type error that it can use
// to record errors.
func ObjectUnmarshaler(o design.Object, context, source, target string) string {
	return objectUnmarshalerR(o, nil, context, source, target, 1)
}
func objectUnmarshalerR(o design.Object, required []string, context, source, target string, depth int) string {
	data := map[string]interface{}{
		"type":     o,
		"required": required,
		"context":  context,
		"source":   source,
		"target":   target,
		"depth":    depth,
	}
	return runTemplate(unmObjectOrUserT, data)
}

// NamedTypeUnmarshaler produces the Go code that initializes a named type from its JSON
// representation.
// This include running any validation defined on the type.
// source is the name of the variable that contains the raw interface{} value and target the
// name of the variable to initialize.
// context is used to keep track of recursion to produce helpful error messages in case of type
// mismatch or validation error.
// The generated code assumes that there is a variable called "err" of type error that it can use
// to record errors.
func NamedTypeUnmarshaler(t design.DataType, context, source, target string) string {
	return namedTypeUnmarshalerR(t, context, source, target, 1)
}
func namedTypeUnmarshalerR(t design.DataType, context, source, target string, depth int) string {
	data := map[string]interface{}{
		"type":    t,
		"context": context,
		"source":  source,
		"target":  target,
		"depth":   depth,
	}
	return runTemplate(unmObjectOrUserT, data)
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
	case *design.MediaTypeDefinition:
		return Goify(actual.TypeName, true)
	case *design.UserTypeDefinition:
		return Goify(actual.TypeName, true)
	case design.Object:
		return GoTypeDef(&design.AttributeDefinition{Type: actual}, tabs, false, false)
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
	if str == "ok" && firstUpper {
		return "OK"
	}
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

// has returns true is slice contains val, false otherwise.
func has(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

// toJSON returns the JSON representation of the given value.
func toJSON(val interface{}) string {
	js, err := json.Marshal(val)
	if err != nil {
		return "<error serializing value>"
	}
	return string(js)
}

// toSlice returns Go code that represents the given slice.
func toSlice(val []interface{}) string {
	elems := make([]string, len(val))
	for i, v := range val {
		elems[i] = fmt.Sprintf("%#v", v)
	}
	return fmt.Sprintf("[]interface{}{%s}", strings.Join(elems, ", "))
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
	mArrayTmpl = `{{tabs .depth}}{{.target}} = make([]interface{}, len({{.source}}))
{{tabs .depth}}for i, r := range {{.source}} {
{{marshalAttribute .elemType (printf "%s[*]" .context) "r" (printf "%s[i]" .target) (add .depth 1)}}
{{tabs .depth}}}
`

	mObjectTmpl = `{{$ctx := .}}{{range $r := .required}}{{$at := index $ctx.o $r}}{{$required := goify $r true}}{{if eq $at.Type.Kind 4}}{{tabs .depth}}if {{.source}}.{{$required}} == "" {
{{tabs .depth}} return fmt.Errorf("missing required attribute \"{{$r}}\"")
{{tabs .depth}}}{{else if gt $at.type.Kind 4}}{{tabs .depth}}if {{.source}}.{{$required}} == nil {
{{tabs .depth}} return fmt.Errorf("missing required attribute \"{{$r}}\"")
{{tabs .depth}}}
{{end}}{{end}}{{tabs .depth}}{{.target}} = map[string]interface{}{
{{range $n, $at := .type}}{{if lt $at.Type.Kind 5}}{{tabs $ctx.depth}}	"{{$n}}": {{$ctx.source}}.{{goify $n true}},
{{end}}{{end}}{{tabs $ctx.depth}}}{{range $n, $at := .type}}{{if gt $at.Type.Kind 4}}
{{tabs $ctx.depth}}if {{$ctx.source}}.{{goify $n true}} != nil {
{{marshalAttribute $at (printf "%s.%s" $ctx.context (goify $n true)) (printf "%s.%s" $ctx.source (goify $n true)) (printf "%s[\"%s\"]" $ctx.target $n) (add $ctx.depth 1)}}{{tabs $ctx.depth}}}{{end}}{{end}}
`

	mLinkTmpl = `{{if .links}}{{$ctx := .}}{{tabs .depth}}links := make(map[string]interface{})
{{range $n, $l := .links}}{{marshalMediaType $l.MediaType (printf "link %s" $n) (printf "%s.%s" $ctx.source (goify $l.Name true)) (printf "links[\"%s\"]" (goify $n true)) $l.View $ctx.depth}}{{end}}{{tabs .depth}}{{.target}}["links"] = links
{{end}}`

	unmPrimitiveTmpl = `{{tabs .depth}}if val, ok := {{.source}}.({{gotyperef .type (add .depth 1)}}); ok {
{{tabs .depth}}	{{.target}} = val
{{tabs .depth}}} else {
{{tabs .depth}}	err = goa.InvalidAttributeTypeError(` + "`" + `{{.context}}` + "`" + `, {{.source}}, "{{gotyperef .type (add .depth 1)}}", err)
{{tabs .depth}}}`

	unmArrayTmpl = `{{tabs .depth}}if val, ok := {{.source}}.([]interface{}); ok {
{{tabs .depth}}	{{.target}} = make([]{{gotyperef .elemType.Type (add .depth 2)}}, len(val))
{{tabs .depth}}	for i, v := range val {
{{tabs .depth}}		{{$temp := tempvar}}var {{$temp}} {{gotyperef .elemType.Type (add .depth 3)}}
{{unmarshalAttribute .elemType (printf "%s[*]" .context) "v" $temp (add .depth 2)}}{{$ctx := .}}{{range $ctx.elemType.Validations}}
{{validate $ctx.elemType (printf "%s[*]" $ctx.context) $temp (add $ctx.depth 2)}}{{end}}
{{tabs .depth}}		{{printf "%s[i]" .target}} = {{$temp}}
{{tabs .depth}}	}
{{tabs .depth}}} else {
{{tabs .depth}}	err = goa.InvalidAttributeTypeError(` + "`" + `{{.context}}` + "`" + `, {{.source}}, "[]interface{}", err)
{{tabs .depth}}}`

	unmObjectTmpl = `{{tabs .depth}}if val, ok := {{.source}}.(map[string]interface{}); ok {
{{tabs .depth}}{{$context := .context}}{{$depth := .depth}}{{$target := .target}}{{$required := .required}}	{{$target}} = new({{gotypename .type (add .depth 1)}})
{{range $name, $att := .type.ToObject}}{{tabs $depth}}	if v, ok := val["{{$name}}"]; ok {
{{tabs $depth}}		{{$temp := tempvar}}var {{$temp}} {{gotyperef $att.Type (add $depth 2)}}
{{unmarshalAttribute $att (printf "%s.%s" $context (goify $name true)) "v" $temp (add $depth 2)}}{{range $att.Validations}}
{{validate $att (printf "%s.%s" $context (goify $name true)) $temp (add $depth 2)}}{{end}}
{{tabs $depth}}		{{printf "%s.%s" $target (goify $name true)}} = {{$temp}}
{{tabs $depth}}	}{{if (has $required $name)}} else {
{{tabs $depth}}		err = goa.MissingAttributeError(` + "`" + `{{.context}}` + "`" + `, "{{$name}}", err)
{{tabs $depth}}	}{{end}}
{{end}}{{tabs $depth}}} else {
{{tabs .depth}}	err = goa.InvalidAttributeTypeError(` + "`" + `{{.context}}` + "`" + `, {{.source}}, "map[string]interface{}", err)
{{tabs .depth}}}`
)
