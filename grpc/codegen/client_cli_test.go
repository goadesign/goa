package codegen

import (
	"bytes"
	"testing"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/grpc/codegen/testdata"
)

func TestClientCLIFiles(t *testing.T) {

	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"payload-with-validations", testdata.PayloadWithValidationsDSL, testdata.PayloadWithValidationsBuildCode},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunGRPCDSL(t, c.DSL)
			fs := ClientCLIFiles("", expr.Root)
			if len(fs) == 0 {
				t.Fatalf("got 0 files, expected 1")
			}
			if len(fs[1].SectionTemplates) == 0 {
				t.Fatalf("got 0 sections, expected at least 1")
			}
			var buf bytes.Buffer
			for _, s := range fs[1].SectionTemplates {
				if err := s.Write(&buf); err != nil {
					t.Fatal(err)
				}
			}
			code := codegen.FormatTestCode(t, buf.String())
			if code != c.Code {
				t.Errorf("invalid code: got\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}
