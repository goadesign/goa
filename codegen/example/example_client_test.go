package example

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example/testdata"
	"goa.design/goa/v3/expr"
)

func TestExampleCLIFiles(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"no-server", testdata.NoServerDSL, testdata.NoServerCLIMainCode},
		{"single-server-single-host", testdata.SingleServerSingleHostDSL, testdata.SingleServerSingleHostCLIMainCode},
		{"single-server-single-host-with-variables", testdata.SingleServerSingleHostWithVariablesDSL, testdata.SingleServerSingleHostWithVariablesCLIMainCode},
		{"single-server-multiple-hosts", testdata.SingleServerMultipleHostsDSL, testdata.SingleServerMultipleHostsCLIMainCode},
		{"single-server-multiple-hosts-with-variables", testdata.SingleServerMultipleHostsWithVariablesDSL, testdata.SingleServerMultipleHostsWithVariablesCLIMainCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			// reset global variable
			Servers = make(ServersData)
			codegen.RunDSL(t, c.DSL)
			fs := CLIFiles("", expr.Root)
			require.Len(t, fs, 1)
			require.Greater(t, len(fs[0].SectionTemplates), 0)
			var buf bytes.Buffer
			for _, s := range fs[0].SectionTemplates[1:] {
				require.NoError(t, s.Write(&buf))
			}
			code := codegen.FormatTestCode(t, "package foo\n"+buf.String())
			assert.Equal(t, c.Code, code)
		})
	}
}
