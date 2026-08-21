package codegen

import (
	"bytes"
	"goa.design/goa/v3/codegen/testutil"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/grpc/codegen/testdata"
)

func TestClientTypeFiles(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"client-payload-with-nested-types", testdata.PayloadWithNestedTypesDSL},
		{"client-payload-with-duplicate-use", testdata.PayloadWithMultipleUseTypesDSL},
		{"client-payload-with-alias-type", testdata.PayloadWithAliasTypeDSL},
		{"client-result-collection", testdata.ResultWithCollectionDSL},
		{"client-alias-validation", testdata.ResultWithAliasValidation},
		{"client-with-errors", testdata.UnaryRPCWithErrorsDSL},
		{"client-bidirectional-streaming-same-type", testdata.BidirectionalStreamingRPCSameTypeDSL},
		{"client-struct-meta-type", testdata.StructMetaTypeDSL},
		{"client-struct-field-name-meta-type", testdata.StructFieldNameMetaTypeDSL},
		{"client-default-fields", testdata.DefaultFieldsDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := RunGRPCDSL(t, c.DSL)
			services := CreateGRPCServices(root)
			fs := ClientTypeFiles(services)
			require.Len(t, fs, 1)
			var buf bytes.Buffer
			for _, s := range fs[0].SectionTemplates[1:] {
				require.NoError(t, s.Write(&buf))
			}
			code := codegen.FormatTestCode(t, "package foo\n"+buf.String())
			testutil.AssertGo(t, "testdata/golden/client_types_"+c.Name+".go.golden", code)
		})
	}
}
