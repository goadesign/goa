package codegen

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"goa.design/goa.v2/design"
)

var (
	transformArrayT *template.Template
	transformMapT   *template.Template
)

// NOTE: can't initialize inline because https://github.com/golang/go/issues/1817
func init() {
	funcMap := template.FuncMap{"transformAttribute": transformAttribute}
	transformArrayT = template.Must(template.New("transformArray").Funcs(funcMap).Parse(transformArrayTmpl))
	transformMapT = template.Must(template.New("transformMap").Funcs(funcMap).Parse(transformMapTmpl))
}

// GoTypeTransform produces Go code that initializes the data structure defined
// by target from an instance of the data structure described by source. The
// data structures can be objects, arrays or maps. The algorithm matches object
// fields by name and ignores object fields in target that don't have a match in
// source. The matching and generated code leverage mapped attributes so that
// attribute names may use the "name:elem" syntax to define the name of the
// design attribute and the name of the corresponding generated Go struct field.
// The function returns an error if target is not compatible with source
// (different type, fields of different type etc).
//
// sourceVar and targetVar contain the name of the variables that hold the
// source and target data structures respectively.
//
// targetPkg contain the name of the Go package that defines the target type in
// case it's not the same package as where the generated code lives.
//
// fromPtrs indicates whether the source data structure uses pointers
// to store all attributes even required ones (i.e. unmarshaled request body)
//
// initDefaults indicates whether fields in the target should be initialized with
// the attribute default values when the source field is a pointer and is nil.
//
// scope is used to compute the name of the user types when initializing fields
// that use them.
func GoTypeTransform(source, target design.DataType, sourceVar, targetVar, targetPkg string,
	fromPtrs, toPtrs, initDefaults bool, scope *NameScope) (string, error) {

	var (
		satt = &design.AttributeExpr{Type: source}
		tatt = &design.AttributeExpr{Type: target}
	)

	code, err := transformAttribute(satt, tatt, sourceVar,
		targetVar, targetPkg, fromPtrs, toPtrs, initDefaults, true, scope)

	if err != nil {
		return "", err
	}

	return code, nil
}

func transformAttribute(source, target *design.AttributeExpr,
	sctx, tctx, targetPkg string, fromPtrs, toPtrs, def, newVar bool, scope *NameScope) (string, error) {

	if err := isCompatible(source.Type, target.Type, sctx, tctx); err != nil {
		return "", err
	}
	var (
		code string
		err  error
	)
	switch {
	case design.IsArray(source.Type):
		code, err = transformArray(design.AsArray(source.Type), design.AsArray(target.Type), sctx, tctx, targetPkg, fromPtrs, toPtrs, def, scope)
	case design.IsMap(source.Type):
		code, err = transformMap(design.AsMap(source.Type), design.AsMap(target.Type), sctx, tctx, targetPkg, fromPtrs, toPtrs, def, scope)
	case design.IsObject(source.Type):
		code, err = transformObject(source, target, sctx, tctx, targetPkg, fromPtrs, toPtrs, def, newVar, scope)
	default:
		assign := "="
		if newVar {
			assign = ":="
		}
		code = fmt.Sprintf("%s %s %s\n", tctx, assign, sctx)
	}
	if err != nil {
		return "", err
	}
	return code, nil
}

func transformObject(source, target *design.AttributeExpr, sctx, tctx, targetPkg string, fromPtrs, toPtrs, def, newVar bool, scope *NameScope) (string, error) {
	src := design.NewMappedAttributeExpr(source)
	tgt := design.NewMappedAttributeExpr(target)
	srcObj := design.AsObject(src.Type)
	tgtObj := design.AsObject(tgt.Type)

	// Map source object attribute names to target object attributes
	attributeMap := make(map[string]*design.AttributeExpr)
	for _, nat := range *srcObj {
		if att := tgtObj.Attribute(nat.Name); att != nil {
			attributeMap[nat.Name] = att
		}
	}

	buffer := &bytes.Buffer{}
	var initCode string
	for _, natt := range *srcObj {
		n := natt.Name
		if _, ok := attributeMap[n]; !ok {
			continue
		}
		srcAtt := srcObj.Attribute(n)
		if !design.IsPrimitive(srcAtt.Type) {
			continue
		}
		srcField := sctx + "." + Goify(src.ElemName(n), true)
		deref := ""
		if (fromPtrs || src.IsPrimitivePointer(n, !fromPtrs)) && tgt.IsRequired(n) {
			deref = "*"
		}
		if toPtrs && !src.IsPrimitivePointer(n, true) {
			deref = "&"
		}
		initCode += fmt.Sprintf("\n%s: %s%s,", Goify(tgt.ElemName(n), true), deref, srcField)
	}
	if initCode != "" {
		initCode += "\n"
	}
	assign := "="
	if newVar {
		assign = ":="
	}
	buffer.WriteString(fmt.Sprintf("%s %s &%s{%s}\n", tctx, assign,
		scope.GoFullTypeName(target, targetPkg), initCode))
	for _, natt := range *srcObj {
		n := natt.Name
		att, ok := attributeMap[n]
		if !ok {
			// no match in target object, skip
			continue
		}
		srcAtt := srcObj.Attribute(n)
		srcField := sctx + "." + Goify(src.ElemName(n), true)
		tgtField := tctx + "." + Goify(tgt.ElemName(n), true)
		if err := isCompatible(srcAtt.Type, att.Type, srcField, tgtField); err != nil {
			return "", err
		}

		var (
			code string
			err  error
		)
		switch {
		case design.IsArray(srcAtt.Type):
			code, err = transformArray(design.AsArray(srcAtt.Type), design.AsArray(att.Type), srcField, tgtField, targetPkg, fromPtrs, toPtrs, def, scope)
		case design.IsMap(srcAtt.Type):
			code, err = transformMap(design.AsMap(srcAtt.Type), design.AsMap(att.Type), srcField, tgtField, targetPkg, fromPtrs, toPtrs, def, scope)
		case design.IsObject(srcAtt.Type):
			code, err = transformObject(srcAtt, att, srcField, tgtField, targetPkg, fromPtrs, toPtrs, def, false, scope)
		}
		if !src.IsRequired(n) && (!tgt.IsPrimitivePointer(n, false) || tgt.HasDefaultValue(n) && def) {
			hasTransform := code != ""
			if hasTransform {
				code = fmt.Sprintf("if %s != nil {\n\t%s", srcField, code)
			} else {
				code = fmt.Sprintf("if %s == nil {", srcField)
			}
			if tgt.HasDefaultValue(n) && def {
				if hasTransform {
					code = fmt.Sprintf("%s} else {\n", code)
				}
				if tgt.IsPrimitivePointer(n, false) {
					code = fmt.Sprintf("%s\n\ttmp := %#v\n\t%s = &tmp\n", code, att.DefaultValue, tgtField)
				} else {
					code = fmt.Sprintf("%s\n\t%s = %#v\n", code, tgtField, att.DefaultValue)
				}
			}
			code += "}\n"
		}

		if err != nil {
			return "", err
		}
		buffer.WriteString(code)
	}

	return buffer.String(), nil
}

func transformArray(source, target *design.Array, sctx, tctx, targetPkg string, fromPtrs, toPtrs, def bool, scope *NameScope) (string, error) {
	if err := isCompatible(source.ElemType.Type, target.ElemType.Type, sctx+"[0]", tctx+"[0]"); err != nil {
		return "", err
	}
	data := map[string]interface{}{
		"Source":       sctx,
		"Target":       tctx,
		"NewVar":       !strings.Contains(sctx, "."),
		"ElemTypeRef":  scope.GoFullTypeRef(target.ElemType, targetPkg),
		"SourceElem":   source.ElemType,
		"TargetElem":   target.ElemType,
		"TargetPkg":    targetPkg,
		"FromPtrs":     fromPtrs,
		"ToPtrs":       toPtrs,
		"InitDefaults": def,
		"Scope":        scope,
	}
	var buf bytes.Buffer
	if err := transformArrayT.Execute(&buf, data); err != nil {
		panic(err) // bug
	}
	code := buf.String()

	return code, nil
}

func transformMap(source, target *design.Map, sctx, tctx, targetPkg string, fromPtrs, toPtrs, def bool, scope *NameScope) (string, error) {
	if err := isCompatible(source.KeyType.Type, target.KeyType.Type, sctx+".key", tctx+".key"); err != nil {
		return "", err
	}
	if err := isCompatible(source.ElemType.Type, target.ElemType.Type, sctx+"[*]", tctx+"[*]"); err != nil {
		return "", err
	}
	data := map[string]interface{}{
		"Source":       sctx,
		"Target":       tctx,
		"NewVar":       !strings.Contains(sctx, "."),
		"KeyTypeRef":   scope.GoFullTypeRef(target.KeyType, targetPkg),
		"ElemTypeRef":  scope.GoFullTypeRef(target.ElemType, targetPkg),
		"SourceKey":    source.KeyType,
		"TargetKey":    target.KeyType,
		"SourceElem":   source.ElemType,
		"TargetElem":   target.ElemType,
		"TargetPkg":    targetPkg,
		"FromPtrs":     fromPtrs,
		"ToPtrs":       toPtrs,
		"InitDefaults": def,
		"Scope":        scope,
	}
	var buf bytes.Buffer
	if err := transformMapT.Execute(&buf, data); err != nil {
		panic(err) // bug
	}
	return buf.String(), nil
}

// isCompatible returns an error if a and b are not both objects, both arrays,
// both maps or both the same primitive type. actx and bctx are used to build
// the error message if any.
func isCompatible(a, b design.DataType, actx, bctx string) error {
	switch {
	case design.IsObject(a):
		if !design.IsObject(b) {
			return fmt.Errorf("%s is an object but %s type is %s", actx, bctx, b.Name())
		}
	case design.IsArray(a):
		if !design.IsArray(b) {
			return fmt.Errorf("%s is an array but %s type is %s", actx, bctx, b.Name())
		}
	case design.IsMap(a):
		if !design.IsMap(b) {
			return fmt.Errorf("%s is a hash but %s type is %s", actx, bctx, b.Name())
		}
	default:
		if a.Kind() != b.Kind() {
			return fmt.Errorf("%s is a %s but %s type is %s", actx, a.Name(), bctx, b.Name())
		}
	}

	return nil
}

const transformArrayTmpl = `{{ .Target}} {{ if .NewVar }}:{{ end }}= make([]{{ .ElemTypeRef }}, len({{ .Source }}))
for i, val := range {{ .Source }} {
	{{ transformAttribute .SourceElem .TargetElem "val" (printf "%s[i]" .Target) .TargetPkg .FromPtrs .ToPtrs .InitDefaults false .Scope -}}
}
`

const transformMapTmpl = `{{ .Target }} {{ if .NewVar }}:{{ end }}= make(map[{{ .KeyTypeRef }}]{{ .ElemTypeRef }}, len({{ .Source }}))
for key, val := range {{ .Source }} {
	{{ transformAttribute .SourceKey .TargetKey "key" "tk" .TargetPkg  .FromPtrs .ToPtrs .InitDefaults true .Scope -}}
	{{ transformAttribute .SourceElem .TargetElem "val" "tv" .TargetPkg .FromPtrs .ToPtrs .InitDefaults true .Scope -}}
	{{ .Target }}[tk] = tv
}
`
