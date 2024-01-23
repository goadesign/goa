package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestClientInit(t *testing.T) {
	cases := []struct {
		Name       string
		DSL        func()
		Code       string
		FileCount  int
		SectionNum int
	}{
		{"multiple endpoints", testdata.ServerMultiEndpointsDSL, testdata.MultipleEndpointsClientInitCode, 2, 2},
		{"streaming", testdata.StreamingResultDSL, testdata.StreamingClientInitCode, 3, 2},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			fs := ClientFiles("", expr.Root)
			require.Len(t, fs, c.FileCount)
			sections := fs[0].SectionTemplates
			require.Greater(t, len(sections), c.SectionNum)
			code := codegen.SectionCode(t, sections[c.SectionNum])
			assert.Equal(t, c.Code, code)
		})
	}
}
