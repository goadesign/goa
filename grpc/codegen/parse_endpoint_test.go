package codegen

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/grpc/codegen/testdata"
)

func TestParseEndpointWithInterceptors(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{
			Name: "endpoint-with-interceptors",
			DSL:  testdata.InterceptorsDSL,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunGRPCDSL(t, c.DSL)
			fs := ClientCLIFiles("", expr.Root)
			require.Greater(t, len(fs), 1, "expected at least 2 files")
			require.NotEmpty(t, fs[0].SectionTemplates)
			var buf bytes.Buffer
			for _, s := range fs[0].SectionTemplates {
				require.NoError(t, s.Write(&buf))
			}
			code := codegen.FormatTestCode(t, buf.String())
			golden := filepath.Join("testdata", "endpoint-"+c.Name+".golden")
			compareOrUpdateGolden(t, code, golden)
		})
	}
}
