// This file checks the handler wrappers and extra routes that plugins declare
// before Goa chooses generated package names.
package codegen

import (
	"cmp"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

type extensionNameOrder string

// ComparePackageName gives extension declarations a stable order in tests.
func (o extensionNameOrder) ComparePackageName(other codegen.PackageNameOrder) int {
	return cmp.Compare(string(o), string(other.(extensionNameOrder)))
}

func TestPlanDeclaresServerExtensions(t *testing.T) {
	root := extensionRoot(t)
	plan, generation, servicePlan := plannedHTTPPlan(t, root, false)
	serviceExpr := root.API.HTTP.Services[0]

	first, err := plan.DeclareServerHandlerWrapper(serviceExpr, "WrapHandler", extensionNameOrder("first"))
	require.NoError(t, err)
	second, err := plan.DeclareServerHandlerWrapper(serviceExpr, "WrapHandler", extensionNameOrder("second"))
	require.NoError(t, err)
	endpointWrapper, err := plan.DeclareServerEndpointHandlerWrapper(serviceExpr.HTTPEndpoints[0], "wrapEndpoint", extensionNameOrder("endpoint"))
	require.NoError(t, err)
	secondEndpointWrapper, err := plan.DeclareServerEndpointHandlerWrapper(serviceExpr.HTTPEndpoints[0], "wrapEndpoint", extensionNameOrder("second endpoint"))
	require.NoError(t, err)
	descriptions := []ServerMountPoint{{Method: "Preflight", Verb: "OPTIONS", Pattern: "/items/{id}"}}
	mount, err := plan.DeclareServerMount(serviceExpr, "MountPreflight", extensionNameOrder("mount"), descriptions)
	require.NoError(t, err)
	descriptions[0].Method = "changed"

	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, plan.Link())

	data := plan.services.Get("Files")
	require.Equal(t, "WrapHandler", first.Name())
	require.Equal(t, "WrapHandler2", second.Name())
	require.Equal(t, []*codegen.NameDeclaration{first, second}, data.ServerHandlerWrappers)
	require.Equal(t, "wrapEndpoint", endpointWrapper.Name())
	require.Equal(t, "wrapEndpoint2", secondEndpointWrapper.Name())
	require.Equal(t, []*codegen.NameDeclaration{first, second, endpointWrapper, secondEndpointWrapper}, data.Endpoints[0].ServerHandlerWrappers)
	for _, fileServer := range data.FileServers {
		require.Equal(t, []*codegen.NameDeclaration{first, second}, fileServer.ServerHandlerWrappers)
	}
	require.Equal(t, mount, data.ServerMounts[0].Declaration)
	require.Equal(t, []ServerMountPoint{{Method: "Preflight", Verb: "OPTIONS", Pattern: "/items/{id}"}}, data.ServerMounts[0].MountPoints)
}

func TestPlanRejectsInvalidServerExtensions(t *testing.T) {
	root := extensionRoot(t)
	plan, generation, servicePlan := plannedHTTPPlan(t, root, false)
	serviceExpr := root.API.HTTP.Services[0]
	foreignRoot := extensionRoot(t)
	foreignService := foreignRoot.API.HTTP.Services[0]
	foreignEndpoint := foreignService.HTTPEndpoints[0]

	tests := []struct {
		name string
		call func() error
		want string
	}{
		{"nil service", func() error {
			_, err := plan.DeclareServerHandlerWrapper(nil, "Wrap", extensionNameOrder("nil"))
			return err
		}, "HTTP server extension requires a service from this plan"},
		{"foreign service", func() error {
			_, err := plan.DeclareServerHandlerWrapper(foreignService, "Wrap", extensionNameOrder("foreign"))
			return err
		}, "HTTP service does not belong to this plan"},
		{"nil endpoint", func() error {
			_, err := plan.DeclareServerEndpointHandlerWrapper(nil, "wrap", extensionNameOrder("nil endpoint"))
			return err
		}, "HTTP server endpoint wrapper requires an endpoint from this plan"},
		{"foreign endpoint", func() error {
			_, err := plan.DeclareServerEndpointHandlerWrapper(foreignEndpoint, "wrap", extensionNameOrder("foreign endpoint"))
			return err
		}, "HTTP endpoint does not belong to this plan"},
		{"empty preferred name", func() error {
			_, err := plan.DeclareServerHandlerWrapper(serviceExpr, "", extensionNameOrder("empty"))
			return err
		}, "package name must not be empty"},
		{"nil order", func() error {
			_, err := plan.DeclareServerHandlerWrapper(serviceExpr, "Wrap", nil)
			return err
		}, `generated package "generated.local/gen/http/files/server" cannot declare preferred function "Wrap": package name order must be a stable concrete named value type`},
		{"no mount descriptions", func() error {
			_, err := plan.DeclareServerMount(serviceExpr, "Mount", extensionNameOrder("none"), nil)
			return err
		}, "HTTP server mount requires at least one mount point"},
		{"empty method", func() error {
			_, err := plan.DeclareServerMount(serviceExpr, "Mount", extensionNameOrder("method"), []ServerMountPoint{{Verb: "OPTIONS", Pattern: "/"}})
			return err
		}, "HTTP server mount point 0 has an empty method"},
		{"empty verb", func() error {
			_, err := plan.DeclareServerMount(serviceExpr, "Mount", extensionNameOrder("verb"), []ServerMountPoint{{Method: "Preflight", Pattern: "/"}})
			return err
		}, "HTTP server mount point 0 has an empty verb"},
		{"empty pattern", func() error {
			_, err := plan.DeclareServerMount(serviceExpr, "Mount", extensionNameOrder("pattern"), []ServerMountPoint{{Method: "Preflight", Verb: "OPTIONS"}})
			return err
		}, "HTTP server mount point 0 has an empty pattern"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.EqualError(t, test.call(), test.want)
		})
	}

	require.NoError(t, generation.Freeze())
	_, err := plan.DeclareServerHandlerWrapper(serviceExpr, "Late", extensionNameOrder("late"))
	require.EqualError(t, err, "HTTP server extension cannot be declared after generation freeze")
	_, err = plan.DeclareServerEndpointHandlerWrapper(serviceExpr.HTTPEndpoints[0], "late", extensionNameOrder("late endpoint"))
	require.EqualError(t, err, "HTTP server extension cannot be declared after generation freeze")
	require.NoError(t, servicePlan.Link())
	require.NoError(t, plan.Link())
	_, err = plan.DeclareServerHandlerWrapper(serviceExpr, "Linked", extensionNameOrder("linked"))
	require.EqualError(t, err, "HTTP server extension cannot be declared after plan linking")
	_, err = plan.DeclareServerEndpointHandlerWrapper(serviceExpr.HTTPEndpoints[0], "linked", extensionNameOrder("linked endpoint"))
	require.EqualError(t, err, "HTTP server extension cannot be declared after plan linking")
}

func TestJSONRPCPlanRejectsServerExtensions(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.Service("RPC", func() {
			dsl.Method("Read", func() { dsl.JSONRPC(func() {}) })
		})
	})
	plan, _, _ := plannedHTTPPlan(t, root, true)
	_, err := plan.DeclareServerHandlerWrapper(root.API.JSONRPC.Services[0], "Wrap", extensionNameOrder("rpc"))
	require.EqualError(t, err, "JSON-RPC HTTP plans do not support server extensions")
	_, err = plan.DeclareServerEndpointHandlerWrapper(root.API.JSONRPC.Services[0].HTTPEndpoints[0], "wrap", extensionNameOrder("rpc endpoint"))
	require.EqualError(t, err, "JSON-RPC HTTP plans do not support server extensions")
}

// extensionRoot builds the HTTP service used to test endpoint and file wrappers.
func extensionRoot(t *testing.T) *expr.RootExpr {
	t.Helper()
	return expr.RunDSL(t, func() {
		dsl.Service("Files", func() {
			dsl.Method("Read", func() {
				dsl.Payload(func() { dsl.Attribute("id", dsl.String) })
				dsl.HTTP(func() { dsl.GET("/items/{id}") })
			})
			dsl.Files("/assets/{*path}", "assets")
			dsl.Files("/old", "old.html", func() { dsl.Redirect("/new", 301) })
		})
	})
}

// plannedHTTPPlan creates an HTTP or JSON-RPC plan without linking it so each
// test can add server functions before generated names become final.
func plannedHTTPPlan(t *testing.T, root *expr.RootExpr, jsonrpc bool) (*Plan, *codegen.Generation, *service.Plan) {
	t.Helper()
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	input := PlanInput{Root: root, Service: servicePlan}
	var plans []*Plan
	if jsonrpc {
		plans, err = NewJSONRPCPlans(generation, input)
	} else {
		plans, err = NewPlans(generation, input)
	}
	require.NoError(t, err)
	require.Len(t, plans, 1)
	return plans[0], generation, servicePlan
}
