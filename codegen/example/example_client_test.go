package example

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example/testdata"
	"goa.design/goa/v3/expr"
)

func TestExampleCLIFiles(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"no-server", testdata.NoServerDSL},
		{"single-server-single-host", testdata.SingleServerSingleHostDSL},
		{"single-server-single-host-with-variables", testdata.SingleServerSingleHostWithVariablesDSL},
		{"single-server-multiple-hosts", testdata.SingleServerMultipleHostsDSL},
		{"single-server-multiple-hosts-with-variables", testdata.SingleServerMultipleHostsWithVariablesDSL},
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
			golden := filepath.Join("testdata", "client-"+c.Name+".golden")
			compareOrUpdateGolden(t, code, golden)
		})
	}
}
