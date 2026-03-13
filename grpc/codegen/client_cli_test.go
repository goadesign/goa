package codegen

import (
	"bytes"
	"goa.design/goa/v3/codegen/testutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/grpc/codegen/testdata"
)

func TestClientCLIFiles(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"payload-with-validations", testdata.PayloadWithValidationsDSL},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := RunGRPCDSL(t, c.DSL)
			services := CreateGRPCServices(root)
			fs := ClientCLIFiles("", services)
			require.Greater(t, len(fs), 1, "expected at least 2 files")
			require.NotEmpty(t, fs[1].SectionTemplates)
			var buf bytes.Buffer
			for _, s := range fs[1].SectionTemplates {
				require.NoError(t, s.Write(&buf))
			}
			code := codegen.FormatTestCode(t, buf.String())
			testutil.AssertGo(t, "testdata/golden/client_cli_"+c.Name+".go.golden", code)
		})
	}
}

func TestConstructorUnionUnaryRPCClientCLIFiles(t *testing.T) {
	root := RunGRPCDSL(t, testdata.ConstructorUnionUnaryRPCDSL)
	services := CreateGRPCServices(root)
	fs := ClientCLIFiles("", services)
	require.Greater(t, len(fs), 1, "expected parser and payload builder files for constructor union gRPC service")

	var builder bytes.Buffer
	for _, s := range fs[1].SectionTemplates {
		require.NoError(t, s.Write(&builder))
	}
	builderCode := codegen.FormatTestCode(t, builder.String())
	if !strings.Contains(builderCode, "json.Unmarshal") {
		t.Errorf("expected gRPC CLI payload builder to decode constructor union payload from JSON, got %q", builderCode)
	}
	if !strings.Contains(builderCode, "BuildShowPayload") {
		t.Errorf("expected gRPC CLI payload builder to expose constructor union payload builder, got %q", builderCode)
	}
}
