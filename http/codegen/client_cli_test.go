package codegen

import (
	"testing"

	"goa.design/goa.v2/codegen"
	. "goa.design/goa.v2/http/codegen/testing"
	httpdesign "goa.design/goa.v2/http/design"
)

func TestParseEndpoint(t *testing.T) {

	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"no-payload", MultiNoPayloadDSL, MultiNoPayloadCode},
		{"simple", MultiSimpleDSL, MultiSimpleCode},
		{"multi", MultiDSL, MultiCode},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			fs := ClientCLIFile(httpdesign.Root)
			sections := fs.Sections("")
			code := codegen.SectionCode(t, sections[3])
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}
