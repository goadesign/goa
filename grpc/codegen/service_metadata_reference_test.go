// This file verifies that gRPC metadata casts use frozen service declaration
// references instead of rebuilding type names from DSL locations.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestMetadataFieldTypeRefUsesFrozenServiceDeclaration(t *testing.T) {
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
	require.Equal(t, "shared.Value", metadata[0].FieldTypeRef)
	require.Equal(t, metadata[0].FieldTypeRef, initArgsFromMetadata(metadata)[0].FieldTypeRef)
}
