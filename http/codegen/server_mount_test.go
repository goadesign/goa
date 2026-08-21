package codegen

import (
	"testing"

	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/expr"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/codegentest"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestServerMount(t *testing.T) {
	const genpkg = "gen"
	cases := []struct {
		Name       string
		DSL        func()
		SectionNum int

		SectionName string
	}{
		{"simple routing constructor", testdata.ServerSimpleRoutingDSL, 0, "server-mount"},
		{"simple routing with a redirect constructor", testdata.ServerSimpleRoutingWithRedirectDSL, 0, "server-mount"},
		{"multiple_files_constructor", testdata.ServerMultipleFilesDSL, 0, "server-mount"},
		{"multiple_files_mounter", testdata.ServerMultipleFilesDSL, 3, "server-files"},
		{"multiple_files_constructor_w_prefix_path", testdata.ServerMultipleFilesWithPrefixPathDSL, 0, "server-mount"},
		{"multiple_files_mounter_w_prefix_path", testdata.ServerMultipleFilesWithPrefixPathDSL, 3, "server-files"},
		{"multiple_files_with_a_redirect_constructor", testdata.ServerMultipleFilesWithRedirectDSL, 0, "server-mount"},
		{"multiple_files_with_a_redirect_mounter", testdata.ServerMultipleFilesWithRedirectDSL, 3, "server-files"},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := expr.RunDSL(t, c.DSL)
			services := CreateHTTPServices(root)
			fs := ServerFiles(services)
			sections := codegentest.Sections(fs, "server.go", c.SectionName)
			require.Greater(t, len(sections), c.SectionNum)
			code := codegen.SectionCode(t, sections[c.SectionNum])
			testutil.AssertGo(t, "testdata/golden/server_mount_"+c.Name+".go.golden", code)
		})
	}
}
