package codegen

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"goa.design/goa/expr"
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

	// too many args...

	targs struct {
		sourceVar, targetVar string
		sourcePkg, targetPkg string
		unmarshal            bool
		scope                *NameScope
	}

	thargs struct {
		sourcePkg, targetPkg string
		unmarshal            bool
		scope                *NameScope
	}
)

// NOTE: can't initialize inline because https://github.com/golang/go/issues/1817
func init() {
	funcMap := template.FuncMap{"transformAttribute": transformAttributeHelper}
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
// sourcePkg and targetPkg contain the name of the Go package that defines the
// source or target type respectively in case it's not the same package as where
// the generated code lives.
//
// unmarshal if true indicates whether the code is being generated to
// initialize a type from unmarshaled data, otherwise to initialize a type that
// is marshaled:
//
//   if unmarshal is true (unmarshal)
//     - The source type uses pointers for all fields - even required ones.
//     - The target type do not use pointers for primitive fields that have
//			 default values even when not required.
//
//   if unmarshal is false (marshal)
//     - The source type used do not use pointers for primitive fields that
//			 have default values even when not required.
//     - The target type fields are initialized with their default values
//			 (if any) when source field is a primitive pointer and nil or a
//			 non-primitive type and nil.
//
// scope is used to compute the name of the user types when initializing fields
// that use them.
//
func GoTypeTransform(source, target expr.DataType, sourceVar, targetVar, sourcePkg, targetPkg string, unmarshal bool, scope *NameScope) (string, []*TransformFunctionData, error) {

	var (
		satt = &expr.AttributeExpr{Type: source}
		tatt = &expr.AttributeExpr{Type: target}
	)

	a := targs{sourceVar, targetVar, sourcePkg, targetPkg, unmarshal, scope}
	code, err := transformAttribute(satt, tatt, true, a)
	if err != nil {
		return "", nil, err
	}

	b := thargs{sourcePkg, targetPkg, unmarshal, scope}
	funcs, err := transformAttributeHelpers(source, target, b)
	if err != nil {
		return "", nil, err
	}

	return strings.TrimRight(code, "\n"), funcs, nil
}

func transformAttribute(source, target *expr.AttributeExpr, newVar bool, a targs) (string, error) {
	if err := isCompatible(source.Type, target.Type, a.sourceVar, a.targetVar); err != nil {
		return "", err
	}
	var (
		code string
		err  error
	)
	switch {
	case expr.IsArray(source.Type):
		code, err = transformArray(expr.AsArray(source.Type), expr.AsArray(target.Type), newVar, a)
	case expr.IsMap(source.Type):
		code, err = transformMap(expr.AsMap(source.Type), expr.AsMap(target.Type), newVar, a)
	case expr.IsObject(source.Type):
		code, err = transformObject(source, target, newVar, a)
	default:
		assign := "="
		if newVar {
			assign = ":="
		}
		if _, ok := target.Type.(expr.UserType); ok {
			// Primitive user type, these are used for error results
			cast := a.scope.GoFullTypeRef(target, a.targetPkg)
			return fmt.Sprintf("%s %s %s(%s)\n", a.targetVar, assign, cast, a.sourceVar), nil
		}
		code = fmt.Sprintf("%s %s %s\n", a.targetVar, assign, a.sourceVar)
	}
	if err != nil {
		return "", err
	}
	return code, nil
}

func transformObject(source, target *expr.AttributeExpr, newVar bool, a targs) (string, error) {
	buffer := &bytes.Buffer{}
	var (
		initCode     string
		postInitCode string
	)
	walkMatches(source, target, func(src, tgt *expr.MappedAttributeExpr, srcAtt, _ *expr.AttributeExpr, n string) {
		if !expr.IsPrimitive(srcAtt.Type) {
			return
		}
		srcPtr := a.unmarshal || src.IsPrimitivePointer(n, !a.unmarshal)
		tgtPtr := tgt.IsPrimitivePointer(n, true)
		deref := ""
		srcField := a.sourceVar + "." + Goify(src.ElemName(n), true)
		if srcPtr && !tgtPtr {
			if !source.IsRequired(n) {
				postInitCode += fmt.Sprintf("if %s != nil {\n\t%s.%s = %s\n}\n",
					srcField, a.targetVar, Goify(tgt.ElemName(n), true), "*"+srcField)
				return
			}
			deref = "*"
		} else if !srcPtr && tgtPtr {
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
	deref := "&"
	// if the target is a raw struct no need to return a pointer
	if _, ok := target.Type.(*expr.Object); ok {
		deref = ""
	}
	buffer.WriteString(fmt.Sprintf("%s %s %s%s{%s}\n", a.targetVar, assign, deref,
		a.scope.GoFullTypeName(target, a.targetPkg), initCode))
	buffer.WriteString(postInitCode)
	var err error
	walkMatches(source, target, func(src, tgt *expr.MappedAttributeExpr, srcAtt, tgtAtt *expr.AttributeExpr, n string) {
		b := a
		b.sourceVar = a.sourceVar + "." + GoifyAtt(srcAtt, src.ElemName(n), true)
		b.targetVar = a.targetVar + "." + GoifyAtt(tgtAtt, tgt.ElemName(n), true)
		err = isCompatible(srcAtt.Type, tgtAtt.Type, b.sourceVar, b.targetVar)
		if err != nil {
			return
		}

		var (
			code string
		)
		_, ok := srcAtt.Type.(expr.UserType)
		switch {
		case expr.IsArray(srcAtt.Type):
			code, err = transformArray(expr.AsArray(srcAtt.Type), expr.AsArray(tgtAtt.Type), false, b)
		case expr.IsMap(srcAtt.Type):
			code, err = transformMap(expr.AsMap(srcAtt.Type), expr.AsMap(tgtAtt.Type), false, b)
		case ok:
			code = fmt.Sprintf("%s = %s(%s)\n", b.targetVar, transformHelperName(srcAtt, tgtAtt, b), b.sourceVar)
		case expr.IsObject(srcAtt.Type):
			code, err = transformAttribute(srcAtt, tgtAtt, false, b)
		}
		if err != nil {
			return
		}

		// We need to check for a nil source if it holds a reference
		// (pointer to primitive or an object, array or map) and is not
		// required. We also want to always check when unmarshaling if
		// the attribute type is not a primitive: either it's a user
		// type and we want to avoid calling transform helper functions
		// with nil value (if unmarshaling then requiredness has been
		// validated) or it's an object, map or array and we need to
		// check for nil to avoid making empty arrays and maps and to
		// avoid derefencing nil.
		var checkNil bool
		{
			isRef := !expr.IsPrimitive(srcAtt.Type) && !src.IsRequired(n) || src.IsPrimitivePointer(n, !b.unmarshal)
			marshalNonPrimitive := !b.unmarshal && !expr.IsPrimitive(srcAtt.Type)
			checkNil = isRef || marshalNonPrimitive
		}
		if code != "" && checkNil {
			code = fmt.Sprintf("if %s != nil {\n\t%s}\n", b.sourceVar, code)
		}

		// Default value handling.
		//
		// There are 2 cases: one when generating marshaler code
		// (a.unmarshal is false) and the other when generating
		// unmarshaler code (a.unmarshal is true).
		//
		// When generating marshaler code we want to be lax and not
		// assume that required fields are set in case they have a
		// default value, instead the generated code is going to set the
		// fields to their default value (only applies to non-primitive
		// attributes).
		//
		// When generating unmarshaler code we rely on validations
		// running prior to this code so assume required fields are set.
		if tgt.HasDefaultValue(n) {
			if b.unmarshal {
				code += fmt.Sprintf("if %s == nil {\n\t", b.sourceVar)
				if tgt.IsPrimitivePointer(n, true) {
					code += fmt.Sprintf("var tmp %s = %#v\n\t%s = &tmp\n", GoNativeTypeName(tgtAtt.Type), tgtAtt.DefaultValue, b.targetVar)
				} else {
					code += fmt.Sprintf("%s = %#v\n", b.targetVar, tgtAtt.DefaultValue)
				}
				code += "}\n"
			} else if src.IsPrimitivePointer(n, true) || !expr.IsPrimitive(srcAtt.Type) {
				code += fmt.Sprintf("if %s == nil {\n\t", b.sourceVar)
				if tgt.IsPrimitivePointer(n, true) {
					code += fmt.Sprintf("var tmp %s = %#v\n\t%s = &tmp\n", GoNativeTypeName(tgtAtt.Type), tgtAtt.DefaultValue, b.targetVar)
				} else {
					code += fmt.Sprintf("%s = %#v\n", b.targetVar, tgtAtt.DefaultValue)
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

func transformArray(source, target *expr.Array, newVar bool, a targs) (string, error) {
	if err := isCompatible(source.ElemType.Type, target.ElemType.Type, a.sourceVar+"[0]", a.targetVar+"[0]"); err != nil {
		return "", err
	}
	data := map[string]interface{}{
		"Source":      a.sourceVar,
		"Target":      a.targetVar,
		"NewVar":      newVar,
		"ElemTypeRef": a.scope.GoFullTypeRef(target.ElemType, a.targetPkg),
		"SourceElem":  source.ElemType,
		"TargetElem":  target.ElemType,
		"SourcePkg":   a.sourcePkg,
		"TargetPkg":   a.targetPkg,
		"Unmarshal":   a.unmarshal,
		"Scope":       a.scope,
		"LoopVar":     string(105 + strings.Count(a.targetVar, "[")),
	}
	var buf bytes.Buffer
	if err := transformArrayT.Execute(&buf, data); err != nil {
		panic(err) // bug
	}
	code := buf.String()

	return code, nil
}

func transformMap(source, target *expr.Map, newVar bool, a targs) (string, error) {
	if err := isCompatible(source.KeyType.Type, target.KeyType.Type, a.sourceVar+".key", a.targetVar+".key"); err != nil {
		return "", err
	}
	if err := isCompatible(source.ElemType.Type, target.ElemType.Type, a.sourceVar+"[*]", a.targetVar+"[*]"); err != nil {
		return "", err
	}
	data := map[string]interface{}{
		"Source":      a.sourceVar,
		"Target":      a.targetVar,
		"NewVar":      newVar,
		"KeyTypeRef":  a.scope.GoFullTypeRef(target.KeyType, a.targetPkg),
		"ElemTypeRef": a.scope.GoFullTypeRef(target.ElemType, a.targetPkg),
		"SourceKey":   source.KeyType,
		"TargetKey":   target.KeyType,
		"SourceElem":  source.ElemType,
		"TargetElem":  target.ElemType,
		"SourcePkg":   a.sourcePkg,
		"TargetPkg":   a.targetPkg,
		"Unmarshal":   a.unmarshal,
		"Scope":       a.scope,
		"LoopVar":     "",
	}
	if depth := mapDepth(target); depth > 0 {
		data["LoopVar"] = string(97 + depth)
	}
	var buf bytes.Buffer
	if err := transformMapT.Execute(&buf, data); err != nil {
		panic(err) // bug
	}
	return buf.String(), nil
}

// mapDepth returns the level of nested maps. If map not nested, it returns 0.
func mapDepth(mp *expr.Map) int {
	return traverseMap(mp.ElemType.Type, 0)
}

func traverseMap(dt expr.DataType, depth int, seen ...map[string]struct{}) int {
	if mp := expr.AsMap(dt); mp != nil {
		depth++
		depth = traverseMap(mp.ElemType.Type, depth, seen...)
	} else if ar := expr.AsArray(dt); ar != nil {
		depth = traverseMap(ar.ElemType.Type, depth, seen...)
	} else if mo := expr.AsObject(dt); mo != nil {
		var s map[string]struct{}
		if len(seen) > 0 {
			s = seen[0]
		} else {
			s = make(map[string]struct{})
			seen = append(seen, s)
		}
		key := dt.Name()
		if u, ok := dt.(expr.UserType); ok {
			key = u.ID()
		}
		if _, ok := s[key]; ok {
			return depth
		}
		s[key] = struct{}{}
		var level int
		for _, nat := range *mo {
			// if object type has attributes of type map then find out the attribute that has
			// the deepest level of nested maps
			lvl := 0
			lvl = traverseMap(nat.Attribute.Type, lvl, seen...)
			if lvl > level {
				level = lvl
			}
		}
		depth += level
	}
	return depth
}

func transformAttributeHelpers(source, target expr.DataType, a thargs, seen ...map[string]*TransformFunctionData) ([]*TransformFunctionData, error) {
	var (
		helpers []*TransformFunctionData
		err     error
	)
	// Do not generate a transform function for the top most user type.
	switch {
	case expr.IsArray(source):
		source = expr.AsArray(source).ElemType.Type
		target = expr.AsArray(target).ElemType.Type
		helpers, err = transformAttributeHelpers(source, target, a, seen...)
	case expr.IsMap(source):
		sm := expr.AsMap(source)
		tm := expr.AsMap(target)
		source = sm.ElemType.Type
		target = tm.ElemType.Type
		helpers, err = transformAttributeHelpers(source, target, a, seen...)
		if err == nil {
			var other []*TransformFunctionData
			source = sm.KeyType.Type
			target = tm.KeyType.Type
			other, err = transformAttributeHelpers(source, target, a, seen...)
			helpers = append(helpers, other...)
		}
	case expr.IsObject(source):
		helpers, err = transformObjectHelpers(source, target, a, seen...)
	}
	if err != nil {
		return nil, err
	}
	return helpers, nil
}

func transformObjectHelpers(source, target expr.DataType, a thargs, seen ...map[string]*TransformFunctionData) ([]*TransformFunctionData, error) {
	var (
		helpers []*TransformFunctionData
		err     error

		satt = &expr.AttributeExpr{Type: source}
		tatt = &expr.AttributeExpr{Type: target}
	)
	walkMatches(satt, tatt, func(src, tgt *expr.MappedAttributeExpr, srcAtt, tgtAtt *expr.AttributeExpr, n string) {
		if err != nil {
			return
		}
		h, err2 := collectHelpers(srcAtt, tgtAtt, a, src.IsRequired(n), seen...)
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

// isCompatible returns an error if a and b are not both objects, both arrays,
// both maps or both the same primitive type. actx and bctx are used to build
// the error message if any.
func isCompatible(a, b expr.DataType, actx, bctx string) error {
	switch {
	case expr.IsObject(a):
		if !expr.IsObject(b) {
			return fmt.Errorf("%s is an object but %s type is %s", actx, bctx, b.Name())
		}
	case expr.IsArray(a):
		if !expr.IsArray(b) {
			return fmt.Errorf("%s is an array but %s type is %s", actx, bctx, b.Name())
		}
	case expr.IsMap(a):
		if !expr.IsMap(b) {
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
func collectHelpers(source, target *expr.AttributeExpr, a thargs, req bool, seen ...map[string]*TransformFunctionData) ([]*TransformFunctionData, error) {
	var data []*TransformFunctionData
	switch {
	case expr.IsArray(source.Type):
		helpers, err := transformAttributeHelpers(
			expr.AsArray(source.Type).ElemType.Type,
			expr.AsArray(target.Type).ElemType.Type,
			a, seen...)
		if err != nil {
			return nil, err
		}
		data = append(data, helpers...)
	case expr.IsMap(source.Type):
		helpers, err := transformAttributeHelpers(
			expr.AsMap(source.Type).KeyType.Type,
			expr.AsMap(target.Type).KeyType.Type,
			a, seen...)
		if err != nil {
			return nil, err
		}
		data = append(data, helpers...)
		helpers, err = transformAttributeHelpers(
			expr.AsMap(source.Type).ElemType.Type,
			expr.AsMap(target.Type).ElemType.Type,
			a, seen...)
		if err != nil {
			return nil, err
		}
		data = append(data, helpers...)
	case expr.IsObject(source.Type):
		if ut, ok := source.Type.(expr.UserType); ok {
			name := transformHelperName(source, target, targs{unmarshal: a.unmarshal, scope: a.scope})
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
			code, err := transformAttribute(ut.Attribute(), target, true,
				targs{"v", "res", a.sourcePkg, a.targetPkg, a.unmarshal, a.scope})
			if err != nil {
				return nil, err
			}
			if !req {
				code = "if v == nil {\n\treturn nil\n}\n" + code
			}
			t := &TransformFunctionData{
				Name:          name,
				ParamTypeRef:  a.scope.GoFullTypeRef(source, a.sourcePkg),
				ResultTypeRef: a.scope.GoFullTypeRef(target, a.targetPkg),
				Code:          code,
			}
			s[name] = t
			data = append(data, t)
		}
		var err error
		walkMatches(source, target, func(srcm, _ *expr.MappedAttributeExpr, src, tgt *expr.AttributeExpr, n string) {
			var helpers []*TransformFunctionData
			helpers, err = collectHelpers(src, tgt, a, srcm.IsRequired(n), seen...)
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

func walkMatches(source, target *expr.AttributeExpr, walker func(src, tgt *expr.MappedAttributeExpr, srcc, tgtc *expr.AttributeExpr, n string)) {
	src := expr.NewMappedAttributeExpr(source)
	tgt := expr.NewMappedAttributeExpr(target)
	srcObj := expr.AsObject(src.Type)
	tgtObj := expr.AsObject(tgt.Type)
	// Map source object attribute names to target object attributes
	attributeMap := make(map[string]*expr.AttributeExpr)
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
	for _, h := range newH {
		found := false
		for _, h2 := range oldH {
			if h.Name == h2.Name {
				found = true
				break
			}
		}
		if !found {
			oldH = append(oldH, h)
		}
	}
	return oldH
}

func transformHelperName(satt, tatt *expr.AttributeExpr, a targs) string {
	var (
		sname  string
		tname  string
		prefix string
	)
	{
		sname = a.scope.GoTypeName(satt)
		if _, ok := satt.Meta["goa.external"]; ok {
			// type belongs to external package so name could clash
			sname += "Ext"
		}
		tname = a.scope.GoTypeName(tatt)
		if _, ok := tatt.Meta["goa.external"]; ok {
			// type belongs to external package so name could clash
			tname += "Ext"
		}
		prefix = "marshal"
		if a.unmarshal {
			prefix = "unmarshal"
		}
	}
	return Goify(prefix+sname+"To"+tname, false)
}

// used by template
func transformAttributeHelper(source, target *expr.AttributeExpr, sourceVar, targetVar, sourcePkg, targetPkg string, unmarshal, newVar bool, scope *NameScope) (string, error) {
	return transformAttribute(source, target, newVar, targs{sourceVar, targetVar, sourcePkg, targetPkg, unmarshal, scope})
}

const transformArrayTmpl = `{{ .Target}} {{ if .NewVar }}:{{ end }}= make([]{{ .ElemTypeRef }}, len({{ .Source }}))
for {{ .LoopVar }}, val := range {{ .Source }} {
	{{ transformAttribute .SourceElem .TargetElem "val" (printf "%s[%s]" .Target .LoopVar) .SourcePkg .TargetPkg .Unmarshal false .Scope -}}
}
`

const transformMapTmpl = `{{ .Target }} {{ if .NewVar }}:{{ end }}= make(map[{{ .KeyTypeRef }}]{{ .ElemTypeRef }}, len({{ .Source }}))
for key, val := range {{ .Source }} {
	{{ transformAttribute .SourceKey .TargetKey "key" "tk" .SourcePkg .TargetPkg .Unmarshal true .Scope -}}
	{{ transformAttribute .SourceElem .TargetElem "val" (printf "tv%s" .LoopVar) .SourcePkg .TargetPkg .Unmarshal true .Scope -}}
	{{ .Target }}[tk] = {{ printf "tv%s" .LoopVar }}
}
`
