package codegen

import (
	"testing"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestPayloadConstructor(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"query-bool", testdata.PayloadQueryBoolDSL, testdata.PayloadQueryBoolConstructorCode},
		{"query-bool-validate", testdata.PayloadQueryBoolValidateDSL, testdata.PayloadQueryBoolValidateConstructorCode},
		{"query-int", testdata.PayloadQueryIntDSL, testdata.PayloadQueryIntConstructorCode},
		{"query-int-validate", testdata.PayloadQueryIntValidateDSL, testdata.PayloadQueryIntValidateConstructorCode},
		{"query-int32", testdata.PayloadQueryInt32DSL, testdata.PayloadQueryInt32ConstructorCode},
		{"query-int32-validate", testdata.PayloadQueryInt32ValidateDSL, testdata.PayloadQueryInt32ValidateConstructorCode},
		{"query-int64", testdata.PayloadQueryInt64DSL, testdata.PayloadQueryInt64ConstructorCode},
		{"query-int64-validate", testdata.PayloadQueryInt64ValidateDSL, testdata.PayloadQueryInt64ValidateConstructorCode},
		{"query-uint", testdata.PayloadQueryUIntDSL, testdata.PayloadQueryUIntConstructorCode},
		{"query-uint-validate", testdata.PayloadQueryUIntValidateDSL, testdata.PayloadQueryUIntValidateConstructorCode},
		{"query-uint32", testdata.PayloadQueryUInt32DSL, testdata.PayloadQueryUInt32ConstructorCode},
		{"query-uint32-validate", testdata.PayloadQueryUInt32ValidateDSL, testdata.PayloadQueryUInt32ValidateConstructorCode},
		{"query-uint64", testdata.PayloadQueryUInt64DSL, testdata.PayloadQueryUInt64ConstructorCode},
		{"query-uint64-validate", testdata.PayloadQueryUInt64ValidateDSL, testdata.PayloadQueryUInt64ValidateConstructorCode},
		{"query-float32", testdata.PayloadQueryFloat32DSL, testdata.PayloadQueryFloat32ConstructorCode},
		{"query-float32-validate", testdata.PayloadQueryFloat32ValidateDSL, testdata.PayloadQueryFloat32ValidateConstructorCode},
		{"query-float64", testdata.PayloadQueryFloat64DSL, testdata.PayloadQueryFloat64ConstructorCode},
		{"query-float64-validate", testdata.PayloadQueryFloat64ValidateDSL, testdata.PayloadQueryFloat64ValidateConstructorCode},
		{"query-string", testdata.PayloadQueryStringDSL, testdata.PayloadQueryStringConstructorCode},
		{"query-string-validate", testdata.PayloadQueryStringValidateDSL, testdata.PayloadQueryStringValidateConstructorCode},
		{"query-bytes", testdata.PayloadQueryBytesDSL, testdata.PayloadQueryBytesConstructorCode},
		{"query-bytes-validate", testdata.PayloadQueryBytesValidateDSL, testdata.PayloadQueryBytesValidateConstructorCode},
		{"query-any", testdata.PayloadQueryAnyDSL, testdata.PayloadQueryAnyConstructorCode},
		{"query-any-validate", testdata.PayloadQueryAnyValidateDSL, testdata.PayloadQueryAnyValidateConstructorCode},
		{"query-array-bool", testdata.PayloadQueryArrayBoolDSL, testdata.PayloadQueryArrayBoolConstructorCode},
		{"query-array-bool-validate", testdata.PayloadQueryArrayBoolValidateDSL, testdata.PayloadQueryArrayBoolValidateConstructorCode},
		{"query-array-int", testdata.PayloadQueryArrayIntDSL, testdata.PayloadQueryArrayIntConstructorCode},
		{"query-array-int-validate", testdata.PayloadQueryArrayIntValidateDSL, testdata.PayloadQueryArrayIntValidateConstructorCode},
		{"query-array-int32", testdata.PayloadQueryArrayInt32DSL, testdata.PayloadQueryArrayInt32ConstructorCode},
		{"query-array-int32-validate", testdata.PayloadQueryArrayInt32ValidateDSL, testdata.PayloadQueryArrayInt32ValidateConstructorCode},
		{"query-array-int64", testdata.PayloadQueryArrayInt64DSL, testdata.PayloadQueryArrayInt64ConstructorCode},
		{"query-array-int64-validate", testdata.PayloadQueryArrayInt64ValidateDSL, testdata.PayloadQueryArrayInt64ValidateConstructorCode},
		{"query-array-uint", testdata.PayloadQueryArrayUIntDSL, testdata.PayloadQueryArrayUIntConstructorCode},
		{"query-array-uint-validate", testdata.PayloadQueryArrayUIntValidateDSL, testdata.PayloadQueryArrayUIntValidateConstructorCode},
		{"query-array-uint32", testdata.PayloadQueryArrayUInt32DSL, testdata.PayloadQueryArrayUInt32ConstructorCode},
		{"query-array-uint32-validate", testdata.PayloadQueryArrayUInt32ValidateDSL, testdata.PayloadQueryArrayUInt32ValidateConstructorCode},
		{"query-array-uint64", testdata.PayloadQueryArrayUInt64DSL, testdata.PayloadQueryArrayUInt64ConstructorCode},
		{"query-array-uint64-validate", testdata.PayloadQueryArrayUInt64ValidateDSL, testdata.PayloadQueryArrayUInt64ValidateConstructorCode},
		{"query-array-float32", testdata.PayloadQueryArrayFloat32DSL, testdata.PayloadQueryArrayFloat32ConstructorCode},
		{"query-array-float32-validate", testdata.PayloadQueryArrayFloat32ValidateDSL, testdata.PayloadQueryArrayFloat32ValidateConstructorCode},
		{"query-array-float64", testdata.PayloadQueryArrayFloat64DSL, testdata.PayloadQueryArrayFloat64ConstructorCode},
		{"query-array-float64-validate", testdata.PayloadQueryArrayFloat64ValidateDSL, testdata.PayloadQueryArrayFloat64ValidateConstructorCode},
		{"query-array-string", testdata.PayloadQueryArrayStringDSL, testdata.PayloadQueryArrayStringConstructorCode},
		{"query-array-string-validate", testdata.PayloadQueryArrayStringValidateDSL, testdata.PayloadQueryArrayStringValidateConstructorCode},
		{"query-array-bytes", testdata.PayloadQueryArrayBytesDSL, testdata.PayloadQueryArrayBytesConstructorCode},
		{"query-array-bytes-validate", testdata.PayloadQueryArrayBytesValidateDSL, testdata.PayloadQueryArrayBytesValidateConstructorCode},
		{"query-array-any", testdata.PayloadQueryArrayAnyDSL, testdata.PayloadQueryArrayAnyConstructorCode},
		{"query-array-any-validate", testdata.PayloadQueryArrayAnyValidateDSL, testdata.PayloadQueryArrayAnyValidateConstructorCode},
		{"query-map-string-string", testdata.PayloadQueryMapStringStringDSL, testdata.PayloadQueryMapStringStringConstructorCode},
		{"query-map-string-string-validate", testdata.PayloadQueryMapStringStringValidateDSL, testdata.PayloadQueryMapStringStringValidateConstructorCode},
		{"query-map-string-bool", testdata.PayloadQueryMapStringBoolDSL, testdata.PayloadQueryMapStringBoolConstructorCode},
		{"query-map-string-bool-validate", testdata.PayloadQueryMapStringBoolValidateDSL, testdata.PayloadQueryMapStringBoolValidateConstructorCode},
		{"query-map-bool-string", testdata.PayloadQueryMapBoolStringDSL, testdata.PayloadQueryMapBoolStringConstructorCode},
		{"query-map-bool-string-validate", testdata.PayloadQueryMapBoolStringValidateDSL, testdata.PayloadQueryMapBoolStringValidateConstructorCode},
		{"query-map-bool-bool", testdata.PayloadQueryMapBoolBoolDSL, testdata.PayloadQueryMapBoolBoolConstructorCode},
		{"query-map-bool-bool-validate", testdata.PayloadQueryMapBoolBoolValidateDSL, testdata.PayloadQueryMapBoolBoolValidateConstructorCode},
		{"query-map-string-array-string", testdata.PayloadQueryMapStringArrayStringDSL, testdata.PayloadQueryMapStringArrayStringConstructorCode},
		{"query-map-string-array-string-validate", testdata.PayloadQueryMapStringArrayStringValidateDSL, testdata.PayloadQueryMapStringArrayStringValidateConstructorCode},
		{"query-map-string-array-bool", testdata.PayloadQueryMapStringArrayBoolDSL, testdata.PayloadQueryMapStringArrayBoolConstructorCode},
		{"query-map-string-array-bool-validate", testdata.PayloadQueryMapStringArrayBoolValidateDSL, testdata.PayloadQueryMapStringArrayBoolValidateConstructorCode},
		{"query-map-bool-array-string", testdata.PayloadQueryMapBoolArrayStringDSL, testdata.PayloadQueryMapBoolArrayStringConstructorCode},
		{"query-map-bool-array-string-validate", testdata.PayloadQueryMapBoolArrayStringValidateDSL, testdata.PayloadQueryMapBoolArrayStringValidateConstructorCode},
		{"query-map-bool-array-bool", testdata.PayloadQueryMapBoolArrayBoolDSL, testdata.PayloadQueryMapBoolArrayBoolConstructorCode},
		{"query-map-bool-array-bool-validate", testdata.PayloadQueryMapBoolArrayBoolValidateDSL, testdata.PayloadQueryMapBoolArrayBoolValidateConstructorCode},

		{"query-string-mapped", testdata.PayloadQueryStringMappedDSL, testdata.PayloadQueryStringMappedConstructorCode},

		{"path-string", testdata.PayloadPathStringDSL, testdata.PayloadPathStringConstructorCode},
		{"path-string-validate", testdata.PayloadPathStringValidateDSL, testdata.PayloadPathStringValidateConstructorCode},
		{"path-array-string", testdata.PayloadPathArrayStringDSL, testdata.PayloadPathArrayStringConstructorCode},
		{"path-array-string-validate", testdata.PayloadPathArrayStringValidateDSL, testdata.PayloadPathArrayStringValidateConstructorCode},

		{"header-string", testdata.PayloadHeaderStringDSL, testdata.PayloadHeaderStringConstructorCode},
		{"header-string-validate", testdata.PayloadHeaderStringValidateDSL, testdata.PayloadHeaderStringValidateConstructorCode},
		{"header-array-string", testdata.PayloadHeaderArrayStringDSL, testdata.PayloadHeaderArrayStringConstructorCode},
		{"header-array-string-validate", testdata.PayloadHeaderArrayStringValidateDSL, testdata.PayloadHeaderArrayStringValidateConstructorCode},

		{"body-query-object", testdata.PayloadBodyQueryObjectDSL, testdata.PayloadBodyQueryObjectConstructorCode},
		{"body-query-object-validate", testdata.PayloadBodyQueryObjectValidateDSL, testdata.PayloadBodyQueryObjectValidateConstructorCode},
		{"body-query-user", testdata.PayloadBodyQueryUserDSL, testdata.PayloadBodyQueryUserConstructorCode},
		{"body-query-user-validate", testdata.PayloadBodyQueryUserValidateDSL, testdata.PayloadBodyQueryUserValidateConstructorCode},

		{"body-path-object", testdata.PayloadBodyPathObjectDSL, testdata.PayloadBodyPathObjectConstructorCode},
		{"body-path-object-validate", testdata.PayloadBodyPathObjectValidateDSL, testdata.PayloadBodyPathObjectValidateConstructorCode},
		{"body-path-user", testdata.PayloadBodyPathUserDSL, testdata.PayloadBodyPathUserConstructorCode},
		{"body-path-user-validate", testdata.PayloadBodyPathUserValidateDSL, testdata.PayloadBodyPathUserValidateConstructorCode},

		{"body-query-path-object", testdata.PayloadBodyQueryPathObjectDSL, testdata.PayloadBodyQueryPathObjectConstructorCode},
		{"body-query-path-object-validate", testdata.PayloadBodyQueryPathObjectValidateDSL, testdata.PayloadBodyQueryPathObjectValidateConstructorCode},
		{"body-query-path-user", testdata.PayloadBodyQueryPathUserDSL, testdata.PayloadBodyQueryPathUserConstructorCode},
		{"body-query-path-user-validate", testdata.PayloadBodyQueryPathUserValidateDSL, testdata.PayloadBodyQueryPathUserValidateConstructorCode},

		{"body-user-inner", testdata.PayloadBodyUserInnerDSL, testdata.PayloadBodyUserInnerConstructorCode},
		{"body-user-inner-default", testdata.PayloadBodyUserInnerDefaultDSL, testdata.PayloadBodyUserInnerDefaultConstructorCode},
		{"body-user-inner-origin", testdata.PayloadBodyUserOriginDSL, testdata.PayloadBodyUserOriginConstructorCode},
		{"body-inline-array-user", testdata.PayloadBodyInlineArrayUserDSL, testdata.PayloadBodyInlineArrayUserConstructorCode},
		{"body-inline-map-user", testdata.PayloadBodyInlineMapUserDSL, testdata.PayloadBodyInlineMapUserConstructorCode},
		{"body-inline-recursive-user", testdata.PayloadBodyInlineRecursiveUserDSL, testdata.PayloadBodyInlineRecursiveUserConstructorCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			if len(expr.Root.API.HTTP.Services) != 1 {
				t.Fatalf("got %d file(s), expected 1", len(expr.Root.API.HTTP.Services))
			}
			fs := serverType("", expr.Root.API.HTTP.Services[0], make(map[string]struct{}))
			sections := fs.SectionTemplates
			var section *codegen.SectionTemplate
			for _, s := range sections {
				if s.Source == serverTypeInitT {
					section = s
				}
			}
			if section == nil {
				t.Fatalf("could not find payload init section")
			}
			code := codegen.SectionCode(t, section)
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}
