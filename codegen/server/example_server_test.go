package server

import (
	"bytes"
	"testing"

	"goa.design/goa/codegen"
	"goa.design/goa/codegen/server/testdata"
	"goa.design/goa/codegen/service"
	"goa.design/goa/expr"
)

func TestExampleServerFiles(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"no-server", testdata.NoServerDSL, testdata.NoServerServerMainCode},
		{"single-server-single-host", testdata.SingleServerSingleHostDSL, testdata.SingleServerSingleHostServerMainCode},
		{"single-server-single-host-with-variables", testdata.SingleServerSingleHostWithVariablesDSL, testdata.SingleServerSingleHostWithVariablesServerMainCode},
		{"server-hosting-service-with-file-server", testdata.ServerHostingServiceWithFileServerDSL, testdata.ServerHostingServiceWithFileServerServerMainCode},
		{"server-hosting-service-subset", testdata.ServerHostingServiceSubsetDSL, testdata.ServerHostingServiceSubsetServerMainCode},
		{"server-hosting-multiple-services", testdata.ServerHostingMultipleServicesDSL, testdata.ServerHostingMultipleServicesServerMainCode},
		{"single-server-multiple-hosts", testdata.SingleServerMultipleHostsDSL, testdata.SingleServerMultipleHostsServerMainCode},
		{"single-server-multiple-hosts-with-variables", testdata.SingleServerMultipleHostsWithVariablesDSL, testdata.SingleServerMultipleHostsWithVariablesServerMainCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			// reset global variable
			service.Services = make(service.ServicesData)
			Servers = make(ServersData)
			codegen.RunDSL(t, c.DSL)
			fs := ExampleServerFiles("", expr.Root)
			if len(fs) == 0 {
				t.Fatalf("got 0 files, expected 1")
			}
			if len(fs[0].SectionTemplates) == 0 {
				t.Fatalf("got 0 sections, expected at least 1")
			}
			var buf bytes.Buffer
			for _, s := range fs[0].SectionTemplates[1:] {
				if err := s.Write(&buf); err != nil {
					t.Fatal(err)
				}
			}
			code := codegen.FormatTestCode(t, "package foo\n"+buf.String())
			if code != c.Code {
				t.Errorf("invalid code for %s: got\n%s\ngot vs. expected:\n%s", fs[0].Path, code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}
