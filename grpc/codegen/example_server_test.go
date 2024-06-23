package codegen

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	ctestdata "goa.design/goa/v3/codegen/example/testdata"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

var updateGolden = false

func init() {
	flag.BoolVar(&updateGolden, "w", false, "update golden files")
}

func compareOrUpdateGolden(t *testing.T, code, golden string) {
	t.Helper()
	if updateGolden {
		require.NoError(t, os.MkdirAll(filepath.Dir(golden), 0755))
		require.NoError(t, os.WriteFile(golden, []byte(code), 0644))
		return
	}
	data, err := os.ReadFile(golden)
	require.NoError(t, err)
	assert.Equal(t, string(data), code)
}

func TestExampleServerFiles(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"no-server", ctestdata.NoServerDSL},
		{"server-hosting-service-subset", ctestdata.ServerHostingServiceSubsetDSL},
		{"server-hosting-multiple-services", ctestdata.ServerHostingMultipleServicesDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			// reset global variable
			GRPCServices = make(ServicesData)
			service.Services = make(service.ServicesData)
			example.Servers = make(example.ServersData)
			codegen.RunDSL(t, c.DSL)
			fs := ExampleServerFiles("", expr.Root)
			require.Greater(t, len(fs), 0)
			require.Greater(t, len(fs[0].SectionTemplates), 0)
			var buf bytes.Buffer
			for _, s := range fs[0].SectionTemplates[1:] {
				require.NoError(t, s.Write(&buf))
			}
			code := codegen.FormatTestCode(t, "package foo\n"+buf.String())
			golden := filepath.Join("testdata", "server-"+c.Name+".golden")
			compareOrUpdateGolden(t, code, golden)
		})
	}
}
