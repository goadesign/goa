package codegen

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"goa.design/goa/expr"
)

var transformGoArrayT, transformGoMapT, transformGoArrayElemT, transformGoMapElemT *template.Template

// NOTE: can't initialize inline because https://github.com/golang/go/issues/1817
func init() {
	fm := template.FuncMap{"transformAttribute": transformAttribute, "transformHelperName": transformHelperName}
	transformGoArrayT = template.Must(template.New("transformGoArray").Funcs(fm).Parse(transformGoArrayTmpl))
	transformGoMapT = template.Must(template.New("transformGoMap").Funcs(fm).Parse(transformGoMapTmpl))
	transformGoArrayElemT = template.Must(template.New("transformGoArrayElem").Funcs(fm).Parse(transformGoArrayElemTmpl))
	transformGoMapElemT = template.Must(template.New("transformGoMapElem").Funcs(fm).Parse(transformGoMapElemTmpl))
}

// GoTransform produces Go code that initializes the data structure defined
// by target from an instance of the data structure described by source.
// The data structures can be objects, arrays or maps. The algorithm
// matches object fields by name and ignores object fields in target that
// don't have a match in source. The matching and generated code leverage
// mapped attributes so that attribute names may use the "name:elem"
// syntax to define the name of the design attribute and the name of the
// corresponding generated Go struct field. The function returns an error
// if target is not compatible with source (different type, fields of
// different type etc).
//
// source and target are the attributes used in the transformation
//
// sourceVar and targetVar are the variable names used in the transformation
//
// sourceCtx and targetCtx are the attribute contexts for the source and target
// attributes
//
// prefix is the transformation helper function prefix
//
func GoTransform(source, target *expr.AttributeExpr, sourceVar, targetVar string, sourceCtx, targetCtx *AttributeContext, prefix string) (string, []*TransformFunctionData, error) {
	ta := &TransformAttrs{
		SourceCtx: sourceCtx,
		TargetCtx: targetCtx,
		Prefix:    prefix,
	}

	code, err := transformAttribute(source, target, sourceVar, targetVar, true, ta)
	if err != nil {
		return "", nil, err
	}

	funcs, err := transformAttributeHelpers(source, target, ta, make(map[string]*TransformFunctionData))
	if err != nil {
		return "", nil, err
	}

	return strings.TrimRight(code, "\n"), funcs, nil
}

// transformAttribute returns the code to transform source attribute to target
// attribute. It returns an error if source and target are not compatible for
// transformation.
func transformAttribute(source, target *expr.AttributeExpr, sourceVar, targetVar string, newVar bool, ta *TransformAttrs) (string, error) {
	var err error
	{
		if err = IsCompatible(source.Type, target.Type, sourceVar, targetVar); err != nil {
			return "", err
		}
	}

	var code string
	{
		switch {
		case expr.IsArray(source.Type):
			code, err = transformArray(expr.AsArray(source.Type), expr.AsArray(target.Type), sourceVar, targetVar, newVar, ta)
		case expr.IsMap(source.Type):
			code, err = transformMap(expr.AsMap(source.Type), expr.AsMap(target.Type), sourceVar, targetVar, newVar, ta)
		case expr.IsObject(source.Type):
			code, err = transformObject(source, target, sourceVar, targetVar, newVar, ta)
		default:
			assign := "="
			if newVar {
				assign = ":="
			}
			if _, ok := target.Type.(expr.UserType); ok {
				// Primitive user type, these are used for error results
				cast := ta.TargetCtx.Scope.Ref(target, ta.TargetCtx.Pkg)
				return fmt.Sprintf("%s %s %s(%s)\n", targetVar, assign, cast, sourceVar), nil
			}
			code = fmt.Sprintf("%s %s %s\n", targetVar, assign, sourceVar)
		}
	}
	if err != nil {
		return "", err
	}
	return code, nil
}

// transformObject generates Go code to transform source object to target
// object.
func transformObject(source, target *expr.AttributeExpr, sourceVar, targetVar string, newVar bool, ta *TransformAttrs) (string, error) {
	var (
		initCode     string
		postInitCode string
	)
	{
		// walk through primitives first to initialize the struct
		walkMatches(source, target, func(srcMatt, tgtMatt *expr.MappedAttributeExpr, srcc, tgtc *expr.AttributeExpr, n string) {
			if !expr.IsPrimitive(srcc.Type) {
				return
			}
			var (
				deref string

				srcPtr   = ta.SourceCtx.IsPrimitivePointer(n, srcMatt.AttributeExpr)
				tgtPtr   = ta.TargetCtx.IsPrimitivePointer(n, tgtMatt.AttributeExpr)
				srcField = sourceVar + "." + GoifyAtt(srcc, srcMatt.ElemName(n), true)
				tgtField = GoifyAtt(tgtc, tgtMatt.ElemName(n), true)
			)
			{
				switch {
				case srcPtr && !tgtPtr:
					if !srcMatt.IsRequired(n) {
						postInitCode += fmt.Sprintf("if %s != nil {\n\t%s.%s = %s\n}\n", srcField, targetVar, tgtField, "*"+srcField)
						return
					}
					deref = "*"
				case !srcPtr && tgtPtr:
					deref = "&"
				}
			}
			initCode += fmt.Sprintf("\n%s: %s%s,", tgtField, deref, srcField)
		})
		if initCode != "" {
			initCode += "\n"
		}
	}

	buffer := &bytes.Buffer{}
	deref := "&"
	assign := "="
	if newVar {
		assign = ":="
	}
	buffer.WriteString(fmt.Sprintf("%s %s %s%s{%s}\n", targetVar, assign, deref, ta.TargetCtx.Scope.Name(target, ta.TargetCtx.Pkg), initCode))
	buffer.WriteString(postInitCode)

	// iterate through attributes to initialize rest of the struct fields and
	// handle default values
	var err error
	walkMatches(source, target, func(srcMatt, tgtMatt *expr.MappedAttributeExpr, srcc, tgtc *expr.AttributeExpr, n string) {
		if err = IsCompatible(srcc.Type, tgtc.Type, sourceVar, targetVar); err != nil {
			return
		}

		var (
			code string

			srcVar = sourceVar + "." + GoifyAtt(srcc, srcMatt.ElemName(n), true)
			tgtVar = targetVar + "." + GoifyAtt(tgtc, tgtMatt.ElemName(n), true)
		)
		{
			_, ok := srcc.Type.(expr.UserType)
			switch {
			case expr.IsArray(srcc.Type):
				code, err = transformArrayElem(expr.AsArray(srcc.Type), expr.AsArray(tgtc.Type), srcVar, tgtVar, false, ta)
			case expr.IsMap(srcc.Type):
				code, err = transformMapElem(expr.AsMap(srcc.Type), expr.AsMap(tgtc.Type), srcVar, tgtVar, false, ta)
			case ok:
				code = fmt.Sprintf("%s = %s(%s)\n", tgtVar, transformHelperName(srcc, tgtc, ta), srcVar)
			case expr.IsObject(srcc.Type):
				code, err = transformAttribute(srcc, tgtc, srcVar, tgtVar, false, ta)
			}
		}
		if err != nil {
			return
		}

		// We need to check for a nil source if it holds a reference (pointer to
		// primitive or an object, array or map) and is not required. We also want
		// to always check nil if the attribute is not a primitive; it's a
		// 1) user type and we want to avoid calling transform helper functions
		// with nil value
		// 2) it's an object, map or array to avoid making empty arrays and maps
		// and to avoid derefencing nil.
		var checkNil bool
		{
			isRef := !expr.IsPrimitive(srcc.Type) && !srcMatt.IsRequired(n) || ta.SourceCtx.IsPrimitivePointer(n, srcMatt.AttributeExpr) && expr.IsPrimitive(srcc.Type)
			marshalNonPrimitive := !expr.IsPrimitive(srcc.Type) && ta.SourceCtx.UseDefault && ta.TargetCtx.UseDefault
			checkNil = isRef || marshalNonPrimitive
		}
		if code != "" && checkNil {
			code = fmt.Sprintf("if %s != nil {\n\t%s}\n", srcVar, code)
		}

		// Default value handling. We need to handle default values if the target
		// type uses default values (i.e. attributes with default values are
		// non-pointers) and has a default value set.
		if tdef := tgtc.DefaultValue; tdef != nil && ta.TargetCtx.UseDefault {
			if (ta.SourceCtx.IsPrimitivePointer(n, srcMatt.AttributeExpr) || !expr.IsPrimitive(srcc.Type)) && !srcMatt.IsRequired(n) {
				code += fmt.Sprintf("if %s == nil {\n\t", srcVar)
				if ta.TargetCtx.IsPrimitivePointer(n, tgtMatt.AttributeExpr) && expr.IsPrimitive(tgtc.Type) {
					code += fmt.Sprintf("var tmp %s = %#v\n\t%s = &tmp\n", GoNativeTypeName(tgtc.Type), tdef, tgtVar)
				} else {
					code += fmt.Sprintf("%s = %#v\n", tgtVar, tdef)
				}
				code += "}\n"
			}
		}
		buffer.WriteString(code)
	})
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}

// transformArray generates Go code to transform source array to target array.
func transformArray(source, target *expr.Array, sourceVar, targetVar string, newVar bool, ta *TransformAttrs) (string, error) {
	if err := IsCompatible(source.ElemType.Type, target.ElemType.Type, sourceVar+"[0]", targetVar+"[0]"); err != nil {
		return "", err
	}
	data := map[string]interface{}{
		"ElemTypeRef":    ta.TargetCtx.Scope.Ref(target.ElemType, ta.TargetCtx.Pkg),
		"SourceElem":     source.ElemType,
		"TargetElem":     target.ElemType,
		"SourceVar":      sourceVar,
		"TargetVar":      targetVar,
		"NewVar":         newVar,
		"TransformAttrs": ta,
		"LoopVar":        string(105 + strings.Count(targetVar, "[")),
	}
	var buf bytes.Buffer
	if err := transformGoArrayT.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// transformArrayElem generates Go code to transform an object field of type array.
func transformArrayElem(source, target *expr.Array, sourceVar, targetVar string, newVar bool, ta *TransformAttrs) (string, error) {
	st, tt := source.ElemType.Type, target.ElemType.Type
	if err := IsCompatible(st, tt, sourceVar+"[0]", targetVar+"[0]"); err != nil {
		return "", err
	}
	if _, ok := st.(expr.UserType); ok {
		data := map[string]interface{}{
			"ElemTypeRef":    ta.TargetCtx.Scope.Ref(target.ElemType, ta.TargetCtx.Pkg),
			"SourceElem":     source.ElemType,
			"TargetElem":     target.ElemType,
			"SourceVar":      sourceVar,
			"TargetVar":      targetVar,
			"NewVar":         newVar,
			"TransformAttrs": ta,
			"LoopVar":        string(105 + strings.Count(targetVar, "[")),
		}
		var buf bytes.Buffer
		if err := transformGoArrayElemT.Execute(&buf, data); err != nil {
			return "", err
		}
		return buf.String(), nil
	}
	return transformArray(source, target, sourceVar, targetVar, newVar, ta)
}

// transformMap generates Go code to transform source map to target map.
func transformMap(source, target *expr.Map, sourceVar, targetVar string, newVar bool, ta *TransformAttrs) (string, error) {
	if err := IsCompatible(source.KeyType.Type, target.KeyType.Type, sourceVar+"[key]", targetVar+"[key]"); err != nil {
		return "", err
	}
	if err := IsCompatible(source.ElemType.Type, target.ElemType.Type, sourceVar+"[*]", targetVar+"[*]"); err != nil {
		return "", err
	}
	data := map[string]interface{}{
		"KeyTypeRef":     ta.TargetCtx.Scope.Ref(target.KeyType, ta.TargetCtx.Pkg),
		"ElemTypeRef":    ta.TargetCtx.Scope.Ref(target.ElemType, ta.TargetCtx.Pkg),
		"SourceKey":      source.KeyType,
		"TargetKey":      target.KeyType,
		"SourceElem":     source.ElemType,
		"TargetElem":     target.ElemType,
		"SourceVar":      sourceVar,
		"TargetVar":      targetVar,
		"NewVar":         newVar,
		"TransformAttrs": ta,
		"LoopVar":        "",
	}
	if depth := MapDepth(target); depth > 0 {
		data["LoopVar"] = string(97 + depth)
	}
	var buf bytes.Buffer
	if err := transformGoMapT.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// transformMapElem generates Go code to transform an object field of type map.
func transformMapElem(source, target *expr.Map, sourceVar, targetVar string, newVar bool, ta *TransformAttrs) (string, error) {
	if err := IsCompatible(source.KeyType.Type, target.KeyType.Type, sourceVar+"[key]", targetVar+"[key]"); err != nil {
		return "", err
	}
	if err := IsCompatible(source.ElemType.Type, target.ElemType.Type, sourceVar+"[*]", targetVar+"[*]"); err != nil {
		return "", err
	}
	if _, ok := target.ElemType.Type.(expr.UserType); ok {
		data := map[string]interface{}{
			"KeyTypeRef":     ta.TargetCtx.Scope.Ref(target.KeyType, ta.TargetCtx.Pkg),
			"ElemTypeRef":    ta.TargetCtx.Scope.Ref(target.ElemType, ta.TargetCtx.Pkg),
			"SourceKey":      source.KeyType,
			"TargetKey":      target.KeyType,
			"SourceElem":     source.ElemType,
			"TargetElem":     target.ElemType,
			"SourceVar":      sourceVar,
			"TargetVar":      targetVar,
			"NewVar":         newVar,
			"TransformAttrs": ta,
			"LoopVar":        "",
		}
		if depth := MapDepth(target); depth > 0 {
			data["LoopVar"] = string(97 + depth)
		}
		var buf bytes.Buffer
		if err := transformGoMapElemT.Execute(&buf, data); err != nil {
			return "", err
		}
		return buf.String(), nil
	}
	return transformMap(source, target, sourceVar, targetVar, newVar, ta)
}

// transformAttributeHelpers returns the Go transform functions and their definitions
// that may be used in code produced by Transform. It returns an error if source and
// target are incompatible (different types, fields of different type etc).
//
// source, target are the source and target attributes used in transformation
//
// ta holds the transform attributes
//
// seen keeps track of generated transform functions to avoid recursion
//
func transformAttributeHelpers(source, target *expr.AttributeExpr, ta *TransformAttrs, seen map[string]*TransformFunctionData) ([]*TransformFunctionData, error) {
	var (
		helpers []*TransformFunctionData
		err     error
	)
	{
		// Do not generate a transform function for the top most user type.
		switch {
		case expr.IsArray(source.Type):
			source = expr.AsArray(source.Type).ElemType
			target = expr.AsArray(target.Type).ElemType
			helpers, err = transformAttributeHelpers(source, target, ta, seen)
		case expr.IsMap(source.Type):
			sm := expr.AsMap(source.Type)
			tm := expr.AsMap(target.Type)
			helpers, err = transformAttributeHelpers(sm.ElemType, tm.ElemType, ta, seen)
			if err == nil {
				var other []*TransformFunctionData
				other, err = transformAttributeHelpers(sm.KeyType, tm.KeyType, ta, seen)
				helpers = append(helpers, other...)
			}
		case expr.IsObject(source.Type):
			walkMatches(source, target, func(srcMatt, tgtMatt *expr.MappedAttributeExpr, srcc, tgtc *expr.AttributeExpr, n string) {
				if err != nil {
					return
				}
				h, err2 := collectHelpers(srcc, tgtc, srcMatt.IsRequired(n), ta, seen)
				if err2 != nil {
					err = err2
					return
				}
				helpers = append(helpers, h...)
			})
		}
	}
	if err != nil {
		return nil, err
	}
	return helpers, nil
}

// collectHelpers recursively traverses the given attributes and return the
// transform helper functions required to generate the transform code.
func collectHelpers(source, target *expr.AttributeExpr, req bool, ta *TransformAttrs, seen map[string]*TransformFunctionData) ([]*TransformFunctionData, error) {
	var (
		data []*TransformFunctionData
	)
	switch {
	case expr.IsArray(source.Type):
		st, tt := expr.AsArray(source.Type).ElemType, expr.AsArray(target.Type).ElemType
		if _, ok := st.Type.(expr.UserType); ok {
			// Handle array of user types explicitly to avoid infinite recursions
			// when the user type has an attribute of type array of itself.
			d, err := generateHelper(st, tt, req, ta, seen)
			if err != nil {
				return nil, err
			}
			data = append(data, d)
		} else {
			helpers, err := transformAttributeHelpers(st, tt, ta, seen)
			if err != nil {
				return nil, err
			}
			data = append(data, helpers...)
		}
	case expr.IsMap(source.Type):
		se, te := expr.AsMap(source.Type).ElemType, expr.AsMap(target.Type).ElemType
		sk, tk := expr.AsMap(source.Type).KeyType, expr.AsMap(target.Type).KeyType
		helpers, err := transformAttributeHelpers(sk, tk, ta, seen)
		if err != nil {
			return nil, err
		}
		data = append(data, helpers...)
		if _, ok := se.Type.(expr.UserType); ok {
			// Handle map of user types explicitly to avoid infinite recursions
			// when the user type has an attribute of type map of itself.
			d, err := generateHelper(se, te, req, ta, seen)
			if err != nil {
				return nil, err
			}
			data = append(data, d)
		} else {
			helpers, err = transformAttributeHelpers(se, te, ta, seen)
			if err != nil {
				return nil, err
			}
			data = append(data, helpers...)
		}
	case expr.IsObject(source.Type):
		if _, ok := source.Type.(expr.UserType); ok {
			d, err := generateHelper(source, target, req, ta, seen)
			if err != nil {
				return nil, err
			}
			data = append(data, d)
		}

		// collect helpers
		var err error
		{
			walkMatches(source, target, func(srcMatt, _ *expr.MappedAttributeExpr, srcc, tgtc *expr.AttributeExpr, n string) {
				name := transformHelperName(source, target, ta)
				if _, ok := seen[name]; ok {
					return
				}
				var helpers []*TransformFunctionData
				helpers, err = collectHelpers(srcc, tgtc, srcMatt.IsRequired(n), ta, seen)
				if err != nil {
					return
				}
				data = append(data, helpers...)
			})
		}
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

// generateHelper generates the code that transform instances of source into
// target.
func generateHelper(source, target *expr.AttributeExpr, req bool, ta *TransformAttrs, seen map[string]*TransformFunctionData) (*TransformFunctionData, error) {
	name := transformHelperName(source, target, ta)
	if _, ok := seen[name]; ok {
		return nil, nil
	}
	code, err := transformAttribute(source.Type.(expr.UserType).Attribute(), target, "v", "res", true, ta)
	if err != nil {
		return nil, err
	}
	if !req {
		code = "if v == nil {\n\treturn nil\n}\n" + code
	}
	tfd := &TransformFunctionData{
		Name:          name,
		ParamTypeRef:  ta.SourceCtx.Scope.Ref(source, ta.SourceCtx.Pkg),
		ResultTypeRef: ta.TargetCtx.Scope.Ref(target, ta.TargetCtx.Pkg),
		Code:          code,
	}
	seen[name] = tfd
	return tfd, nil
}

// walkMatches iterates through the source attribute expression and executes
// the walker function.
func walkMatches(source, target *expr.AttributeExpr, walker func(src, tgt *expr.MappedAttributeExpr, srcc, tgtc *expr.AttributeExpr, n string)) {
	srcMatt := expr.NewMappedAttributeExpr(source)
	tgtMatt := expr.NewMappedAttributeExpr(target)
	srcObj := expr.AsObject(srcMatt.Type)
	tgtObj := expr.AsObject(tgtMatt.Type)
	for _, nat := range *srcObj {
		if att := tgtObj.Attribute(nat.Name); att != nil {
			walker(srcMatt, tgtMatt, nat.Attribute, att, nat.Name)
		}
	}
}

// transformHelperName returns the transformation function name to initialize a
// target user type from an instance of a source user type.
func transformHelperName(source, target *expr.AttributeExpr, ta *TransformAttrs) string {
	var (
		sname  string
		tname  string
		prefix string
	)
	{
		sname = Goify(ta.SourceCtx.Scope.Name(source, ta.SourceCtx.Pkg), true)
		tname = Goify(ta.TargetCtx.Scope.Name(target, ta.TargetCtx.Pkg), true)
		prefix = ta.Prefix
		if prefix == "" {
			prefix = "transform"
		}
	}
	return Goify(prefix+sname+"To"+tname, false)
}

const (
	transformGoArrayTmpl = `{{ .TargetVar }} {{ if .NewVar }}:={{ else }}={{ end }} make([]{{ .ElemTypeRef }}, len({{ .SourceVar }}))
for {{ .LoopVar }}, val := range {{ .SourceVar }} {
  {{ transformAttribute .SourceElem .TargetElem "val" (printf "%s[%s]" .TargetVar .LoopVar) false .TransformAttrs -}}
}
`

	transformGoArrayElemTmpl = `{{ .TargetVar }} {{ if .NewVar }}:={{ else }}={{ end }} make([]{{ .ElemTypeRef }}, len({{ .SourceVar }}))
for {{ .LoopVar }}, val := range {{ .SourceVar }} {
  {{ .TargetVar }}[{{ .LoopVar }}] = {{ transformHelperName .SourceElem .TargetElem .TransformAttrs }}(val)
}
`

	transformGoMapTmpl = `{{ .TargetVar }} {{ if .NewVar }}:={{ else }}={{ end }} make(map[{{ .KeyTypeRef }}]{{ .ElemTypeRef }}, len({{ .SourceVar }}))
for key, val := range {{ .SourceVar }} {
  {{ transformAttribute .SourceKey .TargetKey "key" "tk" true .TransformAttrs -}}
  {{ transformAttribute .SourceElem .TargetElem "val" (printf "tv%s" .LoopVar) true .TransformAttrs -}}
  {{ .TargetVar }}[tk] = {{ printf "tv%s" .LoopVar }}
}
`

	transformGoMapElemTmpl = `{{ .TargetVar }} {{ if .NewVar }}:={{ else }}={{ end }} make(map[{{ .KeyTypeRef }}]{{ .ElemTypeRef }}, len({{ .SourceVar }}))
for key, val := range {{ .SourceVar }} {
  {{ transformAttribute .SourceKey .TargetKey "key" "tk" true .TransformAttrs -}}
  {{ .TargetVar }}[tk] = {{ transformHelperName .SourceElem .TargetElem .TransformAttrs }}(val)
}
`
)
