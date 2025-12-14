package codegen

import (
	"goa.design/goa/v3/codegen/testutil"
	"testing"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestClientCLIFiles(t *testing.T) {
	cases := []struct {
		Name         string
		DSL          func()
		FileIndex    int
		SectionIndex int
	}{
		{"no-payload-parse", testdata.MultiNoPayloadDSL, 0, 3},
		{"simple-parse", testdata.MultiSimpleDSL, 0, 3},
		{"multi-parse", testdata.MultiDSL, 0, 3},
		{"multi-required-payload", testdata.MultiRequiredPayloadDSL, 0, 3},
		{"skip-request-body-encode-decode", testdata.SkipRequestBodyEncodeDecodeDSL, 0, 3},
		{"streaming-parse", testdata.StreamingMultipleServicesDSL, 0, 3},
		{"simple-build", testdata.MultiSimpleDSL, 1, 1},
		{"multi-build", testdata.MultiDSL, 1, 1},
		{"bool-build", testdata.PayloadQueryBoolDSL, 1, 1},
		{"uint32-build", testdata.PayloadQueryUInt32DSL, 1, 1},
		{"uint64-build", testdata.PayloadQueryUIntDSL, 1, 1},
		{"string-build", testdata.PayloadQueryStringDSL, 1, 1},
		{"string-required-build", testdata.PayloadQueryStringValidateDSL, 1, 1},
		{"string-default-build", testdata.PayloadQueryStringDefaultDSL, 1, 1},
		{"body-query-path-object-build", testdata.PayloadBodyQueryPathObjectDSL, 1, 1},
		{"param-validation-build", testdata.ParamValidateDSL, 1, 1},
		{"payload-primitive-type", testdata.PayloadBodyPrimitiveBoolValidateDSL, 0, 3},
		{"payload-array-primitive-type", testdata.PayloadBodyPrimitiveArrayStringValidateDSL, 0, 3},
		{"payload-array-user-type", testdata.PayloadBodyInlineArrayUserDSL, 1, 1},
		{"payload-map-user-type", testdata.PayloadBodyInlineMapUserDSL, 1, 1},
		{"payload-object-type", testdata.PayloadBodyInlineObjectDSL, 1, 1},
		{"payload-object-default-type", testdata.PayloadBodyInlineObjectDefaultDSL, 1, 1},
		{"map-query", testdata.PayloadMapQueryPrimitiveArrayDSL, 0, 3},
		{"map-query-object", testdata.PayloadMapQueryObjectDSL, 1, 1},
		{"empty-body-build", testdata.PayloadBodyPrimitiveFieldEmptyDSL, 1, 1},
		{"with-params-and-headers-dsl", testdata.WithParamsAndHeadersBlockDSL, 1, 1},
		{"body-custom-name", testdata.PayloadBodyCustomNameDSL, 1, 1},
		{"path-custom-name", testdata.PayloadPathCustomNameDSL, 1, 1},
		{"query-custom-name", testdata.PayloadQueryCustomNameDSL, 1, 1},
		{"header-custom-name", testdata.PayloadHeaderCustomNameDSL, 1, 1},
		{"cookie-custom-name", testdata.PayloadCookieCustomNameDSL, 1, 1},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := RunHTTPDSL(t, c.DSL)
			services := CreateHTTPServices(root)
			fs := ClientCLIFiles("", services)
			sections := fs[c.FileIndex].SectionTemplates
			code := codegen.SectionCode(t, sections[c.SectionIndex])
			testutil.AssertGo(t, "testdata/golden/client_cli_"+c.Name+".go.golden", code)
		})
	}
}
