// This file generates Go transformations between compatible design types.
// Recursive helpers carry each side's package owner through nested named
// declarations so emitted references select the planned Go package.
package codegen

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"text/template"

	"goa.design/goa/v3/expr"
)

var transformGoArrayT, transformGoMapT, transformGoUnionT *template.Template

// NOTE: can't initialize inline because https://github.com/golang/go/issues/1817
func init() {
	fm := template.FuncMap{
		"transformAttribute":  TransformAttribute,
		"transformHelperName": TransformHelperName,
	}
	transformGoArrayT = template.Must(template.New("transformGoArray").Funcs(fm).Parse(codegenTemplates.Read(transformGoArrayTmplName)))
	transformGoMapT = template.Must(template.New("transformGoMap").Funcs(fm).Parse(codegenTemplates.Read(transformGoMapTmplName)))
	transformGoUnionT = template.Must(template.New("transformGoUnion").Funcs(fm).Parse(codegenTemplates.Read(transformGoUnionTmplName)))
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
	return GoTransformWithAttrs(source, target, sourceVar, targetVar, &TransformAttrs{
		SourceCtx: sourceCtx,
		TargetCtx: targetCtx,
		Prefix:    prefix,
	}, newVar)
}

// GoTransformWithAttrs is GoTransform with a caller built TransformAttrs. It
// lets generators customize the transformation via TransformAttrs.Hooks, see
// TransformHooks.
func GoTransformWithAttrs(source, target *expr.AttributeExpr, sourceVar, targetVar string, ta *TransformAttrs, newVar bool) (string, []*TransformFunctionData, error) {
	code, err := TransformAttribute(source, target, sourceVar, targetVar, newVar, ta)
	if err != nil {
		return "", nil, err
	}

	funcs, err := collectHelpers(source, target, true, true, ta, make(map[string]*TransformFunctionData))
	if err != nil {
		return "", nil, err
	}

	return strings.TrimRight(code, "\n"), funcs, nil
}

// TransformAttribute returns the code to transform source attribute to target
// attribute. It returns an error if source and target are not compatible for
// transformation. It is exported so that TransformHooks implementations can
// recurse into the transform engine from the code they render.
func TransformAttribute(source, target *expr.AttributeExpr, sourceVar, targetVar string, newVar bool, ta *TransformAttrs) (string, error) {
	var prelude string
	if h := ta.Hooks; h != nil && h.UnwrapPair != nil {
		var dir *WrapDirective
		source, target, dir = h.UnwrapPair(source, target)
		prelude = dir.apply(&sourceVar, &targetVar, &newVar)
	}
	ta = enterTransformAttrs(source, target, ta)
	if err := IsCompatible(source.Type, target.Type, sourceVar, targetVar); err != nil {
		return "", err
	}
	var (
		code string
		err  error
	)
	switch {
	case expr.IsArray(source.Type):
		if h := ta.Hooks; h != nil && h.TransformArray != nil {
			code, err = h.TransformArray(expr.AsArray(source.Type), expr.AsArray(target.Type), sourceVar, targetVar, newVar, ta)
		} else {
			code, err = transformArray(expr.AsArray(source.Type), expr.AsArray(target.Type), sourceVar, targetVar, newVar, ta)
		}
	case expr.IsMap(source.Type):
		if h := ta.Hooks; h != nil && h.TransformMap != nil {
			code, err = h.TransformMap(expr.AsMap(source.Type), expr.AsMap(target.Type), sourceVar, targetVar, newVar, ta)
		} else {
			code, err = transformMap(expr.AsMap(source.Type), expr.AsMap(target.Type), sourceVar, targetVar, newVar, ta)
		}
	case expr.IsUnion(source.Type):
		if h := ta.Hooks; h != nil && h.TransformUnion != nil {
			code, err = h.TransformUnion(source, target, sourceVar, targetVar, newVar, nil, nil, ta)
		} else {
			code, err = transformUnion(source, target, sourceVar, targetVar, newVar, ta)
		}
	case expr.IsObject(source.Type):
		code, err = transformObject(source, target, sourceVar, targetVar, newVar, ta)
	default:
		code = transformPrimitive(source, target, sourceVar, targetVar, newVar, ta)
	}
	if err != nil {
		return "", err
	}
	return prelude + code, nil
}

// TransformHelperName returns the transformation function name to initialize a
// target user type from an instance of a source user type. It is exported so
// that TransformHooks implementations can compute the names of the helper
// functions the engine collects.
func TransformHelperName(source, target *expr.AttributeExpr, ta *TransformAttrs) string {
	var (
		sname  string
		tname  string
		prefix string
	)
	{
		if h := ta.Hooks; h != nil && h.HelperNameAttrs != nil {
			source, target = h.HelperNameAttrs(source, target)
		}
		ta = enterTransformAttrs(source, target, ta)
		sname = Goify(ta.SourceCtx.Scope.Name(source, ta.SourceCtx.Pkg(source), ta.SourceCtx.Pointer, ta.SourceCtx.UseDefault), true)
		tname = Goify(ta.TargetCtx.Scope.Name(target, ta.TargetCtx.Pkg(target), ta.TargetCtx.Pointer, ta.TargetCtx.UseDefault), true)
		prefix = ta.Prefix
		if prefix == "" {
			prefix = "transform"
		}
	}
	return Goify(prefix+sname+"To"+tname, false)
}

// transformPrimitive returns the code to transform source primitive type to
// target primitive type. The caller (TransformAttribute) already verified that
// source and target are compatible.
func transformPrimitive(source, target *expr.AttributeExpr, sourceVar, targetVar string, newVar bool, ta *TransformAttrs) string {
	assign := "="
	if newVar {
		assign = ":="
	}

	if h := ta.Hooks; h != nil && h.ConvertPrimitive != nil {
		if exp, ok := h.ConvertPrimitive(source, target, sourceVar, false, false, ta); ok {
			return fmt.Sprintf("%s %s %s\n", targetVar, assign, exp)
		}
	}
	srcRef := ta.SourceCtx.Scope.Ref(source, ta.SourceCtx.Pkg(source))
	tgtRef := ta.TargetCtx.Scope.Ref(target, ta.TargetCtx.Pkg(target))
	if srcRef != tgtRef {
		return fmt.Sprintf("%s %s %s(%s)\n", targetVar, assign, tgtRef, sourceVar)
	}
	return fmt.Sprintf("%s %s %s\n", targetVar, assign, sourceVar)
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
				srcField   = sourceVar + "." + ta.SourceCtx.Scope.Field(srcc, srcMatt.ElemName(n), true)
				tgtField   = ta.TargetCtx.Scope.Field(tgtc, tgtMatt.ElemName(n), true)
				_, isSrcUT = srcc.Type.(expr.UserType)
				_, isTgtUT = tgtc.Type.(expr.UserType)
			)
			{
				var (
					convExp string
					hasConv bool
				)
				if h := ta.Hooks; h != nil && h.ConvertPrimitive != nil {
					convExp, hasConv = h.ConvertPrimitive(srcc, tgtc, srcField, srcPtr, tgtPtr, ta)
				}
				switch {
				case hasConv && (isSrcUT || isTgtUT || convExp != srcField), !hasConv && (isSrcUT || isTgtUT):
					if hasConv {
						exp = convExp
					} else {
						deref := ""
						if srcPtr {
							deref = "*"
						}
						exp = fmt.Sprintf("%s(%s%s)", ta.TargetCtx.Scope.Ref(tgtc, ta.TargetCtx.Pkg(tgtc)), deref, srcField)
					}
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
	if h := ta.Hooks; h != nil && h.ObjectDeref != nil {
		if d, ok := h.ObjectDeref(target); ok {
			deref = d
		}
	}
	assign := "="
	if newVar {
		assign = ":="
	}
	name := ta.TargetCtx.Scope.Name(target, ta.TargetCtx.Pkg(target), ta.TargetCtx.Pointer, ta.TargetCtx.UseDefault)
	fmt.Fprintf(buffer, "%s %s %s%s{%s}\n", targetVar, assign, deref, name, initCode)
	buffer.WriteString(postInitCode)

	// iterate through attributes to initialize rest of the struct fields and
	// handle default values
	walkMatches(source, target, func(srcMatt, tgtMatt *expr.MappedAttributeExpr, srcc, tgtc *expr.AttributeExpr, n string) {
		h := ta.Hooks
		if h != nil && h.FieldPairAttrs != nil {
			srcc, tgtc = h.FieldPairAttrs(srcc, tgtc)
		}
		var (
			srcVar = sourceVar + "." + ta.SourceCtx.Scope.Field(srcc, srcMatt.ElemName(n), true)
			tgtVar = targetVar + "." + ta.TargetCtx.Scope.Field(tgtc, tgtMatt.ElemName(n), true)
		)
		var dir *WrapDirective
		if h != nil && h.UnwrapPair != nil {
			srcc, tgtc, dir = h.UnwrapPair(srcc, tgtc)
		}
		if err = IsCompatible(srcc.Type, tgtc.Type, sourceVar, targetVar); err != nil {
			return
		}

		var code string
		{
			// The wrap directive (if any) only redirects the code
			// transforming the field value: nil guards and default value
			// handling keep using the unwrapped field variables.
			dispatchSrcVar, dispatchTgtVar, dispatchNewVar := srcVar, tgtVar, false
			prelude := dir.apply(&dispatchSrcVar, &dispatchTgtVar, &dispatchNewVar)
			var postlude string
			if expr.IsUnion(tgtc.Type) && ta.TargetCtx.IsFieldPointer(n, tgtMatt.AttributeExpr) {
				unionVar := Goify(tgtMatt.ElemName(n), false) + "Value"
				unionRef := ta.TargetCtx.Scope.Name(tgtc, ta.TargetCtx.Pkg(tgtc), false, ta.TargetCtx.UseDefault)
				prelude += fmt.Sprintf("var %s %s\n", unionVar, unionRef)
				dispatchTgtVar = unionVar
				postlude = fmt.Sprintf("%s = &%s\n", tgtVar, unionVar)
			}
			_, ok := srcc.Type.(expr.UserType)
			switch {
			case expr.IsArray(srcc.Type):
				if h != nil && h.TransformArray != nil {
					code, err = h.TransformArray(expr.AsArray(srcc.Type), expr.AsArray(tgtc.Type), dispatchSrcVar, dispatchTgtVar, dispatchNewVar, ta)
				} else {
					code, err = transformArray(expr.AsArray(srcc.Type), expr.AsArray(tgtc.Type), dispatchSrcVar, dispatchTgtVar, dispatchNewVar, ta)
				}
			case expr.IsMap(srcc.Type):
				if h != nil && h.TransformMap != nil {
					code, err = h.TransformMap(expr.AsMap(srcc.Type), expr.AsMap(tgtc.Type), dispatchSrcVar, dispatchTgtVar, dispatchNewVar, ta)
				} else {
					code, err = transformMap(expr.AsMap(srcc.Type), expr.AsMap(tgtc.Type), dispatchSrcVar, dispatchTgtVar, dispatchNewVar, ta)
				}
			case expr.IsUnion(srcc.Type):
				if h != nil && h.TransformUnion != nil {
					code, err = h.TransformUnion(srcc, tgtc, dispatchSrcVar, dispatchTgtVar, dispatchNewVar, source, target, ta)
				} else {
					code, err = transformUnion(srcc, tgtc, dispatchSrcVar, dispatchTgtVar, dispatchNewVar, ta)
				}
			case ok:
				if !expr.IsPrimitive(srcc.Type) {
					code = fmt.Sprintf("%s = %s(%s)\n", dispatchTgtVar, TransformHelperName(srcc, tgtc, ta), dispatchSrcVar)
				}
			case expr.IsObject(srcc.Type):
				code, err = TransformAttribute(srcc, tgtc, dispatchSrcVar, dispatchTgtVar, dispatchNewVar, ta)
			}
			if code != "" {
				code = prelude + code + postlude
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
		var guarded bool
		if h != nil && h.GuardCondition != nil {
			var cond string
			if cond, guarded = h.GuardCondition(srcc, srcVar, srcMatt.IsRequired(n), ta.SourceCtx.IsFieldPointer(n, srcMatt.AttributeExpr)); guarded && cond != "" && code != "" {
				code = fmt.Sprintf("%s\t%s}\n", cond, code)
			}
		}
		if !guarded {
			var checkNil bool
			{
				isRef := !expr.IsPrimitive(srcc.Type) && !srcMatt.IsRequired(n) || ta.SourceCtx.IsPrimitivePointer(n, srcMatt.AttributeExpr) && expr.IsPrimitive(srcc.Type)
				marshalNonPrimitive := !expr.IsPrimitive(srcc.Type) && ta.SourceCtx.UseDefault && ta.TargetCtx.UseDefault
				checkNil = isRef || marshalNonPrimitive
			}
			if code != "" && checkNil {
				cond := fmt.Sprintf("if %s != nil {\n", srcVar)
				// A pointer-backed union uses nil as its sole absence value. Preserve a
				// non-nil zero union so validation rejects its missing discriminator.
				if expr.IsUnion(srcc.Type) && !ta.SourceCtx.IsFieldPointer(n, srcMatt.AttributeExpr) {
					cond = fmt.Sprintf("if %s.Kind() != \"\" {\n", srcVar)
				}
				code = fmt.Sprintf("%s\t%s}", cond, code)
				if expr.IsArray(srcc.Type) && srcMatt.IsRequired(n) {
					code += fmt.Sprintf("else {\n\t%s = []%s{}\n}\n", tgtVar, ta.TargetCtx.Scope.Ref(expr.AsArray(tgtc.Type).ElemType, ta.TargetCtx.Pkg(expr.AsArray(tgtc.Type).ElemType)))
				} else {
					code += "\n"
				}
			}
		}

		// Default value handling. We need to handle default values if the target
		// type uses default values (i.e. attributes with default values are
		// non-pointers) and has a default value set.
		if tdef := tgtMatt.GetDefault(n); tdef != nil && ta.TargetCtx.UseDefault && !ta.TargetCtx.Pointer && !srcMatt.IsRequired(n) {
			switch {
			case ta.SourceCtx.IsPrimitivePointer(n, srcMatt.AttributeExpr) || !expr.IsPrimitive(srcc.Type):
				// source attribute is a primitive pointer or not a primitive
				code += fmt.Sprintf("if %s == nil {\n\t", srcVar)
				switch {
				case ta.TargetCtx.IsPrimitivePointer(n, tgtMatt.AttributeExpr) && expr.IsPrimitive(tgtc.Type):
					typeName := GoNativeTypeName(tgtc.Type)
					if h != nil && h.ZeroTypeName != nil {
						if zn, ok := h.ZeroTypeName(tgtc); ok {
							typeName = zn
						}
					}
					code += fmt.Sprintf("var tmp %s = %#v\n\t%s = &tmp\n", typeName, tdef, tgtVar)
				case expr.IsArray(tgtc.Type):
					arr := expr.AsArray(tgtc.Type)
					if expr.IsAlias(arr.ElemType.Type) {
						// Render typed array default literals for aliased element types,
						// e.g. []pkg.EnumType{pkg.EnumType("val")}.
						elemRef := ta.TargetCtx.Scope.Ref(arr.ElemType, ta.TargetCtx.Pkg(arr.ElemType))
						rv := reflect.ValueOf(tdef)
						if rv.Kind() != reflect.Slice {
							panic(fmt.Sprintf("unsupported default value type %T for aliased array element", tdef)) // bug
						}
						items := make([]string, rv.Len())
						for i := range items {
							items[i] = fmt.Sprintf("%s(%#v)", elemRef, rv.Index(i).Interface())
						}
						if len(items) > 0 {
							code += fmt.Sprintf("%s = []%s{%s}\n", tgtVar, elemRef, strings.Join(items, ", "))
						}
					} else {
						// Non-alias element type: use raw default without casting elements
						code += fmt.Sprintf("%s = %#v\n", tgtVar, tdef)
					}
				default:
					code += fmt.Sprintf("%s = %#v\n", tgtVar, tdef)
				}
				code += "}\n"
			case expr.IsPrimitive(srcc.Type) && srcMatt.HasDefaultValue(n) && ta.SourceCtx.UseDefault:
				// source attribute is a primitive with default value
				// (the field is not a pointer in this case)
				code += "{\n\t"
				var (
					zeroName string
					nilable  bool
				)
				if h != nil && h.ZeroTypeName != nil {
					var ok bool
					if zeroName, ok = h.ZeroTypeName(tgtc); ok {
						nilable = typeStringIsNilable(zeroName)
					}
				}
				if zeroName == "" {
					if typeName, _ := GetMetaType(tgtc); typeName != "" {
						zeroName = typeName
						nilable = typeStringIsNilable(typeName)
					} else if _, ok := tgtc.Type.(expr.UserType); ok {
						// aliased primitive
						zeroName = ta.TargetCtx.Scope.Ref(tgtc, ta.TargetCtx.Pkg(tgtc))
					} else {
						zeroName = GoNativeTypeName(tgtc.Type)
					}
				}
				if !nilable {
					code += fmt.Sprintf("var zero %s\n\t", zeroName)
					code += fmt.Sprintf("if %s == zero ", tgtVar)
				} else {
					code += fmt.Sprintf("if %s == nil ", tgtVar)
				}
				code += fmt.Sprintf("{\n\t%s = %#v\n}\n", tgtVar, tdef)
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

// typeStringIsNilable takes a go type as a string and checks for a '[]' or
// 'map[' prefix to see if it's a nilable primitive type.
func typeStringIsNilable(typeName string) bool {
	return strings.HasPrefix(typeName, "[]") || strings.HasPrefix(typeName, "map[")
}

// transformArray generates Go code to transform source array to target array.
func transformArray(source, target *expr.Array, sourceVar, targetVar string, newVar bool, ta *TransformAttrs) (string, error) {
	if err := IsCompatible(source.ElemType.Type, target.ElemType.Type, sourceVar+"[0]", targetVar+"[0]"); err != nil {
		return "", err
	}
	data := map[string]any{
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
	data := map[string]any{
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
	if !expr.IsUnion(target.Type) {
		return "", fmt.Errorf("cannot transform union %s to non-union %s", source.Type.Name(), target.Type.Name())
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

	// Unions are generated as concrete sum-type structs with Kind/AsX/SetX
	// helpers. Transform by branching on the runtime Kind discriminator.
	unionPkg := ta.TargetCtx.Pkg(target)
	typeRef := ta.TargetCtx.Scope.Ref(target, unionPkg)

	// Use deterministic temp var: 'obj' at top-level, 'tmp' for nested
	// assignments. A "obj." prefix means this transform is emitted inside a
	// case body of an enclosing union transform which already declared 'obj'.
	tempVarName := "obj"
	if strings.HasPrefix(targetVar, "obj.") {
		tempVarName = "tmp"
	}

	cases := make([]map[string]any, 0, len(srcUnion.Values))
	for i, st := range srcUnion.Values {
		tt := tgtUnion.Values[i]
		branchAttrs := &TransformAttrs{
			SourceCtx: ta.SourceCtx.Enter(st.Attribute),
			TargetCtx: ta.TargetCtx.Enter(tt.Attribute),
			Prefix:    ta.Prefix,
			Hooks:     ta.Hooks,
		}
		useHelper := false
		if _, ok := st.Attribute.Type.(expr.UserType); ok && expr.IsObject(st.Attribute.Type) {
			if _, ok := tt.Attribute.Type.(expr.UserType); ok && expr.IsObject(tt.Attribute.Type) {
				useHelper = true
			}
		}
		cases = append(cases, map[string]any{
			"CaseName":        st.Name,
			"SourceFieldName": Goify(st.Name, true),
			"TargetFieldName": Goify(tt.Name, true),
			"SourceAttr":      st.Attribute,
			"TargetAttr":      tt.Attribute,
			"TargetCastType":  branchAttrs.TargetCtx.Scope.Ref(tt.Attribute, branchAttrs.TargetCtx.Pkg(tt.Attribute)),
			"UseHelper":       useHelper,
			"HelperName":      TransformHelperName(st.Attribute, tt.Attribute, branchAttrs),
		})
	}

	data := map[string]any{
		"SourceVar":       sourceVar,
		"TargetVar":       targetVar,
		"NewVar":          newVar,
		"TypeRef":         typeRef,
		"TargetIsPointer": strings.HasPrefix(typeRef, "*"),
		"ValueTypeRef":    strings.TrimPrefix(typeRef, "*"),
		"TempVarName":     tempVarName,
		"Cases":           cases,
		"TransformAttrs":  ta,
	}

	var buf bytes.Buffer
	if err := transformGoUnionT.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// collectHelpers recurses through the given attributes and returns the
// transform helper functions required by the code GoTransform produces. The
// top-level call (topLevel true) does not generate a helper for the top-most
// user type because the generated code inlines that transformation; children
// of composite top-level types always get helpers.
//
// seen keeps track of generated transform functions to avoid infinite
// recursion on recursive types.
func collectHelpers(source, target *expr.AttributeExpr, req, topLevel bool, ta *TransformAttrs, seen map[string]*TransformFunctionData) (helpers []*TransformFunctionData, err error) {
	if h := ta.Hooks; h != nil && h.UnwrapPair != nil {
		source, target, _ = h.UnwrapPair(source, target)
	}
	ta = enterTransformAttrs(source, target, ta)
	if topLevel {
		req = true
	} else {
		name := TransformHelperName(source, target, ta)
		if _, ok := seen[name]; ok {
			return helpers, err
		}
		if _, ok := source.Type.(expr.UserType); ok && expr.IsObject(source.Type) {
			var h *TransformFunctionData
			if h, err = generateHelper(source, target, req, ta, seen); h != nil {
				helpers = append(helpers, h)
			}
		}
	}
	// Renderers which inline composite element construction do not call
	// element transform helpers: skip helper generation for the elements
	// themselves by treating them as top-level attributes.
	elemTop := ta.Hooks != nil && ta.Hooks.InlineCompositeElems
	var other []*TransformFunctionData
	switch {
	case expr.IsArray(source.Type):
		if other, err = collectHelpers(expr.AsArray(source.Type).ElemType, expr.AsArray(target.Type).ElemType, req, elemTop, ta, seen); err == nil {
			helpers = append(helpers, other...)
		}
	case expr.IsMap(source.Type):
		sm, tm := expr.AsMap(source.Type), expr.AsMap(target.Type)
		if other, err = collectHelpers(sm.ElemType, tm.ElemType, req, elemTop, ta, seen); err == nil {
			helpers = append(helpers, other...)
			if other, err = collectHelpers(sm.KeyType, tm.KeyType, req, elemTop, ta, seen); err == nil {
				helpers = append(helpers, other...)
			}
		}
	case expr.IsUnion(source.Type):
		tt := expr.AsUnion(target.Type)
		if tt == nil {
			return helpers, err
		}
		for i, st := range expr.AsUnion(source.Type).Values {
			if other, err = collectHelpers(st.Attribute, tt.Values[i].Attribute, req, false, ta, seen); err == nil {
				helpers = append(helpers, other...)
			}
		}
	case expr.IsObject(source.Type):
		if expr.IsUnion(target.Type) {
			return helpers, err
		}
		walkMatches(source, target, func(srcMatt, _ *expr.MappedAttributeExpr, srcc, tgtc *expr.AttributeExpr, n string) {
			if err != nil {
				return
			}
			if other, err = collectHelpers(srcc, tgtc, srcMatt.IsRequired(n), false, ta, seen); err == nil {
				helpers = append(helpers, other...)
			}
		})
	}
	return helpers, err
}

// enterTransformAttrs returns transform attributes whose source and target
// resolvers independently own the attributes being transformed.
func enterTransformAttrs(source, target *expr.AttributeExpr, attributes *TransformAttrs) *TransformAttrs {
	entered := *attributes
	entered.SourceCtx = attributes.SourceCtx.Enter(source)
	entered.TargetCtx = attributes.TargetCtx.Enter(target)
	return &entered
}

// generateHelper generates the code that transforms instances of source into
// target. Both source and target must be user types. The caller
// (collectHelpers) guarantees no helper was generated yet for the pair.
func generateHelper(source, target *expr.AttributeExpr, req bool, ta *TransformAttrs, seen map[string]*TransformFunctionData) (*TransformFunctionData, error) {
	name := TransformHelperName(source, target, ta)

	code, err := TransformAttribute(source, target, "v", "res", true, ta)
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
