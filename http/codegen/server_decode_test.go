package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestDecode(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"decode-path-custom-float32", testdata.PayloadPathCustomFloat32DSL},
		{"decode-path-custom-float64", testdata.PayloadPathCustomFloat64DSL},
		{"decode-path-custom-int", testdata.PayloadPathCustomIntDSL},
		{"decode-path-custom-int32", testdata.PayloadPathCustomInt32DSL},
		{"decode-path-custom-int64", testdata.PayloadPathCustomInt64DSL},
		{"decode-path-custom-uint", testdata.PayloadPathCustomUIntDSL},
		{"decode-path-custom-uint32", testdata.PayloadPathCustomUInt32DSL},
		{"decode-path-custom-uint64", testdata.PayloadPathCustomUInt64DSL},
		{"decode-path-custom-text-unmarshaler", testdata.PayloadPathCustomTextUnmarshalerDSL},
		{"decode-path-custom-text-unmarshaler-format", testdata.PayloadPathCustomTextUnmarshalerFormatDSL},
		{"decode-query-custom-text-unmarshaler", testdata.PayloadQueryCustomTextUnmarshalerDSL},
		{"decode-query-custom-text-unmarshaler-optional", testdata.PayloadQueryCustomTextUnmarshalerOptionalDSL},
		{"decode-query-bool", testdata.PayloadQueryBoolDSL},
		{"decode-query-bool-validate", testdata.PayloadQueryBoolValidateDSL},
		{"decode-query-int", testdata.PayloadQueryIntDSL},
		{"decode-query-int-validate", testdata.PayloadQueryIntValidateDSL},
		{"decode-query-int32", testdata.PayloadQueryInt32DSL},
		{"decode-query-int32-validate", testdata.PayloadQueryInt32ValidateDSL},
		{"decode-query-int64", testdata.PayloadQueryInt64DSL},
		{"decode-query-int64-validate", testdata.PayloadQueryInt64ValidateDSL},
		{"decode-query-uint", testdata.PayloadQueryUIntDSL},
		{"decode-query-uint-validate", testdata.PayloadQueryUIntValidateDSL},
		{"decode-query-uint32", testdata.PayloadQueryUInt32DSL},
		{"decode-query-uint32-validate", testdata.PayloadQueryUInt32ValidateDSL},
		{"decode-query-uint64", testdata.PayloadQueryUInt64DSL},
		{"decode-query-uint64-validate", testdata.PayloadQueryUInt64ValidateDSL},
		{"decode-query-float32", testdata.PayloadQueryFloat32DSL},
		{"decode-query-float32-validate", testdata.PayloadQueryFloat32ValidateDSL},
		{"decode-query-float64", testdata.PayloadQueryFloat64DSL},
		{"decode-query-float64-validate", testdata.PayloadQueryFloat64ValidateDSL},
		{"decode-query-string", testdata.PayloadQueryStringDSL},
		{"decode-query-string-validate", testdata.PayloadQueryStringValidateDSL},
		{"decode-query-string-not-required-validate", testdata.PayloadQueryStringNotRequiredValidateDSL},
		{"decode-query-bytes", testdata.PayloadQueryBytesDSL},
		{"decode-query-bytes-validate", testdata.PayloadQueryBytesValidateDSL},
		{"decode-query-any", testdata.PayloadQueryAnyDSL},
		{"decode-query-any-validate", testdata.PayloadQueryAnyValidateDSL},
		{"decode-query-array-bool", testdata.PayloadQueryArrayBoolDSL},
		{"decode-query-array-bool-validate", testdata.PayloadQueryArrayBoolValidateDSL},
		{"decode-query-array-int", testdata.PayloadQueryArrayIntDSL},
		{"decode-query-array-int-validate", testdata.PayloadQueryArrayIntValidateDSL},
		{"decode-query-array-int32", testdata.PayloadQueryArrayInt32DSL},
		{"decode-query-array-int32-validate", testdata.PayloadQueryArrayInt32ValidateDSL},
		{"decode-query-array-int64", testdata.PayloadQueryArrayInt64DSL},
		{"decode-query-array-int64-validate", testdata.PayloadQueryArrayInt64ValidateDSL},
		{"decode-query-array-uint", testdata.PayloadQueryArrayUIntDSL},
		{"decode-query-array-uint-validate", testdata.PayloadQueryArrayUIntValidateDSL},
		{"decode-query-array-uint32", testdata.PayloadQueryArrayUInt32DSL},
		{"decode-query-array-uint32-validate", testdata.PayloadQueryArrayUInt32ValidateDSL},
		{"decode-query-array-uint64", testdata.PayloadQueryArrayUInt64DSL},
		{"decode-query-array-uint64-validate", testdata.PayloadQueryArrayUInt64ValidateDSL},
		{"decode-query-array-float32", testdata.PayloadQueryArrayFloat32DSL},
		{"decode-query-array-float32-validate", testdata.PayloadQueryArrayFloat32ValidateDSL},
		{"decode-query-array-float64", testdata.PayloadQueryArrayFloat64DSL},
		{"decode-query-array-float64-validate", testdata.PayloadQueryArrayFloat64ValidateDSL},
		{"decode-query-array-string", testdata.PayloadQueryArrayStringDSL},
		{"decode-query-array-string-validate", testdata.PayloadQueryArrayStringValidateDSL},
		{"decode-query-array-bytes", testdata.PayloadQueryArrayBytesDSL},
		{"decode-query-array-bytes-validate", testdata.PayloadQueryArrayBytesValidateDSL},
		{"decode-query-array-any", testdata.PayloadQueryArrayAnyDSL},
		{"decode-query-array-any-validate", testdata.PayloadQueryArrayAnyValidateDSL},
		{"decode-query-map-string-string", testdata.PayloadQueryMapStringStringDSL},
		{"decode-query-map-string-string-validate", testdata.PayloadQueryMapStringStringValidateDSL},
		{"decode-query-map-string-bool", testdata.PayloadQueryMapStringBoolDSL},
		{"decode-query-map-string-bool-validate", testdata.PayloadQueryMapStringBoolValidateDSL},
		{"decode-query-map-bool-string", testdata.PayloadQueryMapBoolStringDSL},
		{"decode-query-map-bool-string-validate", testdata.PayloadQueryMapBoolStringValidateDSL},
		{"decode-query-map-bool-bool", testdata.PayloadQueryMapBoolBoolDSL},
		{"decode-query-map-bool-bool-validate", testdata.PayloadQueryMapBoolBoolValidateDSL},
		{"decode-query-map-string-array-string", testdata.PayloadQueryMapStringArrayStringDSL},
		{"decode-query-map-string-array-string-validate", testdata.PayloadQueryMapStringArrayStringValidateDSL},
		{"decode-query-map-string-array-bool", testdata.PayloadQueryMapStringArrayBoolDSL},
		{"decode-query-map-string-array-bool-validate", testdata.PayloadQueryMapStringArrayBoolValidateDSL},
		{"decode-query-map-bool-array-string", testdata.PayloadQueryMapBoolArrayStringDSL},
		{"decode-query-map-bool-array-string-validate", testdata.PayloadQueryMapBoolArrayStringValidateDSL},
		{"decode-query-map-bool-array-bool", testdata.PayloadQueryMapBoolArrayBoolDSL},
		{"decode-query-map-bool-array-bool-validate", testdata.PayloadQueryMapBoolArrayBoolValidateDSL},

		{"decode-query-primitive-string-validate", testdata.PayloadQueryPrimitiveStringValidateDSL},
		{"decode-query-primitive-bool-validate", testdata.PayloadQueryPrimitiveBoolValidateDSL},
		{"decode-query-primitive-array-string-validate", testdata.PayloadQueryPrimitiveArrayStringValidateDSL},
		{"decode-query-primitive-array-bool-validate", testdata.PayloadQueryPrimitiveArrayBoolValidateDSL},
		{"decode-query-primitive-map-string-array-string-validate", testdata.PayloadQueryPrimitiveMapStringArrayStringValidateDSL},
		{"decode-query-primitive-map-string-bool-validate", testdata.PayloadQueryPrimitiveMapStringBoolValidateDSL},
		{"decode-query-primitive-map-bool-array-bool-validate", testdata.PayloadQueryPrimitiveMapBoolArrayBoolValidateDSL},
		{"decode-query-map-string-map-int-string-validate", testdata.PayloadQueryMapStringMapIntStringValidateDSL},
		{"decode-query-map-int-map-string-array-int-validate", testdata.PayloadQueryMapIntMapStringArrayIntValidateDSL},

		{"decode-query-string-mapped", testdata.PayloadQueryStringMappedDSL},

		{"decode-query-string-default", testdata.PayloadQueryStringDefaultDSL},
		{"decode-query-string-slice-default", testdata.PayloadQueryStringSliceDefaultDSL},
		{"decode-query-string-default-validate", testdata.PayloadQueryStringDefaultValidateDSL},
		{"decode-query-primitive-string-default", testdata.PayloadQueryPrimitiveStringDefaultDSL},
		{"decode-query-string-extended-payload", testdata.PayloadExtendedQueryStringDSL},

		{"decode-path-string", testdata.PayloadPathStringDSL},
		{"decode-path-string-validate", testdata.PayloadPathStringValidateDSL},
		{"decode-path-array-string", testdata.PayloadPathArrayStringDSL},
		{"decode-path-array-string-validate", testdata.PayloadPathArrayStringValidateDSL},

		{"decode-path-primitive-string-validate", testdata.PayloadPathPrimitiveStringValidateDSL},
		{"decode-path-primitive-string-formatip-validate", testdata.PayloadPathPrimitiveStringFormatIPValidateDSL},
		{"decode-path-primitive-bool-validate", testdata.PayloadPathPrimitiveBoolValidateDSL},
		{"decode-path-primitive-array-string-validate", testdata.PayloadPathPrimitiveArrayStringValidateDSL},
		{"decode-path-primitive-array-bool-validate", testdata.PayloadPathPrimitiveArrayBoolValidateDSL},

		{"decode-header-string", testdata.PayloadHeaderStringDSL},
		{"decode-header-string-validate", testdata.PayloadHeaderStringValidateDSL},
		{"decode-header-array-string", testdata.PayloadHeaderArrayStringDSL},
		{"decode-header-array-string-validate", testdata.PayloadHeaderArrayStringValidateDSL},

		{"decode-header-primitive-string-validate", testdata.PayloadHeaderPrimitiveStringValidateDSL},
		{"decode-header-primitive-bool-validate", testdata.PayloadHeaderPrimitiveBoolValidateDSL},
		{"decode-header-primitive-array-string-validate", testdata.PayloadHeaderPrimitiveArrayStringValidateDSL},
		{"decode-header-primitive-array-bool-validate", testdata.PayloadHeaderPrimitiveArrayBoolValidateDSL},

		{"decode-header-string-default", testdata.PayloadHeaderStringDefaultDSL},
		{"decode-header-string-default-validate", testdata.PayloadHeaderStringDefaultValidateDSL},
		{"decode-header-primitive-string-default", testdata.PayloadHeaderPrimitiveStringDefaultDSL},

		{"decode-cookie-string", testdata.PayloadCookieStringDSL},
		{"decode-cookie-string-validate", testdata.PayloadCookieStringValidateDSL},

		{"decode-cookie-primitive-string-validate", testdata.PayloadCookiePrimitiveStringValidateDSL},
		{"decode-cookie-primitive-bool-validate", testdata.PayloadCookiePrimitiveBoolValidateDSL},

		{"decode-cookie-string-default", testdata.PayloadCookieStringDefaultDSL},
		{"decode-cookie-string-default-validate", testdata.PayloadCookieStringDefaultValidateDSL},
		{"decode-cookie-primitive-string-default", testdata.PayloadCookiePrimitiveStringDefaultDSL},

		{"decode-body-string", testdata.PayloadBodyStringDSL},
		{"decode-body-string-validate", testdata.PayloadBodyStringValidateDSL},
		{"decode-body-user", testdata.PayloadBodyUserDSL},
		{"decode-body-user-required", testdata.PayloadBodyUserRequiredDSL},
		{"decode-body-user-nested", testdata.PayloadBodyNestedUserDSL},
		{"decode-body-user-validate", testdata.PayloadBodyUserValidateDSL},
		{"decode-body-object", testdata.PayloadBodyObjectDSL},
		{"decode-body-object-required", testdata.PayloadBodyObjectRequiredDSL},
		{"decode-body-object-validate", testdata.PayloadBodyObjectValidateDSL},
		{"decode-body-union", testdata.PayloadBodyUnionDSL},
		{"decode-body-union-validate", testdata.PayloadBodyUnionValidateDSL},
		{"decode-body-union-user", testdata.PayloadBodyUnionUserDSL},
		{"decode-body-union-user-validate", testdata.PayloadBodyUnionUserValidateDSL},
		{"decode-body-array-string", testdata.PayloadBodyArrayStringDSL},
		{"decode-body-array-string-validate", testdata.PayloadBodyArrayStringValidateDSL},
		{"decode-body-array-user", testdata.PayloadBodyArrayUserDSL},
		{"decode-body-array-user-validate", testdata.PayloadBodyArrayUserValidateDSL},
		{"decode-body-map-string", testdata.PayloadBodyMapStringDSL},
		{"decode-body-map-string-validate", testdata.PayloadBodyMapStringValidateDSL},
		{"decode-body-map-user", testdata.PayloadBodyMapUserDSL},
		{"decode-body-map-user-validate", testdata.PayloadBodyMapUserValidateDSL},
		{"decode-deep-user", testdata.PayloadDeepUserDSL},
		{"decode-body-nested-uuid-field", testdata.PayloadBodyNestedUUIDFieldDSL},

		{"decode-body-primitive-string-validate", testdata.PayloadBodyPrimitiveStringValidateDSL},
		{"decode-body-primitive-bool-validate", testdata.PayloadBodyPrimitiveBoolValidateDSL},
		{"decode-body-primitive-array-string-validate", testdata.PayloadBodyPrimitiveArrayStringValidateDSL},
		{"decode-body-primitive-array-bool-validate", testdata.PayloadBodyPrimitiveArrayBoolValidateDSL},

		{"decode-body-primitive-array-user-required", testdata.PayloadBodyPrimitiveArrayUserRequiredDSL},
		{"decode-body-primitive-array-user-validate", testdata.PayloadBodyPrimitiveArrayUserValidateDSL},
		{"decode-body-primitive-field-array-user", testdata.PayloadBodyPrimitiveFieldArrayUserDSL},
		{"decode-body-extend-primitive-field-array-user", testdata.PayloadExtendBodyPrimitiveFieldArrayUserDSL},
		{"decode-body-extend-primitive-field-string", testdata.PayloadExtendBodyPrimitiveFieldStringDSL},
		{"decode-body-primitive-field-array-user-validate", testdata.PayloadBodyPrimitiveFieldArrayUserValidateDSL},

		{"decode-body-query-object", testdata.PayloadBodyQueryObjectDSL},
		{"decode-body-query-object-validate", testdata.PayloadBodyQueryObjectValidateDSL},
		{"decode-body-query-user", testdata.PayloadBodyQueryUserDSL},
		{"decode-body-query-user-validate", testdata.PayloadBodyQueryUserValidateDSL},

		{"decode-body-path-object", testdata.PayloadBodyPathObjectDSL},
		{"decode-body-path-object-validate", testdata.PayloadBodyPathObjectValidateDSL},
		{"decode-body-path-user", testdata.PayloadBodyPathUserDSL},
		{"decode-body-path-user-validate", testdata.PayloadBodyPathUserValidateDSL},

		{"decode-body-query-path-object", testdata.PayloadBodyQueryPathObjectDSL},
		{"decode-body-query-path-object-validate", testdata.PayloadBodyQueryPathObjectValidateDSL},
		{"decode-body-query-path-user", testdata.PayloadBodyQueryPathUserDSL},
		{"decode-body-query-path-user-validate", testdata.PayloadBodyQueryPathUserValidateDSL},

		{"decode-map-query-primitive-primitive", testdata.PayloadMapQueryPrimitivePrimitiveDSL},
		{"decode-map-query-primitive-array", testdata.PayloadMapQueryPrimitiveArrayDSL},
		{"decode-map-query-object", testdata.PayloadMapQueryObjectDSL},
		{"decode-multipart-body-primitive", testdata.PayloadMultipartPrimitiveDSL},
		{"decode-multipart-body-user-type", testdata.PayloadMultipartUserTypeDSL},
		{"decode-multipart-body-array-type", testdata.PayloadMultipartArrayTypeDSL},
		{"decode-multipart-body-map-type", testdata.PayloadMultipartMapTypeDSL},
		{"decode-with-params-and-headers-dsl", testdata.WithParamsAndHeadersBlockDSL},

		{"decode-query-int-alias", testdata.QueryIntAliasDSL},
		{"decode-query-int-alias-validate", testdata.QueryIntAliasValidateDSL},
		{"decode-query-array-alias", testdata.QueryArrayAliasDSL},
		{"decode-query-array-alias-validate", testdata.QueryArrayAliasValidateDSL},
		{"decode-query-map-alias", testdata.QueryMapAliasDSL},
		{"decode-query-map-alias-validate", testdata.QueryMapAliasValidateDSL},
		{"decode-query-array-nested-alias-validate", testdata.QueryArrayNestedAliasValidateDSL},
		{"decode-header-int-alias", testdata.HeaderIntAliasDSL},
		{"decode-path-int-alias", testdata.PathIntAliasDSL},

		{"decode-body-custom-name", testdata.PayloadBodyCustomNameDSL},
		{"decode-path-custom-name", testdata.PayloadPathCustomNameDSL},
		{"decode-query-custom-name", testdata.PayloadQueryCustomNameDSL},
		{"decode-header-custom-name", testdata.PayloadHeaderCustomNameDSL},
		{"decode-cookie-custom-name", testdata.PayloadCookieCustomNameDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := expr.RunDSL(t, c.DSL)
			services := CreateHTTPServices(root)
			fs := ServerFiles("", services)
			require.Len(t, fs, 2)
			sections := fs[1].SectionTemplates
			require.Greater(t, len(sections), 2)
			code := codegen.SectionCode(t, sections[2])
			require.NotContains(t, code, "return nil,")
			testutil.AssertGo(t, "testdata/golden/server_decode_"+c.Name+".go.golden", code)
		})
	}
}
