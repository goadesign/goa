package codegen

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

var (
	// renderGoArrayT is the template rendering protocol buffer array
	// transformations driven by the shared transform engine.
	renderGoArrayT *template.Template
	// renderGoMapT is the template rendering protocol buffer map
	// transformations driven by the shared transform engine.
	renderGoMapT *template.Template
	// renderGoUnionToProtoT is the template rendering Go union to protobuf
	// oneof transformations.
	renderGoUnionToProtoT *template.Template
	// renderGoUnionFromProtoT is the template rendering protobuf oneof to
	// Go union transformations.
	renderGoUnionFromProtoT *template.Template
)

// NOTE: can't initialize inline because https://github.com/golang/go/issues/1817
func init() {
	fm := template.FuncMap{"transformAttribute": codegen.TransformAttribute}
	renderGoArrayT = template.Must(template.New("renderGoArray").Funcs(fm).Parse(grpcTemplates.Read(grpcTransformGoArrayT)))
	renderGoMapT = template.Must(template.New("renderGoMap").Funcs(fm).Parse(grpcTemplates.Read(grpcTransformGoMapT)))
	renderGoUnionToProtoT = template.Must(template.New("renderGoUnionToProto").Parse(grpcTemplates.Read(grpcTransformGoUnionToProtoT)))
	renderGoUnionFromProtoT = template.Must(template.New("renderGoUnionFromProto").Parse(grpcTemplates.Read(grpcTransformGoUnionFromProtoT)))
}

// protoHooks returns the transform hooks that specialize the shared Go
// transform engine for protocol buffer transformations. proto is true when
// the transformation initializes a protocol buffer type from a service type
// and false when it initializes a service type from a protocol buffer type.
// targetCtx is the target attribute context of the transformation; it is used
// to name the synthetic wrapper messages initialized by the generated code.
func protoHooks(proto bool, targetCtx *codegen.AttributeContext) *codegen.TransformHooks {
	return &codegen.TransformHooks{
		UnwrapPair: func(src, tgt *expr.AttributeExpr) (*expr.AttributeExpr, *expr.AttributeExpr, *codegen.WrapDirective) {
			if proto {
				if isWrappedAttr(tgt) {
					name := targetCtx.Scope.Name(tgt, targetCtx.Pkg(tgt), targetCtx.Pointer, targetCtx.UseDefault)
					return src, unwrapAttr(expr.DupAtt(tgt)), &codegen.WrapDirective{WrapTarget: true, InitTypeName: name, FieldName: "Field"}
				}
				return src, tgt, nil
			}
			if isWrappedAttr(src) {
				return unwrapAttr(expr.DupAtt(src)), tgt, &codegen.WrapDirective{FieldName: "Field"}
			}
			return src, tgt, nil
		},
		FieldPairAttrs: func(src, tgt *expr.AttributeExpr) (*expr.AttributeExpr, *expr.AttributeExpr) {
			return unAlias(src), unAlias(tgt)
		},
		ConvertPrimitive: func(src, tgt *expr.AttributeExpr, srcVar string, srcPtr, tgtPtr bool, ta *codegen.TransformAttrs) (string, bool) {
			exp := convertType(src, tgt, srcPtr, tgtPtr, srcVar, proto, ta)
			if _, isSrcUT := src.Type.(expr.UserType); isSrcUT && !proto {
				// If the source is an alias type and the code is initializing a
				// service type then we must cast to the alias type.
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
		TransformUnion: func(source, target *expr.AttributeExpr, sourceVar, targetVar string, _ bool, srcParent, tgtParent *expr.AttributeExpr, ta *codegen.TransformAttrs) (string, error) {
			if proto {
				// Go service union fields are interfaces (nil when absent); do
				// not dereference.
				return renderUnionToProtoTransform(source, target, sourceVar, targetVar, false, oneofMessageName(tgtParent, proto, ta), ta)
			}
			// Service unions in Goa are represented as interface types, not
			// *interface. Always assign concrete values to the interface (no
			// pointer-to-interface).
			return renderUnionFromProtoTransform(source, target, sourceVar, targetVar, oneofMessageName(srcParent, proto, ta), ta)
		},
		HelperNameAttrs: func(src, tgt *expr.AttributeExpr) (*expr.AttributeExpr, *expr.AttributeExpr) {
			// Do not consider package overrides for protogen generated types.
			if proto {
				tgt = expr.DupAtt(tgt)
				codegen.Walk(tgt, func(att *expr.AttributeExpr) error { // nolint: errcheck
					delete(att.Meta, "struct:pkg:path")
					return nil
				})
			} else {
				src = expr.DupAtt(src)
				codegen.Walk(src, func(att *expr.AttributeExpr) error { // nolint: errcheck
					delete(att.Meta, "struct:pkg:path")
					return nil
				})
			}
			return src, tgt
		},
		GuardCondition: func(src *expr.AttributeExpr, srcVar string, _, srcPtr bool) (string, bool) {
			// Non-primitives are always guarded (proto3 message fields are
			// always nilable).
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
			// if the target is a raw struct no need to return a pointer
			if _, ok := tgt.Type.(*expr.Object); ok {
				return "", true
			}
			return "&", true
		},
		InlineCompositeElems: true,
	}
}

// oneofMessageName returns the reference to the protoc generated Go type of
// the message containing the oneof being transformed: protoc builds the union
// wrapper struct type names from the parent message type name. parent is the
// attribute of the object owning the union field, nil when the union is
// transformed directly in which case there is no message context.
func oneofMessageName(parent *expr.AttributeExpr, proto bool, ta *codegen.TransformAttrs) string {
	if parent == nil {
		return ""
	}
	if _, ok := parent.Type.(expr.UserType); !ok {
		return ""
	}
	if proto {
		return ta.TargetCtx.Scope.Name(parent, ta.TargetCtx.Pkg(parent), false, false)
	}
	return ta.SourceCtx.Scope.Ref(parent, ta.SourceCtx.Pkg(parent))
}

// renderArrayTransform renders the code transforming the source array held by
// sourceVar into the target array held by targetVar. proto is true when the
// target is the protocol buffer type. Wrapped element types are passed
// through to the engine which unwraps them when recursing.
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

	data := map[string]any{
		"ElemTypeRef":    targetRef,
		"SourceElem":     source.ElemType,
		"TargetElem":     elem,
		"SourceVar":      sourceVar,
		"TargetVar":      targetVar,
		"NewVar":         newVar,
		"TransformAttrs": ta,
		"LoopVar":        string(rune(105 + strings.Count(targetVar, "["))),
		"ValVar":         valVar,
	}
	var buf bytes.Buffer
	if err := renderGoArrayT.Execute(&buf, data); err != nil {
		return "", err
	}
	return ensureTrailingNewline(buf.String()), nil
}

// renderMapTransform renders the code transforming the source map held by
// sourceVar into the target map held by targetVar. proto is true when the
// target is the protocol buffer type. Wrapped element types are passed
// through to the engine which unwraps them when recursing; map keys cannot be
// nested in protocol buffers so only elements may be wrapped.
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
	data := map[string]any{
		"KeyTypeRef":     ta.TargetCtx.Scope.Ref(kt, ta.TargetCtx.Pkg(kt)),
		"ElemTypeRef":    ta.TargetCtx.Scope.Ref(et, ta.TargetCtx.Pkg(et)),
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
	if depth := codegen.MapDepth(target); depth > 0 {
		data["LoopVar"] = string(rune(97 + depth))
	}
	var buf bytes.Buffer
	if err := renderGoMapT.Execute(&buf, data); err != nil {
		return "", err
	}
	return ensureTrailingNewline(buf.String()), nil
}

// renderUnionToProtoTransform renders the code transforming the source Goa
// union held by sourceVar into the protoc generated oneof field held by
// targetVar. message is the reference to the protoc generated Go type of the
// message containing the oneof.
func renderUnionToProtoTransform(source, target *expr.AttributeExpr, sourceVar, targetVar string, sourcePtr bool, message string, ta *codegen.TransformAttrs) (string, error) {
	if err := codegen.IsCompatible(source.Type, target.Type, sourceVar, targetVar); err != nil {
		return "", err
	}
	src, tgt := expr.AsUnion(source.Type), expr.AsUnion(target.Type)
	cases := make([]map[string]any, 0, len(src.Values))
	for i, sv := range src.Values {
		tv := tgt.Values[i]
		fieldName := ta.TargetCtx.Scope.Field(tv.Attribute, tv.Name, true)
		cases = append(cases, map[string]any{
			"TypeTag":           sv.Name,
			"SourceFieldName":   codegen.Goify(sv.Name, true),
			"TargetWrapperType": protocOneofWrapperRef(message, fieldName),
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

// renderUnionFromProtoTransform renders the code transforming the protoc
// generated oneof field held by sourceVar into the target Goa union held by
// targetVar. message is the reference to the protoc generated Go type of the
// message containing the oneof.
func renderUnionFromProtoTransform(source, target *expr.AttributeExpr, sourceVar, targetVar, message string, ta *codegen.TransformAttrs) (string, error) {
	if err := codegen.IsCompatible(source.Type, target.Type, sourceVar, targetVar); err != nil {
		return "", err
	}
	src, tgt := expr.AsUnion(source.Type), expr.AsUnion(target.Type)
	cases := make([]map[string]any, 0, len(src.Values))
	for i, sv := range src.Values {
		tv := tgt.Values[i]
		sourceFieldName := ta.SourceCtx.Scope.Field(sv.Attribute, sv.Name, true)
		cases = append(cases, map[string]any{
			"SourceValueTypeRef": protocOneofWrapperRef(message, sourceFieldName),
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

// ensureTrailingNewline appends a newline to code when missing so that the
// rendered transformations compose with the code the engine emits around
// them.
func ensureTrailingNewline(code string) string {
	if code != "" && !strings.HasSuffix(code, "\n") {
		code += "\n"
	}
	return code
}
