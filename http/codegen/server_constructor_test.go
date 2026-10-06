// These tests validate plugin declarations before rendering server source.
// Required dependencies must have distinct private names, retain their planned
// types, and produce the same ordered arguments when registration order changes.
package codegen

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	ctestdata "goa.design/goa/v3/codegen/example/testdata"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestPlanOrdersServerConstructorDependencies(t *testing.T) {
	root := extensionRoot(t)
	plan, generation, servicePlan := plannedHTTPPlan(t, root, false)
	service := root.API.HTTP.Services[0]
	dependencyType := constructorTestType(t)
	zulu, err := plan.DeclareServerConstructorDependency(service, "zulu", dependencyType, "NewDependency", extensionNameOrder("zulu"))
	require.NoError(t, err)
	alpha, err := plan.DeclareServerConstructorDependency(service, "alpha", dependencyType, "NewDependency", extensionNameOrder("alpha"))
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, plan.Link())
	data, ok := plan.Service(service)
	require.True(t, ok)
	require.Len(t, data.ConstructorDependencies, 2)
	require.Equal(t, "alpha", data.ConstructorDependencies[0].Name)
	require.Equal(t, "zulu", data.ConstructorDependencies[1].Name)
	require.Equal(t, "*http.Client", data.ConstructorDependencies[0].TypeRef)
	require.Equal(t, "NewDependency", alpha.Name())
	require.Equal(t, "NewDependency2", zulu.Name())
	_, err = plan.DeclareServerConstructorDependency(service, "late", dependencyType, "NewLate", extensionNameOrder("late"))
	require.ErrorContains(t, err, "before generation freeze and plan linking")
}

func TestJSONRPCConstructorDependenciesAreIndependent(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.Service("RPC", func() {
			dsl.Method("Read", func() { dsl.JSONRPC(func() {}) })
		})
	})
	plan, generation, servicePlan := plannedHTTPPlan(t, root, true)
	_, err := plan.DeclareServerConstructorDependency(root.API.JSONRPC.Services[0], "http", constructorTestType(t), "NewHTTP", extensionNameOrder("http"))
	require.ErrorContains(t, err, "conflicts with the HTTP package")
	_, err = plan.DeclareServerConstructorDependency(root.API.JSONRPC.Services[0], "client", constructorTestType(t), "NewClient", extensionNameOrder("client"))
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, plan.Link())
	data, ok := plan.JSONRPCService("RPC")
	require.True(t, ok)
	require.Len(t, data.ConstructorDependencies, 1)
	data.ConstructorDependencies[0].Name = "changed"
	again, ok := plan.JSONRPCService("RPC")
	require.True(t, ok)
	require.Equal(t, "client", again.ConstructorDependencies[0].Name)
}

func TestPlanRejectsInvalidServerConstructorDependencies(t *testing.T) {
	for _, name := range []string{"", "Public", "_", "bad-name", "e", "endpoints", "mux", "decoder", "encoder", "errhandler", "formatter", "upgrader", "configurer", "s", "http"} {
		t.Run(name, func(t *testing.T) {
			root := extensionRoot(t)
			plan, _, _ := plannedHTTPPlan(t, root, false)
			_, err := plan.DeclareServerConstructorDependency(root.API.HTTP.Services[0], name, constructorTestType(t), "NewDependency", extensionNameOrder("invalid"))
			require.Error(t, err)
		})
	}
	root := extensionRoot(t)
	plan, _, _ := plannedHTTPPlan(t, root, false)
	service := root.API.HTTP.Services[0]
	dependencyType := constructorTestType(t)
	_, err := plan.DeclareServerConstructorDependency(nil, "client", dependencyType, "NewDependency", extensionNameOrder("nil"))
	require.ErrorContains(t, err, "requires a service from this plan")
	foreign := extensionRoot(t).API.HTTP.Services[0]
	_, err = plan.DeclareServerConstructorDependency(foreign, "client", dependencyType, "NewDependency", extensionNameOrder("foreign"))
	require.ErrorContains(t, err, "requires a service from this plan")
	_, err = plan.DeclareServerConstructorDependency(service, "client", nil, "NewDependency", extensionNameOrder("type"))
	require.ErrorContains(t, err, "requires a retained Go type")
	_, err = plan.DeclareServerConstructorDependency(service, "client", dependencyType, "", extensionNameOrder("factory"))
	require.Error(t, err)
	_, err = plan.DeclareServerConstructorDependency(service, "client", dependencyType, "NewDependency", nil)
	require.Error(t, err)
	_, err = plan.DeclareServerConstructorDependency(service, "client", dependencyType, "NewDependency", extensionNameOrder("first"))
	require.NoError(t, err)
	_, err = plan.DeclareServerConstructorDependency(service, "client", dependencyType, "NewDependency", extensionNameOrder("second"))
	require.ErrorContains(t, err, "already declared")
}

func TestPlanRejectsDesignedConstructorArgumentCollisions(t *testing.T) {
	for _, kind := range []string{"multipart", "file"} {
		t.Run(kind, func(t *testing.T) {
			root := releasedHTTPNamesRoot(t)
			baseline, generation, servicePlan := plannedHTTPPlan(t, root, false)
			require.NoError(t, generation.Freeze())
			require.NoError(t, servicePlan.Link())
			require.NoError(t, baseline.Link())
			data, ok := baseline.Service(root.API.HTTP.Services[0])
			require.True(t, ok)
			var argument string
			if kind == "file" {
				argument = data.FileServers[0].ArgName
			} else {
				for _, endpoint := range data.Endpoints {
					if endpoint.MultipartRequestDecoder != nil {
						argument = endpoint.MultipartRequestDecoder.VarName
						break
					}
				}
			}
			require.NotEmpty(t, argument)
			root = releasedHTTPNamesRoot(t)
			plan, generation, servicePlan := plannedHTTPPlan(t, root, false)
			_, err := plan.DeclareServerConstructorDependency(root.API.HTTP.Services[0], argument, constructorTestType(t), "NewDependency", extensionNameOrder(kind))
			require.NoError(t, err)
			require.NoError(t, generation.Freeze())
			require.NoError(t, servicePlan.Link())
			require.ErrorContains(t, plan.Link(), "conflicts with")
		})
	}
}

func TestConstructorDependencyImportsOnlyConfiguredServers(t *testing.T) {
	root := codegen.RunDSL(t, ctestdata.ServerHostingServiceSubsetDSL)
	plan, generation, servicePlan := plannedHTTPPlan(t, root, false)
	examplePlan, err := example.NewPlan(generation, servicePlan)
	require.NoError(t, err)
	examples, err := NewExamplePlan(plan, examplePlan)
	require.NoError(t, err)
	_, err = plan.DeclareServerConstructorDependency(root.API.HTTP.Service("IgnoredService"), "client", constructorTestType(t), "NewIgnoredClient", extensionNameOrder("ignored"))
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, plan.Link())
	files := examples.ServerFiles()
	require.Len(t, files, 2)
	imports := importPaths(files[0].SectionTemplates[0].Data.(map[string]any)["Imports"].([]*codegen.ImportSpec))
	require.NotContains(t, imports, path.Dir(generation.GenPkg()))
	require.NotContains(t, renderedHTTPFile(t, files[0]), "NewIgnoredClient")
	require.Contains(t, renderedHTTPFile(t, files[1]), "NewIgnoredClient")
}

// constructorTestType supplies an imported pointer type through Goa's existing
// custom-type metadata, so constructor tests exercise the normal import planner.
func constructorTestType(t *testing.T) *codegen.GoTypePlan {
	t.Helper()
	result, err := codegen.PlanGoType(&expr.AttributeExpr{
		Type: expr.String,
		Meta: expr.MetaExpr{"struct:field:type": {"*http.Client", "net/http", "http"}},
	}, codegen.GoTypePlanOptions{Owner: "generated.local"})
	require.NoError(t, err)
	return result
}
