package codegen

import (
	"bytes"
	"testing"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	ctestdata "goa.design/goa/v3/codegen/example/testdata"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestExampleCLIFiles(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"no-server", ctestdata.NoServerDSL, testdata.ExampleCLICode},
		{"server-hosting-service-subset", ctestdata.ServerHostingServiceSubsetDSL, testdata.ExampleCLICode},
		{"server-hosting-multiple-services", ctestdata.ServerHostingMultipleServicesDSL, testdata.ExampleCLICode},
		{"streaming", testdata.StreamingResultDSL, testdata.StreamingExampleCLICode},
		{"streaming-multiple-services", testdata.StreamingMultipleServicesDSL, testdata.StreamingMultipleServicesExampleCLICode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			// reset global variable
			example.Servers = make(example.ServersData)
			codegen.RunDSL(t, c.DSL)
			fs := ExampleCLIFiles("", expr.Root)
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
