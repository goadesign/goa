// This file compares generated server source for plugin-declared handler
// wrappers and additional routes with reviewed golden files.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/codegentest"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestServerExtensions(t *testing.T) {
	root := extensionRoot(t)
	plan, generation, servicePlan := plannedHTTPPlan(t, root, false)
	serviceExpr := root.API.HTTP.Services[0]
	_, err := plan.DeclareServerHandlerWrapper(serviceExpr, "First", extensionNameOrder("first"))
	require.NoError(t, err)
	_, err = plan.DeclareServerHandlerWrapper(serviceExpr, "Second", extensionNameOrder("second"))
	require.NoError(t, err)
	_, err = plan.DeclareServerEndpointHandlerWrapper(serviceExpr.HTTPEndpoints[0], "wrapEndpoint", extensionNameOrder("endpoint"))
	require.NoError(t, err)
	_, err = plan.DeclareServerMount(serviceExpr, "MountPreflight", extensionNameOrder("mount"), []ServerMountPoint{
		{Method: "Preflight item", Verb: "OPTIONS", Pattern: "/items/{id}"},
		{Method: "Preflight assets", Verb: "OPTIONS", Pattern: "/assets/{*path}"},
	})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, plan.Link())

	files := plan.ServerFiles()
	for _, test := range []struct {
		section      string
		sectionIndex int
		golden       string
	}{
		{"server-mount", 0, "testdata/golden/server_extensions_mount.go.golden"},
		{"server-handler", 0, "testdata/golden/server_extensions_endpoint_helper.go.golden"},
		{"server-files", 0, "testdata/golden/server_extensions_file_helper.go.golden"},
		{"server-files", 1, "testdata/golden/server_extensions_redirect_helper.go.golden"},
		{"server-init", 0, "testdata/golden/server_extensions_init.go.golden"},
	} {
		sections := codegentest.Sections(files, "server.go", test.section)
		require.Greater(t, len(sections), test.sectionIndex)
		testutil.AssertGo(t, test.golden, codegen.SectionCode(t, sections[test.sectionIndex]))
	}
}

func TestServerExtensionMountPointEscaping(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.Service("Escape", func() {
			dsl.Method("Ping", func() { dsl.HTTP(func() { dsl.GET("/") }) })
		})
	})
	plan, generation, servicePlan := plannedHTTPPlan(t, root, false)
	_, err := plan.DeclareServerMount(root.API.HTTP.Services[0], "MountQuoted", extensionNameOrder("quoted"), []ServerMountPoint{{
		Method:  "Quoted \"method\"\nnext",
		Verb:    "CUSTOM\\VERB",
		Pattern: "/quoted/\"value\"\\next\nline",
	}})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, plan.Link())

	sections := codegentest.Sections(plan.ServerFiles(), "server.go", "server-init")
	require.Len(t, sections, 1)
	testutil.AssertGo(t, "testdata/golden/server_extensions_escaping.go.golden", codegen.SectionCode(t, sections[0]))
}
