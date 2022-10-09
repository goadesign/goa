package codegen

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"goa.design/goa/v3/expr"
)

var transformGoArrayT, transformGoMapT, transformGoUnionT, transformGoUnionToObjectT, transformGoObjectToUnionT *template.Template

// NOTE: can't initialize inline because https://github.com/golang/go/issues/1817
func init() {
	fm := template.FuncMap{"transformAttribute": transformAttribute, "transformHelperName": transformHelperName}
	transformGoArrayT = template.Must(template.New("transformGoArray").Funcs(fm).Parse(transformGoArrayTmpl))
	transformGoMapT = template.Must(template.New("transformGoMap").Funcs(fm).Parse(transformGoMapTmpl))
	transformGoUnionT = template.Must(template.New("transformGoUnion").Funcs(fm).Parse(transformGoUnionTmpl))
	transformGoUnionToObjectT = template.Must(template.New("transformGoUnionToObject").Funcs(fm).Parse(transformGoUnionToObjectTmpl))
	transformGoObjectToUnionT = template.Must(template.New("transformGoObjectToUnion").Funcs(fm).Parse(transformGoObjectToUnionTmpl))
}

// GoTransform produces Go code that initializes the data structure defined
// by target from an instance of the data structure described by source.
// The data structures can be objects, arrays or maps. The algorithm
// matches object fields by name and ignores object fields in target that
// don't have a match in source. The matching and generated code leverage
// mapped attributes so that attribute names may use the "name:elem"
// syntax to define the name of the design attribute and the name of the
// corresponding generated Go struct field. The object field may also differ
// in that they may be pointers in one case and not the other. The function
// returns an error if target is not compatible with source (different type,
// fields of different type etc).
//
// As a special case GoTransform can map union types from and to object types
// with two attributes, one called "Value" which stores the value and one called
// "Type" which is of type string and contains the value type name (union types
// are otherwise implemented as a struct containing a single field: the current
// value - however having the kind explicitly stored is required to serialize to
// JSON for example).
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
// newVar if true initializes a target variable with the generated Go code
// using `:=` operator. If false, it assigns Go code to the target variable
// using `=`.
func GoTransform(source, target *expr.AttributeExpr, sourceVar, targetVar string, sourceCtx, targetCtx *AttributeContext, prefix string, newVar bool) (string, []*TransformFunctionData, error) {
	ta := &TransformAttrs{
		SourceCtx: sourceCtx,
		TargetCtx: targetCtx,
		Prefix:    prefix,
	}

	code, err := transformAttribute(source, target, sourceVar, targetVar, newVar, ta)
	if err != nil {
		return "", nil, err
	}

	funcs, err := transformAttributeHelpers(source, target, ta, make(map[string]*TransformFunctionData))
	if err != nil {
		return "", nil, err
	}

	return strings.TrimRight(code, "\n"), funcs, nil
}

// transformPrimitive returns the code to transform source primtive type to
// target primitive type. It returns an error if source and target are not
// compatible for transformation.
func transformPrimitive(source, target *expr.AttributeExpr, sourceVar, targetVar string, newVar bool, ta *TransformAttrs) (string, error) {
	if err := IsCompatible(source.Type, target.Type, sourceVar, targetVar); err != nil {
		return "", err
	}
	assign := "="
	if newVar {
		assign = ":="
	}
	if source.Type.Name() != target.Type.Name() {
		cast := ta.TargetCtx.Scope.Ref(target, ta.TargetCtx.Pkg(target))
		return fmt.Sprintf("%s %s %s(%s)\n", targetVar, assign, cast, sourceVar), nil
	}
	return fmt.Sprintf("%s %s %s\n", targetVar, assign, sourceVar), nil
}

// transformAttribute returns the code to transform source attribute to target
// attribute. It returns an error if source and target are not compatible for
// transformation.
func transformAttribute(source, target *expr.AttributeExpr, sourceVar, targetVar string, newVar bool, ta *TransformAttrs) (code string, err error) {
	if err = IsCompatible(source.Type, target.Type, sourceVar, targetVar); err != nil {
		return
	}
	switch {
	case expr.IsArray(source.Type):
		code, err = transformArray(expr.AsArray(source.Type), expr.AsArray(target.Type), sourceVar, targetVar, newVar, ta)
	case expr.IsMap(source.Type):
		code, err = transformMap(expr.AsMap(source.Type), expr.AsMap(target.Type), sourceVar, targetVar, newVar, ta)
	case expr.IsUnion(source.Type):
		code, err = transformUnion(source, target, sourceVar, targetVar, newVar, ta)
	case expr.IsObject(source.Type):
		code, err = transformObject(source, target, sourceVar, targetVar, newVar, ta)
	default:
		code, err = transformPrimitive(source, target, sourceVar, targetVar, newVar, ta)
	}
	return
}

// transformObject generates Go code to transform source object to target
// object.
func transformObject(source, target *expr.AttributeExpr, sourceVar, targetVar string, newVar bool, ta *TransformAttrs) (string, error) {
	var (
		initCode     string
		postInitCode string
		err          error
	)
	{
		if expr.IsUnion(target.Type) {
			return transformObjectToUnion(source, target, sourceVar, targetVar, newVar, ta)
		}

		// walk through primitives first to initialize the struct
		walkMatches(source, target, func(srcMatt, tgtMatt *expr.MappedAttributeExpr, srcc, tgtc *expr.AttributeExpr, n string) {
			if !expr.IsPrimitive(srcc.Type) {
				return
			}
			// Source and/or target could be primitive user type. Make sure the
			// aliased type is compatible for transformation.
			if err = IsCompatible(srcc.Type, tgtc.Type, sourceVar, targetVar); err != nil {
				return
			}
			var (
				exp string

				srcPtr     = ta.SourceCtx.IsPrimitivePointer(n, srcMatt.AttributeExpr)
				tgtPtr     = ta.TargetCtx.IsPrimitivePointer(n, tgtMatt.AttributeExpr)
				srcField   = sourceVar + "." + GoifyAtt(srcc, srcMatt.ElemName(n), true)
				tgtField   = GoifyAtt(tgtc, tgtMatt.ElemName(n), true)
				_, isSrcUT = srcc.Type.(expr.UserType)
				_, isTgtUT = tgtc.Type.(expr.UserType)
			)
			{
				switch {
				case isSrcUT || isTgtUT:
					deref := ""
					if srcPtr {
						deref = "*"
					}
					exp = fmt.Sprintf("%s(%s%s)", ta.TargetCtx.Scope.Ref(tgtc, ta.TargetCtx.Pkg(tgtc)), deref, srcField)
					if srcPtr && !srcMatt.IsRequired(n) {
						postInitCode += fmt.Sprintf("if %s != nil {\n", srcField)
						if tgtPtr {
							tmp := Goify(tgtMatt.ElemName(n), false)
							postInitCode += fmt.Sprintf("%s := %s\n%s.%s = &%s\n", tmp, exp, targetVar, tgtField, tmp)
						} else {
							postInitCode += fmt.Sprintf("%s.%s = %s\n", targetVar, tgtField, exp)
						}
						postInitCode += "}\n"
						return
					} else if tgtPtr {
						tmp := Goify(tgtMatt.ElemName(n), false)
						postInitCode += fmt.Sprintf("%s := %s\n%s.%s = &%s\n", tmp, exp, targetVar, tgtField, tmp)
						return
					}
				case srcPtr && !tgtPtr:
					exp = "*" + srcField
					if !srcMatt.IsRequired(n) {
						postInitCode += fmt.Sprintf("if %s != nil {\n\t%s.%s = %s\n}\n", srcField, targetVar, tgtField, exp)
						return
					}
				case !srcPtr && tgtPtr:
					exp = "&" + srcField
				default:
					exp = srcField
				}
			}
			initCode += fmt.Sprintf("\n%s: %s,", tgtField, exp)
		})
		if initCode != "" {
			initCode += "\n"
		}
	}
	if err != nil {
		return "", err
	}

	buffer := &bytes.Buffer{}
	deref := "&"
	assign := "="
	if newVar {
		assign = ":="
	}
	name := ta.TargetCtx.Scope.Name(target, ta.TargetCtx.Pkg(target), ta.TargetCtx.Pointer, ta.TargetCtx.UseDefault)
	buffer.WriteString(fmt.Sprintf("%s %s %s%s{%s}\n", targetVar, assign, deref, name, initCode))
	buffer.WriteString(postInitCode)

	// iterate through attributes to initialize rest of the struct fields and
	// handle default values
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
				code, err = transformArray(expr.AsArray(srcc.Type), expr.AsArray(tgtc.Type), srcVar, tgtVar, false, ta)
			case expr.IsMap(srcc.Type):
				code, err = transformMap(expr.AsMap(srcc.Type), expr.AsMap(tgtc.Type), srcVar, tgtVar, false, ta)
			case expr.IsUnion(srcc.Type):
				code, err = transformUnion(srcc, tgtc, srcVar, tgtVar, false, ta)
			case ok:
				if ta.TargetCtx.IsInterface {
					ref := ta.TargetCtx.Scope.Ref(target, ta.TargetCtx.Pkg(target))
					tgtVar = targetVar + ".(" + ref + ")." + GoifyAtt(tgtc, tgtMatt.ElemName(n), true)
				}
				if !expr.IsPrimitive(srcc.Type) {
					code = fmt.Sprintf("%s = %s(%s)\n", tgtVar, transformHelperName(srcc, tgtc, ta), srcVar)
				}
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
		if tdef := tgtMatt.GetDefault(n); tdef != nil && ta.TargetCtx.UseDefault && !ta.TargetCtx.Pointer && !srcMatt.IsRequired(n) {
			switch {
			case ta.SourceCtx.IsPrimitivePointer(n, srcMatt.AttributeExpr) || !expr.IsPrimitive(srcc.Type):
				// source attribute is a primitive pointer or not a primitive
				code += fmt.Sprintf("if %s == nil {\n\t", srcVar)
				if ta.TargetCtx.IsPrimitivePointer(n, tgtMatt.AttributeExpr) && expr.IsPrimitive(tgtc.Type) {
					code += fmt.Sprintf("var tmp %s = %#v\n\t%s = &tmp\n", GoNativeTypeName(tgtc.Type), tdef, tgtVar)
				} else {
					code += fmt.Sprintf("%s = %#v\n", tgtVar, tdef)
				}
				code += "}\n"
			case expr.IsPrimitive(srcc.Type) && srcMatt.HasDefaultValue(n) && ta.SourceCtx.UseDefault:
				// source attribute is a primitive with default value
				// (the field is not a pointer in this case)
				code += "{\n\t"
				if typeName, _ := GetMetaType(tgtc); typeName != "" {
					code += fmt.Sprintf("var zero %s\n\t", typeName)
				} else if _, ok := tgtc.Type.(expr.UserType); ok {
					// aliased primitive
					code += fmt.Sprintf("var zero %s\n\t", ta.TargetCtx.Scope.Ref(tgtc, ta.TargetCtx.Pkg(tgtc)))
				} else {
					code += fmt.Sprintf("var zero %s\n\t", GoNativeTypeName(tgtc.Type))
				}
				code += fmt.Sprintf("if %s == zero {\n\t%s = %#v\n}\n", tgtVar, tgtVar, tdef)
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
		"ElemTypeRef":    ta.TargetCtx.Scope.Ref(target.ElemType, ta.TargetCtx.Pkg(target.ElemType)),
		"SourceElem":     source.ElemType,
		"TargetElem":     target.ElemType,
		"SourceVar":      sourceVar,
		"TargetVar":      targetVar,
		"NewVar":         newVar,
		"TransformAttrs": ta,
		"LoopVar":        string(rune(105 + strings.Count(targetVar, "["))),
		"IsStruct":       expr.IsObject(target.ElemType.Type),
	}
	var buf bytes.Buffer
	if err := transformGoArrayT.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
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
		"KeyTypeRef":     ta.TargetCtx.Scope.Ref(target.KeyType, ta.TargetCtx.Pkg(target.KeyType)),
		"ElemTypeRef":    ta.TargetCtx.Scope.Ref(target.ElemType, ta.TargetCtx.Pkg(target.ElemType)),
		"SourceKey":      source.KeyType,
		"TargetKey":      target.KeyType,
		"SourceElem":     source.ElemType,
		"TargetElem":     target.ElemType,
		"SourceVar":      sourceVar,
		"TargetVar":      targetVar,
		"NewVar":         newVar,
		"TransformAttrs": ta,
		"LoopVar":        "",
		"IsKeyStruct":    expr.IsObject(target.KeyType.Type),
		"IsElemStruct":   expr.IsObject(target.ElemType.Type),
	}
	if depth := MapDepth(target); depth > 0 {
		data["LoopVar"] = string(rune(97 + depth))
	}
	var buf bytes.Buffer
	if err := transformGoMapT.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// transformUnion generates Go code to transform source union to target union.
//
// Note: transport to/from service transforms are always object to union or
// union to object. The only case a transform is union to union is when
// converting a projected type from/to a service type.
func transformUnion(source, target *expr.AttributeExpr, sourceVar, targetVar string, newVar bool, ta *TransformAttrs) (string, error) {
	if expr.IsObject(target.Type) {
		return transformUnionToObject(source, target, sourceVar, targetVar, newVar, ta)
	}
	srcUnion, tgtUnion := expr.AsUnion(source.Type), expr.AsUnion(target.Type)
	if len(srcUnion.Values) != len(tgtUnion.Values) {
		return "", fmt.Errorf("cannot transform union: number of union types differ (%s has %d, %s has %d)",
			source.Type.Name(), len(srcUnion.Values), target.Type.Name(), len(tgtUnion.Values))
	}
	for i, st := range srcUnion.Values {
		if err := IsCompatible(st.Attribute.Type, tgtUnion.Values[i].Attribute.Type, sourceVar, targetVar); err != nil {
			return "", fmt.Errorf("cannot transform union %s to %s: type at index %d: %w",
				source.Type.Name(), target.Type.Name(), i, err)
		}
	}
	sourceTypeRefs := make([]string, len(srcUnion.Values))
	for i, st := range srcUnion.Values {
		sourceTypeRefs[i] = ta.TargetCtx.Scope.Ref(st.Attribute, ta.SourceCtx.Pkg(st.Attribute))
	}
	targetTypeNames := make([]string, len(tgtUnion.Values))
	for i, tt := range tgtUnion.Values {
		targetTypeNames[i] = ta.TargetCtx.Scope.Name(tt.Attribute, ta.TargetCtx.Pkg(tt.Attribute), ta.TargetCtx.Pointer, ta.TargetCtx.Pointer)
	}

	// Need to type assert targetVar before assigning field values.
	ta.TargetCtx.IsInterface = true

	data := map[string]interface{}{
		"SourceTypeRefs": sourceTypeRefs,
		"SourceTypes":    srcUnion.Values,
		"TargetTypes":    tgtUnion.Values,
		"SourceVar":      sourceVar,
		"TargetVar":      targetVar,
		"NewVar":         newVar,
		"TypeRef":        ta.TargetCtx.Scope.Ref(target, ta.TargetCtx.Pkg(target)),
		"TargetTypeName": ta.TargetCtx.Scope.Name(target, ta.TargetCtx.Pkg(target), ta.TargetCtx.Pointer, ta.TargetCtx.UseDefault),
		"TransformAttrs": ta,
	}
	var buf bytes.Buffer
	if err := transformGoUnionT.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func transformUnionToObject(source, target *expr.AttributeExpr, sourceVar, targetVar string, newVar bool, ta *TransformAttrs) (string, error) {
	obj := expr.AsObject(target.Type)
	if (*obj)[0].Attribute.Type != expr.String {
		return "", fmt.Errorf("union to object transform requires first field to be string")
	}
	if (*obj)[1].Attribute.Type != expr.String {
		return "", fmt.Errorf("union to object transform requires second field to be string")
	}
	srcUnion := expr.AsUnion(source.Type)
	sourceTypeRefs := make([]string, len(srcUnion.Values))
	sourceTypeNames := make([]string, len(srcUnion.Values))
	for i, st := range srcUnion.Values {
		sourceTypeRefs[i] = ta.SourceCtx.Scope.Ref(st.Attribute, ta.SourceCtx.Pkg(st.Attribute))
		sourceTypeNames[i] = st.Name
	}
	data := map[string]interface{}{
		"NewVar":          newVar,
		"TargetVar":       targetVar,
		"TypeRef":         ta.TargetCtx.Scope.Ref(target, ta.TargetCtx.Pkg(target)),
		"SourceVar":       sourceVar,
		"SourceTypeRefs":  sourceTypeRefs,
		"SourceTypeNames": sourceTypeNames,
		"TargetTypeName":  ta.TargetCtx.Scope.Name(target, ta.TargetCtx.Pkg(target), ta.TargetCtx.Pointer, ta.TargetCtx.UseDefault),
	}
	var buf bytes.Buffer
	if err := transformGoUnionToObjectT.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func transformObjectToUnion(source, target *expr.AttributeExpr, sourceVar, targetVar string, newVar bool, ta *TransformAttrs) (string, error) {
	obj := expr.AsObject(source.Type)
	if (*obj)[0].Attribute.Type != expr.String {
		return "", fmt.Errorf("union to object transform requires first field to be string")
	}
	if (*obj)[1].Attribute.Type != expr.String {
		return "", fmt.Errorf("union to object transform requires second field to be string")
	}

	sourceVarDeref := sourceVar
	if ta.SourceCtx.Pointer {
		sourceVarDeref = "*" + sourceVar
	}
	tgtUnion := expr.AsUnion(target.Type)
	unionTypes := make([]string, len(tgtUnion.Values))
	targetTypeRefs := make([]string, len(tgtUnion.Values))
	for i, tt := range tgtUnion.Values {
		unionTypes[i] = tt.Name
		targetTypeRefs[i] = ta.TargetCtx.Scope.Ref(tt.Attribute, ta.TargetCtx.Pkg(tt.Attribute))
	}
	data := map[string]interface{}{
		"NewVar":         newVar,
		"TargetVar":      targetVar,
		"TypeRef":        ta.TargetCtx.Scope.Ref(target, ta.TargetCtx.Pkg(target)),
		"SourceVar":      sourceVar,
		"SourceVarDeref": sourceVarDeref,
		"UnionTypes":     unionTypes,
		"TargetTypeRefs": targetTypeRefs,
		"TargetTypeName": ta.TargetCtx.Scope.Name(target, ta.TargetCtx.Pkg(target), ta.TargetCtx.Pointer, ta.TargetCtx.UseDefault),
		"Pointer":        ta.SourceCtx.Pointer,
	}
	var buf bytes.Buffer
	if err := transformGoObjectToUnionT.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// transformAttributeHelpers returns the Go transform functions and their definitions
// that may be used in code produced by Transform. It returns an error if source and
// target are incompatible (different types, fields of different type etc).
// transformAttributeHelpers recurses through the attribute types and calls
// collectHelpers for each child attribute. collectHelpers actually produces the
// transform helper functions for the given attribute.
//
// source, target are the source and target attributes used in transformation
//
// ta holds the transform attributes
//
// seen keeps track of generated transform functions to avoid infinite recursion.
func transformAttributeHelpers(source, target *expr.AttributeExpr, ta *TransformAttrs, seen map[string]*TransformFunctionData) (helpers []*TransformFunctionData, err error) {
	// Do not generate a transform function for the top most user type.
	var other []*TransformFunctionData
	switch {
	case expr.IsArray(source.Type):
		if other, err = collectHelpers(expr.AsArray(source.Type).ElemType, expr.AsArray(target.Type).ElemType, true, ta, seen); err == nil {
			helpers = append(helpers, other...)
		}
	case expr.IsMap(source.Type):
		sm, tm := expr.AsMap(source.Type), expr.AsMap(target.Type)
		if other, err = collectHelpers(sm.ElemType, tm.ElemType, true, ta, seen); err == nil {
			helpers = append(helpers, other...)
			if other, err = collectHelpers(sm.KeyType, tm.KeyType, true, ta, seen); err == nil {
				helpers = append(helpers, other...)
			}
		}
	case expr.IsUnion(source.Type):
		tt := expr.AsUnion(target.Type)
		if tt == nil {
			return
		}
		for i, st := range expr.AsUnion(source.Type).Values {
			if other, err = collectHelpers(st.Attribute, tt.Values[i].Attribute, true, ta, seen); err == nil {
				helpers = append(helpers, other...)
			}
		}
	case expr.IsObject(source.Type):
		if expr.IsUnion(target.Type) {
			return
		}
		walkMatches(source, target, func(srcMatt, _ *expr.MappedAttributeExpr, srcc, tgtc *expr.AttributeExpr, n string) {
			if err != nil {
				return
			}
			if other, err = collectHelpers(srcc, tgtc, srcMatt.IsRequired(n), ta, seen); err == nil {
				helpers = append(helpers, other...)
			}
		})
	}
	return
}

// collectHelpers recurses through the given attributes and returns the transform
// helper functions required to generate the transform code. If the attributes type
// is array or map then the recursion is done via transformAttributeHelpers so that
// the tope level conversion function is skipped as the generate code does not make
// use of it (since it inlines that top-level transformation).
func collectHelpers(source, target *expr.AttributeExpr, req bool, ta *TransformAttrs, seen map[string]*TransformFunctionData) (helpers []*TransformFunctionData, err error) {
	name := transformHelperName(source, target, ta)
	if _, ok := seen[name]; ok {
		return
	}
	if _, ok := source.Type.(expr.UserType); ok && expr.IsObject(source.Type) {
		var h *TransformFunctionData
		if h, err = generateHelper(source, target, req, ta, seen); h != nil {
			helpers = append(helpers, h)
		}
	}
	var other []*TransformFunctionData
	switch {
	case expr.IsArray(source.Type):
		if other, err = collectHelpers(expr.AsArray(source.Type).ElemType, expr.AsArray(target.Type).ElemType, req, ta, seen); err == nil {
			helpers = append(helpers, other...)
		}
	case expr.IsMap(source.Type):
		sm, tm := expr.AsMap(source.Type), expr.AsMap(target.Type)
		if other, err = collectHelpers(sm.ElemType, tm.ElemType, req, ta, seen); err == nil {
			helpers = append(helpers, other...)
			if other, err = collectHelpers(sm.KeyType, tm.KeyType, req, ta, seen); err == nil {
				helpers = append(helpers, other...)
			}
		}
	case expr.IsUnion(source.Type):
		tt := expr.AsUnion(target.Type)
		if tt == nil {
			return
		}
		for i, st := range expr.AsUnion(source.Type).Values {
			if other, err = collectHelpers(st.Attribute, tt.Values[i].Attribute, req, ta, seen); err == nil {
				helpers = append(helpers, other...)
			}
		}
	case expr.IsObject(source.Type):
		if expr.IsUnion(target.Type) {
			return
		}
		walkMatches(source, target, func(srcMatt, _ *expr.MappedAttributeExpr, srcc, tgtc *expr.AttributeExpr, n string) {
			if err != nil {
				return
			}
			if other, err = collectHelpers(srcc, tgtc, srcMatt.IsRequired(n), ta, seen); err == nil {
				helpers = append(helpers, other...)
			}
		})
	}
	return
}

// generateHelper generates the code that transform instances of source into
// target. Both source and targe must be user types or generateHelper panics.
// generateHelper returns nil if a helper has already been generated for the
// pair source, target.
func generateHelper(source, target *expr.AttributeExpr, req bool, ta *TransformAttrs, seen map[string]*TransformFunctionData) (*TransformFunctionData, error) {
	name := transformHelperName(source, target, ta)
	if _, ok := seen[name]; ok {
		return nil, nil
	}

	// Reset need for type assertion for union types because we are
	// generating the code to transform the concrete type.
	ta.TargetCtx.IsInterface = false

	code, err := transformAttribute(source.Type.(expr.UserType).Attribute(), target, "v", "res", true, ta)
	if err != nil {
		return nil, err
	}
	if !req && !expr.IsPrimitive(source.Type) {
		code = "if v == nil {\n\treturn nil\n}\n" + code
	}
	tfd := &TransformFunctionData{
		Name:          name,
		ParamTypeRef:  ta.SourceCtx.Scope.Ref(source, ta.SourceCtx.Pkg(source)),
		ResultTypeRef: ta.TargetCtx.Scope.Ref(target, ta.TargetCtx.Pkg(target)),
		Code:          code,
	}
	seen[name] = tfd
	return tfd, nil
}

// walkMatches iterates through the attributes of source and looks for
// attributes with identical names in target. walkMatches calls the walker
// function for each pair of matched attributes. Both source and target must be
// objects or else walkMatches panics.
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
		sname = Goify(ta.SourceCtx.Scope.Name(source, ta.SourceCtx.Pkg(source), ta.SourceCtx.Pointer, ta.SourceCtx.UseDefault), true)
		tname = Goify(ta.TargetCtx.Scope.Name(target, ta.TargetCtx.Pkg(target), ta.TargetCtx.Pointer, ta.TargetCtx.UseDefault), true)
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
{{ if .IsStruct -}}
	{{ .TargetVar }}[{{ .LoopVar }}] = {{ transformHelperName .SourceElem .TargetElem .TransformAttrs }}(val)
{{ else -}}
	{{ transformAttribute .SourceElem .TargetElem "val" (printf "%s[%s]" .TargetVar .LoopVar) false .TransformAttrs -}}
{{ end -}}
}
`

	transformGoMapTmpl = `{{ .TargetVar }} {{ if .NewVar }}:={{ else }}={{ end }} make(map[{{ .KeyTypeRef }}]{{ .ElemTypeRef }}, len({{ .SourceVar }}))
for key, val := range {{ .SourceVar }} {
{{ if .IsKeyStruct -}}
	tk := {{ transformHelperName .SourceKey .TargetKey .TransformAttrs -}}(val)
{{ else -}}
  {{ transformAttribute .SourceKey .TargetKey "key" "tk" true .TransformAttrs -}}
{{ end -}}
{{ if .IsElemStruct -}}
	{{ .TargetVar }}[tk] = {{ transformHelperName .SourceElem .TargetElem .TransformAttrs -}}(val)
{{ else -}}
	{{ transformAttribute .SourceElem .TargetElem "val" (printf "tv%s" .LoopVar) true .TransformAttrs -}}
	{{ .TargetVar }}[tk] = {{ printf "tv%s" .LoopVar -}}
{{ end -}}
}
`

	transformGoUnionTmpl = `{{ if .NewVar }}var {{ .TargetVar }} {{ .TypeRef }}
{{ end }}switch actual := {{ .SourceVar }}.(type) {
	{{- range $i, $ref := .SourceTypeRefs }}
	case {{ $ref }}:
		{{- transformAttribute (index $.SourceTypes $i).Attribute (index $.TargetTypes $i).Attribute "actual" $.TargetVar false $.TransformAttrs -}}
	{{- end }}
}
`

	transformGoUnionToObjectTmpl = `{{ if .NewVar }}var {{ .TargetVar }} {{ .TypeRef }}
{{ end }}js, _ := json.Marshal({{ .SourceVar }})
var name string
switch {{ .SourceVar }}.(type) {
	{{- range $i, $ref := .SourceTypeRefs }}
	case {{ $ref }}:
		name = {{ printf "%q" (index $.SourceTypeNames $i) }}
	{{- end }}
}
{{ .TargetVar }} = &{{ .TargetTypeName }}{
	Type: name,
	Value: string(js),
}
`

	transformGoObjectToUnionTmpl = `{{ if .NewVar }}var {{ .TargetVar }} {{ .TypeRef }}
{{ end }}switch {{ .SourceVarDeref }}.Type {
	{{- range $i, $name := .UnionTypes }}
	case {{ printf "%q" $name }}:
		var val {{ index $.TargetTypeRefs $i }}
		json.Unmarshal([]byte({{ if $.Pointer }}*{{ end }}{{ $.SourceVar }}.Value), &val)
		{{ $.TargetVar }} = val
	{{- end }}
}
`
)
