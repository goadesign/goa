package codegen

import (
	"bytes"
	"testing"

	"goa.design/goa/codegen"
	"goa.design/goa/codegen/server"
	ctestdata "goa.design/goa/codegen/server/testdata"
	"goa.design/goa/expr"
	"goa.design/goa/grpc/codegen/testdata"
)

func TestExampleCLIFiles(t *testing.T) {
	cases := []struct {
		Name    string
		DSL     func()
		PkgPath string
		Code    string
	}{
		{"no-server", ctestdata.NoServerDSL, "", testdata.ExampleCLIImport + "\n" + testdata.ExampleCLICode},
		{"server-hosting-service-subset", ctestdata.ServerHostingServiceSubsetDSL, "", testdata.ExampleSingleHostCLIImport + "\n" + testdata.ExampleCLICode},
		{"server-hosting-multiple-services", ctestdata.ServerHostingMultipleServicesDSL, "", testdata.ExampleSingleHostCLIImport + "\n" + testdata.ExampleCLICode},
		{"no-server-pkgpath", ctestdata.NoServerDSL, "my/pkg/path", testdata.ExamplePkgPathCLIImport + "\n" + testdata.ExampleCLICode},
		{"server-hosting-service-subset-pkgpath", ctestdata.ServerHostingServiceSubsetDSL, "my/pkg/path", testdata.ExampleSingleHostPkgPathCLIImport + "\n" + testdata.ExampleCLICode},
		{"server-hosting-multiple-services-pkgpath", ctestdata.ServerHostingMultipleServicesDSL, "my/pkg/path", testdata.ExampleSingleHostPkgPathCLIImport + "\n" + testdata.ExampleCLICode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			// reset global variable
			server.Servers = make(server.ServersData)
			codegen.RunDSL(t, c.DSL)
			fs := ExampleCLIFiles(c.PkgPath, expr.Root)
			if len(fs) == 0 {
				t.Fatalf("got 0 files, expected 1")
			}
			if len(fs[0].SectionTemplates) == 0 {
				t.Fatalf("got 0 sections, expected at least 1")
			}
			var buf bytes.Buffer
			for _, s := range fs[0].SectionTemplates {
				if err := s.Write(&buf); err != nil {
					t.Fatal(err)
				}
			}
			code := codegen.FormatTestCode(t, buf.String())
			if code != c.Code {
				t.Errorf("invalid code for %s: got\n%s\ngot vs. expected:\n%s", fs[0].Path, code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}
