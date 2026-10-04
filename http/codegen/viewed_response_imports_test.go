// These tests render HTTP clients whose response view is chosen from the
// response header. The client decoder reports an unknown view with Goa's enum
// error, so the planned codec imports must include the goa package.
package codegen

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

// TestHTTPViewedResponseImportsGoa checks the planned client codec imports and
// compiles the generated service, client and server.
func TestHTTPViewedResponseImportsGoa(t *testing.T) {
	tests := []struct {
		name       string
		collection bool
		fixedView  bool
		goa        bool
	}{
		{name: "result_type", goa: true},
		{name: "collection", collection: true, goa: true},
		{name: "fixed_view", fixedView: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := viewedResponseRoot(t, test.collection, test.fixedView)
			plan := linkedHTTPPlanForRoot(t, root)
			imports := plannedHTTPFileImports(t, plan.ClientFiles(), "/client/encode_decode.go")
			require.Equal(t, test.goa, slices.Contains(imports, codegen.GoaImport("").Path))

			serviceFiles, err := service.Files(plan.servicePlan)
			require.NoError(t, err)
			files := slices.Clone(serviceFiles)
			files = append(files, plan.ClientFiles()...)
			files = append(files, plan.ClientTypeFiles()...)
			files = append(files, plan.ServerFiles()...)
			files = append(files, plan.ServerTypeFiles()...)
			files = append(files, plan.PathFiles()...)
			runGeneratedMixedSSECompile(t, files, "./gen/...")
		})
	}
}

// viewedResponseRoot returns a method whose result type has two views. Unless
// fixedView is set, the HTTP design leaves the view to the response.
func viewedResponseRoot(t *testing.T, collection, fixedView bool) *expr.RootExpr {
	t.Helper()
	return expr.RunDSL(t, func() {
		item := dsl.ResultType("application/vnd.item", func() {
			dsl.TypeName("Item")
			dsl.Attribute("name", dsl.String)
			dsl.View("default", func() {
				dsl.Attribute("name")
			})
			dsl.View("tiny", func() {
				dsl.Attribute("name")
			})
		})
		dsl.Service("Views", func() {
			dsl.Method("List", func() {
				var result any = item
				if collection {
					result = dsl.CollectionOf(item)
				}
				if fixedView {
					dsl.Result(result, func() {
						dsl.View("tiny")
					})
				} else {
					dsl.Result(result)
				}
				dsl.HTTP(func() {
					dsl.GET("/")
				})
			})
		})
	})
}
