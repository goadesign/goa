package codegen

import (
	"testing"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestClientCLIFiles(t *testing.T) {

	cases := []struct {
		Name         string
		DSL          func()
		Code         string
		FileIndex    int
		SectionIndex int
	}{
		{"no-payload-parse", testdata.MultiNoPayloadDSL, testdata.MultiNoPayloadParseCode, 0, 3},
		{"simple-parse", testdata.MultiSimpleDSL, testdata.MultiSimpleParseCode, 0, 3},
		{"multi-parse", testdata.MultiDSL, testdata.MultiParseCode, 0, 3},
		{"multi-required-payload", testdata.MultiRequiredPayloadDSL, testdata.MultiRequiredPayloadParseCode, 0, 3},
		{"streaming-parse", testdata.StreamingMultipleServicesDSL, testdata.StreamingParseCode, 0, 3},
		{"simple-build", testdata.MultiSimpleDSL, testdata.MultiSimpleBuildCode, 1, 1},
		{"multi-build", testdata.MultiDSL, testdata.MultiBuildCode, 1, 1},
		{"bool-build", testdata.PayloadQueryBoolDSL, testdata.QueryBoolBuildCode, 1, 1},
		{"uint32-build", testdata.PayloadQueryUInt32DSL, testdata.QueryUInt32BuildCode, 1, 1},
		{"uint64-build", testdata.PayloadQueryUIntDSL, testdata.QueryUIntBuildCode, 1, 1},
		{"string-build", testdata.PayloadQueryStringDSL, testdata.QueryStringBuildCode, 1, 1},
		{"string-required-build", testdata.PayloadQueryStringValidateDSL, testdata.QueryStringRequiredBuildCode, 1, 1},
		{"string-default-build", testdata.PayloadQueryStringDefaultDSL, testdata.QueryStringDefaultBuildCode, 1, 1},
		{"body-query-path-object-build", testdata.PayloadBodyQueryPathObjectDSL, testdata.BodyQueryPathObjectBuildCode, 1, 1},
		{"param-validation-build", testdata.ParamValidateDSL, testdata.ParamValidateBuildCode, 1, 1},
		{"payload-primitive-type", testdata.PayloadBodyPrimitiveBoolValidateDSL, testdata.PayloadPrimitiveTypeParseCode, 0, 3},
		{"payload-array-primitive-type", testdata.PayloadBodyPrimitiveArrayStringValidateDSL, testdata.PayloadArrayPrimitiveTypeParseCode, 0, 3},
		{"payload-array-user-type", testdata.PayloadBodyInlineArrayUserDSL, testdata.PayloadArrayUserTypeBuildCode, 1, 1},
		{"payload-map-user-type", testdata.PayloadBodyInlineMapUserDSL, testdata.PayloadMapUserTypeBuildCode, 1, 1},
		{"payload-object-type", testdata.PayloadBodyInlineObjectDSL, testdata.PayloadObjectBuildCode, 1, 1},
		{"payload-object-default-type", testdata.PayloadBodyInlineObjectDefaultDSL, testdata.PayloadObjectDefaultBuildCode, 1, 1},
		{"map-query", testdata.PayloadMapQueryPrimitiveArrayDSL, testdata.MapQueryParseCode, 0, 3},
		{"map-query-object", testdata.PayloadMapQueryObjectDSL, testdata.MapQueryObjectBuildCode, 1, 1},
		{"empty-body-build", testdata.PayloadBodyPrimitiveFieldEmptyDSL, testdata.EmptyBodyBuildCode, 1, 1},
		{"with-params-and-headers-dsl", testdata.WithParamsAndHeadersBlockDSL, testdata.WithParamsAndHeadersBlockBuildCode, 1, 1},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			fs := ClientCLIFiles("", expr.Root)
			sections := fs[c.FileIndex].SectionTemplates
			code := codegen.SectionCode(t, sections[c.SectionIndex])
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}
