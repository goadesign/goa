package codegen

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"
	"unicode"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/dslengine"
)

// TransformMapKey is the name of the metadata used to specify the key for mapping fields when
// generating the code that transforms one data structure into another.
const TransformMapKey = "transform:key"

var (
	// TempCount holds the value appended to variable names to make them unique.
	TempCount int

	// Templates used by GoTypeTransform
	transformT       *template.Template
	transformArrayT  *template.Template
	transformHashT   *template.Template
	transformObjectT *template.Template
)

// Initialize all templates
func init() {
	var err error
	fn := template.FuncMap{
		"tabs":               Tabs,
		"add":                func(a, b int) int { return a + b },
		"goify":              Goify,
		"gotyperef":          GoTypeRef,
		"gotypename":         GoTypeName,
		"transformAttribute": transformAttribute,
		"transformArray":     transformArray,
		"transformHash":      transformHash,
		"transformObject":    transformObject,
		"typeName":           typeName,
	}
	if transformT, err = template.New("transform").Funcs(fn).Parse(transformTmpl); err != nil {
		panic(err) // bug
	}
	if transformArrayT, err = template.New("transformArray").Funcs(fn).Parse(transformArrayTmpl); err != nil {
		panic(err) // bug
	}
	if transformHashT, err = template.New("transformHash").Funcs(fn).Parse(transformHashTmpl); err != nil {
		panic(err) // bug
	}
	if transformObjectT, err = template.New("transformObject").Funcs(fn).Parse(transformObjectTmpl); err != nil {
		panic(err) // bug
	}
}

// GoTypeDef returns the Go code that defines a Go type which matches the data structure
// definition (the part that comes after `type foo`).
// tabs is the number of tab character(s) used to tabulate the definition however the first
// line is never indented.
// jsonTags controls whether to produce json tags.
// private controls whether the field is a pointer or not. All fields in the struct are
//   pointers for a private struct.
func GoTypeDef(ds design.DataStructure, tabs int, jsonTags, private bool) string {
	def := ds.Definition()
	if tname, ok := def.Metadata["struct:field:type"]; ok {
		if len(tname) > 0 {
			return tname[0]
		}
	}
	t := def.Type
	switch actual := t.(type) {
	case design.Primitive:
		return GoTypeName(t, nil, tabs, private)
	case *design.Array:
		d := GoTypeDef(actual.ElemType, tabs, jsonTags, private)
		if actual.ElemType.Type.IsObject() {
			d = "*" + d
		}
		return "[]" + d
	case *design.Hash:
		keyDef := GoTypeDef(actual.KeyType, tabs, jsonTags, private)
		if actual.KeyType.Type.IsObject() {
			keyDef = "*" + keyDef
		}
		elemDef := GoTypeDef(actual.ElemType, tabs, jsonTags, private)
		if actual.ElemType.Type.IsObject() {
			elemDef = "*" + elemDef
		}
		return fmt.Sprintf("map[%s]%s", keyDef, elemDef)
	case design.Object:
		return goTypeDefObject(actual, def, tabs, jsonTags, private)
	case *design.UserTypeDefinition:
		return GoTypeName(actual, actual.AllRequired(), tabs, private)
	case *design.MediaTypeDefinition:
		return GoTypeName(actual, actual.AllRequired(), tabs, private)
	default:
		panic("goa bug: unknown data structure type")
	}
}

// goTypeDefObject returns the Go code that defines a Go struct.
func goTypeDefObject(obj design.Object, def *design.AttributeDefinition, tabs int, jsonTags, private bool) string {
	var buffer bytes.Buffer
	buffer.WriteString("struct {\n")
	keys := make([]string, len(obj))
	i := 0
	for n := range obj {
		keys[i] = n
		i++
	}
	sort.Strings(keys)
	for _, name := range keys {
		WriteTabs(&buffer, tabs+1)
		field := obj[name]
		typedef := GoTypeDef(field, tabs+1, jsonTags, private)
		if (field.Type.IsPrimitive() && private) || field.Type.IsObject() || def.IsPrimitivePointer(name) {
			typedef = "*" + typedef
		}
		fname := GoifyAtt(field, name, true)
		var tags string
		if jsonTags {
			tags = attributeTags(def, field, name, private)
		}
		desc := obj[name].Description
		if desc != "" {
			desc = strings.Replace(desc, "\n", "\n\t// ", -1)
			desc = fmt.Sprintf("// %s\n\t", desc)
		}
		buffer.WriteString(fmt.Sprintf("%s%s %s%s\n", desc, fname, typedef, tags))
	}
	WriteTabs(&buffer, tabs)
	buffer.WriteString("}")
	return buffer.String()
}

// attributeTags computes the struct field tags.
func attributeTags(parent, att *design.AttributeDefinition, name string, private bool) string {
	var elems []string
	keys := make([]string, len(att.Metadata))
	i := 0
	for k := range att.Metadata {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	for _, key := range keys {
		val := att.Metadata[key]
		if strings.HasPrefix(key, "struct:tag:") {
			name := key[11:]
			value := strings.Join(val, ",")
			elems = append(elems, fmt.Sprintf("%s:\"%s\"", name, value))
		}
	}
	if len(elems) > 0 {
		return " `" + strings.Join(elems, " ") + "`"
	}
	// Default algorithm
	var omit string
	if private || (!parent.IsRequired(name) && !parent.HasDefaultValue(name)) {
		omit = ",omitempty"
	}
	return fmt.Sprintf(" `form:\"%s%s\" json:\"%s%s\" xml:\"%s%s\"`", name, omit, name, omit, name, omit)
}

// GoTypeRef returns the Go code that refers to the Go type which matches the given data type
// (the part that comes after `var foo`)
// required only applies when referring to a user type that is an object defined inline. In this
// case the type (Object) does not carry the required field information defined in the parent
// (anonymous) attribute.
// tabs is used to properly tabulate the object struct fields and only applies to this case.
// This function assumes the type is in the same package as the code accessing it.
func GoTypeRef(t design.DataType, required []string, tabs int, private bool) string {
	tname := GoTypeName(t, required, tabs, private)
	if mt, ok := t.(*design.MediaTypeDefinition); ok {
		if mt.IsError() {
			return "error"
		}
	}
	if t.IsObject() {
		return "*" + tname
	}
	return tname
}

// GoTypeName returns the Go type name for a data type.
// tabs is used to properly tabulate the object struct fields and only applies to this case.
// This function assumes the type is in the same package as the code accessing it.
// required only applies when referring to a user type that is an object defined inline. In this
// case the type (Object) does not carry the required field information defined in the parent
// (anonymous) attribute.
func GoTypeName(t design.DataType, required []string, tabs int, private bool) string {
	switch actual := t.(type) {
	case design.Primitive:
		return GoNativeType(t)
	case *design.Array:
		return "[]" + GoTypeRef(actual.ElemType.Type, actual.ElemType.AllRequired(), tabs+1, private)
	case design.Object:
		att := &design.AttributeDefinition{Type: actual}
		if len(required) > 0 {
			requiredVal := &dslengine.ValidationDefinition{Required: required}
			att.Validation.Merge(requiredVal)
		}
		return GoTypeDef(att, tabs, false, private)
	case *design.Hash:
		return fmt.Sprintf(
			"map[%s]%s",
			GoTypeRef(actual.KeyType.Type, actual.KeyType.AllRequired(), tabs+1, private),
			GoTypeRef(actual.ElemType.Type, actual.ElemType.AllRequired(), tabs+1, private),
		)
	case *design.UserTypeDefinition:
		return Goify(actual.TypeName, !private)
	case *design.MediaTypeDefinition:
		if actual.IsError() {
			return "error"
		}
		return Goify(actual.TypeName, !private)
	default:
		panic(fmt.Sprintf("goa bug: unknown type %#v", actual))
	}
}

// GoNativeType returns the Go built-in type from which instances of t can be initialized.
func GoNativeType(t design.DataType) string {
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
		case design.DateTimeKind:
			return "time.Time"
		case design.UUIDKind:
			return "uuid.UUID"
		case design.AnyKind:
			return "interface{}"
		default:
			panic(fmt.Sprintf("goa bug: unknown primitive type %#v", actual))
		}
	case *design.Array:
		return "[]" + GoNativeType(actual.ElemType.Type)
	case design.Object:
		return "map[string]interface{}"
	case *design.Hash:
		return fmt.Sprintf("map[%s]%s", GoNativeType(actual.KeyType.Type), GoNativeType(actual.ElemType.Type))
	case *design.MediaTypeDefinition:
		return GoNativeType(actual.Type)
	case *design.UserTypeDefinition:
		return GoNativeType(actual.Type)
	default:
		panic(fmt.Sprintf("goa bug: unknown type %#v", actual))
	}
}

// GoTypeDesc returns the description of a type.  If no description is defined
// for the type, one will be generated.
func GoTypeDesc(t design.DataType, upper bool) string {
	switch actual := t.(type) {
	case *design.UserTypeDefinition:
		if actual.Description != "" {
			return strings.Replace(actual.Description, "\n", "\n// ", -1)
		}

		return Goify(actual.TypeName, upper) + " user type."
	case *design.MediaTypeDefinition:
		if actual.Description != "" {
			return strings.Replace(actual.Description, "\n", "\n// ", -1)
		}
		name := Goify(actual.TypeName, upper)
		if actual.View != "default" {
			name += Goify(actual.View, true)
		}

		switch elem := actual.UserTypeDefinition.AttributeDefinition.Type.(type) {
		case *design.Array:
			elemName := GoTypeName(elem.ElemType.Type, nil, 0, !upper)
			if actual.View != "default" {
				elemName += Goify(actual.View, true)
			}
			return fmt.Sprintf("%s media type is a collection of %s.", name, elemName)
		default:
			return name + " media type."
		}
	default:
		return ""
	}
}

var commonInitialisms = map[string]bool{
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

// GoifyAtt honors any struct:field:name metadata set on the attribute and calls Goify with the tag
// value if present or the given name otherwise.
func GoifyAtt(att *design.AttributeDefinition, name string, firstUpper bool) string {
	if tname, ok := att.Metadata["struct:field:name"]; ok {
		if len(tname) > 0 {
			name = tname[0]
		}
	}
	return Goify(name, firstUpper)
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

// Reserved golang keywords and package names
var Reserved = map[string]bool{
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

	// stdlib and goa packages used by generated code
	"fmt":  true,
	"http": true,
	"json": true,
	"os":   true,
	"url":  true,
	"time": true,
}

// validIdentifier returns true if the rune is a letter or number
func validIdentifier(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

// fixReserved appends an underscore on to Go reserved keywords.
func fixReserved(w string) string {
	if Reserved[w] {
		w += "_"
	}
	return w
}

// GoTypeTransform produces Go code that initializes the data structure defined by target from an
// instance of the data structure described by source. The algorithm matches object fields by name
// or using the value of the "transform:key" attribute metadata when present.
// The function returns an error if target is not compatible with source (different type, fields of
// different type etc). It ignores fields in target that don't have a match in source.
func GoTypeTransform(source, target *design.UserTypeDefinition, targetPkg, funcName string) (string, error) {
	var impl string
	var err error
	switch {
	case source.IsObject():
		if !target.IsObject() {
			return "", fmt.Errorf("source is an object but target type is %s", target.Type.Name())
		}
		impl, err = transformObject(source.ToObject(), target.ToObject(), targetPkg, target.TypeName, "source", "target", 1)
	case source.IsArray():
		if !target.IsArray() {
			return "", fmt.Errorf("source is an array but target type is %s", target.Type.Name())
		}
		impl, err = transformArray(source.ToArray(), target.ToArray(), targetPkg, "source", "target", 1)
	case source.IsHash():
		if !target.IsHash() {
			return "", fmt.Errorf("source is a hash but target type is %s", target.Type.Name())
		}
		impl, err = transformHash(source.ToHash(), target.ToHash(), targetPkg, "source", "target", 1)
	default:
		panic("cannot transform primitive types") // bug
	}

	if err != nil {
		return "", err
	}
	t := GoTypeRef(target, nil, 0, false)
	if strings.HasPrefix(t, "*") && len(targetPkg) > 0 {
		t = fmt.Sprintf("*%s.%s", targetPkg, t[1:])
	}
	data := map[string]interface{}{
		"Name":      funcName,
		"Source":    source,
		"Target":    target,
		"TargetRef": t,
		"TargetPkg": targetPkg,
		"Impl":      impl,
	}
	return RunTemplate(transformT, data), nil
}

// GoTypeTransformName generates a valid Go identifer that is adequate for naming the type
// transform function that creates an instance of the data structure described by target from an
// instance of the data strucuture described by source.
func GoTypeTransformName(source, target *design.UserTypeDefinition, suffix string) string {
	return fmt.Sprintf("%sTo%s%s", Goify(source.TypeName, true), Goify(target.TypeName, true), Goify(suffix, true))
}

// WriteTabs is a helper function that writes count tabulation characters to buf.
func WriteTabs(buf *bytes.Buffer, count int) {
	for i := 0; i < count; i++ {
		buf.WriteByte('\t')
	}
}

// Tempvar generates a unique variable name.
func Tempvar() string {
	TempCount++
	return fmt.Sprintf("tmp%d", TempCount)
}

// RunTemplate executs the given template with the given input and returns
// the rendered string.
func RunTemplate(tmpl *template.Template, data interface{}) string {
	var b bytes.Buffer
	err := tmpl.Execute(&b, data)
	if err != nil {
		panic(err) // should never happen, bug if it does.
	}
	return b.String()
}

func transformAttribute(source, target *design.AttributeDefinition, targetPkg, sctx, tctx string, depth int) (string, error) {
	if source.Type.Kind() != target.Type.Kind() {
		return "", fmt.Errorf("incompatible attribute types: %s is of type %s but %s is of type %s",
			sctx, source.Type.Name(), tctx, target.Type.Name())
	}
	switch {
	case source.Type.IsArray():
		return transformArray(source.Type.ToArray(), target.Type.ToArray(), targetPkg, sctx, tctx, depth)
	case source.Type.IsHash():
		return transformHash(source.Type.ToHash(), target.Type.ToHash(), targetPkg, sctx, tctx, depth)
	case source.Type.IsObject():
		return transformObject(source.Type.ToObject(), target.Type.ToObject(), targetPkg, typeName(target), sctx, tctx, depth)
	default:
		return fmt.Sprintf("%s%s = %s\n", Tabs(depth), tctx, sctx), nil
	}
}

func transformObject(source, target design.Object, targetPkg, targetType, sctx, tctx string, depth int) (string, error) {
	attributeMap, err := computeMapping(source, target, sctx, tctx)
	if err != nil {
		return "", err
	}

	// First validate that all attributes are compatible - doing that in a template doesn't make
	// sense.
	for s, t := range attributeMap {
		sourceAtt := source[s]
		targetAtt := target[t]
		if sourceAtt.Type.Kind() != targetAtt.Type.Kind() {
			return "", fmt.Errorf("incompatible attribute types: %s.%s is of type %s but %s.%s is of type %s",
				sctx, source.Name(), sourceAtt.Type.Name(), tctx, target.Name(), targetAtt.Type.Name())
		}
	}

	// We're good - generate
	data := map[string]interface{}{
		"AttributeMap": attributeMap,
		"Source":       source,
		"Target":       target,
		"TargetPkg":    targetPkg,
		"TargetType":   targetType,
		"SourceCtx":    sctx,
		"TargetCtx":    tctx,
		"Depth":        depth,
	}
	return RunTemplate(transformObjectT, data), nil
}

func transformArray(source, target *design.Array, targetPkg, sctx, tctx string, depth int) (string, error) {
	if source.ElemType.Type.Kind() != target.ElemType.Type.Kind() {
		return "", fmt.Errorf("incompatible attribute types: %s is an array with elements of type %s but %s is an array with elements of type %s",
			sctx, source.ElemType.Type.Name(), tctx, target.ElemType.Type.Name())
	}
	data := map[string]interface{}{
		"Source":    source,
		"Target":    target,
		"TargetPkg": targetPkg,
		"SourceCtx": sctx,
		"TargetCtx": tctx,
		"Depth":     depth,
	}
	return RunTemplate(transformArrayT, data), nil
}

func transformHash(source, target *design.Hash, targetPkg, sctx, tctx string, depth int) (string, error) {
	if source.ElemType.Type.Kind() != target.ElemType.Type.Kind() {
		return "", fmt.Errorf("incompatible attribute types: %s is a hash with elements of type %s but %s is a hash with elements of type %s",
			sctx, source.ElemType.Type.Name(), tctx, target.ElemType.Type.Name())
	}
	if source.KeyType.Type.Kind() != target.KeyType.Type.Kind() {
		return "", fmt.Errorf("incompatible attribute types: %s is a hash with keys of type %s but %s is a hash with keys of type %s",
			sctx, source.KeyType.Type.Name(), tctx, target.KeyType.Type.Name())
	}
	data := map[string]interface{}{
		"Source":    source,
		"Target":    target,
		"TargetPkg": targetPkg,
		"SourceCtx": sctx,
		"TargetCtx": tctx,
		"Depth":     depth,
	}
	return RunTemplate(transformHashT, data), nil
}

// computeMapping returns a map that indexes the target type definition object attributes with the
// corresponding source type definition object attributes. An attribute is associated with another
// attribute if their map key match. The map key of an attribute is the value of the TransformMapKey
// metadata if present, the attribute name otherwise.
// The function returns an error if the TransformMapKey metadata is malformed (has no value).
func computeMapping(source, target design.Object, sctx, tctx string) (map[string]string, error) {
	attributeMap := make(map[string]string)
	sourceMap := make(map[string]string)
	targetMap := make(map[string]string)
	for name, att := range source {
		key := name
		if keys, ok := att.Metadata[TransformMapKey]; ok {
			if len(keys) == 0 {
				return nil, fmt.Errorf("invalid metadata transform key: missing value on attribte %s of %s", name, sctx)
			}
			key = keys[0]
		}
		sourceMap[key] = name
	}
	for name, att := range target {
		key := name
		if keys, ok := att.Metadata[TransformMapKey]; ok {
			if len(keys) == 0 {
				return nil, fmt.Errorf("invalid metadata transform key: missing value on attribute %s of %s", name, tctx)
			}
			key = keys[0]
		}
		targetMap[key] = name
	}
	for key, attName := range sourceMap {
		if targetAtt, ok := targetMap[key]; ok {
			attributeMap[attName] = targetAtt
		}
	}
	return attributeMap, nil
}

// toSlice returns Go code that represents the given slice.
func toSlice(val []interface{}) string {
	elems := make([]string, len(val))
	for i, v := range val {
		elems[i] = fmt.Sprintf("%#v", v)
	}
	return fmt.Sprintf("[]interface{}{%s}", strings.Join(elems, ", "))
}

// typeName returns the type name of the given attribute if it is a named type, empty string otherwise.
func typeName(att *design.AttributeDefinition) (name string) {
	if ut, ok := att.Type.(*design.UserTypeDefinition); ok {
		name = Goify(ut.TypeName, true)
	} else if mt, ok := att.Type.(*design.MediaTypeDefinition); ok {
		name = Goify(mt.TypeName, true)
	}
	return
}

const transformTmpl = `func {{ .Name }}(source {{ gotyperef .Source nil 0 false }}) (target {{ .TargetRef }}) {
{{ .Impl }}	return
}
`

const transformObjectTmpl = `{{ tabs .Depth }}{{ .TargetCtx }} = new({{ if .TargetPkg }}{{ .TargetPkg }}.{{ end }}{{ if .TargetType }}{{ .TargetType }}{{ else }}{{ gotyperef .Target.Type .Target.AllRequired 1 false }}{{ end }})
{{ range $source, $target := .AttributeMap }}{{/*
*/}}{{ $sourceAtt := index $.Source $source }}{{ $targetAtt := index $.Target $target }}{{/*
*/}}{{ $source := goify $source true }}{{ $target := goify $target true }}{{/*
*/}}{{     if $sourceAtt.Type.IsArray }}{{ transformArray  $sourceAtt.Type.ToArray  $targetAtt.Type.ToArray  $.TargetPkg (printf "%s.%s" $.SourceCtx $source) (printf "%s.%s" $.TargetCtx $target) $.Depth }}{{/*
*/}}{{ else if $sourceAtt.Type.IsHash }}{{  transformHash   $sourceAtt.Type.ToHash   $targetAtt.Type.ToHash   $.TargetPkg (printf "%s.%s" $.SourceCtx $source) (printf "%s.%s" $.TargetCtx $target) $.Depth }}{{/*
*/}}{{ else if $sourceAtt.Type.IsObject }}{{ transformObject $sourceAtt.Type.ToObject $targetAtt.Type.ToObject $.TargetPkg (typeName $targetAtt) (printf "%s.%s" $.SourceCtx $source) (printf "%s.%s" $.TargetCtx $target) $.Depth }}{{/*
*/}}{{ else }}{{ tabs $.Depth }}{{ $.TargetCtx }}.{{ $target }} = {{ $.SourceCtx }}.{{ $source }}
{{ end }}{{ end }}`

const transformArrayTmpl = `{{ tabs .Depth }}{{ .TargetCtx}} = make([]{{ gotyperef .Target.ElemType.Type nil 0 false }}, len({{ .SourceCtx }}))
{{ tabs .Depth }}for i, v := range {{ .SourceCtx }} {
{{ transformAttribute .Source.ElemType .Target.ElemType .TargetPkg (printf "%s[i]" .SourceCtx) (printf "%s[i]" .TargetCtx) (add .Depth 1) }}{{/*
*/}}{{ tabs .Depth }}}
`

const transformHashTmpl = `{{ tabs .Depth }}{{ .TargetCtx }} = make(map[{{ gotyperef .Target.KeyType.Type nil 0 false }}]{{ gotyperef .Target.ElemType.Type nil 0 false }}, len({{ .SourceCtx }}))
{{ tabs .Depth }}for k, v := range {{ .SourceCtx }} {
{{ tabs .Depth }}	var tk {{ gotyperef .Target.KeyType.Type nil 0 false }}
{{ transformAttribute .Source.KeyType .Target.KeyType .TargetPkg "k" "tk" (add .Depth 1) }}{{/*
*/}}{{ tabs .Depth }}	var tv {{ gotyperef .Target.ElemType.Type nil 0 false }}
{{ transformAttribute .Source.ElemType .Target.ElemType .TargetPkg "v" "tv" (add .Depth 1) }}{{/*
*/}}{{ tabs .Depth }}	{{ .TargetCtx }}[tk] = tv
{{ tabs .Depth }}}
`
