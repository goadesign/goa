package codegen

import (
	"bytes"
	"testing"

	"goa.design/goa/codegen"
	"goa.design/goa/codegen/server"
	ctestdata "goa.design/goa/codegen/server/testdata"
	"goa.design/goa/codegen/service"
	"goa.design/goa/expr"
	"goa.design/goa/http/codegen/testdata"
)

func TestExampleServerFiles(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"no-server", ctestdata.NoServerDSL, testdata.NoServerServerHandleCode},
		{"server-hosting-service-with-file-server", ctestdata.ServerHostingServiceWithFileServerDSL, testdata.ServerHostingServiceWithFileServerHandlerCode},
		{"server-hosting-service-subset", ctestdata.ServerHostingServiceSubsetDSL, testdata.ServerHostingServiceSubsetServerHandleCode},
		{"server-hosting-multiple-services", ctestdata.ServerHostingMultipleServicesDSL, testdata.ServerHostingMultipleServicesServerHandleCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			// reset global variable
			HTTPServices = make(ServicesData)
			service.Services = make(service.ServicesData)
			server.Servers = make(server.ServersData)
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
