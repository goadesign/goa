package example

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example/testdata"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

func TestExampleServerFiles(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"no-server", testdata.NoServerDSL, testdata.NoServerServerMainCode},
		{"same-api-service-name", testdata.SameAPIServiceNameDSL, testdata.SameAPIServiceNameServerMainCode},
		{"single-server-single-host", testdata.SingleServerSingleHostDSL, testdata.SingleServerSingleHostServerMainCode},
		{"single-server-single-host-with-variables", testdata.SingleServerSingleHostWithVariablesDSL, testdata.SingleServerSingleHostWithVariablesServerMainCode},
		{"server-hosting-service-with-file-server", testdata.ServerHostingServiceWithFileServerDSL, testdata.ServerHostingServiceWithFileServerServerMainCode},
		{"server-hosting-service-subset", testdata.ServerHostingServiceSubsetDSL, testdata.ServerHostingServiceSubsetServerMainCode},
		{"server-hosting-multiple-services", testdata.ServerHostingMultipleServicesDSL, testdata.ServerHostingMultipleServicesServerMainCode},
		{"single-server-multiple-hosts", testdata.SingleServerMultipleHostsDSL, testdata.SingleServerMultipleHostsServerMainCode},
		{"single-server-multiple-hosts-with-variables", testdata.SingleServerMultipleHostsWithVariablesDSL, testdata.SingleServerMultipleHostsWithVariablesServerMainCode},
		{"service-name-with-spaces", testdata.NamesWithSpacesDSL, testdata.NamesWithSpacesServerMainCode},
		{"service-for-only-http", testdata.ServiceForOnlyHTTPDSL, testdata.ServiceForOnlyHTTPServerMainCode},
		{"sercice-for-only-grpc", testdata.ServiceForOnlyGRPCDSL, testdata.ServiceForOnlyGRPCServerMainCode},
		{"service-for-http-and-part-of-grpc", testdata.ServiceForHTTPAndPartOfGRPCDSL, testdata.ServiceForHTTPAndPartOfGRPCServerMainCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			// reset global variable
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
			assert.Equal(t, c.Code, code)
		})
	}
}
