package codegen

import (
	"testing"

	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/expr"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestServerInit(t *testing.T) {
	const genpkg = "gen"
	cases := []struct {
		Name       string
		DSL        func()
		FileCount  int
		SectionNum int
	}{
		{"multiple endpoints", testdata.ServerMultiEndpointsDSL, 2, 3},
		{"multiple bases", testdata.ServerMultiBasesDSL, 2, 3},
		{"file server", testdata.ServerFileServerDSL, 1, 3},
		{"file server with a redirect", testdata.ServerFileServerWithRedirectDSL, 1, 3},
		{"file server with root path", testdata.ServerFileServerRootPathDSL, 1, 3},
		{"mixed", testdata.ServerMixedDSL, 2, 3},
		{"multipart", testdata.ServerMultipartDSL, 2, 4},
		{"streaming", testdata.StreamingResultDSL, 3, 3},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := expr.RunDSL(t, c.DSL)
			services := CreateHTTPServices(root)
			fs := ServerFiles(services)
			require.Len(t, fs, c.FileCount)
			sections := fs[0].SectionTemplates
			require.Greater(t, len(sections), c.SectionNum)
			code := codegen.SectionCode(t, sections[c.SectionNum])
			testutil.AssertGo(t, "testdata/golden/server_init_"+c.Name+".go.golden", code)
		})
	}
}
