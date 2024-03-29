package codegen

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/codegentest"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestServerMount(t *testing.T) {
	const genpkg = "gen"
	cases := []struct {
		Name       string
		DSL        func()
		Code       string
		SectionNum int

		SectionName string
	}{
		{"simple routing constructor", testdata.ServerSimpleRoutingDSL, testdata.ServerSimpleRoutingConstructorCode, 0, "server-mount"},
		{"simple routing with a redirect constructor", testdata.ServerSimpleRoutingWithRedirectDSL, testdata.ServerSimpleRoutingConstructorCode, 0, "server-mount"},
		{"multiple files constructor", testdata.ServerMultipleFilesDSL, testdata.ServerMultipleFilesConstructorCode, 0, "server-mount"},
		{"multiple files mounter", testdata.ServerMultipleFilesDSL, testdata.ServerMultipleFilesMounterCode, 3, "server-files"},
		{"multiple files constructor /w prefix path", testdata.ServerMultipleFilesWithPrefixPathDSL, testdata.ServerMultipleFilesWithPrefixPathConstructorCode, 0, "server-mount"},
		{"multiple files mounter /w prefix path", testdata.ServerMultipleFilesWithPrefixPathDSL, testdata.ServerMultipleFilesWithPrefixPathMounterCode, 3, "server-files"},
		{"multiple files with a redirect constructor", testdata.ServerMultipleFilesWithRedirectDSL, testdata.ServerMultipleFilesWithRedirectConstructorCode, 0, "server-mount"},
		{"multiple files with a redirect mounter", testdata.ServerMultipleFilesWithRedirectDSL, testdata.ServerMultipleFilesMounterCode, 3, "server-files"},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			fs := ServerFiles(genpkg, expr.Root)
			sections := codegentest.Sections(fs, filepath.Join("", "server.go"), c.SectionName)
			require.Greater(t, len(sections), c.SectionNum)
			code := codegen.SectionCode(t, sections[c.SectionNum])
			assert.Equal(t, c.Code, code)
		})
	}
}
