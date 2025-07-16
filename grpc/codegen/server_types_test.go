package codegen

import (
	"goa.design/goa/v3/codegen/testutil"
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/grpc/codegen/testdata"
)

func TestServerTypeFiles(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"server-payload-with-nested-types", testdata.PayloadWithNestedTypesDSL},
		{"server-payload-with-duplicate-use", testdata.PayloadWithMultipleUseTypesDSL},
		{"server-payload-with-alias-type", testdata.PayloadWithAliasTypeDSL},
		{"server-payload-with-mixed-attributes", testdata.PayloadWithMixedAttributesDSL},
		{"server-payload-with-custom-type-package", testdata.PayloadWithCustomTypePackageDSL},
		{"server-result-collection", testdata.ResultWithCollectionDSL},
		{"server-with-errors", testdata.UnaryRPCWithErrorsDSL},
		{"server-elem-validation", testdata.ElemValidationDSL},
		{"server-alias-validation", testdata.AliasValidationDSL},
		{"server-struct-meta-type", testdata.StructMetaTypeDSL},
		{"server-struct-field-name-meta-type", testdata.StructFieldNameMetaTypeDSL},
		{"server-default-fields", testdata.DefaultFieldsDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := RunGRPCDSL(t, c.DSL)
			services := CreateGRPCServices(root)
			fs := ServerTypeFiles("", services)
			require.Len(t, fs, 1)
			var buf bytes.Buffer
			for _, s := range fs[0].SectionTemplates[1:] {
				require.NoError(t, s.Write(&buf))
			}
			code := codegen.FormatTestCode(t, "package foo\n"+buf.String())
			testutil.AssertGo(t, "testdata/golden/server_types_"+c.Name+".go.golden", code)
		})
	}
}
