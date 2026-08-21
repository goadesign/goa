// This file verifies HTTP import planning participates in the shared
// generation lifecycle before service aliases are frozen.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestPlanReservesStaticAliasesBeforeFreeze(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.Service("Path", func() {
			dsl.Method("Read", func() {})
		})
	})
	generation := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, service.Plan(root, generation))
	require.NoError(t, Plan(generation))
	require.NoError(t, generation.Freeze())
	services, err := service.NewServicesData(root, generation)
	require.NoError(t, err)

	require.Equal(t, "path2", services.ServiceImport("Path").Name)
}

func TestPlanRejectsFrozenGeneration(t *testing.T) {
	generation := codegen.NewGeneration("generated.local/gen", nil)
	require.NoError(t, generation.Freeze())

	require.Error(t, Plan(generation))
}
