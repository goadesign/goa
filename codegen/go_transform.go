package codegen

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"goa.design/goa/expr"
)

var (
	transformGoArrayT *template.Template
	transformGoMapT   *template.Template
)

// NOTE: can't initialize inline because https://github.com/golang/go/issues/1817
func init() {
	funcMap := template.FuncMap{"transformAttribute": transformAttributeHelper}
	transformGoArrayT = template.Must(template.New("transformGoArray").Funcs(funcMap).Parse(transformGoArrayTmpl))
	transformGoMapT = template.Must(template.New("transformGoMap").Funcs(funcMap).Parse(transformGoMapTmpl))
}

type (
	// GoTransformer defines the fields to generate Go code when transforming
	// a source attribute to a target attribute. It implements the Transformer
	// interface.
	GoTransformer struct {
		*AttributeTransformer
	}
)

// NewGoTransformer returns a new transformer that generates Go code during
// transformation.
func NewGoTransformer(prefix string) Transformer {
	return &GoTransformer{
		AttributeTransformer: &AttributeTransformer{HelperPrefix: prefix},
	}
}

// Transform returns the code to transform source attribute to
// target attribute. It returns an error if source and target are not
// compatible for transformation.
func (g *GoTransformer) Transform(source, target AttributeAnalyzer, ta *TransformAttrs) (string, error) {
	return GoAttributeTransform(source, target, ta, g)
}

// TransformHelpers returns the transform functions required to transform
// source attribute to target attribute. It returns an error if source and
// target are incompatible.
func (g *GoTransformer) TransformHelpers(source, target AttributeAnalyzer) ([]*TransformFunctionData, error) {
	return GoTransformHelpers(source, target, g)
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
// prefix is the transformation helper function prefix
//
func GoTransform(source, target AttributeAnalyzer, sourceVar, targetVar, prefix string) (string, []*TransformFunctionData, error) {
	t := NewGoTransformer(prefix)
	return Transform(source, target, sourceVar, targetVar, t)
}

// GoAttributeTransform generates Go code to transform source attribute to
// target attribute.
//
// source, target are the source and target attributes
//
// ta is the transform attributes to assist in the transformation
//
// t is the Transfomer used in the transformation
//
func GoAttributeTransform(source, target AttributeAnalyzer, ta *TransformAttrs, t Transformer) (string, error) {
	var (
		err error

		sourceType = source.Attribute().Type
		targetType = target.Attribute().Type
	)
	if source, target, ta, err = t.MakeCompatible(source, target, ta, ""); err != nil {
		return "", err
	}

	var code string
	{
		switch {
		case expr.IsArray(sourceType):
			code, err = GoArrayTransform(
				expr.AsArray(sourceType),
				expr.AsArray(targetType),
				source, target, ta, t)
		case expr.IsMap(sourceType):
			code, err = GoMapTransform(
				expr.AsMap(sourceType),
				expr.AsMap(targetType),
				source, target, ta, t)
		case expr.IsObject(sourceType):
			code, err = GoObjectTransform(source, target, ta, t)
		default:
			assign := "="
			if ta.NewVar {
				assign = ":="
			}
			if _, ok := targetType.(expr.UserType); ok {
				// Primitive user type, these are used for error results
				cast := target.Ref(true)
				return fmt.Sprintf("%s %s %s(%s)\n", ta.TargetVar, assign, cast, ta.SourceVar), nil
			}
			srcField, _ := t.ConvertType(ta.SourceVar, sourceType)
			code = fmt.Sprintf("%s %s %s\n", ta.TargetVar, assign, srcField)
		}
	}
	if err != nil {
		return "", err
	}
	return code, nil
}

// GoObjectTransform generates Go code to transform source object to
// target object.
//
// source, target are the source and target attributes of object type
//
// ta is the transform attributes to assist in the transformation
//
// t is the Transfomer used in the transformation
//
func GoObjectTransform(source, target AttributeAnalyzer, ta *TransformAttrs, t Transformer) (string, error) {
	if t := source.Attribute().Type; !expr.IsObject(t) {
		return "", fmt.Errorf("source is not an object type: received %T", t)
	}
	if t := target.Attribute().Type; !expr.IsObject(t) {
		return "", fmt.Errorf("target is not an object type: received %T", t)
	}
	var (
		initCode     string
		postInitCode string
	)
	{
		// iterate through primitive attributes to initialize the struct
		walkMatches(source, target, func(srcMatt, tgtMatt *expr.MappedAttributeExpr, srcc, tgtc AttributeAnalyzer, n string) {
			if !expr.IsPrimitive(srcc.Attribute().Type) {
				return
			}
			srcField := ta.SourceVar + "." + srcc.Identifier(srcMatt.ElemName(n), true)
			tgtField := tgtc.Identifier(tgtMatt.ElemName(n), true)
			srcPtr := srcc.IsPointer()
			tgtPtr := tgtc.IsPointer()
			deref := ""
			if srcPtr && !tgtPtr {
				deref = "*"
				if !srcc.IsRequired() {
					srcFieldConv, _ := t.ConvertType("*"+srcField, srcc.Attribute().Type)
					postInitCode += fmt.Sprintf("if %s != nil {\n\t%s.%s = %s\n}\n", srcField, ta.TargetVar, tgtField, srcFieldConv)
					return
				}
			} else if !srcPtr && tgtPtr {
				deref = "&"
			}
			srcFieldConv, ok := t.ConvertType(srcField, srcc.Attribute().Type)
			if ok {
				// type conversion required. Add it in postinit code.
				tgtName := tgtc.Identifier(tgtMatt.ElemName(n), false)
				postInitCode += fmt.Sprintf("%sptr := %s\n%s.%s = %s%sptr\n", tgtName, srcFieldConv, ta.TargetVar, tgtField, deref, tgtName)
				return
			}
			initCode += fmt.Sprintf("\n%s: %s%s,", tgtField, deref, srcFieldConv)
		})
		if initCode != "" {
			initCode += "\n"
		}
	}

	buffer := &bytes.Buffer{}
	deref := "&"
	// if the target is a raw struct no need to return a pointer
	if _, ok := target.Attribute().Type.(*expr.Object); ok {
		deref = ""
	}
	assign := "="
	if ta.NewVar {
		assign = ":="
	}
	buffer.WriteString(fmt.Sprintf("%s %s %s%s{%s}\n", ta.TargetVar, assign, deref, target.Name(true), initCode))
	buffer.WriteString(postInitCode)

	// iterate through non-primitive attributes to initialize rest of the
	// struct fields
	var err error
	walkMatches(source, target, func(srcMatt, tgtMatt *expr.MappedAttributeExpr, srcc, tgtc AttributeAnalyzer, n string) {
		srccAtt := srcc.Attribute()
		tgtcAtt := tgtc.Attribute()
		if srcc, tgtc, ta, err = t.MakeCompatible(srcc, tgtc, ta, ""); err != nil {
			return
		}

		var (
			code string

			newTA = &TransformAttrs{
				SourceVar: ta.SourceVar + "." + srcc.Identifier(srcMatt.ElemName(n), true),
				TargetVar: ta.TargetVar + "." + tgtc.Identifier(tgtMatt.ElemName(n), true),
				NewVar:    false,
			}
		)
		{
			_, ok := srccAtt.Type.(expr.UserType)
			switch {
			case expr.IsArray(srccAtt.Type):
				code, err = GoArrayTransform(
					expr.AsArray(srccAtt.Type),
					expr.AsArray(tgtcAtt.Type),
					srcc, tgtc, newTA, t)
			case expr.IsMap(srccAtt.Type):
				code, err = GoMapTransform(
					expr.AsMap(srccAtt.Type),
					expr.AsMap(tgtcAtt.Type),
					srcc, tgtc, newTA, t)
			case ok:
				code = fmt.Sprintf("%s = %s(%s)\n", newTA.TargetVar, t.HelperName(srcc, tgtc), newTA.SourceVar)
			case expr.IsObject(srccAtt.Type):
				code, err = t.Transform(srcc, tgtc, newTA)
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
			checkNil = srcc.IsPointer()
			if !checkNil && !expr.IsPrimitive(srccAtt.Type) {
				if !srcc.IsRequired() && srcc.DefaultValue() == nil {
					checkNil = true
				}
			}
		}
		if code != "" && checkNil {
			code = fmt.Sprintf("if %s != nil {\n\t%s}\n", newTA.SourceVar, code)
		}

		// Default value handling. We need to handle default values if the target
		// type uses default values (i.e. attributes with default values are
		// non-pointers) and has a default value set.
		if tdef := tgtc.DefaultValue(); tdef != nil {
			if srcc.IsPointer() {
				code += fmt.Sprintf("if %s == nil {\n\t", newTA.SourceVar)
				if tgtc.IsPointer() {
					code += fmt.Sprintf("var tmp %s = %#v\n\t%s = &tmp\n", tgtc.Def(), tdef, newTA.TargetVar)
				} else {
					code += fmt.Sprintf("%s = %#v\n", newTA.TargetVar, tdef)
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

// GoArrayTransform generates Go code to transform source array to
// target array.
//
// sourceArr, targetArr are the source and target arrays
//
// source, target are the source and target analyzers
//
// ta is the transform attributes to assist in the transformation
//
// t is the Transfomer used in the transformation
//
func GoArrayTransform(sourceArr, targetArr *expr.Array, source, target AttributeAnalyzer, ta *TransformAttrs, t Transformer) (string, error) {
	var (
		err error
	)
	sourceElem := source.Dup(sourceArr.ElemType, true)
	targetElem := target.Dup(targetArr.ElemType, true)
	tgtElemRef := targetElem.Ref(true)
	if sourceElem, targetElem, ta, err = t.MakeCompatible(sourceElem, targetElem, ta, "[0]"); err != nil {
		return "", err
	}
	data := map[string]interface{}{
		"Transformer": t,
		"ElemTypeRef": tgtElemRef,
		"SourceElem":  sourceElem,
		"TargetElem":  targetElem,
		"SourceVar":   ta.SourceVar,
		"TargetVar":   ta.TargetVar,
		"NewVar":      ta.NewVar,
		"LoopVar":     string(105 + strings.Count(ta.TargetVar, "[")),
	}
	var buf bytes.Buffer
	if err := transformGoArrayT.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GoMapTransform generates Go code to transform source map to target map.
//
// sourceMap, targetMap are the source and target maps
//
// source, target are the source and target analyzers
//
// ta is the transform attributes to assist in the transformation
//
// t is the Transfomer used in the transformation
//
func GoMapTransform(sourceMap, targetMap *expr.Map, source, target AttributeAnalyzer, ta *TransformAttrs, t Transformer) (string, error) {
	var (
		err error
	)
	sourceKey := source.Dup(sourceMap.KeyType, true)
	targetKey := target.Dup(targetMap.KeyType, true)
	tgtKeyRef := targetKey.Ref(true)
	if sourceKey, targetKey, _, err = t.MakeCompatible(sourceKey, targetKey, ta, "[key]"); err != nil {
		return "", err
	}

	sourceElem := source.Dup(sourceMap.ElemType, true)
	targetElem := target.Dup(targetMap.ElemType, true)
	tgtElemRef := targetElem.Ref(true)
	if sourceElem, targetElem, ta, err = t.MakeCompatible(sourceElem, targetElem, ta, "[*]"); err != nil {
		return "", err
	}
	data := map[string]interface{}{
		"Transformer": t,
		"KeyTypeRef":  tgtKeyRef,
		"ElemTypeRef": tgtElemRef,
		"SourceKey":   sourceKey,
		"TargetKey":   targetKey,
		"SourceElem":  sourceElem,
		"TargetElem":  targetElem,
		"SourceVar":   ta.SourceVar,
		"TargetVar":   ta.TargetVar,
		"NewVar":      ta.NewVar,
		"LoopVar":     "",
	}
	if depth := mapDepth(targetMap); depth > 0 {
		data["LoopVar"] = string(97 + depth)
	}
	var buf bytes.Buffer
	if err := transformGoMapT.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GoTransformHelpers returns the Go transform functions required to transform
// source attribute to target attribute. It returns an error if source and
// target are incompatible.
//
// source, target are the source and target attributes used in transformation
//
// t is the transformer used in the transformation
//
// seen keeps track of generated transform functions to avoid recursion
//
func GoTransformHelpers(source, target AttributeAnalyzer, t Transformer, seen ...map[string]*TransformFunctionData) ([]*TransformFunctionData, error) {
	var (
		err error

		ta = &TransformAttrs{}
	)
	if source, target, ta, err = t.MakeCompatible(source, target, ta, ""); err != nil {
		return nil, err
	}

	var (
		helpers []*TransformFunctionData

		sourceType = source.Attribute().Type
		targetType = target.Attribute().Type
	)
	{
		// Do not generate a transform function for the top most user type.
		switch {
		case expr.IsArray(sourceType):
			source = source.Dup(expr.AsArray(sourceType).ElemType, true)
			target = target.Dup(expr.AsArray(targetType).ElemType, true)
			helpers, err = GoTransformHelpers(source, target, t, seen...)
		case expr.IsMap(sourceType):
			sm := expr.AsMap(sourceType)
			tm := expr.AsMap(targetType)
			source = source.Dup(sm.ElemType, true)
			target = target.Dup(tm.ElemType, true)
			helpers, err = GoTransformHelpers(source, target, t, seen...)
			if err == nil {
				var other []*TransformFunctionData
				source = source.Dup(sm.KeyType, true)
				target = target.Dup(tm.KeyType, true)
				other, err = GoTransformHelpers(source, target, t, seen...)
				helpers = AppendHelpers(helpers, other)
			}
		case expr.IsObject(sourceType):
			helpers, err = GoObjectTransformHelpers(source, target, t, seen...)
		}
	}
	if err != nil {
		return nil, err
	}
	return helpers, nil
}

// GoObjectTransformHelpers collects the helper functions required to transform
// source attribute of object type to target attribute of object type.
func GoObjectTransformHelpers(source, target AttributeAnalyzer, t Transformer, seen ...map[string]*TransformFunctionData) ([]*TransformFunctionData, error) {
	if t := source.Attribute().Type; !expr.IsObject(t) {
		return nil, fmt.Errorf("source is not an object type: received %T", t)
	}
	if t := target.Attribute().Type; !expr.IsObject(t) {
		return nil, fmt.Errorf("target is not an object type: received %T", t)
	}
	var (
		helpers []*TransformFunctionData
		err     error
	)
	walkMatches(source, target, func(srcMatt, tgtMatt *expr.MappedAttributeExpr, srcc, tgtc AttributeAnalyzer, n string) {
		if err != nil {
			return
		}
		h, err2 := collectHelpers(srcc, tgtc, t, seen...)
		if err2 != nil {
			err = err2
			return
		}
		helpers = AppendHelpers(helpers, h)
	})
	if err != nil {
		return nil, err
	}
	return helpers, nil
}

// collectHelpers recursively traverses the given attributes and return the
// transform helper functions required to generate the transform code.
func collectHelpers(source, target AttributeAnalyzer, t Transformer, seen ...map[string]*TransformFunctionData) ([]*TransformFunctionData, error) {
	var (
		data []*TransformFunctionData

		sourceType = source.Attribute().Type
		targetType = target.Attribute().Type
	)
	switch {
	case expr.IsArray(sourceType):
		source = source.Dup(expr.AsArray(sourceType).ElemType, true)
		target = target.Dup(expr.AsArray(targetType).ElemType, true)
		helpers, err := GoTransformHelpers(source, target, t, seen...)
		if err != nil {
			return nil, err
		}
		data = AppendHelpers(data, helpers)
	case expr.IsMap(sourceType):
		source = source.Dup(expr.AsMap(sourceType).KeyType, true)
		target = target.Dup(expr.AsMap(targetType).KeyType, true)
		helpers, err := GoTransformHelpers(source, target, t, seen...)
		if err != nil {
			return nil, err
		}
		data = AppendHelpers(data, helpers)
		source = source.Dup(expr.AsMap(sourceType).ElemType, true)
		target = target.Dup(expr.AsMap(targetType).ElemType, true)
		helpers, err = GoTransformHelpers(source, target, t, seen...)
		if err != nil {
			return nil, err
		}
		data = AppendHelpers(data, helpers)
	case expr.IsObject(sourceType):
		if ut, ok := sourceType.(expr.UserType); ok {
			name := t.HelperName(source, target)
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
			code, err := t.Transform(
				source.Dup(ut.Attribute(), true), target,
				&TransformAttrs{SourceVar: "v", TargetVar: "res", NewVar: true})
			if err != nil {
				return nil, err
			}
			if !source.IsRequired() {
				code = "if v == nil {\n\treturn nil\n}\n" + code
			}
			tfd := &TransformFunctionData{
				Name:          t.HelperName(source, target),
				ParamTypeRef:  source.Ref(true),
				ResultTypeRef: target.Ref(true),
				Code:          code,
			}
			s[name] = tfd
			data = AppendHelpers(data, []*TransformFunctionData{tfd})
		}

		// collect helpers
		var err error
		{
			walkMatches(source, target, func(srcMatt, _ *expr.MappedAttributeExpr, srcc, tgtc AttributeAnalyzer, n string) {
				var helpers []*TransformFunctionData
				helpers, err = collectHelpers(srcc, tgtc, t, seen...)
				if err != nil {
					return
				}
				data = AppendHelpers(data, helpers)
			})
		}
		if err != nil {
			return nil, err
		}
	}
	return data, nil
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

// walkMatches iterates through the source attribute expression and executes
// the walker function.
func walkMatches(source, target AttributeAnalyzer, walker func(src, tgt *expr.MappedAttributeExpr, srcc, tgtc AttributeAnalyzer, n string)) {
	srcMatt := expr.NewMappedAttributeExpr(source.Attribute())
	tgtMatt := expr.NewMappedAttributeExpr(target.Attribute())
	srcObj := expr.AsObject(srcMatt.Type)
	tgtObj := expr.AsObject(tgtMatt.Type)
	for _, nat := range *srcObj {
		if att := tgtObj.Attribute(nat.Name); att != nil {
			srcc := source.Dup(nat.Attribute, srcMatt.IsRequired(nat.Name))
			tgtc := target.Dup(att, tgtMatt.IsRequired(nat.Name))
			walker(srcMatt, tgtMatt, srcc, tgtc, nat.Name)
		}
	}
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

// used by template
func transformAttributeHelper(source, target AttributeAnalyzer, sourceVar, targetVar string, newVar bool, t Transformer) (string, error) {
	ta := &TransformAttrs{
		SourceVar: sourceVar,
		TargetVar: targetVar,
		NewVar:    newVar,
	}
	return t.Transform(source, target, ta)
}

const (
	transformGoArrayTmpl = `{{ .TargetVar }} {{ if .NewVar }}:={{ else }}={{ end }} make([]{{ .ElemTypeRef }}, len({{ .SourceVar }}))
for {{ .LoopVar }}, val := range {{ .SourceVar }} {
  {{ transformAttribute .SourceElem .TargetElem "val" (printf "%s[%s]" .TargetVar .LoopVar) false .Transformer -}}
}
`

	transformGoMapTmpl = `{{ .TargetVar }} {{ if .NewVar }}:={{ else }}={{ end }} make(map[{{ .KeyTypeRef }}]{{ .ElemTypeRef }}, len({{ .SourceVar }}))
for key, val := range {{ .SourceVar }} {
  {{ transformAttribute .SourceKey .TargetKey "key" "tk" true .Transformer -}}
  {{ transformAttribute .SourceElem .TargetElem "val" (printf "tv%s" .LoopVar) true .Transformer -}}
  {{ .TargetVar }}[tk] = {{ printf "tv%s" .LoopVar }}
}
`
)
