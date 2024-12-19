package codegen

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	ctestdata "goa.design/goa/v3/codegen/example/testdata"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/grpc/codegen/testdata"
)

func TestExampleCLIFiles(t *testing.T) {
	cases := []struct {
		Name    string
		DSL     func()
		PkgPath string
	}{
		{"no-server", ctestdata.NoServerDSL, ""},
		{"server-hosting-service-subset", ctestdata.ServerHostingServiceSubsetDSL, ""},
		{"server-hosting-multiple-services", ctestdata.ServerHostingMultipleServicesDSL, ""},
		{"no-server-pkgpath", ctestdata.NoServerDSL, "my/pkg/path"},
		{"server-hosting-service-subset-pkgpath", ctestdata.ServerHostingServiceSubsetDSL, "my/pkg/path"},
		{"server-hosting-multiple-services-pkgpath", ctestdata.ServerHostingMultipleServicesDSL, "my/pkg/path"},
		{"interceptors", testdata.InterceptorsDSL, ""},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			// reset global variable
			example.Servers = make(example.ServersData)
			codegen.RunDSL(t, c.DSL)
			fs := ExampleCLIFiles(c.PkgPath, expr.Root)
			require.Greater(t, len(fs), 0)
			require.Greater(t, len(fs[0].SectionTemplates), 0)
			var buf bytes.Buffer
			for _, s := range fs[0].SectionTemplates {
				require.NoError(t, s.Write(&buf))
			}
			code := codegen.FormatTestCode(t, buf.String())
			golden := filepath.Join("testdata", "client-"+c.Name+".golden")
			compareOrUpdateGolden(t, code, golden)
		})
	}
}
