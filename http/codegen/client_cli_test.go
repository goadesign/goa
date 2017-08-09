package codegen

import (
	"testing"

	"goa.design/goa.v2/codegen"
	. "goa.design/goa.v2/http/codegen/testing"
	httpdesign "goa.design/goa.v2/http/design"
)

func TestClientCLIFiles(t *testing.T) {

	cases := []struct {
		Name         string
		DSL          func()
		Code         string
		FileIndex    int
		SectionIndex int
	}{
		{"no-payload-parse", MultiNoPayloadDSL, MultiNoPayloadParseCode, 0, 3},
		{"simple-parse", MultiSimpleDSL, MultiSimpleParseCode, 0, 3},
		{"multi-parse", MultiDSL, MultiParseCode, 0, 3},
		{"simple-build", MultiSimpleDSL, MultiSimpleBuildCode, 1, 1},
		{"multi-build", MultiDSL, MultiBuildCode, 1, 1},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			fs := ClientCLIFiles(httpdesign.Root)
			sections := fs[c.FileIndex].Sections("")
			code := codegen.SectionCode(t, sections[c.SectionIndex])
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}
