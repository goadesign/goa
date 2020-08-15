package codegen

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type transformAttrs struct {
	*codegen.TransformAttrs
	// proto if true indicates that the transformation code is used to initialize
	// a protocol buffer type from a service type. If false, the transformation
	// code is used to initialize a service type from a protocol buffer type.
	proto bool
	// targetInit is the initialization code for the target type for nested
	// map and array types.
	targetInit string
	// wrapped indicates whether the source or target is in a wrapped state.
	// See grpc/docs/FAQ.md. `wrapped` is true and `proto` is true indicates
	// the target attribute is in wrapped state. `wrapped` is true and `proto`
	// is false indicates the source attribute is in wrapped state.
	wrapped bool
}

var (
	// transformGoArrayT is the template to generate Go array transformation
	// code.
	transformGoArrayT *template.Template
	// transformGoMapT is the template to generate Go map transformation code.
	transformGoMapT *template.Template
)

// NOTE: can't initialize inline because https://github.com/golang/go/issues/1817
func init() {
	fm := template.FuncMap{"transformAttribute": transformAttribute}
	transformGoArrayT = template.Must(template.New("transformGoArray").Funcs(fm).Parse(transformGoArrayTmpl))
	transformGoMapT = template.Must(template.New("transformGoMap").Funcs(fm).Parse(transformGoMapTmpl))
}

// protoBufTransform produces Go code to initialize a data structure defined
// by target from an instance of data structure defined by source. The source
// or target is a protocol buffer type.
//
// source, target are the source and target attributes used in transformation
//
// sourceVar, targetVar are the source and target variables
//
// sourceCtx, targetCtx are the source and target attribute contexts
//
// `proto` param if true indicates that the target is a protocol buffer type
//
// newVar if true initializes a target variable with the generated Go code
// using `:=` operator. If false, it assigns Go code to the target variable
// using `=`.
//
func protoBufTransform(source, target *expr.AttributeExpr, sourceVar, targetVar string, sourceCtx, targetCtx *codegen.AttributeContext, proto, newVar bool) (string, []*codegen.TransformFunctionData, error) {
	source = unAlias(source)
	target = unAlias(target)
	var prefix string
	{
		prefix = "protobuf"
		if proto {
			prefix = "svc"
		}
	}
	ta := &transformAttrs{
		TransformAttrs: &codegen.TransformAttrs{
			SourceCtx: sourceCtx,
			TargetCtx: targetCtx,
			Prefix:    prefix,
		},
		proto: proto,
	}

	code, err := transformAttribute(source, target, sourceVar, targetVar, newVar, ta)
	if err != nil {
		return "", nil, err
	}

	funcs, err := transformAttributeHelpers(source, target, ta)
	if err != nil {
		return "", nil, err
	}

	return strings.TrimRight(code, "\n"), funcs, nil
}

// transformAttribute returns the code to initialize a target data structure
// from an instance of source data structure. It returns an error if source and
// target are not compatible for transformation (different types, fields of
// different type).
func transformAttribute(source, target *expr.AttributeExpr, sourceVar, targetVar string, newVar bool, ta *transformAttrs) (string, error) {
	var (
		initCode string
		err      error
	)

	if err := codegen.IsCompatible(source.Type, target.Type, sourceVar, targetVar); err != nil {
		if ta.proto {
			name := ta.TargetCtx.Scope.Name(target, ta.TargetCtx.Pkg, ta.TargetCtx.Pointer, ta.TargetCtx.UseDefault)
			initCode += fmt.Sprintf("%s := &%s{}\n", targetVar, name)
			targetVar += ".Field"
			newVar = false
			target = unwrapAttr(expr.DupAtt(target))
		} else {
			source = unwrapAttr(expr.DupAtt(source))
			sourceVar += ".Field"
		}
		if err = codegen.IsCompatible(source.Type, target.Type, sourceVar, targetVar); err != nil {
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
			srcField := convertType(source, target, sourceVar, ta)
			code = fmt.Sprintf("%s %s %s\n", targetVar, assign, srcField)
		}
	}
	if err != nil {
		return "", err
	}
	return initCode + code, nil
}

// transformObject returns the code to transform source attribute of object
// type to target attribute of object type. It returns an error if source
// and target are not compatible for transformation.
func transformObject(source, target *expr.AttributeExpr, sourceVar, targetVar string, newVar bool, ta *transformAttrs) (string, error) {
	var (
		initCode     string
		postInitCode string
	)
	{
		// iterate through primitive attributes to initialize the struct
		walkMatches(source, target, func(srcMatt, tgtMatt *expr.MappedAttributeExpr, srcc, tgtc *expr.AttributeExpr, n string) {
			if !expr.IsPrimitive(srcc.Type) {
				return
			}
			var (
				srcField = sourceVar + "." + ta.SourceCtx.Scope.Field(srcc, srcMatt.ElemName(n), true)
				tgtField = ta.TargetCtx.Scope.Field(tgtc, tgtMatt.ElemName(n), true)
				srcPtr   = ta.SourceCtx.IsPrimitivePointer(n, srcMatt.AttributeExpr)
				tgtPtr   = ta.TargetCtx.IsPrimitivePointer(n, tgtMatt.AttributeExpr)
			)
			srcFieldConv := convertType(srcc, tgtc, srcField, ta)
			switch {
			case srcPtr && !tgtPtr:
				postInitCode += fmt.Sprintf("if %s != nil {\n\t%s.%s = %s\n}\n", srcField, targetVar, tgtField, convertType(srcc, tgtc, "*"+srcField, ta))
				return
			case !srcPtr && tgtPtr:
				// In protocol buffer version 3, there is no concept of an optional
				// message field, i.e., a message field is always initialized with the
				// type's zero value - https://developers.google.com/protocol-buffers/docs/proto3#default
				// So if the attribute in the service type is an optional attribute and
				// the corresponding protocol buffer field contains a zero value then
				// we set the optional attribute as nil.
				// We don't check for zero values for booleans since by default, protocol
				// buffer sets boolean fields as false.
				if !srcMatt.IsRequired(n) && srcc.Type != expr.Boolean {
					postInitCode += fmt.Sprintf("if %s {\n\t", checkZeroValue(srcc.Type, srcField, true))
					if srcField != srcFieldConv {
						// type conversion required. Add it in postinit code.
						tgtName := codegen.Goify(tgtField, false)
						postInitCode += fmt.Sprintf("%sptr := %s\n%s.%s = &%sptr", tgtName, srcFieldConv, targetVar, tgtField, tgtName)
					} else {
						postInitCode += fmt.Sprintf("%s.%s = &%s", targetVar, tgtField, srcFieldConv)
					}
					postInitCode += "\n}\n"
					return
				} else if srcField != srcFieldConv {
					// type conversion required. Add it in postinit code.
					tgtName := codegen.Goify(tgtField, false)
					postInitCode += fmt.Sprintf("%sptr := %s\n%s.%s = &%sptr\n", tgtName, srcFieldConv, targetVar, tgtField, tgtName)
					return
				}
				srcFieldConv = "&" + srcFieldConv
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
	if _, ok := target.Type.(*expr.Object); ok {
		deref = ""
	}
	assign := "="
	if newVar {
		assign = ":="
	}
	tname := ta.TargetCtx.Scope.Name(target, ta.TargetCtx.Pkg, ta.TargetCtx.Pointer, ta.TargetCtx.UseDefault)
	buffer.WriteString(fmt.Sprintf("%s %s %s%s{%s}\n", targetVar, assign, deref, tname, initCode))
	buffer.WriteString(postInitCode)

	// iterate through attributes to initialize rest of the struct fields and
	// handle default values
	var err error
	walkMatches(source, target, func(srcMatt, tgtMatt *expr.MappedAttributeExpr, srcc, tgtc *expr.AttributeExpr, n string) {
		srcc = unAlias(srcc)
		tgtc = unAlias(tgtc)
		var (
			code string

			srcVar = sourceVar + "." + ta.SourceCtx.Scope.Field(srcc, srcMatt.ElemName(n), true)
			tgtVar = targetVar + "." + ta.TargetCtx.Scope.Field(tgtc, tgtMatt.ElemName(n), true)
		)
		{
			if err = codegen.IsCompatible(srcc.Type, tgtc.Type, "", ""); err != nil {
				if ta.proto {
					ta.targetInit = ta.TargetCtx.Scope.Name(tgtc, ta.TargetCtx.Pkg, ta.TargetCtx.Pointer, ta.TargetCtx.UseDefault)
					tgtc = unwrapAttr(tgtc)
				} else {
					srcc = unwrapAttr(srcc)
				}
				ta.wrapped = true
				if err = codegen.IsCompatible(srcc.Type, tgtc.Type, "", ""); err != nil {
					return
				}
			}
			_, ok := srcc.Type.(expr.UserType)
			switch {
			case expr.IsArray(srcc.Type):
				code, err = transformArray(expr.AsArray(srcc.Type), expr.AsArray(tgtc.Type), srcVar, tgtVar, false, ta)
			case expr.IsMap(srcc.Type):
				code, err = transformMap(expr.AsMap(srcc.Type), expr.AsMap(tgtc.Type), srcVar, tgtVar, false, ta)
			case ok:
				code = fmt.Sprintf("%s = %s\n", tgtVar, convertType(srcc, tgtc, srcVar, ta))
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
			checkNil = !expr.IsPrimitive(srcc.Type) || ta.SourceCtx.IsPrimitivePointer(n, srcMatt.AttributeExpr)
		}
		if code != "" && checkNil {
			code = fmt.Sprintf("if %s != nil {\n\t%s}\n", srcVar, code)
		}

		// Default value handling. We need to handle default values if the target
		// type uses default values (i.e. attributes with default values are
		// non-pointers) and has a default value set.
		if tdef := tgtc.DefaultValue; tdef != nil && ta.TargetCtx.UseDefault {
			if ta.proto {
				// We set default values in protocol buffer type only if the source type
				// uses pointers to hold default values.
				if ta.SourceCtx.IsPrimitivePointer(n, srcMatt.AttributeExpr) {
					code += fmt.Sprintf("if %s == nil {\n\t%s = %#v\n}\n", srcVar, tgtVar, tdef)
				} else if !expr.IsPrimitive(srcc.Type) && !srcMatt.IsRequired(n) {
					code += fmt.Sprintf("if %s {\n\t%s = %#v\n}\n", checkZeroValue(srcc.Type, srcVar, false), tgtVar, tdef)
				}
			} else {
				// In protocol buffer version 3, the optional attributes are always
				// initialized with their zero values - https://developers.google.com/protocol-buffers/docs/proto3#default
				// Therefore, we always initialize optional fields in the target with
				// their default values if the corresponding source fields have zero
				// values.
				// We don't set default values for booleans since by default, protocol
				// buffer sets boolean fields as false. Changing them to the default
				// value is counter-intuitive.
				if !srcMatt.IsRequired(n) && srcc.Type != expr.Boolean {
					code += fmt.Sprintf("if %s {\n\t", checkZeroValue(srcc.Type, srcVar, false))
					if ta.TargetCtx.IsPrimitivePointer(n, tgtMatt.AttributeExpr) && expr.IsPrimitive(tgtc.Type) {
						code += fmt.Sprintf("var tmp %s = %#v\n\t%s = &tmp\n", codegen.GoNativeTypeName(tgtc.Type), tdef, tgtVar)
					} else {
						code += fmt.Sprintf("%s = %#v\n", tgtVar, tdef)
					}
					code += "}\n"
				}
			}
		}
		buffer.WriteString(code)
	})
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}

// transformArray returns the code to transform source attribute of array
// type to target attribute of array type. It returns an error if source
// and target are not compatible for transformation.
func transformArray(source, target *expr.Array, sourceVar, targetVar string, newVar bool, ta *transformAttrs) (string, error) {
	targetRef := ta.TargetCtx.Scope.Ref(target.ElemType, ta.TargetCtx.Pkg)

	var (
		code string
		err  error
	)

	// If targetInit is set, the target array element is in a nested state.
	// See grpc/docs/FAQ.md.
	if ta.targetInit != "" {
		assign := "="
		if newVar {
			assign = ":="
		}
		code = fmt.Sprintf("%s %s &%s{}\n", targetVar, assign, ta.targetInit)
		ta.targetInit = ""
	}
	if ta.wrapped {
		if ta.proto {
			targetVar += ".Field"
			newVar = false
		} else {
			sourceVar += ".Field"
		}
		ta.wrapped = false
	}

	src := source.ElemType
	tgt := target.ElemType
	if err = codegen.IsCompatible(src.Type, tgt.Type, "[0]", "[0]"); err != nil {
		if ta.proto {
			ta.targetInit = ta.TargetCtx.Scope.Name(tgt, ta.TargetCtx.Pkg, ta.TargetCtx.Pointer, ta.TargetCtx.UseDefault)
			tgt = unwrapAttr(expr.DupAtt(tgt))
		} else {
			src = unwrapAttr(expr.DupAtt(src))
		}
		ta.wrapped = true
		if err = codegen.IsCompatible(src.Type, tgt.Type, "[0]", "[0]"); err != nil {
			return "", err
		}
	}

	data := map[string]interface{}{
		"ElemTypeRef":    targetRef,
		"SourceElem":     src,
		"TargetElem":     tgt,
		"SourceVar":      sourceVar,
		"TargetVar":      targetVar,
		"NewVar":         newVar,
		"TransformAttrs": ta,
		"LoopVar":        string(rune(105 + strings.Count(targetVar, "["))),
	}
	var buf bytes.Buffer
	if err := transformGoArrayT.Execute(&buf, data); err != nil {
		return "", err
	}
	return code + buf.String(), nil
}

// transformMap returns the code to transform source attribute of map
// type to target attribute of map type. It returns an error if source
// and target are not compatible for transformation.
func transformMap(source, target *expr.Map, sourceVar, targetVar string, newVar bool, ta *transformAttrs) (string, error) {
	// Target map key cannot be nested in protocol buffers. So no need to worry
	// about unwrapping.
	if err := codegen.IsCompatible(source.KeyType.Type, target.KeyType.Type, sourceVar+"[key]", targetVar+"[key]"); err != nil {
		return "", err
	}

	targetKeyRef := ta.TargetCtx.Scope.Ref(target.KeyType, ta.TargetCtx.Pkg)
	targetElemRef := ta.TargetCtx.Scope.Ref(target.ElemType, ta.TargetCtx.Pkg)

	var (
		code string
		err  error
	)

	// If targetInit is set, the target map element is in a nested state.
	// See grpc/docs/FAQ.md.
	if ta.targetInit != "" {
		assign := "="
		if newVar {
			assign = ":="
		}
		code = fmt.Sprintf("%s %s &%s{}\n", targetVar, assign, ta.targetInit)
		ta.targetInit = ""
	}
	if ta.wrapped {
		if ta.proto {
			targetVar += ".Field"
			newVar = false
		} else {
			sourceVar += ".Field"
		}
		ta.wrapped = false
	}

	src := source.ElemType
	tgt := target.ElemType
	if err = codegen.IsCompatible(src.Type, tgt.Type, "[*]", "[*]"); err != nil {
		if ta.proto {
			ta.targetInit = ta.TargetCtx.Scope.Name(tgt, ta.TargetCtx.Pkg, ta.TargetCtx.Pointer, ta.TargetCtx.UseDefault)
			tgt = unwrapAttr(expr.DupAtt(tgt))
		} else {
			src = unwrapAttr(expr.DupAtt(src))
		}
		ta.wrapped = true
		if err = codegen.IsCompatible(src.Type, tgt.Type, "[*]", "[*]"); err != nil {
			return "", err
		}
	}
	data := map[string]interface{}{
		"KeyTypeRef":     targetKeyRef,
		"ElemTypeRef":    targetElemRef,
		"SourceKey":      source.KeyType,
		"TargetKey":      target.KeyType,
		"SourceElem":     src,
		"TargetElem":     tgt,
		"SourceVar":      sourceVar,
		"TargetVar":      targetVar,
		"NewVar":         newVar,
		"TransformAttrs": ta,
		"LoopVar":        "",
	}
	if depth := codegen.MapDepth(target); depth > 0 {
		data["LoopVar"] = string(rune(97 + depth))
	}
	var buf bytes.Buffer
	if err := transformGoMapT.Execute(&buf, data); err != nil {
		return "", err
	}
	return code + buf.String(), nil
}

// convertType produces code to initialize a target type from a source type
// held by sourceVar.
// NOTE: For Int and UInt kinds, protocol buffer Go compiler generates
// int32 and uint32 respectively whereas goa v2 generates int and uint.
func convertType(source, target *expr.AttributeExpr, sourceVar string, ta *transformAttrs) string {
	if _, ok := source.Type.(expr.UserType); ok {
		// return a function name for the conversion
		sourcePrimitive, targetPrimitive := getPrimitive(source), getPrimitive(target)
		if sourcePrimitive != nil && targetPrimitive != nil && sourcePrimitive.Type == targetPrimitive.Type {
			if ta.proto {
				return fmt.Sprintf("%s(%s)", targetPrimitive.Type.Name(), sourceVar)
			}
			return fmt.Sprintf("%s(%s)", ta.TargetCtx.Scope.Ref(target, ta.TargetCtx.Pkg), sourceVar)
		}
		return fmt.Sprintf("%s(%s)", transformHelperName(source, target, ta), sourceVar)
	}

	if source.Type.Kind() != expr.IntKind && source.Type.Kind() != expr.UIntKind {
		return sourceVar
	}
	if ta.proto {
		return fmt.Sprintf("%s(%s)", protoBufNativeGoTypeName(source.Type), sourceVar)
	}
	return fmt.Sprintf("%s(%s)", codegen.GoNativeTypeName(source.Type), sourceVar)
}

// zeroValure returns the zero value for the given primitive type.
func checkZeroValue(dt expr.DataType, target string, negate bool) string {
	eq := "=="
	if negate {
		eq = "!="
	}
	switch dt.Kind() {
	// don't check for BooleanKind since by default boolean is set to false
	case expr.IntKind, expr.Int32Kind, expr.Int64Kind,
		expr.UIntKind, expr.UInt32Kind, expr.UInt64Kind,
		expr.Float32Kind, expr.Float64Kind:
		return fmt.Sprintf("%s %s 0", target, eq)
	case expr.StringKind:
		return fmt.Sprintf("%s %s \"\"", target, eq)
	case expr.BytesKind, expr.ArrayKind, expr.MapKind:
		return fmt.Sprintf("len(%s) %s 0", target, eq)
	default:
		return fmt.Sprintf("%s %s nil", target, eq)
	}
}

// transformAttributeHelpers returns the Go transform functions and their definitions
// that may be used in code produced by Transform. It returns an error if source and
// target are incompatible (different types, fields of different type etc).
//
// source, target are the source and target attributes used in transformation
//
// ta is the transform attributes
//
// seen keeps track of generated transform functions to avoid recursion
//
func transformAttributeHelpers(source, target *expr.AttributeExpr, ta *transformAttrs, seen ...map[string]*codegen.TransformFunctionData) ([]*codegen.TransformFunctionData, error) {
	var (
		helpers []*codegen.TransformFunctionData
		err     error
	)
	{
		if err = codegen.IsCompatible(source.Type, target.Type, "", ""); err != nil {
			if ta.proto {
				target = unwrapAttr(expr.DupAtt(target))
			} else {
				source = unwrapAttr(expr.DupAtt(source))
			}
			if err = codegen.IsCompatible(source.Type, target.Type, "", ""); err != nil {
				return nil, err
			}
		}
		// Do not generate a transform function for the top most user type.
		switch {
		case expr.IsArray(source.Type):
			source = expr.AsArray(source.Type).ElemType
			target = expr.AsArray(target.Type).ElemType
			helpers, err = transformAttributeHelpers(source, target, ta, seen...)
		case expr.IsMap(source.Type):
			sm := expr.AsMap(source.Type)
			tm := expr.AsMap(target.Type)
			helpers, err = transformAttributeHelpers(sm.ElemType, tm.ElemType, ta, seen...)
			if err == nil {
				var other []*codegen.TransformFunctionData
				other, err = transformAttributeHelpers(sm.KeyType, tm.KeyType, ta, seen...)
				helpers = append(helpers, other...)
			}
		case expr.IsObject(source.Type):
			walkMatches(source, target, func(srcMatt, tgtMatt *expr.MappedAttributeExpr, srcc, tgtc *expr.AttributeExpr, n string) {
				if err != nil {
					return
				}
				if err = codegen.IsCompatible(srcc.Type, tgtc.Type, "", ""); err != nil {
					if ta.proto {
						tgtc = unwrapAttr(tgtc)
					} else {
						srcc = unwrapAttr(srcc)
					}
					if err = codegen.IsCompatible(srcc.Type, tgtc.Type, "", ""); err != nil {
						return
					}
				}
				h, err2 := collectHelpers(srcc, tgtc, srcMatt.IsRequired(n), ta, seen...)
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
func collectHelpers(source, target *expr.AttributeExpr, req bool, ta *transformAttrs, seen ...map[string]*codegen.TransformFunctionData) ([]*codegen.TransformFunctionData, error) {
	var (
		data []*codegen.TransformFunctionData
	)
	switch {
	case expr.IsArray(source.Type):
		helpers, err := transformAttributeHelpers(
			expr.AsArray(source.Type).ElemType,
			expr.AsArray(target.Type).ElemType,
			ta, seen...)
		if err != nil {
			return nil, err
		}
		data = append(data, helpers...)
	case expr.IsMap(source.Type):
		helpers, err := transformAttributeHelpers(
			expr.AsMap(source.Type).KeyType,
			expr.AsMap(target.Type).KeyType,
			ta, seen...)
		if err != nil {
			return nil, err
		}
		data = append(data, helpers...)
		helpers, err = transformAttributeHelpers(
			expr.AsMap(source.Type).ElemType,
			expr.AsMap(target.Type).ElemType,
			ta, seen...)
		if err != nil {
			return nil, err
		}
		data = append(data, helpers...)
	case expr.IsObject(source.Type):
		if ut, ok := source.Type.(expr.UserType); ok {
			name := transformHelperName(source, target, ta)
			var s map[string]*codegen.TransformFunctionData
			if len(seen) > 0 {
				s = seen[0]
			} else {
				s = make(map[string]*codegen.TransformFunctionData)
				seen = append(seen, s)
			}
			if _, ok := s[name]; ok {
				return nil, nil
			}
			code, err := transformAttribute(ut.Attribute(), target, "v", "res", true, ta)
			if err != nil {
				return nil, err
			}
			if !req {
				code = "if v == nil {\n\treturn nil\n}\n" + code
			}
			tfd := &codegen.TransformFunctionData{
				Name:          name,
				ParamTypeRef:  ta.SourceCtx.Scope.Ref(source, ta.SourceCtx.Pkg),
				ResultTypeRef: ta.TargetCtx.Scope.Ref(target, ta.TargetCtx.Pkg),
				Code:          code,
			}
			s[name] = tfd
			data = append(data, tfd)
		}

		// collect helpers
		var err error
		{
			walkMatches(source, target, func(srcMatt, _ *expr.MappedAttributeExpr, srcc, tgtc *expr.AttributeExpr, n string) {
				if err = codegen.IsCompatible(srcc.Type, tgtc.Type, "", ""); err != nil {
					if ta.proto {
						tgtc = unwrapAttr(tgtc)
					} else {
						srcc = unwrapAttr(srcc)
					}
					if err = codegen.IsCompatible(srcc.Type, tgtc.Type, "", ""); err != nil {
						return
					}
				}
				var helpers []*codegen.TransformFunctionData
				helpers, err = collectHelpers(srcc, tgtc, srcMatt.IsRequired(n), ta, seen...)
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
func transformHelperName(source, target *expr.AttributeExpr, ta *transformAttrs) string {
	var (
		sname  string
		tname  string
		prefix string
	)
	{
		sname = codegen.Goify(ta.SourceCtx.Scope.Name(source, ta.SourceCtx.Pkg, ta.TargetCtx.Pointer, ta.TargetCtx.UseDefault), true)
		tname = codegen.Goify(ta.TargetCtx.Scope.Name(target, ta.TargetCtx.Pkg, ta.TargetCtx.Pointer, ta.TargetCtx.UseDefault), true)
		prefix = ta.Prefix
	}
	return codegen.Goify(prefix+sname+"To"+tname, false)
}

// unAlias returns the base AttributeExpr of an aliased one.
func unAlias(at *expr.AttributeExpr) *expr.AttributeExpr {
	if prim := getPrimitive(at); prim != nil {
		return prim
	}
	return at
}

const (
	transformGoArrayTmpl = `{{ .TargetVar }} {{ if .NewVar }}:={{ else }}={{ end }} make([]{{ .ElemTypeRef }}, len({{ .SourceVar }}))
for {{ .LoopVar }}, val := range {{ .SourceVar }} {
  {{ transformAttribute .SourceElem .TargetElem "val" (printf "%s[%s]" .TargetVar .LoopVar) false .TransformAttrs -}}
}
`

	transformGoMapTmpl = `{{ .TargetVar }} {{ if .NewVar }}:={{ else }}={{ end }} make(map[{{ .KeyTypeRef }}]{{ .ElemTypeRef }}, len({{ .SourceVar }}))
for key, val := range {{ .SourceVar }} {
  {{ transformAttribute .SourceKey .TargetKey "key" "tk" true .TransformAttrs -}}
  {{ transformAttribute .SourceElem .TargetElem "val" (printf "tv%s" .LoopVar) true .TransformAttrs -}}
  {{ .TargetVar }}[tk] = {{ printf "tv%s" .LoopVar }}
}
`
)
