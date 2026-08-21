// This file verifies standalone JSON-RPC planning includes the HTTP codecs and
// helpers that JSON-RPC rendering reuses.
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

func TestPlanIncludesSharedHTTPImportAliases(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.Service("UUID", func() {
			dsl.Method("Read", func() {
				dsl.JSONRPC(func() {})
			})
		})
	})
	generation := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, service.Plan(root, generation))
	require.NoError(t, Plan(generation))
	require.NoError(t, generation.Freeze())
	services, err := service.NewServicesData(root, generation)
	require.NoError(t, err)

	require.Equal(t, "uuid2", services.ServiceImport("UUID").Name)
}
