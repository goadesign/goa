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
// unmarshal indicates whether the code is being generated to initialize a type
// from unmarshaled data or to initialize a type that is marshaled:
//
//     - The source type used to unmarshal uses pointers for all fields - even
//       required ones.
//
//     - The target type used to unmarshal and the source type used to marshal
//       do not use pointers for primitive fields that have default values even
//       when not required.
//
//     - The generated code initializes marshaled type fields with their default
//       values when otherwise nil.
//
// scope is used to compute the name of the user types when initializing fields
// that use them.
//
func GoTypeTransform(source, target design.DataType, sourceVar, targetVar, sourcePkg, targetPkg string,
	unmarshal bool, scope *NameScope) (string, []*TransformFunctionData, error) {

	var (
		satt = &design.AttributeExpr{Type: source}
		tatt = &design.AttributeExpr{Type: target}
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

func transformAttribute(source, target *design.AttributeExpr, newVar bool, a targs) (string, error) {
	if err := isCompatible(source.Type, target.Type, a.sourceVar, a.targetVar); err != nil {
		return "", err
	}
	var (
		code string
		err  error
	)
	switch {
	case design.IsArray(source.Type):
		code, err = transformArray(design.AsArray(source.Type), design.AsArray(target.Type), newVar, a)
	case design.IsMap(source.Type):
		code, err = transformMap(design.AsMap(source.Type), design.AsMap(target.Type), newVar, a)
	case design.IsObject(source.Type):
		code, err = transformObject(source, target, newVar, a)
	default:
		assign := "="
		if newVar {
			assign = ":="
		}
		if _, ok := target.Type.(design.UserType); ok {
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

func transformObject(source, target *design.AttributeExpr, newVar bool, a targs) (string, error) {
	buffer := &bytes.Buffer{}
	var (
		initCode     string
		postInitCode string
	)
	walkMatches(source, target, func(src, tgt *design.MappedAttributeExpr, srcAtt, _ *design.AttributeExpr, n string) {
		if !design.IsPrimitive(srcAtt.Type) {
			return
		}
		srcPtr := a.unmarshal || source.IsPrimitivePointer(n, !a.unmarshal)
		tgtPtr := target.IsPrimitivePointer(n, true)
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
	buffer.WriteString(fmt.Sprintf("%s %s &%s{%s}\n", a.targetVar, assign,
		a.scope.GoFullTypeName(target, a.targetPkg), initCode))
	buffer.WriteString(postInitCode)
	var err error
	walkMatches(source, target, func(src, tgt *design.MappedAttributeExpr, srcAtt, tgtAtt *design.AttributeExpr, n string) {
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
		_, ok := srcAtt.Type.(design.UserType)
		switch {
		case design.IsArray(srcAtt.Type):
			code, err = transformArray(design.AsArray(srcAtt.Type), design.AsArray(tgtAtt.Type), false, b)
		case design.IsMap(srcAtt.Type):
			code, err = transformMap(design.AsMap(srcAtt.Type), design.AsMap(tgtAtt.Type), false, b)
		case ok:
			code = fmt.Sprintf("%s = %s(%s)\n", b.targetVar, transformHelperName(srcAtt, tgtAtt, b), b.sourceVar)
		case design.IsObject(srcAtt.Type):
			code, err = transformAttribute(srcAtt, tgtAtt, false, b)
		}
		if err != nil {
			return
		}

		// Nil check handling.
		//
		// We need to check for a nil source if it holds a reference
		// (pointer to primitive or an object, array or map) and is not
		// required. We also want to always check when unmarshaling is
		// the attribute type is not a primitive: either it's a user
		// type and we want to avoid calling transform helper functions
		// with nil value (if unmarshaling then requiredness has been
		// validated) or it's an object, map or array and we need to
		// check for nil to avoid making empty arrays and maps and to
		// avoid derefencing nil.
		var checkNil bool
		{
			isRef := !design.IsPrimitive(srcAtt.Type) && !src.IsRequired(n) || src.IsPrimitivePointer(n, !b.unmarshal)
			marshalNonPrimitive := !b.unmarshal && !design.IsPrimitive(srcAtt.Type)
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
					code += fmt.Sprintf("tmp := %#v\n\t%s = &tmp\n", tgtAtt.DefaultValue, b.targetVar)
				} else {
					code += fmt.Sprintf("%s = %#v\n", b.targetVar, tgtAtt.DefaultValue)
				}
				code += "}\n"
			} else if src.IsPrimitivePointer(n, true) || !design.IsPrimitive(srcAtt.Type) {
				code += fmt.Sprintf("if %s == nil {\n\t", b.sourceVar)
				code += fmt.Sprintf("%s = %#v\n", b.targetVar, tgtAtt.DefaultValue)
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

func transformArray(source, target *design.Array, newVar bool, a targs) (string, error) {
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
		"LoopVar":     string(105 + strings.Count(a.targetVar, ".")),
	}
	var buf bytes.Buffer
	if err := transformArrayT.Execute(&buf, data); err != nil {
		panic(err) // bug
	}
	code := buf.String()

	return code, nil
}

func transformMap(source, target *design.Map, newVar bool, a targs) (string, error) {
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
		"KeyLoopVar":  "",
		"ValLoopVar":  "",
	}
	depth := 0
	if mapDepth(target.KeyType.Type, &depth); depth > 0 {
		data["KeyLoopVar"] = string(105 + depth)
	}
	depth = 0
	if mapDepth(target.ElemType.Type, &depth); depth > 0 {
		data["ValLoopVar"] = string(105 + depth)
	}
	var buf bytes.Buffer
	if err := transformMapT.Execute(&buf, data); err != nil {
		panic(err) // bug
	}
	return buf.String(), nil
}

func mapDepth(m design.DataType, depth *int, seen ...map[string]struct{}) {
	if mp := design.AsMap(m); mp != nil {
		*depth++
		mapDepth(mp.ElemType.Type, depth, seen...)
	} else if mo := design.AsObject(m); mo != nil {
		var s map[string]struct{}
		if len(seen) > 0 {
			s = seen[0]
			if _, ok := s[m.Name()]; ok {
				return
			}
		} else {
			s = make(map[string]struct{})
			seen = append(seen, s)
		}
		s[m.Name()] = struct{}{}
		for _, nat := range *mo {
			mapDepth(nat.Attribute.Type, depth, seen...)
		}
	}
	return
}

func transformAttributeHelpers(source, target design.DataType, a thargs, seen ...map[string]*TransformFunctionData) ([]*TransformFunctionData, error) {
	var (
		helpers []*TransformFunctionData
		err     error
	)
	// Do not generate a transform function for the top most user type.
	switch {
	case design.IsArray(source):
		source = design.AsArray(source).ElemType.Type
		target = design.AsArray(target).ElemType.Type
		helpers, err = transformAttributeHelpers(source, target, a, seen...)
	case design.IsMap(source):
		sm := design.AsMap(source)
		tm := design.AsMap(target)
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
	case design.IsObject(source):
		helpers, err = transformObjectHelpers(source, target, a, seen...)
	}
	if err != nil {
		return nil, err
	}
	return helpers, nil
}

func transformObjectHelpers(source, target design.DataType, a thargs, seen ...map[string]*TransformFunctionData) ([]*TransformFunctionData, error) {
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
func collectHelpers(source, target *design.AttributeExpr, a thargs, req bool, seen ...map[string]*TransformFunctionData) ([]*TransformFunctionData, error) {
	var data []*TransformFunctionData
	switch {
	case design.IsArray(source.Type):
		helpers, err := transformAttributeHelpers(
			design.AsArray(source.Type).ElemType.Type,
			design.AsArray(target.Type).ElemType.Type,
			a, seen...)
		if err != nil {
			return nil, err
		}
		data = append(data, helpers...)
	case design.IsMap(source.Type):
		helpers, err := transformAttributeHelpers(
			design.AsMap(source.Type).KeyType.Type,
			design.AsMap(target.Type).KeyType.Type,
			a, seen...)
		if err != nil {
			return nil, err
		}
		data = append(data, helpers...)
		helpers, err = transformAttributeHelpers(
			design.AsMap(source.Type).ElemType.Type,
			design.AsMap(target.Type).ElemType.Type,
			a, seen...)
		if err != nil {
			return nil, err
		}
		data = append(data, helpers...)
	case design.IsObject(source.Type):
		if ut, ok := source.Type.(design.UserType); ok {
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
		walkMatches(source, target, func(srcm, _ *design.MappedAttributeExpr, src, tgt *design.AttributeExpr, n string) {
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

func transformHelperName(satt, tatt *design.AttributeExpr, a targs) string {
	var (
		sname  string
		tname  string
		prefix string
	)
	{
		sname = a.scope.GoTypeName(satt)
		tname = a.scope.GoTypeName(tatt)
		prefix = "marshal"
		if a.unmarshal {
			prefix = "unmarshal"
		}
	}
	return Goify(prefix+sname+"To"+tname, false)
}

// used by template
func transformAttributeHelper(source, target *design.AttributeExpr, sourceVar, targetVar, sourcePkg, targetPkg string, unmarshal, newVar bool, scope *NameScope) (string, error) {
	return transformAttribute(source, target, newVar, targs{sourceVar, targetVar, sourcePkg, targetPkg, unmarshal, scope})
}

const transformArrayTmpl = `{{ .Target}} {{ if .NewVar }}:{{ end }}= make([]{{ .ElemTypeRef }}, len({{ .Source }}))
for {{ .LoopVar }}, val := range {{ .Source }} {
	{{ transformAttribute .SourceElem .TargetElem "val" (printf "%s[%s]" .Target .LoopVar) .SourcePkg .TargetPkg .Unmarshal false .Scope -}}
}
`

const transformMapTmpl = `{{ .Target }} {{ if .NewVar }}:{{ end }}= make(map[{{ .KeyTypeRef }}]{{ .ElemTypeRef }}, len({{ .Source }}))
for key, val := range {{ .Source }} {
	{{ transformAttribute .SourceKey .TargetKey "key" (printf "tk%s" .KeyLoopVar) .SourcePkg .TargetPkg  .Unmarshal true .Scope -}}
	{{ transformAttribute .SourceElem .TargetElem "val" (printf "tv%s" .ValLoopVar) .SourcePkg .TargetPkg .Unmarshal true .Scope -}}
	{{ .Target }}[{{ printf "tk%s" .KeyLoopVar }}] = {{ printf "tv%s" .ValLoopVar }}
}
`
