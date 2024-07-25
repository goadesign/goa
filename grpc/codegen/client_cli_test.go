package codegen

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
			require.Greater(t, len(fs), 1, "expected at least 2 files")
			require.NotEmpty(t, fs[1].SectionTemplates)
			var buf bytes.Buffer
			for _, s := range fs[1].SectionTemplates {
				require.NoError(t, s.Write(&buf))
			}
			code := codegen.FormatTestCode(t, buf.String())
			assert.Equal(t, c.Code, code)
		})
	}
}
