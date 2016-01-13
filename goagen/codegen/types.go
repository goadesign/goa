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

	mArrayT           *template.Template
	mObjectT          *template.Template
	mHashT            *template.Template
	mLinkT            *template.Template
	mCollectionT      *template.Template
	mUserImplT        *template.Template
	unmPrimitiveT     *template.Template
	unmUserPrimitiveT *template.Template
	unmUserT          *template.Template
	unmArrayT         *template.Template
	unmObjectT        *template.Template
	unmHashT          *template.Template
	unmUserImplT      *template.Template
)

//  init instantiates the templates.
func init() {
	var err error
	fm := template.FuncMap{
		"marshalAttribute":   attributeMarshalerR,
		"marshalMediaType":   mediaTypeMarshalerR,
		"unmarshalAttribute": attributeUnmarshalerR,
		"gotypename":         GoTypeName,
		"gotyperef":          GoTypeRef,
		"gopkgtyperef":       GoPackageTypeRef,
		"goify":              Goify,
		"gonative":           GoNativeType,
		"tabs":               Tabs,
		"add":                func(a, b int) int { return a + b },
		"tempvar":            Tempvar,
		"has":                has,
	}
	if mArrayT, err = template.New("array marshaler").Funcs(fm).Parse(mArrayTmpl); err != nil {
		panic(err)
	}
	if mObjectT, err = template.New("object marshaler").Funcs(fm).Parse(mObjectTmpl); err != nil {
		panic(err)
	}
	if mHashT, err = template.New("hash marshaler").Funcs(fm).Parse(mHashTmpl); err != nil {
		panic(err)
	}
	if mLinkT, err = template.New("links marshaler").Funcs(fm).Parse(mLinkTmpl); err != nil {
		panic(err)
	}
	if mCollectionT, err = template.New("collection marshaler").Funcs(fm).Parse(mCollectionTmpl); err != nil {
		panic(err)
	}
	if mUserImplT, err = template.New("user marshaler").Funcs(fm).Parse(mUserImplTmpl); err != nil {
		panic(err)
	}
	if unmPrimitiveT, err = template.New("primitive unmarshaler").Funcs(fm).Parse(unmPrimitiveTmpl); err != nil {
		panic(err)
	}
	if unmUserPrimitiveT, err = template.New("user primitive unmarshaler").Funcs(fm).Parse(unmUserPrimitiveTmpl); err != nil {
		panic(err)
	}
	if unmArrayT, err = template.New("array unmarshaler").Funcs(fm).Parse(unmArrayTmpl); err != nil {
		panic(err)
	}
	if unmObjectT, err = template.New("object unmarshaler").Funcs(fm).Parse(unmObjectTmpl); err != nil {
		panic(err)
	}
	if unmHashT, err = template.New("hash unmarshaler").Funcs(fm).Parse(unmHashTmpl); err != nil {
		panic(err)
	}
	if unmUserImplT, err = template.New("user type unmarshaler func").Funcs(fm).Parse(unmUserImplTmpl); err != nil {
		panic(err)
	}
}

// TypeMarshaler produces the Go code that initializes the variable named target which is an
// interface{} with the content of the variable named source which contains an instance of the type
// data structure.
// The code takes care of rendering media types according to the view defined on the attribute
// if any. It also renders media type links. Finally it validates the results using any type
// validation that is defined on the type attributes (if the type contains attributes).
// The generated code assumes that there is a variable called "err" of type error that it can use
// to record errors.
// versioned indicates whether the code is being generated from a version package (true) or from the
// default package defaultPkg (false).
func TypeMarshaler(t design.DataType, versioned bool, defaultPkg, context, source, target string) string {
	return typeMarshalerR(t, versioned, defaultPkg, context, source, target, 1)
}

// MediaTypeMarshaler produces the Go code that initializes the variable named target which holds a
// an interface{} with the content of the variable named source which contains an instance of the
// media type data structure. The code runs any validation defined on the media type definition.
// Also view is used to know which fields to copy and which ones to omit and for fields that are
// media types which view to use to render it. The rendering also takes care of following links.
// The generated code assumes that there is a variable called "err" of type error that it can use
// to record errors.
// versioned indicates whether the code is being generated from a version package (true) or from the
// default package defaultPkg (false).
func MediaTypeMarshaler(mt *design.MediaTypeDefinition, versioned bool, defaultPkg, context, source, target, view string) string {
	return mediaTypeMarshalerR(mt, versioned, defaultPkg, source, target, view, 1)
}

// MediaTypeMarshalerImpl returns the Go code for a function that marshals and validates instances
// of the given media type into raw values using the given view to render the attributes.
func MediaTypeMarshalerImpl(mt *design.MediaTypeDefinition, versioned bool, defaultPkg, view string) string {
	var impl string
	if mt.Type.IsArray() {
		impl = collectionMediaTypeMarshalerImpl(mt, versioned, defaultPkg, view)
	} else {
		impl = mediaTypeMarshalerImpl(mt, versioned, defaultPkg, view)
	}
	data := map[string]interface{}{
		"Name": mediaTypeMarshalerFuncName(mt, view),
		"Type": mt,
		"Impl": impl,
		"View": view,
	}
	return RunTemplate(mUserImplT, data)
}

// UserTypeMarshalerImpl returns the Go code for a function that marshals and validates instances
// of the given user type into raw values using the given view to render the attributes.
func UserTypeMarshalerImpl(u *design.UserTypeDefinition, versioned bool, defaultPkg string) string {
	var impl string
	impl = AttributeMarshaler(u.AttributeDefinition, versioned, defaultPkg, "", "source", "target")
	data := map[string]interface{}{
		"Name": userTypeMarshalerFuncName(u),
		"Type": u,
		"Impl": impl,
	}
	return RunTemplate(mUserImplT, data)
}

// AttributeMarshaler produces the Go code that initiliazes the variable named with the value of
// target which holds an interface{} with the content of the variable named with the value of source
// which contains an instance of the attribute type data structure. The attribute view is used to
// render child attributes if there are any. As with TypeMarshaler the code renders media type links
// and runs any validation defined on the type definition.
// versioned indicates whether the code is being generated from a version package (true) or from the
// default package defaultPkg (false).
//
// The generated code assumes that there is a variable called "err" of type error that it can use
// to record errors.
func AttributeMarshaler(att *design.AttributeDefinition, versioned bool, defaultPkg string, context, source, target string) string {
	validation := RecursiveChecker(att, false, false, source, context, 1)
	marshaler := attributeMarshalerR(att, versioned, defaultPkg, context, source, target, 1)
	if validation != "" {
		return fmt.Sprintf(
			"\tif err2 := %s.Validate(); err2 != nil {\n\terr = goa.ReportError(err, err2)\n\treturn\n\t}\n\t%s",
			source,
			marshaler,
		)
	}
	return marshaler
}

// TypeUnmarshaler produces the Go code that initializes a variable of the given type given
// a deserialized (interface{}) value.
// source is the name of the variable that contains the raw interface{} value and target the
// name of the variable to initialize.
// context is used to keep track of recursion to produce helpful error messages in case of type
// mismatch or validation error.
// The generated code assumes that there is a variable called "err" of type error that it can use
// to record errors.
// versioned indicates whether the code is being generated from a version package (true) or from the
// default package defaultPkg (false).
func TypeUnmarshaler(t design.DataType, versioned bool, defaultPkg, context, source, target string) string {
	return typeUnmarshalerR(t, nil, nil, versioned, defaultPkg, context, source, target, 1)
}

// UserTypeUnmarshalerImpl returns the code implementing the user type unmarshaler function.
func UserTypeUnmarshalerImpl(u *design.UserTypeDefinition, versioned bool, defaultPkg, context string) string {
	validation := RecursiveChecker(u.AttributeDefinition, false, false, "source", context, 1)
	var impl string
	switch {
	case u.IsObject():
		impl = objectUnmarshalerR(u, u.AllRequired(), u.AllNonZero(), versioned, defaultPkg, context, "source", "target", 1)
	case u.IsArray():
		impl = arrayUnmarshalerR(u.ToArray(), versioned, defaultPkg, context, "source", "target", 1)
	case u.IsHash():
		impl = hashUnmarshalerR(u.ToHash(), versioned, defaultPkg, context, "source", "target", 1)
	default:
		return "" // No function for primitive types - they just get casted
	}
	data := map[string]interface{}{
		"Name":         userTypeUnmarshalerFuncName(u),
		"Type":         u,
		"Impl":         impl,
		"MustValidate": strings.TrimSpace(validation) != "",
	}
	return RunTemplate(unmUserImplT, data)
}

// GoTypeDef returns the Go code that defines a Go type which matches the data structure
// definition (the part that comes after `type foo`).
// versioned indicates whether the type is being referenced from a version package (true) or the
// default package (false).
// tabs is the number of tab character(s) used to tabulate the definition however the first
// line is never indented.
// jsonTags controls whether to produce json tags.
func GoTypeDef(ds design.DataStructure, versioned bool, defPkg string, tabs int, jsonTags bool) string {
	var buffer bytes.Buffer
	def := ds.Definition()
	t := def.Type
	switch actual := t.(type) {
	case design.Primitive:
		return GoTypeName(t, nil, tabs)
	case *design.Array:
		d := GoTypeDef(actual.ElemType, versioned, defPkg, tabs, jsonTags)
		if actual.ElemType.Type.IsObject() {
			d = "*" + d
		}
		return "[]" + d
	case *design.Hash:
		keyDef := GoTypeDef(actual.KeyType, versioned, defPkg, tabs, jsonTags)
		if actual.KeyType.Type.IsObject() {
			keyDef = "*" + keyDef
		}
		elemDef := GoTypeDef(actual.ElemType, versioned, defPkg, tabs, jsonTags)
		if actual.ElemType.Type.IsObject() {
			elemDef = "*" + elemDef
		}
		return fmt.Sprintf("map[%s]%s", keyDef, elemDef)
	case design.Object:
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
			field := actual[name]
			typedef := GoTypeDef(field, versioned, defPkg, tabs+1, jsonTags)
			if field.Type.IsObject() || def.IsPrimitivePointer(name) {
				typedef = "*" + typedef
			}
			fname := Goify(name, true)
			var tags string
			if jsonTags {
				var omit string
				if !def.IsRequired(name) {
					omit = ",omitempty"
				}
				tags = fmt.Sprintf(" `json:\"%s%s\"`", name, omit)
			}
			desc := actual[name].Description
			if desc != "" {
				desc = fmt.Sprintf("// %s\n", desc)
			}
			buffer.WriteString(fmt.Sprintf("%s%s %s%s\n", desc, fname, typedef, tags))
		}
		WriteTabs(&buffer, tabs)
		buffer.WriteString("}")
		return buffer.String()
	case *design.UserTypeDefinition:
		return GoPackageTypeName(actual, actual.AllRequired(), versioned, defPkg, tabs)
	case *design.MediaTypeDefinition:
		return GoPackageTypeName(actual, actual.AllRequired(), versioned, defPkg, tabs)
	default:
		panic("goa bug: unknown data structure type")
	}
	return GoTypeDef(ds, versioned, defPkg, tabs, jsonTags)
}

// GoTypeRef returns the Go code that refers to the Go type which matches the given data type
// (the part that comes after `var foo`)
// required only applies when referring to a user type that is an object defined inline. In this
// case the type (Object) does not carry the required field information defined in the parent
// (anonymous) attribute.
// tabs is used to properly tabulate the object struct fields and only applies to this case.
// This function assumes the type is in the same package as the code accessing it.
func GoTypeRef(t design.DataType, required []string, tabs int) string {
	return GoPackageTypeRef(t, required, false, "", tabs)
}

// GoPackageTypeRef returns the Go code that refers to the Go type which matches the given data type.
// versioned indicates whether the type is being referenced from a version package (true) or the
// default package defPkg (false).
// required only applies when referring to a user type that is an object defined inline. In this
// case the type (Object) does not carry the required field information defined in the parent
// (anonymous) attribute.
// tabs is used to properly tabulate the object struct fields and only applies to this case.
func GoPackageTypeRef(t design.DataType, required []string, versioned bool, defPkg string, tabs int) string {
	switch t.(type) {
	case *design.UserTypeDefinition, *design.MediaTypeDefinition:
		var prefix string
		if t.IsObject() {
			prefix = "*"
		}
		return prefix + GoPackageTypeName(t, required, versioned, defPkg, tabs)
	case design.Object:
		return "*" + GoPackageTypeName(t, required, versioned, defPkg, tabs)
	default:
		return GoPackageTypeName(t, required, versioned, defPkg, tabs)
	}
}

// GoTypeName returns the Go type name for a data type.
// tabs is used to properly tabulate the object struct fields and only applies to this case.
// This function assumes the type is in the same package as the code accessing it.
// required only applies when referring to a user type that is an object defined inline. In this
// case the type (Object) does not carry the required field information defined in the parent
// (anonymous) attribute.
func GoTypeName(t design.DataType, required []string, tabs int) string {
	return GoPackageTypeName(t, required, false, "", tabs)
}

// GoPackageTypeName returns the Go type name for a data type.
// versioned indicates whether the type is being referenced from a version package (true) or the
// default package defPkg (false).
// required only applies when referring to a user type that is an object defined inline. In this
// case the type (Object) does not carry the required field information defined in the parent
// (anonymous) attribute.
// tabs is used to properly tabulate the object struct fields and only applies to this case.
func GoPackageTypeName(t design.DataType, required []string, versioned bool, defPkg string, tabs int) string {
	switch actual := t.(type) {
	case design.Primitive:
		return GoNativeType(t)
	case *design.Array:
		return "[]" + GoPackageTypeRef(actual.ElemType.Type, actual.ElemType.AllRequired(), versioned, defPkg, tabs+1)
	case design.Object:
		att := &design.AttributeDefinition{Type: actual}
		if len(required) > 0 {
			requiredVal := &design.RequiredValidationDefinition{Names: required}
			att.Validations = append(att.Validations, requiredVal)
		}
		return GoTypeDef(att, versioned, defPkg, tabs, false)
	case *design.Hash:
		return fmt.Sprintf(
			"map[%s]%s",
			GoPackageTypeRef(actual.KeyType.Type, actual.KeyType.AllRequired(), versioned, defPkg, tabs+1),
			GoPackageTypeRef(actual.ElemType.Type, actual.ElemType.AllRequired(), versioned, defPkg, tabs+1),
		)
	case *design.UserTypeDefinition:
		pkgPrefix := PackagePrefix(actual, versioned, defPkg)
		return pkgPrefix + Goify(actual.TypeName, true)
	case *design.MediaTypeDefinition:
		pkgPrefix := PackagePrefix(actual.UserTypeDefinition, versioned, defPkg)
		return pkgPrefix + Goify(actual.TypeName, true)
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

// Goify makes a valid Go identifier out of any string.
// It does that by removing any non letter and non digit character and by making sure the first
// character is a letter or "_".
// Goify produces a "CamelCase" version of the string, if firstUpper is true the first character
// of the identifier is uppercase otherwise it's lowercase.
func Goify(str string, firstUpper bool) string {
	if str == "ok" && firstUpper {
		return "OK"
	} else if str == "id" && firstUpper {
		return "ID"
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

// Tempvar generates a unique variable name.
func Tempvar() string {
	TempCount++
	return fmt.Sprintf("tmp%d", TempCount)
}

// PackagePrefix returns the package prefix to use to access ut from ver given it lives in the
// package pkg.
func PackagePrefix(ut *design.UserTypeDefinition, versioned bool, pkg string) string {
	if !versioned {
		// If the version is the default version then the user type is in the same package
		// (otherwise the DSL would not be valid).
		return ""
	}
	if len(ut.APIVersions) == 0 {
		// If the type is not versioned but we are accessing it from the non-default version
		// then we need to qualify it with the default version package.
		return pkg + "."
	}
	// If the type is versioned then we must be accessing it from the current version
	// (unversioned definitions cannot use versioned definitions)
	return ""
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

// attributeMarshalerR is the recursive implementation of AttributeMarshaler.
func attributeMarshalerR(att *design.AttributeDefinition, versioned bool, defaultPkg, context, source, target string, depth int) string {
	switch actual := att.Type.(type) {
	case *design.MediaTypeDefinition:
		return mediaTypeMarshalerR(actual, versioned, defaultPkg, source, target, att.View, depth)
	case design.Object:
		return objectMarshalerR(actual, versioned, defaultPkg, context, source, target, depth)
	default:
		return typeMarshalerR(att.Type, versioned, defaultPkg, context, source, target, depth)
	}
}

// ArrayMarshaler produces the Go code that marshals an array for rendering.
// source is the name of the variable that contains the array value and target the name of the
// variable to initialize.
// The generated code assumes that there is a variable called "err" of type error that it can use
// to record errors.
func arrayMarshalerR(a *design.Array, versioned bool, defaultPkg, context, source, target string, depth int) string {
	data := map[string]interface{}{
		"source":     source,
		"target":     target,
		"elemType":   a.ElemType,
		"context":    context,
		"depth":      depth,
		"versioned":  versioned,
		"defaultPkg": defaultPkg,
	}
	return RunTemplate(mArrayT, data)
}

// HashMarshaler produces the Go code that initializes the variable named target which holds a
// map of interface{} to interface{} with the content of the variable named source which contains an
// instance of the hash map. The code runs any validation defined on the hash map key and value
// attribute definitions.
// The generated code assumes that there is a variable called "err" of type error that it can use
// to record errors.
func hashMarshalerR(h *design.Hash, versioned bool, defaultPkg, context, source, target string, depth int) string {
	data := map[string]interface{}{
		"type":       h,
		"context":    context,
		"source":     source,
		"target":     target,
		"depth":      depth,
		"versioned":  versioned,
		"defaultPkg": defaultPkg,
	}
	return RunTemplate(mHashT, data)
}

// ObjectMarshaler produces the Go code that initializes the variable named target which holds a
// map of string to interface{} with the content of the variable named source which contains an
// instance of the object data structure. The code runs any validation defined on the object
// attribute definitions.
// The generated code assumes that there is a variable called "err" of type error that it can use
// to record errors.
func objectMarshalerR(o design.DataType, versioned bool,
	defaultPkg, context, source, target string, depth int) string {
	data := map[string]interface{}{
		"versioned":  versioned,
		"defaultPkg": defaultPkg,
		"type":       o,
		"context":    context,
		"source":     source,
		"target":     target,
		"depth":      depth,
	}
	return RunTemplate(mObjectT, data)
}

// typeMarshalerR implements the recursive function that marshals an instance of a type into a raw
// value.
func typeMarshalerR(t design.DataType, versioned bool, defaultPkg, context, source, target string, depth int) string {
	switch actual := t.(type) {
	case design.Primitive:
		return fmt.Sprintf("%s%s = %s", Tabs(depth), target, source)
	case *design.Array:
		return arrayMarshalerR(actual, versioned, defaultPkg, context, source, target, depth)
	case *design.Hash:
		return hashMarshalerR(actual, versioned, defaultPkg, context, source, target, depth)
	case design.Object:
		return objectMarshalerR(actual.ToObject(), versioned, defaultPkg, context, source, target, depth)
	case *design.UserTypeDefinition:
		if _, ok := actual.Type.(design.Primitive); ok {
			return fmt.Sprintf("%s%s = %s(%s)", Tabs(depth), target, actual.Name(), source)
		}
		return fmt.Sprintf(
			"%s%s, err = %s(%s, err)",
			Tabs(depth),
			target,
			userTypeMarshalerFuncName(actual),
			source,
		)
	default:
		// this should never get called with a MediaType, MediaTypeMarshaler should be
		// called instead so the view is properly taken into account.
		panic(actual)
	}
}

// mediaTypeMarshalerR produces Go code that calls the media type marshaler function.
func mediaTypeMarshalerR(mt *design.MediaTypeDefinition, versioned bool, defaultPkg, source, target, view string, depth int) string {
	prefix := PackagePrefix(mt.UserTypeDefinition, versioned, defaultPkg)
	return fmt.Sprintf(
		`%s%s, err = %s%s(%s, err)`,
		Tabs(depth),
		target,
		prefix,
		mediaTypeMarshalerFuncName(mt, view),
		source,
	)
}

// mediaTypeMarshalerImpl implements the recursive function that marshals an instance of a media
// type into a raw value.
func mediaTypeMarshalerImpl(mt *design.MediaTypeDefinition, versioned bool, defaultPkg, view string) string {
	rendered := mt.AttributeDefinition
	if view == "" {
		view = "default"
	}
	renderLinks := false
	if v, ok := mt.Views[view]; ok {
		var vals []design.ValidationDefinition
		if viewObj := v.Type.ToObject(); viewObj != nil {
			attNames := make(map[string]bool)
			for n := range viewObj {
				if n == "links" {
					renderLinks = true
				} else {
					attNames[n] = true
				}
			}
			vals = make([]design.ValidationDefinition, len(mt.Validations))
			for i, va := range mt.Validations {
				if r, ok := va.(*design.RequiredValidationDefinition); ok {
					var required []string
					for _, n := range r.Names {
						if attNames[n] {
							required = append(required, n)
						}
					}
					vals[i] = &design.RequiredValidationDefinition{Names: required}
				} else {
					vals[i] = va
				}
			}
		}
		rendered = &design.AttributeDefinition{
			Type:        design.DataType(v.Type.ToObject()),
			Validations: vals,
		}
	}
	var linkMarshaler string
	if renderLinks && len(mt.Links) > 0 {
		data := map[string]interface{}{
			"links":      mt.Links,
			"versioned":  versioned,
			"defaultPkg": defaultPkg,
			"context":    "",
			"source":     "source",
			"target":     "target",
			"view":       view,
			"depth":      1,
		}
		linkMarshaler = "\n" + RunTemplate(mLinkT, data)
	}
	final := rendered.Dup()
	o := rendered.Type.ToObject()
	mtObj := mt.Type.ToObject()
	newObj := make(design.Object)
	for n := range o {
		if n != "links" {
			for an, at := range mtObj {
				if an == n {
					newObj[n] = at
					break
				}
			}
		}
	}
	final.Type = newObj
	return AttributeMarshaler(final, versioned, defaultPkg, "", "source", "target") + linkMarshaler
}

func collectionMediaTypeMarshalerImpl(mt *design.MediaTypeDefinition, versioned bool, defaultPkg, view string) string {
	data := map[string]interface{}{
		"versioned":     versioned,
		"defaultPkg":    defaultPkg,
		"context":       "",
		"source":        "source",
		"target":        "target",
		"view":          view,
		"depth":         1,
		"elemMediaType": mt.Type.(*design.Array).ElemType.Type,
	}
	return RunTemplate(mCollectionT, data)
}

// userTypeMarshalerFuncName returns the name for the given media type marshaler function.
func userTypeMarshalerFuncName(u *design.UserTypeDefinition) string {
	return fmt.Sprintf("Marshal%s", GoTypeName(u, u.AllRequired(), 0))
}

// mediaTypeMarshalerFuncName returns the name for the given user type marshaler function.
func mediaTypeMarshalerFuncName(mt *design.MediaTypeDefinition, view string) string {
	name := userTypeMarshalerFuncName(mt.UserTypeDefinition)
	if view == "" || view == "default" {
		return name
	}
	return fmt.Sprintf("%s%s", name, strings.Title(view))
}

// userTypeUnmarshalerFuncName returns the name for the given user type unmarshaler function.
func userTypeUnmarshalerFuncName(u *design.UserTypeDefinition) string {
	return fmt.Sprintf("Unmarshal%s", GoTypeName(u, u.AllRequired(), 0))
}

func typeUnmarshalerR(t design.DataType, required, nonzero []string, versioned bool, defaultPkg, context, source, target string, depth int) string {
	switch actual := t.(type) {
	case design.Primitive:
		return primitiveUnmarshalerR(actual, context, source, target, depth)
	case *design.Array:
		return arrayUnmarshalerR(actual, versioned, defaultPkg, context, source, target, depth)
	case *design.Hash:
		return hashUnmarshalerR(actual, versioned, defaultPkg, context, source, target, depth)
	case design.Object:
		return objectUnmarshalerR(actual, required, nonzero, versioned, defaultPkg, context, source, target, depth)
	case *design.UserTypeDefinition:
		if _, ok := t.(design.Primitive); ok {
			return userPrimitiveUnmarshalerR(actual, context, source, target, depth)
		}
		prefix := PackagePrefix(actual, versioned, defaultPkg)
		return fmt.Sprintf(
			`%s%s, err = %s%s(%s, err)`,
			Tabs(depth),
			target,
			prefix,
			userTypeUnmarshalerFuncName(actual),
			source,
		)
	case *design.MediaTypeDefinition:
		return typeUnmarshalerR(actual.UserTypeDefinition, required, nonzero, versioned, defaultPkg, context, source, target, depth)
	default:
		panic(actual)
	}
}

func userPrimitiveUnmarshalerR(u *design.UserTypeDefinition, context, source, target string, depth int) string {
	data := map[string]interface{}{
		"source":  source,
		"target":  target,
		"type":    u,
		"context": context,
		"depth":   depth,
	}
	return RunTemplate(unmUserPrimitiveT, data)
}

func attributeUnmarshalerR(att *design.AttributeDefinition, required, nonzero []string, versioned bool, defaultPkg, context, source, target string, depth int) string {
	return typeUnmarshalerR(att.Type, required, nonzero, versioned, defaultPkg, context, source, target, depth)
}

// PrimitiveUnmarshaler produces the Go code that initializes a primitive type from its deserialized
// representation.
// source is the name of the variable that contains the raw interface{} value and target the
// name of the variable to initialize.
// The generated code assumes that there is a variable called "err" of type error that it can use
// to record errors.
func primitiveUnmarshalerR(p design.Primitive, context, source, target string, depth int) string {
	data := map[string]interface{}{
		"source":  source,
		"target":  target,
		"type":    p,
		"context": context,
		"depth":   depth,
	}
	return RunTemplate(unmPrimitiveT, data)
}

// ArrayUnmarshaler produces the Go code that initializes an array from its deserialized epresentation.
// source is the name of the variable that contains the raw interface{} value and target the
// name of the variable to initialize.
// The generated code assumes that there is a variable called "err" of type error that it can use
// to record errors.
func arrayUnmarshalerR(a *design.Array, versioned bool, defaultPkg, context, source, target string, depth int) string {
	data := map[string]interface{}{
		"versioned":  versioned,
		"defaultPkg": defaultPkg,
		"source":     source,
		"target":     target,
		"elemType":   a.ElemType,
		"context":    context,
		"depth":      depth,
	}
	return RunTemplate(unmArrayT, data)
}

// HashUnmarshaler produces the Go code that initializes a hash map from its deserialized
// representation.
// source is the name of the variable that contains the raw map[string]interface{} value and target
// the name of the variable to initialize.
// The generated code assumes that there is a variable called "err" of type error that it can use
// to record errors.
func hashUnmarshalerR(h *design.Hash, versioned bool, defaultPkg, context, source, target string, depth int) string {
	data := map[string]interface{}{
		"type":       h,
		"versioned":  versioned,
		"defaultPkg": defaultPkg,
		"context":    context,
		"source":     source,
		"target":     target,
		"depth":      depth,
	}
	return RunTemplate(unmHashT, data)
}

// ObjectUnmarshaler produces the Go code that initializes an object type from its deserialized
// representation.
// source is the name of the variable that contains the raw interface{} value and target the
// name of the variable to initialize.
// The generated code assumes that there is a variable called "err" of type error that it can use
// to record errors.
func objectUnmarshalerR(o design.DataType, required, nonzero []string, versioned bool, defaultPkg, context, source, target string, depth int) string {
	data := map[string]interface{}{
		"type":       o,
		"versioned":  versioned,
		"defaultPkg": defaultPkg,
		"required":   required,
		"nonzero":    nonzero,
		"context":    context,
		"source":     source,
		"target":     target,
		"depth":      depth,
	}
	return RunTemplate(unmObjectT, data)
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

const (
	mArrayTmpl = `{{$tmp := tempvar}}{{tabs .depth}}{{$tmp}} := make([]{{gonative .elemType.Type}}, len({{.source}}))
{{$tmpIndex := tempvar}}{{$tmpElement := tempvar}}{{tabs .depth}}for {{$tmpIndex}}, {{$tmpElement}} := range {{.source}} {
{{marshalAttribute .elemType .versioned .defaultPkg (printf "%s[*]" .context) $tmpElement (printf "%s[%s]" $tmp $tmpIndex) (add .depth 1)}}
{{tabs .depth}}}
{{tabs .depth}}{{.target}} = {{$tmp}}`

	mObjectTmpl = `{{$ctx := .}}{{$depth := .depth}}{{$tmp := tempvar}}{{tabs .depth}}{{$tmp}} := map[string]interface{}{
{{range $n, $at := .type}}{{if $at.Type.IsPrimitive}}{{/*
	## Define basic types inline in the struct definition
*/}}{{tabs $depth}}	"{{$n}}": {{$ctx.source}}.{{goify $n true}},
{{end}}{{end}}{{/* range */}}{{tabs $depth}}}
{{range $n, $at := .type}}{{if (not $at.Type.IsPrimitive)}}{{/*
	## Handle objects, user types and media types (they need an extra temporary variable)
*/}}{{tabs $depth}}if {{$ctx.source}}.{{goify $n true}} != nil {
{{marshalAttribute $at $ctx.versioned $ctx.defaultPkg (printf "%s.%s" $ctx.context (goify $n true)) (printf "%s.%s" $ctx.source (goify $n true)) (printf "%s[\"%s\"]" $tmp $n) (add $depth 1)}}
{{tabs $depth}}}
{{end}}{{end}}{{/*
	## Done
*/}}{{tabs $depth}}{{.target}} = {{$tmp}}`

	mObjectTmplback = `{{$ctx := .}}{{range $r := .required}}{{$at := index $ctx.type $r}}{{$required := goify $r true}}{{/*
*/}}{{if eq $at.Type.Kind 4}}{{tabs $ctx.depth}}if {{$ctx.source}}.{{$required}} == "" {
{{else if (not $at.Type.IsPrimitive)}}{{tabs $ctx.depth}}if {{$ctx.source}}.{{$required}} == nil {
{{end}}{{if or (eq $at.Type.Kind 4) (not $at.Type.IsPrimitive)}}{{tabs $ctx.depth}}	err = goa.MissingAttributeError(` + "`" + `{{$ctx.context}}` + "`" + `, "{{$r}}", err)
{{tabs $ctx.depth}}}
{{end}}{{end}}{{/* range */}}{{/*

*/}}{{if $ctx.required}}{{tabs .depth}}{{if $ctx.checkRequiredError}}if err == nil {
{{end}}{{end}}{{$depth := add .depth (or (and $ctx.required 1) 0)}}{{range $n, $at := .type}}{{/*
*/}}{{$validation := index $ctx.validations $n}}{{if $validation}}{{$validation}}
{{end}}{{end}}{{/* range */}}{{if $ctx.checkValidationError}}{{$depth := add $depth 1}}{{tabs $depth}}if err == nil {
{{end}}{{$tmp := tempvar}}{{tabs $depth}}{{$tmp}} := map[string]interface{}{
{{range $n, $at := .type}}{{if $at.Type.IsPrimitive}}{{/*
	## Define basic types inline in the struct definition
*/}}{{tabs $depth}}	"{{$n}}": {{$ctx.source}}.{{goify $n true}},
{{end}}{{end}}{{/* range */}}{{tabs $depth}}}
{{range $n, $at := .type}}{{if (not $at.Type.IsPrimitive)}}{{/*
	## Handle objects, user types and media types (they need an extra temporary variable)
*/}}{{tabs $depth}}if {{$ctx.source}}.{{goify $n true}} != nil {
{{marshalAttribute $at $ctx.versioned $ctx.defaultPkg (printf "%s.%s" $ctx.context (goify $n true)) (printf "%s.%s" $ctx.source (goify $n true)) (printf "%s[\"%s\"]" $tmp $n) (add $depth 1)}}
{{tabs $depth}}}
{{end}}{{end}}{{/*
	## Done
*/}}{{tabs $depth}}{{.target}} = {{$tmp}}{{if $ctx.checkValidationError}}
{{tabs .depth}}	}{{end}}{{if $ctx.required}}{{if $ctx.checkRequiredError}}
{{tabs .depth}}}{{end}}{{end}}`

	mHashTmpl = `{{tabs .depth}}{{$tmp := tempvar}}{{$tmp}} := make(map[{{gonative .type.ToHash.KeyType.Type}}]{{gonative .type.ToHash.ElemType.Type}}, len({{.source}}))
{{tabs .depth}}for k, v := range {{.source}} {
{{tabs .depth}}	var mk {{gonative .type.ToHash.KeyType.Type}}
{{marshalAttribute .type.ToHash.KeyType .versioned .defaultPkg (printf "%s.keys[*]" .context) "k" "mk" (add .depth 1)}}
{{tabs .depth}}	var mv {{gonative .type.ToHash.ElemType.Type}}
{{marshalAttribute .type.ToHash.ElemType .versioned .defaultPkg (printf "%s.values[*]" .context) "v" "mv" (add .depth 1)}}
{{tabs .depth}}	{{$tmp}}[mk] = mv
{{tabs .depth}}}
{{tabs .depth}}{{.target}} = {{$tmp}}`

	mCollectionTmpl = `{{tabs .depth}}{{.target}} = make([]{{gonative .elemMediaType}}, len({{.source}}))
{{tabs .depth}}for i, res := range {{.source}} {
{{marshalMediaType .elemMediaType .versioned .defaultPkg "res" (printf "%s[i]" .target) .view (add .depth 1)}}
{{tabs .depth}}}`

	mLinkTmpl = `{{if .links}}{{$ctx := .}}{{tabs .depth}}if err == nil {
{{tabs .depth}}	links := make(map[string]interface{})
{{range $n, $l := .links}}{{marshalMediaType $l.MediaType $ctx.versioned $ctx.defaultPkg (printf "%s.%s" $ctx.source (goify $l.Name true)) (printf "links[\"%s\"]" $n) $l.View $ctx.depth}}
{{end}}{{tabs .depth}}	{{.target}}["links"] = links
}{{end}}`

	mUserImplTmpl = `// {{.Name}} validates and renders an instance of {{gotypename .Type nil 0}} into a interface{}{{if .View}}
// using view "{{.View}}".{{end}}
func {{.Name}}(source {{gotyperef .Type .Type.AllRequired 0}}, inErr error) (target {{gonative .Type}}, err error) {
	err = inErr
{{.Impl}}
	return
}`

	unmUserPrimitiveTmpl = `{{tabs .depth}}if val, ok := {{.source}}.({{gonative .type}}); ok {
{{tabs .depth}}	{{.target}} = {{gopkgtyperef .type nil .versioned .defaultPkg 0}}(val)
{{tabs .depth}}} else {
{{tabs .depth}}	err = goa.InvalidAttributeTypeError(` + "`" + `{{.context}}` + "`" + `, {{.source}}, "{{gonative .type}}", err)
{{tabs .depth}}}`

	unmPrimitiveTmpl = `{{if eq .type.Kind 2}}{{tabs .depth}}if f, ok := {{.source}}.(float64); ok {
{{tabs .depth}}	{{.target}} = int(f)
{{else if or (eq .type.Kind 5) (eq .type.Kind 6)}}{{tabs .depth}}{{.target}} = {{.source}}{{else}}{{tabs .depth}}if val, ok := {{.source}}.({{gotyperef .type nil (add .depth 1)}}); ok {
{{tabs .depth}}	{{.target}} = val
{{end}}{{if not (or (eq .type.Kind 5) (eq .type.Kind 6))}}{{tabs .depth}}} else {
{{tabs .depth}}	err = goa.InvalidAttributeTypeError(` + "`" + `{{.context}}` + "`" + `, {{.source}}, "{{gotyperef .type nil (add .depth 1)}}", err)
{{tabs .depth}}}{{end}}`

	unmArrayTmpl = `{{tabs .depth}}if val, ok := {{.source}}.([]interface{}); ok {
{{tabs .depth}}	{{.target}} = make([]{{gopkgtyperef .elemType.Type .elemType.AllRequired .versioned .defaultPkg (add .depth 2)}}, len(val))
{{tabs .depth}}	{{$tmp := tempvar}}for {{$tmp}}, v := range val {
{{unmarshalAttribute .elemType nil nil .versioned .defaultPkg (printf "%s[*]" .context) "v" (printf "%s[%s]" .target $tmp) (add .depth 2)}}{{$ctx := .}}
{{tabs .depth}}	}
{{tabs .depth}}} else {
{{tabs .depth}}	err = goa.InvalidAttributeTypeError(` + "`" + `{{.context}}` + "`" + `, {{.source}}, "array", err)
{{tabs .depth}}}`

	unmObjectTmpl = `{{tabs .depth}}{{$ctx := .}}if val, ok := {{.source}}.(map[string]interface{}); ok {
{{tabs .depth}}{{$context := .context}}{{$depth := .depth}}{{$target := .target}}{{$required := .required}}{{/*
*/}}	{{$target}} = new({{gotypename .type .required (add .depth 1)}})
{{range $name, $att := .type.ToObject}}{{$pointer := and (not (has $ctx.required $name)) (not (has $ctx.nonzero $name)) $att.Type.IsPrimitive}}{{/*
*/}}{{tabs $depth}}	if v, ok := val["{{$name}}"]; ok {
{{tabs $depth}}		{{$temp := tempvar}}var {{$temp}} {{gopkgtyperef $att.Type $att.AllRequired $ctx.versioned $ctx.defaultPkg (add $depth 2)}}
{{unmarshalAttribute $att $att.AllRequired $att.AllNonZero $ctx.versioned $ctx.defaultPkg (printf "%s.%s" $context (goify $name true)) "v" $temp (add $depth 2)}}
{{tabs $depth}}		{{printf "%s.%s" $target (goify $name true)}} = {{if $pointer}}&{{end}}{{$temp}}
{{tabs $depth}}	}{{if (has $required $name)}} else {
{{tabs $depth}}		err = goa.MissingAttributeError(` + "`" + `{{$context}}` + "`" + `, "{{$name}}", err)
{{tabs $depth}}	}{{end}}
{{end}}{{tabs $depth}}} else {
{{tabs .depth}}	err = goa.InvalidAttributeTypeError(` + "`" + `{{.context}}` + "`" + `, {{.source}}, "dictionary", err)
{{tabs .depth}}}`

	unmHashTmpl = `{{tabs .depth}}if val, ok := {{.source}}.(map[string]interface{}); ok {
{{if and (eq .type.KeyType.Type.Kind 4) (eq .type.ElemType.Type.Kind 5)}}{{tabs .depth}}	{{.target}} = val
{{else}}{{tabs .depth}}	{{$tmp := tempvar}}{{$tmp}} := make(map[{{gotypename .type.KeyType.Type .type.KeyType.AllRequired (add .depth 1)}}]{{gotypename .type.ElemType.Type .type.ElemType.AllRequired (add .depth 1)}})
{{tabs .depth}}	for k, v := range val {
{{$k := tempvar}}{{if not (eq .type.KeyType.Type.Kind 4)}}{{tabs .depth}}		{{$ki := tempvar}}var {{$ki}} interface{}
{{tabs .depth}}		err = json.Unmarshal([]byte(k), &{{$ki}})
{{tabs .depth}}		if err != nil {
{{tabs .depth}}			return
{{tabs .depth}}		}
{{tabs .depth}}		var {{$k}} {{gotypename .type.KeyType.Type .type.KeyType.AllRequired (add .depth 2)}}
{{tabs .depth}}		{{unmarshalAttribute .type.KeyType .type.KeyType.AllRequired .type.KeyType.AllNonZero .versioned .defaultPkg (printf "%s.keys[*]" .context) $ki $k (add .depth 2)}}
{{end}}{{$v := tempvar}}{{if not (eq .type.ElemType.Type.Kind 5)}}{{tabs .depth}}		var {{$v}} {{gotypename .type.ElemType.Type .type.ElemType.AllRequired (add .depth 2)}}
{{tabs .depth}}		{{unmarshalAttribute .type.ElemType .type.ElemType.AllRequired .type.ElemType.AllNonZero .versioned .defaultPkg (printf "%s.values[*]" .context) "v" $v (add .depth 2)}}
{{end}}{{tabs .depth}}		{{$tmp}}[{{if eq .type.KeyType.Type.Kind 4}}k{{else}}{{$k}}{{end}}] = {{if eq .type.ElemType.Type.Kind 5}}v{{else}}{{$v}}{{end}}
{{tabs .depth}}	}
{{tabs .depth}}	{{.target}} = {{$tmp}}
{{end}}{{tabs .depth}}} else {
{{tabs .depth}}	err = goa.InvalidAttributeTypeError(` + "`" + `{{.context}}` + "`" + `, {{.source}}, "hash", err)
{{tabs .depth}}}`

	unmUserImplTmpl = `// {{.Name}} unmarshals and validates a raw interface{} into an instance of {{gotypename .Type nil 0}}
func {{.Name}}(source interface{}, inErr error) (target {{gotyperef .Type .Type.AllRequired 0}}, err error) {
	err = inErr
{{.Impl}}
{{if .MustValidate}}	if target != nil {
		err = goa.ReportError(err, target.Validate())
	}
{{end}}	return
}`
)
