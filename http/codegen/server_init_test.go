package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestServerInit(t *testing.T) {
	const genpkg = "gen"
	cases := []struct {
		Name       string
		DSL        func()
		Code       string
		FileCount  int
		SectionNum int
	}{
		{"multiple endpoints", testdata.ServerMultiEndpointsDSL, testdata.ServerMultiEndpointsConstructorCode, 2, 3},
		{"multiple bases", testdata.ServerMultiBasesDSL, testdata.ServerMultiBasesConstructorCode, 2, 3},
		{"file server", testdata.ServerFileServerDSL, testdata.ServerFileServerConstructorCode, 1, 3},
		{"file server with a redirect", testdata.ServerFileServerWithRedirectDSL, testdata.ServerFileServerConstructorCode, 1, 3},
		{"mixed", testdata.ServerMixedDSL, testdata.ServerMixedConstructorCode, 2, 3},
		{"multipart", testdata.ServerMultipartDSL, testdata.ServerMultipartConstructorCode, 2, 4},
		{"streaming", testdata.StreamingResultDSL, testdata.ServerStreamingConstructorCode, 3, 3},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			fs := ServerFiles(genpkg, expr.Root)
			require.Len(t, fs, c.FileCount)
			sections := fs[0].SectionTemplates
			require.Greater(t, len(sections), c.SectionNum)
			code := codegen.SectionCode(t, sections[c.SectionNum])
			assert.Equal(t, c.Code, code)
		})
	}
}
