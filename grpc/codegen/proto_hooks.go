// This file tells the shared Go conversion code how protobuf wrappers,
// collections, unions, and nil values differ from service values.
package codegen

import (
	"bytes"
	"fmt"
	"maps"
	"reflect"
	"strings"
	"text/template"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type (
	// protobufOneofAttributor returns the Go wrapper type generated for one
	// protobuf union branch.
	protobufOneofAttributor interface {
		OneofWrapper(*expr.AttributeExpr) string
	}
)

var (
	// renderGoArrayT writes a conversion between service and protobuf arrays.
	renderGoArrayT *template.Template
	// renderGoMapT writes a conversion between service and protobuf maps.
	renderGoMapT *template.Template
	// renderGoUnionToProtoT writes a service union into a protobuf oneof.
	renderGoUnionToProtoT *template.Template
	// renderGoUnionFromProtoT writes a protobuf oneof into a service union.
	renderGoUnionFromProtoT *template.Template
)

// The templates are initialized here because Go cannot initialize this cycle
// of template functions in the variable declarations.
func init() {
	fm := template.FuncMap{"transformAttribute": codegen.TransformAttribute}
	renderGoArrayT = template.Must(template.New("renderGoArray").Funcs(fm).Parse(grpcTemplates.Read(grpcTransformGoArrayT)))
	renderGoMapT = template.Must(template.New("renderGoMap").Funcs(fm).Parse(grpcTemplates.Read(grpcTransformGoMapT)))
	renderGoUnionToProtoT = template.Must(template.New("renderGoUnionToProto").Parse(grpcTemplates.Read(grpcTransformGoUnionToProtoT)))
	renderGoUnionFromProtoT = template.Must(template.New("renderGoUnionFromProto").Parse(grpcTemplates.Read(grpcTransformGoUnionFromProtoT)))
}

// protoHooks returns the functions that convert between service and protobuf
// values. proto is true when the target is the protobuf value.
func protoHooks(proto bool) *codegen.TransformHooks {
	return &codegen.TransformHooks{
		UnwrapPair: func(src, tgt *expr.AttributeExpr) (*expr.AttributeExpr, *expr.AttributeExpr, *codegen.WrapDirective) {
			if proto {
				if isWrappedAttr(tgt) {
					return src, unwrapAttr(tgt), &codegen.WrapDirective{
						WrapTarget: true,
						Target:     tgt,
						FieldName:  "Field",
					}
				}
				return src, tgt, nil
			}
			if isWrappedAttr(src) {
				return unwrapAttr(src), tgt, &codegen.WrapDirective{FieldName: "Field"}
			}
			return src, tgt, nil
		},
		FieldPairAttrs: func(src, tgt *expr.AttributeExpr) (*expr.AttributeExpr, *expr.AttributeExpr) {
			return unAlias(src), unAlias(tgt)
		},
		ConvertPrimitive: func(src, tgt *expr.AttributeExpr, srcVar string, srcPtr, tgtPtr bool, ta *codegen.TransformAttrs) (string, bool) {
			exp := convertType(src, tgt, srcPtr, tgtPtr, srcVar, proto, ta)
			if _, isSrcUT := src.Type.(expr.UserType); isSrcUT && !proto {
				// A service alias must keep its named Go type.
				deref := ""
				if srcPtr {
					deref = "*"
				}
				exp = fmt.Sprintf("%s(%s%s)", ta.TargetCtx.Scope.Ref(tgt, ta.TargetCtx.Pkg(tgt)), deref, srcVar)
			}
			return exp, true
		},
		TransformArray: func(source, target *expr.Array, sourceVar, targetVar string, newVar bool, ta *codegen.TransformAttrs) (string, error) {
			return renderArrayTransform(source, target, sourceVar, targetVar, newVar, proto, ta)
		},
		TransformMap: func(source, target *expr.Map, sourceVar, targetVar string, newVar bool, ta *codegen.TransformAttrs) (string, error) {
			return renderMapTransform(source, target, sourceVar, targetVar, newVar, proto, ta)
		},
		TransformUnion: func(source, target *expr.AttributeExpr, sourceVar, targetVar string, _ bool, _, _ *expr.AttributeExpr, ta *codegen.TransformAttrs) (string, error) {
			if proto {
				// The service union accessor returns the selected branch value.
				return renderUnionToProtoTransform(source, target, sourceVar, targetVar, false, ta)
			}
			// Select the matching branch in the service union.
			return renderUnionFromProtoTransform(source, target, sourceVar, targetVar, ta)
		},
		PlanUnionHelpers: func(source, target *expr.AttributeExpr, record func(*expr.AttributeExpr, *expr.AttributeExpr)) {
			sourceUnion, targetUnion := expr.AsUnion(source.Type), expr.AsUnion(target.Type)
			for index, sourceBranch := range sourceUnion.Values {
				targetBranch := targetUnion.Values[index]
				if protoUnionBranchUsesHelper(sourceBranch.Attribute, targetBranch.Attribute) {
					record(sourceBranch.Attribute, targetBranch.Attribute)
				}
			}
		},
		SameHelperDefinition: sameGRPCHelperDefinition,
		GuardCondition: func(src *expr.AttributeExpr, srcVar string, _, srcPtr bool) (string, bool) {
			// Protobuf message fields can be nil, so check them before use.
			if expr.IsPrimitive(src.Type) && !srcPtr {
				return "", true
			}
			if proto && expr.IsUnion(src.Type) {
				if srcPtr {
					return fmt.Sprintf("if %s != nil && %s.Kind() != \"\" {\n", srcVar, srcVar), true
				}
				return fmt.Sprintf("if %s.Kind() != \"\" {\n", srcVar), true
			}
			return fmt.Sprintf("if %s != nil {\n", srcVar), true
		},
		ZeroTypeName: func(tgt *expr.AttributeExpr) (string, bool) {
			if proto {
				return protoBufNativeGoTypeName(tgt.Type), true
			}
			return "", false
		},
		ObjectDeref: func(tgt *expr.AttributeExpr) (string, bool) {
			// An unnamed struct value does not need a pointer.
			if _, ok := tgt.Type.(*expr.Object); ok {
				return "", true
			}
			return "&", true
		},
		InlineCompositeElems: true,
	}
}

// sameGRPCHelperDefinition compares every retained attribute fact except the
// protobuf field number. A field number chooses its serialized position but
// cannot change the Go code that copies its value.
func sameGRPCHelperDefinition(firstSource, firstTarget, nextSource, nextTarget *expr.AttributeExpr) bool {
	return sameGRPCHelperAttribute(firstSource, nextSource) &&
		sameGRPCHelperAttribute(firstTarget, nextTarget)
}

// sameGRPCHelperAttribute compares the facts available to conversion hooks.
// Planning copies preserve shared type identity, so a different type cannot
// describe the same function.
func sameGRPCHelperAttribute(first, next *expr.AttributeExpr) bool {
	return first.Type == next.Type &&
		reflect.DeepEqual(first.Bases, next.Bases) &&
		reflect.DeepEqual(first.References, next.References) &&
		first.Description == next.Description &&
		reflect.DeepEqual(first.Docs, next.Docs) &&
		reflect.DeepEqual(first.Validation, next.Validation) &&
		reflect.DeepEqual(grpcHelperMeta(first.Meta), grpcHelperMeta(next.Meta)) &&
		reflect.DeepEqual(first.DefaultValue, next.DefaultValue) &&
		reflect.DeepEqual(first.UserExamples, next.UserExamples)
}

// grpcHelperMeta copies metadata without the protobuf field number.
func grpcHelperMeta(meta expr.MetaExpr) expr.MetaExpr {
	copy := maps.Clone(meta)
	delete(copy, "rpc:tag")
	return copy
}

// renderArrayTransform writes code that copies sourceVar into the target array.
// proto is true when the target is a protobuf array. The shared conversion code
// opens any protobuf wrapper around an array element.
func renderArrayTransform(source, target *expr.Array, sourceVar, targetVar string, newVar, proto bool, ta *codegen.TransformAttrs) (string, error) {
	elem := target.ElemType
	if proto {
		elem = unAlias(elem)
	}
	targetRef := ta.TargetCtx.Scope.Ref(elem, ta.TargetCtx.Pkg(elem))

	valVar := "val"
	if obj := expr.AsObject(source.ElemType.Type); obj != nil && len(*obj) == 0 {
		valVar = ""
	}

	loopVar, childAttrs := ta.EnterCollection()
	childAttrs, err := childAttrs.EnterLocalBlock(sourceVar, targetVar, loopVar, "val")
	if err != nil {
		return "", err
	}
	data := map[string]any{
		"ElemTypeRef":    targetRef,
		"SourceElem":     source.ElemType,
		"TargetElem":     target.ElemType,
		"SourceVar":      sourceVar,
		"TargetVar":      targetVar,
		"NewVar":         newVar,
		"TransformAttrs": childAttrs,
		"LoopVar":        loopVar,
		"ValVar":         valVar,
	}
	var buf bytes.Buffer
	if err := renderGoArrayT.Execute(&buf, data); err != nil {
		return "", err
	}
	return ensureTrailingNewline(buf.String()), nil
}

// renderMapTransform writes code that copies sourceVar into the target map.
// proto is true when the target is a protobuf map. Protobuf map values may use
// wrapper messages; protobuf map keys may not.
func renderMapTransform(source, target *expr.Map, sourceVar, targetVar string, newVar, proto bool, ta *codegen.TransformAttrs) (string, error) {
	if err := codegen.IsCompatible(source.KeyType.Type, target.KeyType.Type, sourceVar+"[key]", targetVar+"[key]"); err != nil {
		return "", err
	}
	kt := target.KeyType
	et := target.ElemType
	if proto {
		kt = unAlias(kt)
		et = unAlias(et)
	}
	loopVar := ""
	if depth := codegen.MapDepth(target); depth > 0 {
		loopVar = string(rune(97 + depth))
	}
	elemTarget := "tv" + loopVar
	blockAttrs, err := ta.EnterLocalBlock(sourceVar, targetVar, "key", "val", "tk", elemTarget)
	if err != nil {
		return "", err
	}
	elemNewVar := true
	if !proto && isWrappedAttr(source.ElemType) {
		elemNewVar = false
	}
	elemTransform, err := codegen.TransformAttribute(source.ElemType, target.ElemType, "val", elemTarget, elemNewVar, blockAttrs)
	if err != nil {
		return "", err
	}
	if !elemNewVar {
		elemTransform = fmt.Sprintf("var %s %s\nif val != nil {\n%s}\n", elemTarget, ta.TargetCtx.Scope.Ref(et, ta.TargetCtx.Pkg(et)), elemTransform)
	}
	data := map[string]any{
		"KeyTypeRef":     ta.TargetCtx.Scope.Ref(kt, ta.TargetCtx.Pkg(kt)),
		"ElemTypeRef":    ta.TargetCtx.Scope.Ref(et, ta.TargetCtx.Pkg(et)),
		"SourceKey":      source.KeyType,
		"TargetKey":      target.KeyType,
		"SourceVar":      sourceVar,
		"TargetVar":      targetVar,
		"NewVar":         newVar,
		"TransformAttrs": blockAttrs,
		"LoopVar":        loopVar,
		"ElemTransform":  elemTransform,
	}
	var buf bytes.Buffer
	if err := renderGoMapT.Execute(&buf, data); err != nil {
		return "", err
	}
	return ensureTrailingNewline(buf.String()), nil
}

// renderUnionToProtoTransform writes a service union from sourceVar into the
// protobuf oneof in targetVar.
func renderUnionToProtoTransform(source, target *expr.AttributeExpr, sourceVar, targetVar string, sourcePtr bool, ta *codegen.TransformAttrs) (string, error) {
	if err := codegen.IsCompatible(source.Type, target.Type, sourceVar, targetVar); err != nil {
		return "", err
	}
	src, tgt := expr.AsUnion(source.Type), expr.AsUnion(target.Type)
	cases := make([]map[string]any, 0, len(src.Values))
	for i, sv := range src.Values {
		tv := tgt.Values[i]
		fieldName := ta.TargetCtx.Scope.Field(tv.Attribute, tv.Name, true)
		scope := ta.TargetCtx.Scope.(protobufOneofAttributor)
		wrapperType := scope.OneofWrapper(tv.Attribute)
		cases = append(cases, map[string]any{
			"TypeTag":           sv.Name,
			"SourceFieldName":   codegen.Goify(sv.Name, true),
			"TargetWrapperType": wrapperType,
			"TargetFieldName":   fieldName,
			"ConvertedValue":    convertType(sv.Attribute, tv.Attribute, false, false, "actual", true, ta),
		})
	}

	data := map[string]any{
		"SourceVar": sourceVar,
		"TargetVar": targetVar,
		"SourcePtr": sourcePtr,
		"Cases":     cases,
	}
	var buf bytes.Buffer
	if err := renderGoUnionToProtoT.Execute(&buf, data); err != nil {
		return "", err
	}
	return ensureTrailingNewline(buf.String()), nil
}

// renderUnionFromProtoTransform writes the protobuf oneof in sourceVar into the
// service union in targetVar.
func renderUnionFromProtoTransform(source, target *expr.AttributeExpr, sourceVar, targetVar string, ta *codegen.TransformAttrs) (string, error) {
	if err := codegen.IsCompatible(source.Type, target.Type, sourceVar, targetVar); err != nil {
		return "", err
	}
	src, tgt := expr.AsUnion(source.Type), expr.AsUnion(target.Type)
	cases := make([]map[string]any, 0, len(src.Values))
	for i, sv := range src.Values {
		tv := tgt.Values[i]
		sourceFieldName := ta.SourceCtx.Scope.Field(sv.Attribute, sv.Name, true)
		scope := ta.SourceCtx.Scope.(protobufOneofAttributor)
		wrapperType := scope.OneofWrapper(sv.Attribute)
		cases = append(cases, map[string]any{
			"SourceValueTypeRef": "*" + wrapperType,
			"TargetFieldName":    codegen.Goify(tv.Name, true),
			"ConvertedValue":     convertType(sv.Attribute, tv.Attribute, false, false, "val."+sourceFieldName, false, ta),
		})
	}
	data := map[string]any{
		"SourceVar": sourceVar,
		"TargetVar": targetVar,
		"Cases":     cases,
	}
	var buf bytes.Buffer
	if err := renderGoUnionFromProtoT.Execute(&buf, data); err != nil {
		return "", err
	}
	return ensureTrailingNewline(buf.String()), nil
}

// ensureTrailingNewline appends a newline when surrounding generated code must
// continue on the next line.
func ensureTrailingNewline(code string) string {
	if code != "" && !strings.HasSuffix(code, "\n") {
		code += "\n"
	}
	return code
}
