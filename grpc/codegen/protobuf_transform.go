// This file writes Go conversions between service values and protobuf values.
package codegen

import (
	"fmt"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// protoBufTransform writes code that copies source into target. One side is a
// service value and the other is a protobuf value. proto is true when the
// target is the protobuf value. newVar chooses between := and =.
func protoBufTransform(source, target *expr.AttributeExpr, sourceCtx, targetCtx *codegen.AttributeContext, proto, newVar bool) (string, []*codegen.TransformFunctionData, error) {
	prefix := "protobuf"
	if proto {
		original := target
		target = expr.DupAtt(target)
		targetCtx.Scope.(*protoBufScope).service.protobuf.plan.bindAttributeCopy(original, target)
		removeMeta(target)
		prefix = "svc"
	} else {
		original := source
		source = expr.DupAtt(source)
		sourceCtx.Scope.(*protoBufScope).service.protobuf.plan.bindAttributeCopy(original, source)
		removeMeta(source)
	}
	ta := &codegen.TransformAttrs{
		SourceCtx: sourceCtx,
		TargetCtx: targetCtx,
		Prefix:    prefix,
		Hooks:     protoHooks(proto),
	}
	return codegen.GoTransformWithAttrs(source, target, "source", "target", ta, newVar)
}

// removeMeta removes service field and package settings from a protobuf copy.
// The protobuf compiler does not use these settings when it writes Go types.
func removeMeta(att *expr.AttributeExpr) {
	err := codegen.Walk(att, func(a *expr.AttributeExpr) error {
		delete(a.Meta, "struct:field:name")
		delete(a.Meta, "struct:field:external")
		delete(a.Meta, "struct.field.external") // Deprecated syntax. Only present for backward compatibility.
		delete(a.Meta, "struct:pkg:path")
		return nil
	})
	if err != nil {
		panic(fmt.Sprintf("remove protobuf metadata: %s", err))
	}
}

// convertType writes the expression that converts srcVar to the target type.
// proto is true when the target is a protobuf value. Protobuf uses int32 and
// uint32 where Goa uses int and uint.
func convertType(src, tgt *expr.AttributeExpr, srcPtr, tgtPtr bool, srcVar string, proto bool, ta *codegen.TransformAttrs) string {
	if protoUnionBranchUsesHelper(src, tgt) {
		return fmt.Sprintf("%s(%s)", codegen.TransformHelperName(src, tgt, ta), srcVar)
	}
	if expr.IsAlias(src.Type) || expr.IsAlias(tgt.Type) {
		srcp, tgtp := unAlias(src), unAlias(tgt)
		if proto {
			return convertPrimitiveToProto(src, tgtp, srcPtr, tgtPtr, srcVar)
		}
		return convertPrimitiveFromProto(srcp, tgt, srcPtr, tgtPtr, srcVar, ta)
	}

	srcType, _ := codegen.GetMetaType(src)
	tgtType, _ := codegen.GetMetaType(tgt)
	if srcType == "" && tgtType == "" && (src.Type != expr.Int) && (src.Type != expr.UInt) && (src.Type != expr.Any) {
		// Any values need a protobuf conversion. Other matching values do not.
		if !proto && srcPtr && !tgtPtr {
			return "*" + srcVar
		}
		return srcVar
	}

	if proto {
		return convertPrimitiveToProto(src, tgt, srcPtr, tgtPtr, srcVar)
	}
	return convertPrimitiveFromProto(src, tgt, srcPtr, tgtPtr, srcVar, ta)
}

// protoUnionBranchUsesHelper reports whether protobuf union rendering emits a
// TransformHelperName call for the branch. Planning and rendering use this
// same rule so their helper order cannot differ.
func protoUnionBranchUsesHelper(source, target *expr.AttributeExpr) bool {
	if expr.IsAlias(source.Type) || expr.IsAlias(target.Type) {
		return unAlias(source).Type != unAlias(target).Type
	}
	_, named := source.Type.(expr.UserType)
	return named
}

const convertGoAnyToProtobufValueFunc = `func() *structpb.Value {
	// Convert Go any to protobuf Value directly
	if %s == nil {
		return structpb.NewNullValue()
	}
	value, err := structpb.NewValue(%s)
	if err != nil {
		panic(fmt.Sprintf("failed to convert value to structpb.Value: %%v", err))
	}
	return value
}()`

const convertProtobufValueToGoAnyFunc = `func() any {
	// Convert protobuf Value to Go any directly
	if %s != nil {
		return %s.AsInterface()
	}
	return nil
}()`

// convertPrimitiveToProto returns the code to convert a primitive type to its
// protocol buffer representation.
func convertPrimitiveToProto(_, tgt *expr.AttributeExpr, srcPtr, _ bool, srcVar string) string {
	// Any values use google.protobuf.Value in protobuf messages.
	if tgt.Type.Kind() == expr.AnyKind {
		if srcPtr {
			srcVar = "*" + srcVar
		}

		return fmt.Sprintf(convertGoAnyToProtobufValueFunc, srcVar, srcVar)
	}

	tgtType := protoBufNativeGoTypeName(tgt.Type)
	if srcPtr {
		srcVar = "*" + srcVar
	}
	return fmt.Sprintf("%s(%s)", tgtType, srcVar)
}

// convertPrimitiveFromProto returns the code to convert the protocol buffer
// representation of a primitive type back to the service type.
func convertPrimitiveFromProto(_, tgt *expr.AttributeExpr, srcPtr, _ bool, srcVar string, ta *codegen.TransformAttrs) string {
	// Any values arrive from protobuf as google.protobuf.Value.
	if tgt.Type.Kind() == expr.AnyKind {
		if srcPtr {
			srcVar = "*" + srcVar
		}

		return fmt.Sprintf(convertProtobufValueToGoAnyFunc, srcVar, srcVar)
	}

	tgtType := ta.TargetCtx.Scope.Ref(tgt, ta.TargetCtx.Pkg(tgt))
	if srcPtr {
		srcVar = "*" + srcVar
	}
	return fmt.Sprintf("%s(%s)", tgtType, srcVar)
}
