package codegen

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"goa.design/goa/expr"
)

var (
	// transformGoArrayT is the template to generate Go array transformation
	// code.
	transformGoArrayT *template.Template
	// TransformGoMapT is the template to generate Go map transformation
	// code.
	transformGoMapT *template.Template
)

// NOTE: can't initialize inline because https://github.com/golang/go/issues/1817
func init() {
	transformGoArrayT = template.Must(template.New("transformGoArray").Funcs(template.FuncMap{
		"transformAttribute": transformAttributeHelper,
		"loopVar":            arrayLoopVar,
	}).Parse(transformGoArrayTmpl))
	transformGoMapT = template.Must(template.New("transformGoMap").Funcs(template.FuncMap{
		"transformAttribute": transformAttributeHelper,
		"loopVar":            mapLoopVar,
	}).Parse(transformGoMapTmpl))
}

type (
	// GoAttribute represents an attribute type that produces Go code.
	GoAttribute struct {
		// Attribute is the underlying attribute expression.
		Attribute *expr.AttributeExpr
		// Pkg is the package name where the attribute type exists.
		Pkg string
		// NameScope is the named scope to produce unique reference to the attribute.
		NameScope *NameScope
	}

	// goTransformer is a Transformer that generates Go code for converting a
	// data structure represented as an attribute expression into a different data
	// structure also represented as an attribute expression.
	goTransformer struct {
		// helperPrefix is the prefix for the helper functions generated during
		// the transformation. The helper functions are named based on this
		// pattern - <helperPrefix><SourceTypeName>To<TargetTypeName>. If no prefix
		// specified, "transform" is used as a prefix by default.
		helperPrefix string
	}
)

// NewGoAttribute returns an attribute that produces Go code.
func NewGoAttribute(att *expr.AttributeExpr, pkg string, scope *NameScope) Attributor {
	return &GoAttribute{
		Attribute: att,
		Pkg:       pkg,
		NameScope: scope,
	}
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
func GoTransform(source, target *ContextualAttribute, sourceVar, targetVar, prefix string) (string, []*TransformFunctionData, error) {
	t := &goTransformer{helperPrefix: prefix}

	code, err := t.Transform(source, target, &TransformAttrs{SourceVar: sourceVar, TargetVar: targetVar, NewVar: true})
	if err != nil {
		return "", nil, err
	}

	funcs, err := GoTransformHelpers(source, target, t, prefix)
	if err != nil {
		return "", nil, err
	}

	return strings.TrimRight(code, "\n"), funcs, nil
}

// GoObjectTransform produces Go code that initializes the data structure
// defined by target object type from an instance of the data structure
// defined by source object type. The algorithm matches object fields by
// name and ignores object fields in target that don't have a match in source.
// The matching and generated code leverage mapped attributes so that attribute
// names may use the "name:elem" syntax to define the name of the design
// attribute and the name of the corresponding generated Go struct field.
// The function returns an error if source or target are not object types
// or has fields of different types.
//
// source and target are the attributes of object type used in the
// transformation
//
// ta is the transform attributes used in the transformation code
//
// t is the transformer used to transform source to target
//
func GoObjectTransform(source, target *ContextualAttribute, ta *TransformAttrs, t Transformer) (string, error) {
	if t := source.Attribute.Expr().Type; !expr.IsObject(t) {
		return "", fmt.Errorf("source is not an object type: received %T", t)
	}
	if t := target.Attribute.Expr().Type; !expr.IsObject(t) {
		return "", fmt.Errorf("target is not an object type: received %T", t)
	}
	var (
		initCode     string
		postInitCode string
	)
	{
		// iterate through primitive attributes to initialize the struct
		walkMatches(source, target, func(srcMatt, tgtMatt *expr.MappedAttributeExpr, srcc, tgtc *ContextualAttribute, n string) {
			if !expr.IsPrimitive(srcc.Attribute.Expr().Type) {
				return
			}
			srcField := ta.SourceVar + "." + srcc.Attribute.Field(srcMatt.ElemName(n), true)
			tgtField := tgtc.Attribute.Field(tgtMatt.ElemName(n), true)
			srcPtr := srcc.IsPointer()
			tgtPtr := tgtc.IsPointer()
			srcFieldConv := t.ConvertType(srcc.Attribute, tgtc.Attribute, srcField)
			switch {
			case srcPtr && !tgtPtr:
				srcFieldConv = t.ConvertType(srcc.Attribute, tgtc.Attribute, "*"+srcField)
				if !srcc.Required {
					postInitCode += fmt.Sprintf("if %s != nil {\n\t%s.%s = %s\n}\n", srcField, ta.TargetVar, tgtField, srcFieldConv)
					return
				}
			case !srcPtr && tgtPtr:
				if srcField != srcFieldConv {
					// type conversion required. Add it in postinit code.
					tgtName := tgtc.Attribute.Field(tgtMatt.ElemName(n), false)
					postInitCode += fmt.Sprintf("%sptr := %s\n%s.%s = &%sptr\n", tgtName, srcFieldConv, ta.TargetVar, tgtField, tgtName)
					return
				}
				srcFieldConv = fmt.Sprintf("&%s", srcField)
			case srcPtr && tgtPtr:
				srcFieldConv = t.ConvertType(srcc.Attribute, tgtc.Attribute, "*"+srcField)
				if "*"+srcField != srcFieldConv {
					// type conversion required. Add it in postinit code.
					tgtName := tgtc.Attribute.Field(tgtMatt.ElemName(n), false)
					postInitCode += fmt.Sprintf("%sptr := %s\n%s.%s = &%sptr\n", tgtName, srcFieldConv, ta.TargetVar, tgtField, tgtName)
					return
				}
				srcFieldConv = srcField
			}
			initCode += fmt.Sprintf("\n%s: %s,", tgtField, srcFieldConv)
		})
		if initCode != "" {
			initCode += "\n"
		}
	}

	buffer := &bytes.Buffer{}
	deref := "&"
	// if the target is a raw struct no need to return a pointer
	if _, ok := target.Attribute.Expr().Type.(*expr.Object); ok {
		deref = ""
	}
	assign := "="
	if ta.NewVar {
		assign = ":="
	}
	buffer.WriteString(fmt.Sprintf("%s %s %s%s{%s}\n", ta.TargetVar, assign, deref, target.Attribute.Name(), initCode))
	buffer.WriteString(postInitCode)

	// iterate through non-primitive attributes to initialize rest of the
	// struct fields
	var err error
	walkMatches(source, target, func(srcMatt, tgtMatt *expr.MappedAttributeExpr, srcc, tgtc *ContextualAttribute, n string) {
		var (
			code string

			newTA = &TransformAttrs{
				SourceVar: ta.SourceVar + "." + srcc.Attribute.Field(srcMatt.ElemName(n), true),
				TargetVar: ta.TargetVar + "." + tgtc.Attribute.Field(tgtMatt.ElemName(n), true),
				NewVar:    false,
			}
		)
		{
			if srcc, tgtc, newTA, err = t.MakeCompatible(srcc, tgtc, newTA, ""); err != nil {
				return
			}
			srccAtt := srcc.Attribute.Expr()
			_, ok := srccAtt.Type.(expr.UserType)
			switch {
			case expr.IsArray(srccAtt.Type):
				code, err = t.TransformArray(srcc, tgtc, newTA)
			case expr.IsMap(srccAtt.Type):
				code, err = t.TransformMap(srcc, tgtc, newTA)
			case ok:
				code = fmt.Sprintf("%s = %s\n", newTA.TargetVar, t.ConvertType(srcc.Attribute, tgtc.Attribute, newTA.SourceVar))
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
			if !checkNil && !expr.IsPrimitive(srcc.Attribute.Expr().Type) {
				if !srcc.Required && srcc.DefaultValue() == nil {
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

// GoTransformHelpers returns the Go transform functions and their definitions
// that may be used in code produced by Transform. It returns an error if source and
// target are incompatible (different types, fields of different type etc).
//
// source, target are the source and target attributes used in transformation
//
// t is the transformer used in the transformation
//
// prefix is the function name prefix
//
// seen keeps track of generated transform functions to avoid recursion
//
func GoTransformHelpers(source, target *ContextualAttribute, t Transformer, prefix string, seen ...map[string]*TransformFunctionData) ([]*TransformFunctionData, error) {
	var (
		err error

		ta = &TransformAttrs{}
	)
	if source, target, ta, err = t.MakeCompatible(source, target, ta, ""); err != nil {
		return nil, err
	}

	var (
		helpers []*TransformFunctionData

		sourceType = source.Attribute.Expr().Type
		targetType = target.Attribute.Expr().Type
	)
	{
		// Do not generate a transform function for the top most user type.
		switch {
		case expr.IsArray(sourceType):
			source = source.Dup(expr.AsArray(sourceType).ElemType, true)
			target = target.Dup(expr.AsArray(targetType).ElemType, true)
			helpers, err = GoTransformHelpers(source, target, t, prefix, seen...)
		case expr.IsMap(sourceType):
			sm := expr.AsMap(sourceType)
			tm := expr.AsMap(targetType)
			source = source.Dup(sm.ElemType, true)
			target = target.Dup(tm.ElemType, true)
			helpers, err = GoTransformHelpers(source, target, t, prefix, seen...)
			if err == nil {
				var other []*TransformFunctionData
				source = source.Dup(sm.KeyType, true)
				target = target.Dup(tm.KeyType, true)
				other, err = GoTransformHelpers(source, target, t, prefix, seen...)
				helpers = append(helpers, other...)
			}
		case expr.IsObject(sourceType):
			walkMatches(source, target, func(srcMatt, tgtMatt *expr.MappedAttributeExpr, srcc, tgtc *ContextualAttribute, n string) {
				if err != nil {
					return
				}
				if srcc, tgtc, ta, err = t.MakeCompatible(srcc, tgtc, ta, ""); err != nil {
					return
				}
				h, err2 := collectHelpers(srcc, tgtc, t, prefix, seen...)
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

// Name returns a valid Go type name for the attribute.
func (g *GoAttribute) Name() string {
	return g.NameScope.GoFullTypeName(g.Attribute, g.Pkg)
}

// Ref returns a valid Go reference to the attribute.
func (g *GoAttribute) Ref() string {
	return g.NameScope.GoFullTypeRef(g.Attribute, g.Pkg)
}

// Scope returns the name scope.
func (g *GoAttribute) Scope() *NameScope {
	return g.NameScope
}

// Expr returns the underlying attribute expression.
func (g *GoAttribute) Expr() *expr.AttributeExpr {
	return g.Attribute
}

// Dup creates a copy of GoAttribute by setting the underlying attribute
// expression.
func (g *GoAttribute) Dup(att *expr.AttributeExpr) Attributor {
	return &GoAttribute{Attribute: att, Pkg: g.Pkg, NameScope: g.NameScope}
}

// Field returns a valid Go field name for the attribute.
func (g *GoAttribute) Field(name string, firstUpper bool) string {
	return GoifyAtt(g.Attribute, name, firstUpper)
}

// Def returns a valid Go definition for the attribute.
func (g *GoAttribute) Def(pointer, useDefault bool) string {
	return g.NameScope.GoTypeDef(g.Attribute, pointer, useDefault)
}

// MakeCompatible checks if target can be transformed to source.
func (g *goTransformer) MakeCompatible(source, target *ContextualAttribute, ta *TransformAttrs, suffix string) (src, tgt *ContextualAttribute, newTA *TransformAttrs, err error) {
	if err = IsCompatible(source.Attribute.Expr().Type, target.Attribute.Expr().Type, ta.SourceVar+suffix, ta.TargetVar+suffix); err != nil {
		return source, target, ta, err
	}
	return source, target, ta, nil
}

// ConvertType produces code to initialize a target type from a source type
// held by sourceVar.
func (g *goTransformer) ConvertType(source, target Attributor, sourceVar string) string {
	if _, ok := source.Expr().Type.(expr.UserType); ok {
		// return a function name for the conversion
		return fmt.Sprintf("%s(%s)", HelperName(source, target, g.helperPrefix), sourceVar)
	}
	// source and target Go types produced by goa are the same kind.
	// Hence no type conversion necessary.
	return sourceVar
}

// Transform returns the code to transform source attribute to
// target attribute. It returns an error if source and target are not
// compatible for transformation.
func (g *goTransformer) Transform(source, target *ContextualAttribute, ta *TransformAttrs) (string, error) {
	var (
		err error

		sourceType = source.Attribute.Expr().Type
		targetType = target.Attribute.Expr().Type
	)
	{
		if err = IsCompatible(sourceType, targetType, ta.SourceVar, ta.TargetVar); err != nil {
			return "", err
		}
	}

	var code string
	{
		switch {
		case expr.IsArray(sourceType):
			code, err = g.TransformArray(source, target, ta)
		case expr.IsMap(sourceType):
			code, err = g.TransformMap(source, target, ta)
		case expr.IsObject(sourceType):
			code, err = g.TransformObject(source, target, ta)
		default:
			assign := "="
			if ta.NewVar {
				assign = ":="
			}
			if _, ok := target.Attribute.Expr().Type.(expr.UserType); ok {
				// Primitive user type, these are used for error results
				cast := target.Attribute.Ref()
				return fmt.Sprintf("%s %s %s(%s)\n", ta.TargetVar, assign, cast, ta.SourceVar), nil
			}
			srcField := g.ConvertType(source.Attribute, target.Attribute, ta.SourceVar)
			code = fmt.Sprintf("%s %s %s\n", ta.TargetVar, assign, srcField)
		}
	}
	if err != nil {
		return "", err
	}
	return code, nil
}

// TransformObject generates Go code to transform source object to
// target object.
//
// source, target are the source and target attributes of object type
//
// ta is the transform attributes to assist in the transformation
//
func (g *goTransformer) TransformObject(source, target *ContextualAttribute, ta *TransformAttrs) (string, error) {
	return GoObjectTransform(source, target, ta, g)
}

// TransformArray generates Go code to transform source array to
// target array.
//
// source, target are the source and target analyzers of array type
//
// ta is the transform attributes to assist in the transformation
//
func (g *goTransformer) TransformArray(source, target *ContextualAttribute, ta *TransformAttrs) (string, error) {
	sourceArr := expr.AsArray(source.Attribute.Expr().Type)
	if sourceArr == nil {
		return "", fmt.Errorf("source is not an array type: received %T", source.Attribute.Expr().Type)
	}
	targetArr := expr.AsArray(target.Attribute.Expr().Type)
	if targetArr == nil {
		return "", fmt.Errorf("target is not an array type: received %T", target.Attribute.Expr().Type)
	}

	source = source.Dup(sourceArr.ElemType, true)
	target = target.Dup(targetArr.ElemType, true)
	if err := IsCompatible(source.Attribute.Expr().Type, target.Attribute.Expr().Type, ta.SourceVar+"[0]", ta.TargetVar+"[0]"); err != nil {
		return "", err
	}
	data := map[string]interface{}{
		"Transformer": g,
		"ElemTypeRef": target.Attribute.Ref(),
		"SourceElem":  source,
		"TargetElem":  target,
		"SourceVar":   ta.SourceVar,
		"TargetVar":   ta.TargetVar,
		"NewVar":      ta.NewVar,
	}
	return RunGoArrayTemplate(data)
}

// RunGoArrayTemplate runs the template to generate Go array code.
func RunGoArrayTemplate(data map[string]interface{}) (string, error) {
	var buf bytes.Buffer
	if err := transformGoArrayT.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// TransformMap generates Go code to transform source map to target map.
//
// source, target are the source and target analyzers
//
// ta is the transform attributes to assist in the transformation
//
// t is the Transfomer used in the transformation
//
func (g *goTransformer) TransformMap(source, target *ContextualAttribute, ta *TransformAttrs) (string, error) {
	sourceMap := expr.AsMap(source.Attribute.Expr().Type)
	if sourceMap == nil {
		return "", fmt.Errorf("source is not a map type: received %T", source.Attribute.Expr().Type)
	}
	targetMap := expr.AsMap(target.Attribute.Expr().Type)
	if targetMap == nil {
		return "", fmt.Errorf("target is not a map type: received %T", target.Attribute.Expr().Type)
	}

	sourceKey := source.Dup(sourceMap.KeyType, true)
	targetKey := target.Dup(targetMap.KeyType, true)
	if err := IsCompatible(sourceKey.Attribute.Expr().Type, targetKey.Attribute.Expr().Type, ta.SourceVar+"[key]", ta.TargetVar+"[key]"); err != nil {
		return "", err
	}
	sourceElem := source.Dup(sourceMap.ElemType, true)
	targetElem := target.Dup(targetMap.ElemType, true)
	if err := IsCompatible(sourceElem.Attribute.Expr().Type, targetElem.Attribute.Expr().Type, ta.SourceVar+"[*]", ta.TargetVar+"[*]"); err != nil {
		return "", err
	}
	data := map[string]interface{}{
		"Transformer": g,
		"KeyTypeRef":  targetKey.Attribute.Ref(),
		"ElemTypeRef": targetElem.Attribute.Ref(),
		"SourceKey":   sourceKey,
		"TargetKey":   targetKey,
		"SourceElem":  sourceElem,
		"TargetElem":  targetElem,
		"SourceVar":   ta.SourceVar,
		"TargetVar":   ta.TargetVar,
		"NewVar":      ta.NewVar,
		"TargetMap":   targetMap,
	}
	return RunGoMapTemplate(data)
}

// RunGoMapTemplate runs the template to generate Go map code.
func RunGoMapTemplate(data map[string]interface{}) (string, error) {
	var buf bytes.Buffer
	if err := transformGoMapT.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// collectHelpers recursively traverses the given attributes and return the
// transform helper functions required to generate the transform code.
func collectHelpers(source, target *ContextualAttribute, t Transformer, prefix string, seen ...map[string]*TransformFunctionData) ([]*TransformFunctionData, error) {
	var (
		data []*TransformFunctionData

		sourceType = source.Attribute.Expr().Type
		targetType = target.Attribute.Expr().Type
	)
	switch {
	case expr.IsArray(sourceType):
		source = source.Dup(expr.AsArray(sourceType).ElemType, true)
		target = target.Dup(expr.AsArray(targetType).ElemType, true)
		helpers, err := GoTransformHelpers(source, target, t, prefix, seen...)
		if err != nil {
			return nil, err
		}
		data = append(data, helpers...)
	case expr.IsMap(sourceType):
		source = source.Dup(expr.AsMap(sourceType).KeyType, true)
		target = target.Dup(expr.AsMap(targetType).KeyType, true)
		helpers, err := GoTransformHelpers(source, target, t, prefix, seen...)
		if err != nil {
			return nil, err
		}
		data = append(data, helpers...)
		source = source.Dup(expr.AsMap(sourceType).ElemType, true)
		target = target.Dup(expr.AsMap(targetType).ElemType, true)
		helpers, err = GoTransformHelpers(source, target, t, prefix, seen...)
		if err != nil {
			return nil, err
		}
		data = append(data, helpers...)
	case expr.IsObject(sourceType):
		if ut, ok := sourceType.(expr.UserType); ok {
			name := HelperName(source.Attribute, target.Attribute, prefix)
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
			if !source.Required {
				code = "if v == nil {\n\treturn nil\n}\n" + code
			}
			tfd := &TransformFunctionData{
				Name:          name,
				ParamTypeRef:  source.Attribute.Ref(),
				ResultTypeRef: target.Attribute.Ref(),
				Code:          code,
			}
			s[name] = tfd
			data = append(data, tfd)
		}

		// collect helpers
		var err error
		{
			walkMatches(source, target, func(srcMatt, _ *expr.MappedAttributeExpr, srcc, tgtc *ContextualAttribute, n string) {
				var helpers []*TransformFunctionData
				helpers, err = collectHelpers(srcc, tgtc, t, prefix, seen...)
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

// walkMatches iterates through the source attribute expression and executes
// the walker function.
func walkMatches(source, target *ContextualAttribute, walker func(src, tgt *expr.MappedAttributeExpr, srcc, tgtc *ContextualAttribute, n string)) {
	srcMatt := expr.NewMappedAttributeExpr(source.Attribute.Expr())
	tgtMatt := expr.NewMappedAttributeExpr(target.Attribute.Expr())
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

// used by template
func transformAttributeHelper(source, target *ContextualAttribute, sourceVar, targetVar string, newVar bool, t Transformer) (string, error) {
	ta := &TransformAttrs{
		SourceVar: sourceVar,
		TargetVar: targetVar,
		NewVar:    newVar,
	}
	return t.Transform(source, target, ta)
}

// used by template
func arrayLoopVar(s string) string {
	return string(105 + strings.Count(s, "["))
}

// used by template
func mapLoopVar(mp *expr.Map) string {
	if depth := mapDepth(mp); depth > 0 {
		return string(97 + depth)
	}
	return ""
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

const (
	transformGoArrayTmpl = `{{ .TargetVar }} {{ if .NewVar }}:={{ else }}={{ end }} make([]{{ .ElemTypeRef }}, len({{ .SourceVar }}))
{{- $loopVar := loopVar .TargetVar }}
for {{ $loopVar }}, val := range {{ .SourceVar }} {
  {{ transformAttribute .SourceElem .TargetElem "val" (printf "%s[%s]" .TargetVar $loopVar) false .Transformer -}}
}
`

	transformGoMapTmpl = `{{ .TargetVar }} {{ if .NewVar }}:={{ else }}={{ end }} make(map[{{ .KeyTypeRef }}]{{ .ElemTypeRef }}, len({{ .SourceVar }}))
{{- $loopVar := loopVar .TargetMap }}
for key, val := range {{ .SourceVar }} {
  {{ transformAttribute .SourceKey .TargetKey "key" "tk" true .Transformer -}}
  {{ transformAttribute .SourceElem .TargetElem "val" (printf "tv%s" $loopVar) true .Transformer -}}
  {{ .TargetVar }}[tk] = {{ printf "tv%s" $loopVar }}
}
`
)
