// This file verifies that gRPC metadata uses detached native wire values and
// canonical conversions to frozen service declarations.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestMetadataConversionUsesDetachedWireAndFrozenServiceDeclaration(t *testing.T) {
	root := expr.RunDSL(t, func() {
		value := dsl.Type("Value", dsl.String, func() {
			dsl.Meta("struct:pkg:path", "domain/shared")
		})
		payload := dsl.Type("Payload", func() {
			dsl.Field(1, "value", value)
			dsl.Required("value")
		})
		dsl.Service("Values", func() {
			dsl.Method("Read", func() {
				dsl.Payload(payload)
				dsl.GRPC(func() {
					dsl.Metadata(func() { dsl.Attribute("value") })
				})
			})
		})
	})

	metadata := CreateGRPCServices(root).Get("Values").Endpoint("Read").Request.Metadata
	require.Len(t, metadata, 1)
	require.False(t, metadata[0].Map)
	require.False(t, metadata[0].MapStringSlice)
	require.Equal(t, "string", metadata[0].TypeRef)
	require.NotContains(t, metadata[0].WireAttribute.Meta, "struct:pkg:path")
	require.Contains(t, metadata[0].EncodeCode, "string(payload.Value)")
	require.Contains(t, initArgsFromMetadata(metadata)[0].InitCode, "shared.Value(value)")
}

func TestMetadataConversionRecursivelyDetachesNamedArrayElements(t *testing.T) {
	root := expr.RunDSL(t, func() {
		value := dsl.Type("Value", dsl.Int, func() {
			dsl.Enum(1, 2)
			dsl.Meta("struct:pkg:path", "domain/shared")
		})
		values := dsl.Type("Values", dsl.ArrayOf(value), func() {
			dsl.Meta("struct:pkg:path", "domain/shared")
		})
		payload := dsl.Type("Payload", func() {
			dsl.Field(1, "values", values)
			dsl.Required("values")
		})
		dsl.Service("Values", func() {
			dsl.Method("Read", func() {
				dsl.Payload(payload)
				dsl.GRPC(func() {
					dsl.Metadata(func() { dsl.Attribute("values") })
				})
			})
		})
	})

	serviceField := root.API.GRPC.Services[0].GRPCEndpoints[0].MethodExpr.Payload.Find("values")
	metadata := CreateGRPCServices(root).Get("Values").Endpoint("Read").Request.Metadata
	require.Len(t, metadata, 1)
	wireArray := expr.AsArray(metadata[0].WireAttribute.Type)
	require.NotNil(t, wireArray)
	require.Equal(t, expr.Int, wireArray.ElemType.Type)
	require.Equal(t, []any{1, 2}, wireArray.ElemType.Validation.Values)
	require.NotContains(t, metadata[0].WireAttribute.Meta, "struct:pkg:path")
	require.NotContains(t, wireArray.ElemType.Meta, "struct:pkg:path")
	require.Contains(t, metadata[0].EncodeCode, "int(val)")
	require.Contains(t, initArgsFromMetadata(metadata)[0].InitCode, "shared.Value(val)")
	require.Contains(t, serviceField.Type.(expr.UserType).Attribute().Meta, "struct:pkg:path")
	require.Contains(t, expr.AsArray(serviceField.Type).ElemType.Type.(expr.UserType).Attribute().Meta, "struct:pkg:path")
}
