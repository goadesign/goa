package codegen

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestBodyTypeDecl(t *testing.T) {
	const genpkg = "gen"

	cases := []struct {
		Name string
		DSL  func()
	}{
		{"body-user-inner", testdata.PayloadBodyUserInnerDSL},
		{"body-path-user-validate", testdata.PayloadBodyPathUserValidateDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := RunHTTPDSL(t, c.DSL)
			services := CreateHTTPServices(root)
			fs := clientType(genpkg, root.API.HTTP.Services[0], make(map[string]struct{}), services)
			section := fs.SectionTemplates[1]
			code := codegen.SectionCode(t, section)
			testutil.AssertGo(t, "testdata/golden/client_body_type_decl_"+c.Name+".go.golden", code)
		})
	}
}

func TestBodyTypeInit(t *testing.T) {
	const genpkg = "gen"
	cases := []struct {
		Name         string
		DSL          func()
		SectionIndex int
	}{
		{"body-user-inner", testdata.PayloadBodyUserInnerDSL, 3},
		{"body-path-user-validate", testdata.PayloadBodyPathUserValidateDSL, 2},
		{"body-primitive-array-user-validate", testdata.PayloadBodyPrimitiveArrayUserValidateDSL, 2},
		{"result-body-user", testdata.ResultBodyObjectHeaderDSL, 2},
		{"result-body-user-required", testdata.ResultBodyUserRequiredDSL, 3},
		{"result-body-inline-object", testdata.ResultBodyInlineObjectDSL, 2},
		{"result-explicit-body-primitive", testdata.ExplicitBodyPrimitiveResultMultipleViewsDSL, 1},
		{"result-explicit-body-user-type", testdata.ExplicitBodyUserResultMultipleViewsDSL, 3},
		{"result-explicit-body-object", testdata.ExplicitBodyUserResultObjectDSL, 3},
		{"result-explicit-body-object-views", testdata.ExplicitBodyUserResultObjectMultipleViewDSL, 3},
		{"body-streaming-aliased-array", testdata.StreamingAliasedArrayDSL, 4},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := RunHTTPDSL(t, c.DSL)
			services := CreateHTTPServices(root)
			fs := clientType(genpkg, root.API.HTTP.Services[0], make(map[string]struct{}), services)
			section := fs.SectionTemplates[c.SectionIndex]
			code := codegen.SectionCode(t, section)
			testutil.AssertGo(t, "testdata/golden/client_body_type_init_"+c.Name+".go.golden", code)
		})
	}
}

func TestClientTypes(t *testing.T) {
	const genpkg = "gen"
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"client-mixed-payload-attrs", testdata.MixedPayloadInBodyDSL},
		{"client-multiple-methods", testdata.MultipleMethodsDSL},
		{"client-multiple-methods-with-array-type-payloads", testdata.MultipleMethodsWithArrayTypePayloadsDSL},
		{"client-payload-extend-validate", testdata.PayloadExtendedValidateDSL},
		{"client-result-type-validate", testdata.ResultTypeValidateDSL},
		{"client-with-result-collection", testdata.ResultWithResultCollectionDSL},
		{"client-with-result-view", testdata.ResultWithResultViewDSL},
		{"client-empty-error-response-body", testdata.EmptyErrorResponseBodyDSL},
		{"client-with-error-custom-pkg", testdata.WithErrorCustomPkgDSL},
		{"client-body-custom-name", testdata.PayloadBodyCustomNameDSL},
		{"client-path-custom-name", testdata.PayloadPathCustomNameDSL},
		{"client-query-custom-name", testdata.PayloadQueryCustomNameDSL},
		{"client-header-custom-name", testdata.PayloadHeaderCustomNameDSL},
		{"client-cookie-custom-name", testdata.PayloadCookieCustomNameDSL},
		{"client-constructor-union", testdata.ConstructorUnionHTTPDSL},
		{"client-constructor-union-custom-keys", testdata.ConstructorUnionCustomKeysHTTPDSL},
		{"client-nested-top-level-constructor-union-custom-keys", testdata.NestedTopLevelConstructorUnionCustomKeysHTTPDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := RunHTTPDSL(t, c.DSL)
			services := CreateHTTPServices(root)
			fs := clientType(genpkg, root.API.HTTP.Services[0], make(map[string]struct{}), services)
			var buf bytes.Buffer
			for _, s := range fs.SectionTemplates[1:] {
				require.NoError(t, s.Write(&buf))
			}
			code := codegen.FormatTestCode(t, "package foo\n"+buf.String())
			testutil.AssertGo(t, "testdata/golden/client_types_"+c.Name+".go.golden", code)
		})
	}
}

func TestClientTypeFiles(t *testing.T) {
	const genpkg = "gen"
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"multiple-services-same-payload-and-result", testdata.MultipleServicesSamePayloadAndResultDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := RunHTTPDSL(t, c.DSL)
			services := CreateHTTPServices(root)
			fw := ClientTypeFiles(genpkg, services)
			for i, fs := range fw {
				var buf bytes.Buffer
				for _, s := range fs.SectionTemplates[1:] {
					require.NoError(t, s.Write(&buf))
				}
				code := codegen.FormatTestCode(t, "package foo\n"+buf.String())
				testutil.AssertGo(t, "testdata/golden/client_type_file_"+c.Name+"_"+string(rune('0'+i))+".go.golden", code)
			}
		})
	}
}
