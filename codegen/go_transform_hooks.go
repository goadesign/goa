package codegen

import (
	"fmt"

	"goa.design/goa/v3/expr"
)

type (
	// TransformHooks are optional extension points consulted by the Go
	// transform engine (GoTransformWithAttrs and the functions it drives).
	// They let transport-specific generators—today the gRPC generator—alter
	// well-defined aspects of the generated transformation code while
	// sharing the engine driver (attribute walking, struct initialization,
	// nil guards, default value handling and helper function collection).
	//
	// All fields are optional: a nil hook (or a nil Hooks pointer
	// altogether) selects the engine default so that consumers which do not
	// set hooks generate exactly the same code as before the hooks were
	// introduced.
	TransformHooks struct {
		// UnwrapPair adapts a source/target attribute pair before
		// compatibility checks and code generation. It returns the
		// attributes to use in place of src and tgt and a non-nil
		// WrapDirective when one side referenced a synthetic wrapper
		// message that was unwrapped (the gRPC generator wraps
		// non-object protobuf message types in single-field messages).
		// The engine applies the directive by initializing the wrapper
		// and redirecting the source or target variable to the wrapper
		// field. UnwrapPair is consulted by TransformAttribute, by
		// transformObject for each matched field pair and by
		// collectHelpers for each attribute pair it recurses into.
		UnwrapPair func(src, tgt *expr.AttributeExpr) (*expr.AttributeExpr, *expr.AttributeExpr, *WrapDirective)

		// FieldPairAttrs normalizes a matched object field attribute
		// pair before the engine generates the field transformation,
		// the nil guard and the default value handling. The gRPC
		// generator resolves primitive alias user types to their
		// underlying primitive attribute. The hook runs before
		// UnwrapPair when transforming object fields.
		FieldPairAttrs func(src, tgt *expr.AttributeExpr) (*expr.AttributeExpr, *expr.AttributeExpr)

		// ConvertPrimitive returns the expression that initializes a
		// target primitive from the source primitive value held by
		// srcVar (e.g. protobuf int32 casts or Go any to
		// structpb.Value conversions). srcPtr and tgtPtr report
		// whether the source and target values are pointers: the
		// returned expression must dereference srcVar when srcPtr is
		// true. ok must be false to use the engine default conversion.
		// The expression may be srcVar itself when no conversion is
		// needed; the engine compares the result against srcVar to
		// decide whether a conversion is taking place.
		ConvertPrimitive func(src, tgt *expr.AttributeExpr, srcVar string, srcPtr, tgtPtr bool, ta *TransformAttrs) (string, bool)

		// TransformArray overrides the rendering of array
		// transformations. The variables and newVar have the same
		// meaning as in TransformAttribute.
		TransformArray func(source, target *expr.Array, sourceVar, targetVar string, newVar bool, ta *TransformAttrs) (string, error)

		// TransformMap overrides the rendering of map transformations.
		// The variables and newVar have the same meaning as in
		// TransformAttribute.
		TransformMap func(source, target *expr.Map, sourceVar, targetVar string, newVar bool, ta *TransformAttrs) (string, error)

		// TransformUnion overrides the rendering of union
		// transformations. srcParent and tgtParent are the attributes
		// of the object being transformed when the union is an object
		// field and nil when the union is transformed directly: the
		// gRPC generator derives the protoc-generated oneof wrapper
		// type names from the parent message type name.
		TransformUnion func(source, target *expr.AttributeExpr, sourceVar, targetVar string, newVar bool, srcParent, tgtParent *expr.AttributeExpr, ta *TransformAttrs) (string, error)

		// HelperNameAttrs normalizes a source/target attribute pair
		// before the transform helper function name is computed. The
		// gRPC generator strips struct:pkg:path metadata from the
		// protobuf side because protoc-generated types ignore package
		// overrides.
		HelperNameAttrs func(src, tgt *expr.AttributeExpr) (*expr.AttributeExpr, *expr.AttributeExpr)

		// GuardCondition returns the condition that guards the
		// transformation code of an object field, e.g.
		// "if p.Name != nil {\n". src is the (possibly normalized)
		// field attribute, srcVar the source field variable, required
		// reports whether the field is required and srcPtr whether the
		// source field is pointer-backed. An empty condition
		// with ok true means the field transformation must not be
		// guarded. ok must be false to use the engine default policy
		// (the gRPC generator always guards non-primitives because
		// proto3 message fields are always nilable).
		GuardCondition func(src *expr.AttributeExpr, srcVar string, required, srcPtr bool) (string, bool)

		// ZeroTypeName returns the Go type name used to declare the
		// zero value or temporary variable in generated default value
		// handling for the given target attribute. ok must be false to
		// use the engine default (metadata type, alias reference or
		// native Go type name). The gRPC generator returns the Go type
		// generated by the protocol buffer compiler.
		ZeroTypeName func(tgt *expr.AttributeExpr) (string, bool)

		// ObjectDeref returns the operator prefixing object
		// initializers ("&" or ""). ok must be false to use the engine
		// default ("&"). The gRPC generator returns "" for raw struct
		// targets which are not generated as pointers.
		ObjectDeref func(tgt *expr.AttributeExpr) (string, bool)

		// InlineCompositeElems reports that the TransformArray and
		// TransformMap renderers inline the transformation of array
		// element, map key and map element types instead of calling
		// transform helper functions. When true the engine does not
		// collect helpers for the element types themselves; it still
		// collects helpers for the user types referenced by the
		// element object fields.
		InlineCompositeElems bool
	}

	// WrapDirective describes how the engine must account for a synthetic
	// wrapper message unwrapped by the UnwrapPair hook.
	WrapDirective struct {
		// WrapTarget is true when the target attribute was the
		// wrapper: the engine initializes the wrapper value and
		// assigns the transformation result to its field. When false
		// the source attribute was the wrapper and the engine reads
		// the value being transformed from the wrapper field.
		WrapTarget bool
		// InitTypeName is the Go type name used to initialize the
		// wrapper when WrapTarget is true.
		InitTypeName string
		// FieldName is the Go name of the wrapper field holding the
		// wrapped value.
		FieldName string
	}
)

// apply rewrites the transformation variables per the directive and returns
// the code that initializes the wrapper when the target is wrapped. A nil
// directive leaves the variables untouched and returns an empty string.
func (d *WrapDirective) apply(sourceVar, targetVar *string, newVar *bool) string {
	if d == nil {
		return ""
	}
	if !d.WrapTarget {
		*sourceVar += "." + d.FieldName
		return ""
	}
	assign := "="
	if *newVar {
		assign = ":="
	}
	code := fmt.Sprintf("%s %s &%s{}\n", *targetVar, assign, d.InitTypeName)
	*targetVar += "." + d.FieldName
	*newVar = false
	return code
}
