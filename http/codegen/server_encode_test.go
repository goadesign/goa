package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestEncode(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"header-bool", testdata.ResultHeaderBoolDSL},
		{"header-int", testdata.ResultHeaderIntDSL},
		{"header-int32", testdata.ResultHeaderInt32DSL},
		{"header-int64", testdata.ResultHeaderInt64DSL},
		{"header-uint", testdata.ResultHeaderUIntDSL},
		{"header-uint32", testdata.ResultHeaderUInt32DSL},
		{"header-uint64", testdata.ResultHeaderUInt64DSL},
		{"header-float32", testdata.ResultHeaderFloat32DSL},
		{"header-float64", testdata.ResultHeaderFloat64DSL},
		{"header-string", testdata.ResultHeaderStringDSL},
		{"header-bytes", testdata.ResultHeaderBytesDSL},
		{"header-any", testdata.ResultHeaderAnyDSL},
		{"header-array-bool", testdata.ResultHeaderArrayBoolDSL},
		{"header-array-int", testdata.ResultHeaderArrayIntDSL},
		{"header-array-int32", testdata.ResultHeaderArrayInt32DSL},
		{"header-array-int64", testdata.ResultHeaderArrayInt64DSL},
		{"header-array-uint", testdata.ResultHeaderArrayUIntDSL},
		{"header-array-uint32", testdata.ResultHeaderArrayUInt32DSL},
		{"header-array-uint64", testdata.ResultHeaderArrayUInt64DSL},
		{"header-array-float32", testdata.ResultHeaderArrayFloat32DSL},
		{"header-array-float64", testdata.ResultHeaderArrayFloat64DSL},
		{"header-array-string", testdata.ResultHeaderArrayStringDSL},
		{"header-array-bytes", testdata.ResultHeaderArrayBytesDSL},
		{"header-array-any", testdata.ResultHeaderArrayAnyDSL},

		{"header-bool-default", testdata.ResultHeaderBoolDefaultDSL},
		{"header-bool-required-default", testdata.ResultHeaderBoolRequiredDefaultDSL},
		{"header-string-default", testdata.ResultHeaderStringDefaultDSL},
		{"header-string-required-default", testdata.ResultHeaderStringRequiredDefaultDSL},
		{"header-array-bool-default", testdata.ResultHeaderArrayBoolDefaultDSL},
		{"header-array-bool-required-default", testdata.ResultHeaderArrayBoolRequiredDefaultDSL},
		{"header-array-string-default", testdata.ResultHeaderArrayStringDefaultDSL},
		{"header-array-string-required-default", testdata.ResultHeaderArrayStringRequiredDefaultDSL},

		{"body-string", testdata.ResultBodyStringDSL},
		{"body-object", testdata.ResultBodyObjectDSL},
		{"body-user", testdata.ResultBodyUserDSL},
		{"body-union", testdata.ResultBodyUnionDSL},
		{"body-result-multiple-views", testdata.ResultBodyMultipleViewsDSL},
		{"body-result-collection-multiple-views", testdata.ResultBodyCollectionDSL},
		{"body-result-collection-explicit-view", testdata.ResultBodyCollectionExplicitViewDSL},
		{"empty-body-result-multiple-views", testdata.EmptyBodyResultMultipleViewsDSL},
		{"body-array-string", testdata.ResultBodyArrayStringDSL},
		{"body-array-user", testdata.ResultBodyArrayUserDSL},
		{"body-primitive-string", testdata.ResultBodyPrimitiveStringDSL},
		{"body-primitive-bool", testdata.ResultBodyPrimitiveBoolDSL},
		{"body-primitive-any", testdata.ResultBodyPrimitiveAnyDSL},
		{"body-primitive-array-string", testdata.ResultBodyPrimitiveArrayStringDSL},
		{"body-primitive-array-bool", testdata.ResultBodyPrimitiveArrayBoolDSL},
		{"body-primitive-array-user", testdata.ResultBodyPrimitiveArrayUserDSL},
		{"body-inline-object", testdata.ResultBodyInlineObjectDSL},

		{"body-header-object", testdata.ResultBodyHeaderObjectDSL},
		{"body-header-user", testdata.ResultBodyHeaderUserDSL},

		{"explicit-body-primitive-result-multiple-views", testdata.ExplicitBodyPrimitiveResultMultipleViewsDSL},
		{"explicit-body-user-result-multiple-views", testdata.ExplicitBodyUserResultMultipleViewsDSL},
		{"explicit-body-result-collection", testdata.ExplicitBodyResultCollectionDSL},
		{"explicit-content-type-result", testdata.ExplicitContentTypeResultDSL},
		{"explicit-content-type-response", testdata.ExplicitContentTypeResponseDSL},

		{"tag-string", testdata.ResultTagStringDSL},
		{"tag-string-required", testdata.ResultTagStringRequiredDSL},
		{"tag-result-multiple-views", testdata.ResultMultipleViewsTagDSL},

		{"empty-server-response", testdata.EmptyServerResponseDSL},
		{"empty-server-response-with-tags", testdata.EmptyServerResponseWithTagsDSL},

		{"skip-response-body-encode-decode", testdata.ResponseEncoderSkipResponseBodyEncodeDecodeDSL},

		{"result-with-custom-pkg-type", testdata.ResultWithCustomPkgTypeDSL},
		{"result-with-embedded-custom-pkg-type", testdata.EmbeddedCustomPkgTypeDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := expr.RunDSL(t, c.DSL)
			services := CreateHTTPServices(root)
			fs := ServerFiles("", services)
			require.Len(t, fs, 2)
			sections := fs[1].SectionTemplates
			require.Greater(t, len(sections), 1)
			code := codegen.SectionCode(t, sections[1])
			testutil.AssertGo(t, "testdata/golden/server_encode_"+c.Name+".go.golden", code)
		})
	}
}

func TestEncodeMarshallingAndUnmarshalling(t *testing.T) {
	cases := []struct {
		Name           string
		DSL            func()
		SectionCount   int
		SectionsOffset int
	}{
		{"embedded-custom-pkg-type", testdata.EmbeddedCustomPkgTypeDSL, 2, 3},
		{"array-alias-extended", testdata.ArrayAliasExtendedDSL, 2, 3},
		{"extension-with-alias", testdata.ExtensionWithAliasDSL, 5, 4},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := expr.RunDSL(t, c.DSL)
			services := CreateHTTPServices(root)
			fs := ServerFiles("", services)
			require.Len(t, fs, 2)
			sections := fs[1].SectionTemplates
			totalSectionsExpected := c.SectionsOffset + c.SectionCount
			require.Len(t, sections, totalSectionsExpected)
			for i := 0; i < c.SectionCount; i++ {
				code := codegen.SectionCode(t, sections[c.SectionsOffset+i])
				testutil.AssertGo(t, "testdata/golden/server_encode_marshal_"+c.Name+"_section"+string(rune('0'+i))+".go.golden", code)
			}
		})
	}
}
