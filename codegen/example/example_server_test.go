package example

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example/testdata"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

// updateGolden is true when -w is passed to `go test`, e.g. `go test ./... -w`
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
		{"no-server", testdata.NoServerDSL},
		{"same-api-service-name", testdata.SameAPIServiceNameDSL},
		{"single-server-single-host", testdata.SingleServerSingleHostDSL},
		{"single-server-single-host-with-variables", testdata.SingleServerSingleHostWithVariablesDSL},
		{"server-hosting-service-with-file-server", testdata.ServerHostingServiceWithFileServerDSL},
		{"server-hosting-service-subset", testdata.ServerHostingServiceSubsetDSL},
		{"server-hosting-multiple-services", testdata.ServerHostingMultipleServicesDSL},
		{"single-server-multiple-hosts", testdata.SingleServerMultipleHostsDSL},
		{"single-server-multiple-hosts-with-variables", testdata.SingleServerMultipleHostsWithVariablesDSL},
		{"service-name-with-spaces", testdata.NamesWithSpacesDSL},
		{"service-for-only-http", testdata.ServiceForOnlyHTTPDSL},
		{"sercice-for-only-grpc", testdata.ServiceForOnlyGRPCDSL},
		{"service-for-http-and-part-of-grpc", testdata.ServiceForHTTPAndPartOfGRPCDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			service.Services = make(service.ServicesData)
			Servers = make(ServersData)
			codegen.RunDSL(t, c.DSL)
			fs := ServerFiles("", expr.Root)
			require.Len(t, fs, 1)
			require.Greater(t, len(fs[0].SectionTemplates), 0)
			var buf bytes.Buffer
			for _, s := range fs[0].SectionTemplates[1:] {
				require.NoError(t, s.Write(&buf))
			}
			code := codegen.FormatTestCode(t, "package foo\n"+buf.String())
			golden := filepath.Join("testdata", "server-"+c.Name+".golden")
			if updateGolden {
				require.NoError(t, os.MkdirAll(filepath.Dir(golden), 0755))
				require.NoError(t, os.WriteFile(golden, []byte(code), 0644))
				return
			}
			data, err := os.ReadFile(golden)
			require.NoError(t, err)
			assert.Equal(t, string(data), code)
		})
	}
}
