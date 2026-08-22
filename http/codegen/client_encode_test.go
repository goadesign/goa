package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestClientEncode(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"query-bool", testdata.PayloadQueryBoolDSL},
		{"query-bool-validate", testdata.PayloadQueryBoolValidateDSL},
		{"query-int", testdata.PayloadQueryIntDSL},
		{"query-int-validate", testdata.PayloadQueryIntValidateDSL},
		{"query-int32", testdata.PayloadQueryInt32DSL},
		{"query-int32-validate", testdata.PayloadQueryInt32ValidateDSL},
		{"query-int64", testdata.PayloadQueryInt64DSL},
		{"query-int64-validate", testdata.PayloadQueryInt64ValidateDSL},
		{"query-uint", testdata.PayloadQueryUIntDSL},
		{"query-uint-validate", testdata.PayloadQueryUIntValidateDSL},
		{"query-uint32", testdata.PayloadQueryUInt32DSL},
		{"query-uint32-validate", testdata.PayloadQueryUInt32ValidateDSL},
		{"query-uint64", testdata.PayloadQueryUInt64DSL},
		{"query-uint64-validate", testdata.PayloadQueryUInt64ValidateDSL},
		{"query-float32", testdata.PayloadQueryFloat32DSL},
		{"query-float32-validate", testdata.PayloadQueryFloat32ValidateDSL},
		{"query-float64", testdata.PayloadQueryFloat64DSL},
		{"query-float64-validate", testdata.PayloadQueryFloat64ValidateDSL},
		{"query-string", testdata.PayloadQueryStringDSL},
		{"query-string-validate", testdata.PayloadQueryStringValidateDSL},
		{"query-bytes", testdata.PayloadQueryBytesDSL},
		{"query-bytes-validate", testdata.PayloadQueryBytesValidateDSL},
		{"query-any", testdata.PayloadQueryAnyDSL},
		{"query-any-validate", testdata.PayloadQueryAnyValidateDSL},
		{"query-array-bool", testdata.PayloadQueryArrayBoolDSL},
		{"query-array-bool-validate", testdata.PayloadQueryArrayBoolValidateDSL},
		{"query-array-int", testdata.PayloadQueryArrayIntDSL},
		{"query-array-int-validate", testdata.PayloadQueryArrayIntValidateDSL},
		{"query-array-int32", testdata.PayloadQueryArrayInt32DSL},
		{"query-array-int32-validate", testdata.PayloadQueryArrayInt32ValidateDSL},
		{"query-array-int64", testdata.PayloadQueryArrayInt64DSL},
		{"query-array-int64-validate", testdata.PayloadQueryArrayInt64ValidateDSL},
		{"query-array-uint", testdata.PayloadQueryArrayUIntDSL},
		{"query-array-uint-validate", testdata.PayloadQueryArrayUIntValidateDSL},
		{"query-array-uint32", testdata.PayloadQueryArrayUInt32DSL},
		{"query-array-uint32-validate", testdata.PayloadQueryArrayUInt32ValidateDSL},
		{"query-array-uint64", testdata.PayloadQueryArrayUInt64DSL},
		{"query-array-uint64-validate", testdata.PayloadQueryArrayUInt64ValidateDSL},
		{"query-array-float32", testdata.PayloadQueryArrayFloat32DSL},
		{"query-array-float32-validate", testdata.PayloadQueryArrayFloat32ValidateDSL},
		{"query-array-float64", testdata.PayloadQueryArrayFloat64DSL},
		{"query-array-float64-validate", testdata.PayloadQueryArrayFloat64ValidateDSL},
		{"query-array-string", testdata.PayloadQueryArrayStringDSL},
		{"query-array-string-validate", testdata.PayloadQueryArrayStringValidateDSL},
		{"query-array-bytes", testdata.PayloadQueryArrayBytesDSL},
		{"query-array-bytes-validate", testdata.PayloadQueryArrayBytesValidateDSL},
		{"query-array-any", testdata.PayloadQueryArrayAnyDSL},
		{"query-array-any-validate", testdata.PayloadQueryArrayAnyValidateDSL},
		{"query-array-alias", testdata.PayloadQueryArrayAliasDSL},
		{"query-map-string-string", testdata.PayloadQueryMapStringStringDSL},
		{"query-map-string-string-validate", testdata.PayloadQueryMapStringStringValidateDSL},
		{"query-map-string-bool", testdata.PayloadQueryMapStringBoolDSL},
		{"query-map-string-bool-validate", testdata.PayloadQueryMapStringBoolValidateDSL},
		{"query-map-bool-string", testdata.PayloadQueryMapBoolStringDSL},
		{"query-map-bool-string-validate", testdata.PayloadQueryMapBoolStringValidateDSL},
		{"query-map-bool-bool", testdata.PayloadQueryMapBoolBoolDSL},
		{"query-map-bool-bool-validate", testdata.PayloadQueryMapBoolBoolValidateDSL},
		{"query-map-string-array-string", testdata.PayloadQueryMapStringArrayStringDSL},
		{"query-map-string-array-string-validate", testdata.PayloadQueryMapStringArrayStringValidateDSL},
		{"query-map-string-array-bool", testdata.PayloadQueryMapStringArrayBoolDSL},
		{"query-map-string-array-bool-validate", testdata.PayloadQueryMapStringArrayBoolValidateDSL},
		{"query-map-bool-array-string", testdata.PayloadQueryMapBoolArrayStringDSL},
		{"query-map-bool-array-string-validate", testdata.PayloadQueryMapBoolArrayStringValidateDSL},
		{"query-map-bool-array-bool", testdata.PayloadQueryMapBoolArrayBoolDSL},
		{"query-map-bool-array-bool-validate", testdata.PayloadQueryMapBoolArrayBoolValidateDSL},

		{"query-primitive-string-validate", testdata.PayloadQueryPrimitiveStringValidateDSL},
		{"query-primitive-bool-validate", testdata.PayloadQueryPrimitiveBoolValidateDSL},
		{"query-primitive-array-string-validate", testdata.PayloadQueryPrimitiveArrayStringValidateDSL},
		{"query-primitive-array-bool-validate", testdata.PayloadQueryPrimitiveArrayBoolValidateDSL},
		{"query-primitive-map-string-array-string-validate", testdata.PayloadQueryPrimitiveMapStringArrayStringValidateDSL},
		{"query-primitive-map-string-bool-validate", testdata.PayloadQueryPrimitiveMapStringBoolValidateDSL},
		{"query-primitive-map-bool-array-bool-validate", testdata.PayloadQueryPrimitiveMapBoolArrayBoolValidateDSL},
		{"query-map-string-map-int-string-validate", testdata.PayloadQueryMapStringMapIntStringValidateDSL},
		{"query-map-int-map-string-array-int-validate", testdata.PayloadQueryMapIntMapStringArrayIntValidateDSL},

		{"query-string-mapped", testdata.PayloadQueryStringMappedDSL},

		{"query-string-default", testdata.PayloadQueryStringDefaultDSL},
		{"query-primitive-string-default", testdata.PayloadQueryPrimitiveStringDefaultDSL},
		{"query-jwt-authorization", testdata.PayloadJWTAuthorizationQueryDSL},

		{"header-string", testdata.PayloadHeaderStringDSL},
		{"header-string-validate", testdata.PayloadHeaderStringValidateDSL},
		{"header-array-string", testdata.PayloadHeaderArrayStringDSL},
		{"header-array-string-validate", testdata.PayloadHeaderArrayStringValidateDSL},
		{"header-int", testdata.PayloadHeaderIntDSL},
		{"header-int-validate", testdata.PayloadHeaderIntValidateDSL},
		{"header-array-int", testdata.PayloadHeaderArrayIntDSL},
		{"header-array-int-validate", testdata.PayloadHeaderArrayIntValidateDSL},

		{"header-primitive-string-validate", testdata.PayloadHeaderPrimitiveStringValidateDSL},
		{"header-primitive-bool-validate", testdata.PayloadHeaderPrimitiveBoolValidateDSL},
		{"header-primitive-array-string-validate", testdata.PayloadHeaderPrimitiveArrayStringValidateDSL},
		{"header-primitive-array-bool-validate", testdata.PayloadHeaderPrimitiveArrayBoolValidateDSL},

		{"header-string-default", testdata.PayloadHeaderStringDefaultDSL},
		{"header-primitive-string-default", testdata.PayloadHeaderPrimitiveStringDefaultDSL},
		{"header-jwt-authorization", testdata.PayloadJWTAuthorizationHeaderDSL},
		{"header-bearer-authorization", testdata.PayloadBearerAuthorizationHeaderDSL},
		{"header-jwt-custom-header", testdata.PayloadJWTAuthorizationCustomHeaderDSL},

		{"body-string", testdata.PayloadBodyStringDSL},
		{"body-string-validate", testdata.PayloadBodyStringValidateDSL},
		{"body-user", testdata.PayloadBodyUserDSL},
		{"body-user-validate", testdata.PayloadBodyUserValidateDSL},
		{"body-array-string", testdata.PayloadBodyArrayStringDSL},
		{"body-array-string-validate", testdata.PayloadBodyArrayStringValidateDSL},
		{"body-array-user", testdata.PayloadBodyArrayUserDSL},
		{"body-array-user-validate", testdata.PayloadBodyArrayUserValidateDSL},
		{"body-map-string", testdata.PayloadBodyMapStringDSL},
		{"body-map-string-validate", testdata.PayloadBodyMapStringValidateDSL},
		{"body-map-user", testdata.PayloadBodyMapUserDSL},
		{"body-map-user-validate", testdata.PayloadBodyMapUserValidateDSL},

		{"body-primitive-string-validate", testdata.PayloadBodyPrimitiveStringValidateDSL},
		{"body-primitive-bool-validate", testdata.PayloadBodyPrimitiveBoolValidateDSL},
		{"body-primitive-array-string-validate", testdata.PayloadBodyPrimitiveArrayStringValidateDSL},
		{"body-primitive-array-bool-validate", testdata.PayloadBodyPrimitiveArrayBoolValidateDSL},

		{"body-primitive-array-user-validate", testdata.PayloadBodyPrimitiveArrayUserValidateDSL},
		{"body-primitive-field-array-user", testdata.PayloadBodyPrimitiveFieldArrayUserDSL},
		{"body-primitive-field-array-user-validate", testdata.PayloadBodyPrimitiveFieldArrayUserValidateDSL},

		{"body-query-object", testdata.PayloadBodyQueryObjectDSL},
		{"body-query-object-validate", testdata.PayloadBodyQueryObjectValidateDSL},
		{"body-query-user", testdata.PayloadBodyQueryUserDSL},
		{"body-query-user-validate", testdata.PayloadBodyQueryUserValidateDSL},

		{"body-path-object", testdata.PayloadBodyPathObjectDSL},
		{"body-path-object-validate", testdata.PayloadBodyPathObjectValidateDSL},
		{"body-path-user", testdata.PayloadBodyPathUserDSL},
		{"body-path-user-validate", testdata.PayloadBodyPathUserValidateDSL},

		{"body-query-path-object", testdata.PayloadBodyQueryPathObjectDSL},
		{"body-query-path-object-validate", testdata.PayloadBodyQueryPathObjectValidateDSL},
		{"body-query-path-user", testdata.PayloadBodyQueryPathUserDSL},
		{"body-query-path-user-validate", testdata.PayloadBodyQueryPathUserValidateDSL},

		{"map-query-primitive-primitive", testdata.PayloadMapQueryPrimitivePrimitiveDSL},
		{"map-query-primitive-array", testdata.PayloadMapQueryPrimitiveArrayDSL},
		{"map-query-object", testdata.PayloadMapQueryObjectDSL},
		{"multipart-body-primitive", testdata.PayloadMultipartPrimitiveDSL},
		{"multipart-body-user-type", testdata.PayloadMultipartUserTypeDSL},
		{"multipart-body-array-type", testdata.PayloadMultipartArrayTypeDSL},
		{"multipart-body-map-type", testdata.PayloadMultipartMapTypeDSL},

		// aliases
		{"query-int-alias", testdata.QueryIntAliasDSL},
		{"query-int-alias-validate", testdata.QueryIntAliasValidateDSL},
		{"query-array-alias-type", testdata.QueryArrayAliasDSL},
		{"query-array-alias-validate", testdata.QueryArrayAliasValidateDSL},
		{"query-map-alias", testdata.QueryMapAliasDSL},
		{"query-map-alias-validate", testdata.QueryMapAliasValidateDSL},
		{"query-array-nested-alias-validate", testdata.QueryArrayNestedAliasValidateDSL},

		{"body-custom-name", testdata.PayloadBodyCustomNameDSL},
		// path-custom-name is not needed because no encoder is created.
		{"query-custom-name", testdata.PayloadQueryCustomNameDSL},
		{"header-custom-name", testdata.PayloadHeaderCustomNameDSL},
		{"cookie-custom-name", testdata.PayloadCookieCustomNameDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := expr.RunDSL(t, c.DSL)
			plan := linkedHTTPPlanForRoot(t, root)
			fs := plan.ClientFiles()
			require.Len(t, fs, 2)
			sections := fs[1].SectionTemplates
			require.Greater(t, len(sections), 2)
			code := codegen.SectionCode(t, sections[2])
			testutil.AssertGo(t, "testdata/golden/client_encode_"+c.Name+".go.golden", code)
		})
	}
}

func TestClientBuildRequest(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"path-string", testdata.PayloadPathStringDSL},
		{"path-string-required", testdata.PayloadPathStringValidateDSL},
		{"path-string-default", testdata.PayloadPathStringDefaultDSL},
		{"path-object", testdata.PayloadPathObjectDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := expr.RunDSL(t, c.DSL)
			plan := linkedHTTPPlanForRoot(t, root)
			fs := plan.ClientFiles()
			require.Len(t, fs, 2)
			sections := fs[1].SectionTemplates
			require.Greater(t, len(sections), 2)
			code := codegen.SectionCode(t, sections[1])
			testutil.AssertGo(t, "testdata/golden/client_build_request_"+c.Name+".go.golden", code)
		})
	}
}
