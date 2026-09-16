// This file verifies other generators can read the exact Go fields already
// chosen for a service method payload.
package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
)

// TestMethodPayloadLayoutReturnsAssignedFieldNames verifies field metadata is
// applied once by the service planner and exposed without rebuilding the name.
func TestMethodPayloadLayoutReturnsAssignedFieldNames(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		query := dsl.Type("ReadQuery", func() {
			dsl.Attribute("cursor", dsl.String, func() {
				dsl.Meta("struct:field:name", "OriginalCursor")
			})
		})
		dsl.Service("Documents", func() {
			dsl.Method("Read", func() {
				dsl.Payload(query)
			})
		})
	})
	plan := mustServicePlan(t, root)

	layout, err := plan.MethodPayloadLayout(root.Service("Documents").Method("Read"))
	require.NoError(t, err)
	require.Equal(t, codegen.GoStruct, layout.Kind())
	require.Len(t, layout.Fields(), 1)
	require.Equal(t, "OriginalCursor", layout.Fields()[0].FieldName(true))
	require.True(t, layout.Fields()[0].IsPointer())
}
