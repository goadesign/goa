// This file verifies other generators can read the exact Go fields already
// chosen for a service method result.
package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
)

// TestMethodResultLayoutReturnsAssignedFieldNames verifies field metadata is
// applied once by the service planner and exposed without rebuilding the name.
func TestMethodResultLayoutReturnsAssignedFieldNames(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		result := dsl.ResultType("ReadResult", func() {
			dsl.Attribute("cursor", dsl.String, func() {
				dsl.Meta("struct:field:name", "OriginalCursor")
			})
		})
		dsl.Service("Documents", func() {
			dsl.Method("Read", func() {
				dsl.Result(result)
			})
		})
	})
	plan := mustServicePlan(t, root)

	layout, err := plan.MethodResultLayout(root.Service("Documents").Method("Read"))
	require.NoError(t, err)
	require.Equal(t, codegen.GoStruct, layout.Kind())
	require.Len(t, layout.Fields(), 1)
	require.Equal(t, "OriginalCursor", layout.Fields()[0].FieldName(true))
	require.True(t, layout.Fields()[0].IsPointer())
}
