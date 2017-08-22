package codegen

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"goa.design/goa/design"
)

var (
	transformArrayT *template.Template
	transformMapT   *template.Template
)

type (
	// TransformFunctionData describes a helper function used to transform
	// user types. These are necessary to prevent potential infinite
	// recursion when a type attribute is defined recursively. For example:
	//
	//     var Recursive = Type("Recursive", func() {
	//         Attribute("r", "Recursive")
	//     }
	//
	// Transforming this type requires generating an intermediary function:
	//
	//     func recursiveToRecursive(r *Recursive) *service.Recursive {
	//         var t service.Recursive
	//         if r.R != nil {
	//             t.R = recursiveToRecursive(r.R)
	//         }
	//    }
	//
	TransformFunctionData struct {
		Name          string
		ParamTypeRef  string
		ResultTypeRef string
		Code          string
	}
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
// fromPtrs and toPtrs indicate whether the source or target data structure
// respectively uses pointers to store all attributes even required ones (i.e.
// unmarshaled request body).
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

// GoTypeTransformHelpers returns the transform helper functions required to
// transform source into target give the other parameters. See GoTypeTransform
// for a description of the parameters. See TransformFunctionData for a
// rationale explaining the need for this function.
func GoTypeTransformHelpers(source, target design.DataType, sPkg, tPkg string, fromPtrs, toPtrs, initDefaults bool, scope *NameScope) ([]*TransformFunctionData, error) {
	return transformAttributeHelpers(source, target, sPkg, tPkg, fromPtrs, toPtrs, initDefaults, scope)
}

func transformAttributeHelpers(source, target design.DataType, sPkg, tPkg string, fromPtrs, toPtrs, initDefaults bool, scope *NameScope) ([]*TransformFunctionData, error) {
	var (
		helpers []*TransformFunctionData
		err     error
	)
	// Do not generate a transform function for the top most user type.
	switch {
	case design.IsArray(source):
		helpers, err = transformAttributeHelpers(
			design.AsArray(source).ElemType.Type,
			design.AsArray(target).ElemType.Type,
			sPkg, tPkg, fromPtrs, toPtrs, initDefaults, scope)
	case design.IsMap(source):
		sm := design.AsMap(source)
		tm := design.AsMap(target)
		helpers, err = transformAttributeHelpers(sm.ElemType.Type, tm.ElemType.Type,
			sPkg, tPkg, fromPtrs, toPtrs, initDefaults, scope)
		if err == nil {
			var other []*TransformFunctionData
			other, err = transformAttributeHelpers(sm.KeyType.Type, tm.KeyType.Type,
				sPkg, tPkg, fromPtrs, toPtrs, initDefaults, scope)
			helpers = append(helpers, other...)
		}
	case design.IsObject(source):
		helpers, err = transformObjectHelpers(source, target, sPkg, tPkg, fromPtrs, toPtrs, initDefaults, scope)
	}
	if err != nil {
		return nil, err
	}
	return helpers, nil
}

func transformObjectHelpers(source, target design.DataType, sPkg, tPkg string, fromPtrs, toPtrs, initDefaults bool, scope *NameScope) ([]*TransformFunctionData, error) {
	var (
		helpers []*TransformFunctionData
		err     error

		satt = &design.AttributeExpr{Type: source}
		tatt = &design.AttributeExpr{Type: target}
	)
	walkMatches(satt, tatt, func(src, tgt *design.MappedAttributeExpr, srcAtt, tgtAtt *design.AttributeExpr, n string) {
		if err != nil {
			return
		}
		h, err2 := collectHelpers(srcAtt, tgtAtt, sPkg, tPkg, fromPtrs, toPtrs, initDefaults, scope)
		if err2 != nil {
			err = err2
			return
		}
		helpers = append(helpers, h...)
	})
	if err != nil {
		return nil, err
	}
	return helpers, nil
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
	buffer := &bytes.Buffer{}
	var initCode string
	walkMatches(source, target, func(src, tgt *design.MappedAttributeExpr, srcAtt, _ *design.AttributeExpr, n string) {
		if !design.IsPrimitive(srcAtt.Type) {
			return
		}
		srcField := sctx + "." + Goify(src.ElemName(n), true)
		deref := ""
		if (fromPtrs || source.IsPrimitivePointer(n, !fromPtrs)) && tgt.IsRequired(n) {
			deref = "*"
		}
		if toPtrs && !source.IsPrimitivePointer(n, true) {
			deref = "&"
		}
		initCode += fmt.Sprintf("\n%s: %s%s,", Goify(tgt.ElemName(n), true), deref, srcField)
	})
	if initCode != "" {
		initCode += "\n"
	}
	assign := "="
	if newVar {
		assign = ":="
	}
	buffer.WriteString(fmt.Sprintf("%s %s &%s{%s}\n", tctx, assign,
		scope.GoFullTypeName(target, targetPkg), initCode))
	var err error
	walkMatches(source, target, func(src, tgt *design.MappedAttributeExpr, srcAtt, tgtAtt *design.AttributeExpr, n string) {
		srcField := sctx + "." + Goify(src.ElemName(n), true)
		tgtField := tctx + "." + Goify(tgt.ElemName(n), true)
		err = isCompatible(srcAtt.Type, tgtAtt.Type, srcField, tgtField)
		if err != nil {
			return
		}

		var (
			code string
		)
		_, ok := srcAtt.Type.(design.UserType)
		switch {
		case ok:
			code = fmt.Sprintf("%s = %s(%s)\n", tgtField, transformHelperName(srcAtt, tgtAtt, fromPtrs, toPtrs, def, scope), srcField)
		case design.IsArray(srcAtt.Type):
			code, err = transformArray(design.AsArray(srcAtt.Type), design.AsArray(tgtAtt.Type), srcField, tgtField, targetPkg, fromPtrs, toPtrs, def, scope)
		case design.IsMap(srcAtt.Type):
			code, err = transformMap(design.AsMap(srcAtt.Type), design.AsMap(tgtAtt.Type), srcField, tgtField, targetPkg, fromPtrs, toPtrs, def, scope)
		case design.IsObject(srcAtt.Type):
			code, err = transformAttribute(srcAtt, tgtAtt, srcField, tgtField, targetPkg, fromPtrs, toPtrs, def, false, scope)
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
					code = fmt.Sprintf("%s\n\ttmp := %#v\n\t%s = &tmp\n", code, tgtAtt.DefaultValue, tgtField)
				} else {
					code = fmt.Sprintf("%s\n\t%s = %#v\n", code, tgtField, tgtAtt.DefaultValue)
				}
			}
			code += "}\n"
		}

		if err != nil {
			return
		}
		buffer.WriteString(code)
	})
	if err != nil {
		return "", err
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
		"LoopVar":      string(105 + strings.Count(sctx, ".")),
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

// collectHelpers recursively traverses the given attributes and return the
// transform helper functions required to generate the transform code.
func collectHelpers(source, target *design.AttributeExpr, sourcePkg, targetPkg string, fromPtrs, toPtrs, def bool, scope *NameScope, seen ...map[string]*TransformFunctionData) ([]*TransformFunctionData, error) {
	var data []*TransformFunctionData
	switch {
	case design.IsArray(source.Type):
		helpers, err := collectHelpers(
			design.AsArray(source.Type).ElemType,
			design.AsArray(target.Type).ElemType,
			sourcePkg, targetPkg, fromPtrs, toPtrs, def, scope, seen...)
		if err != nil {
			return nil, err
		}
		data = append(data, helpers...)
	case design.IsMap(source.Type):
		helpers, err := collectHelpers(
			design.AsMap(source.Type).KeyType,
			design.AsMap(target.Type).KeyType,
			sourcePkg, targetPkg, fromPtrs, toPtrs, def, scope, seen...)
		if err != nil {
			return nil, err
		}
		data = append(data, helpers...)
		helpers, err = collectHelpers(
			design.AsMap(source.Type).ElemType,
			design.AsMap(target.Type).ElemType,
			sourcePkg, targetPkg, fromPtrs, toPtrs, def, scope, seen...)
		if err != nil {
			return nil, err
		}
		data = append(data, helpers...)
	case design.IsObject(source.Type):
		if ut, ok := source.Type.(design.UserType); ok {
			name := transformHelperName(source, target, fromPtrs, toPtrs, def, scope)
			var s map[string]*TransformFunctionData
			if len(seen) > 0 {
				s = seen[0]
			} else {
				s = make(map[string]*TransformFunctionData)
				seen = append(seen, s)
			}
			if _, ok := s[name]; ok {
				return nil, nil
			}
			var code string
			code, err := transformAttribute(ut.Attribute(), target, "v", "res", targetPkg, fromPtrs, toPtrs, def, true, scope)
			if err != nil {
				return nil, err
			}
			t := &TransformFunctionData{
				Name:          name,
				ParamTypeRef:  scope.GoFullTypeRef(source, sourcePkg),
				ResultTypeRef: scope.GoFullTypeRef(target, targetPkg),
				Code:          code,
			}
			s[name] = t
			data = append(data, t)
		}
		var err error
		walkMatches(source, target, func(_, _ *design.MappedAttributeExpr, src, tgt *design.AttributeExpr, n string) {
			var helpers []*TransformFunctionData
			helpers, err = collectHelpers(
				src,
				tgt,
				sourcePkg, targetPkg, fromPtrs, toPtrs, def, scope, seen...)
			if err != nil {
				return
			}
			data = append(data, helpers...)
		})
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

func walkMatches(source, target *design.AttributeExpr, walker func(src, tgt *design.MappedAttributeExpr, srcc, tgtc *design.AttributeExpr, n string)) {
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
	for _, natt := range *srcObj {
		n := natt.Name
		tgtc, ok := attributeMap[n]
		if !ok {
			continue
		}
		walker(src, tgt, natt.Attribute, tgtc, n)
	}
}

// AppendHelpers takes care of only appending helper functions from newH that
// are not already in oldH.
func AppendHelpers(oldH, newH []*TransformFunctionData) []*TransformFunctionData {
	res := oldH
	for _, h := range newH {
		found := false
		for _, h2 := range oldH {
			if h.Name == h2.Name {
				found = true
				continue
			}
		}
		if found {
			continue
		}
		res = append(res, h)
	}
	return res
}

func transformHelperName(satt, tatt *design.AttributeExpr, fromPtrs, toPtrs, def bool, scope *NameScope) string {
	var (
		sname    string
		tname    string
		fps, tps string
		defs     string
	)
	{
		sname = scope.GoTypeName(satt)
		tname = scope.GoTypeName(tatt)
		if fromPtrs {
			fps = "SrcPtr"
		}
		if toPtrs {
			tps = "TgtPtr"
		}
		if !def {
			defs = "NoDefault"
		}
	}
	return Goify(sname+"To"+tname+fps+tps+defs, false)
}

const transformArrayTmpl = `{{ .Target}} {{ if .NewVar }}:{{ end }}= make([]{{ .ElemTypeRef }}, len({{ .Source }}))
for {{ .LoopVar }}, val := range {{ .Source }} {
	{{ transformAttribute .SourceElem .TargetElem "val" (printf "%s[%s]" .Target .LoopVar) .TargetPkg .FromPtrs .ToPtrs .InitDefaults false .Scope -}}
}
`

const transformMapTmpl = `{{ .Target }} {{ if .NewVar }}:{{ end }}= make(map[{{ .KeyTypeRef }}]{{ .ElemTypeRef }}, len({{ .Source }}))
for key, val := range {{ .Source }} {
	{{ transformAttribute .SourceKey .TargetKey "key" "tk" .TargetPkg  .FromPtrs .ToPtrs .InitDefaults true .Scope -}}
	{{ transformAttribute .SourceElem .TargetElem "val" "tv" .TargetPkg .FromPtrs .ToPtrs .InitDefaults true .Scope -}}
	{{ .Target }}[tk] = tv
}
`
