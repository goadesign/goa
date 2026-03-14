package codegen

import (
	"goa.design/goa/v3/codegen/testutil"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestPayloadConstructor(t *testing.T) {
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

		{"query-string-mapped", testdata.PayloadQueryStringMappedDSL},

		{"path-string", testdata.PayloadPathStringDSL},
		{"path-string-validate", testdata.PayloadPathStringValidateDSL},
		{"path-array-string", testdata.PayloadPathArrayStringDSL},
		{"path-array-string-validate", testdata.PayloadPathArrayStringValidateDSL},

		{"header-string", testdata.PayloadHeaderStringDSL},
		{"header-string-validate", testdata.PayloadHeaderStringValidateDSL},
		{"header-array-string", testdata.PayloadHeaderArrayStringDSL},
		{"header-array-string-validate", testdata.PayloadHeaderArrayStringValidateDSL},

		{"body-query-object", testdata.PayloadBodyQueryObjectDSL},
		{"body-query-object-validate", testdata.PayloadBodyQueryObjectValidateDSL},
		{"body-query-user", testdata.PayloadBodyQueryUserDSL},
		{"body-query-user-validate", testdata.PayloadBodyQueryUserValidateDSL},
		{"body-union", testdata.PayloadBodyUnionDSL},
		{"constructor-union", testdata.ConstructorUnionHTTPDSL},
		{"constructor-union-custom-keys", testdata.ConstructorUnionCustomKeysHTTPDSL},
		{"nested-top-level-constructor-union-custom-keys", testdata.NestedTopLevelConstructorUnionCustomKeysHTTPDSL},
		{"body-query-user-union", testdata.PayloadBodyQueryUserUnionDSL},
		{"body-query-user-union-validate", testdata.PayloadBodyQueryUserUnionValidateDSL},

		{"body-path-object", testdata.PayloadBodyPathObjectDSL},
		{"body-path-object-validate", testdata.PayloadBodyPathObjectValidateDSL},
		{"body-path-user", testdata.PayloadBodyPathUserDSL},
		{"body-path-user-validate", testdata.PayloadBodyPathUserValidateDSL},

		{"body-query-path-object", testdata.PayloadBodyQueryPathObjectDSL},
		{"body-query-path-object-validate", testdata.PayloadBodyQueryPathObjectValidateDSL},
		{"body-query-path-user", testdata.PayloadBodyQueryPathUserDSL},
		{"body-query-path-user-validate", testdata.PayloadBodyQueryPathUserValidateDSL},

		{"body-user-inner", testdata.PayloadBodyUserInnerDSL},
		{"body-user-inner-default", testdata.PayloadBodyUserInnerDefaultDSL},
		{"body-user-inner-origin", testdata.PayloadBodyUserOriginDSL},
		{"body-inline-array-user", testdata.PayloadBodyInlineArrayUserDSL},
		{"body-inline-map-user", testdata.PayloadBodyInlineMapUserDSL},
		{"body-inline-recursive-user", testdata.PayloadBodyInlineRecursiveUserDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := RunHTTPDSL(t, c.DSL)
			require.Len(t, root.API.HTTP.Services, 1)
			services := CreateHTTPServices(root)
			fs := serverType("", root.API.HTTP.Services[0], services)
			sections := fs.SectionTemplates
			var section *codegen.SectionTemplate
			for _, s := range sections {
				if s.Source == httpTemplates.Read("server_type_init") {
					section = s
				}
			}
			require.NotNil(t, section)
			code := codegen.SectionCode(t, section)
			testutil.AssertGo(t, "testdata/golden/server_payload_types_"+c.Name+".go.golden", code)
		})
	}
}
