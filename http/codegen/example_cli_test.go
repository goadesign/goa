package codegen

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
