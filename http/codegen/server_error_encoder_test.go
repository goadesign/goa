package codegen

import (
	"testing"

	"goa.design/goa/codegen"
	"goa.design/goa/http/codegen/testdata"
	httpdesign "goa.design/goa/http/design"
)

func TestEncodeError(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"primitive-error-response", testdata.PrimitiveErrorResponseDSL, testdata.PrimitiveErrorResponseEncoderCode},
		{"default-error-response", testdata.DefaultErrorResponseDSL, testdata.DefaultErrorResponseEncoderCode},
		{"service-error-response", testdata.ServiceErrorResponseDSL, testdata.ServiceErrorResponseEncoderCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			fs := ServerFiles("", httpdesign.Root)
			if len(fs) != 2 {
				t.Fatalf("got %d files, expected two", len(fs))
			}
			sections := fs[1].SectionTemplates
			if len(sections) < 2 {
				t.Fatalf("got %d sections, expected at least 2", len(sections))
			}
			code := codegen.SectionCode(t, sections[2])
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}
