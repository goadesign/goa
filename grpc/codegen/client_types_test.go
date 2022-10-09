package codegen

import (
	"bytes"
	"testing"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/grpc/codegen/testdata"
)

func TestClientTypeFiles(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"client-payload-with-nested-types", testdata.PayloadWithNestedTypesDSL, testdata.PayloadWithNestedTypesClientTypeCode},
		{"client-payload-with-duplicate-use", testdata.PayloadWithMultipleUseTypesDSL, testdata.PayloadWithMultipleUseTypesClientTypeCode},
		{"client-payload-with-alias-type", testdata.PayloadWithAliasTypeDSL, testdata.PayloadWithAliasTypeClientTypeCode},
		{"client-result-collection", testdata.ResultWithCollectionDSL, testdata.ResultWithCollectionClientTypeCode},
		{"client-alias-validation", testdata.ResultWithAliasValidation, testdata.ResultWithAliasValidationClientTypeCode},
		{"client-with-errors", testdata.UnaryRPCWithErrorsDSL, testdata.WithErrorsClientTypeCode},
		{"client-bidirectional-streaming-same-type", testdata.BidirectionalStreamingRPCSameTypeDSL, testdata.BidirectionalStreamingRPCSameTypeClientTypeCode},
		{"client-struct-meta-type", testdata.StructMetaTypeDSL, testdata.StructMetaTypeTypeCode},
		{"client-default-fields", testdata.DefaultFieldsDSL, testdata.DefaultFieldsTypeCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunGRPCDSL(t, c.DSL)
			fs := ClientTypeFiles("", expr.Root)
			if len(fs) != 1 {
				t.Fatalf("got %d files, expected one", len(fs))
			}
			var buf bytes.Buffer
			for _, s := range fs[0].SectionTemplates[1:] {
				if err := s.Write(&buf); err != nil {
					t.Fatal(err)
				}
			}
			code := codegen.FormatTestCode(t, "package foo\n"+buf.String())
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}
