package codegen

import (
	"bytes"
	"goa.design/goa/v3/codegen/testutil"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestServerTypes(t *testing.T) {
	const genpkg = "gen"
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"server-mixed-payload-attrs", testdata.MixedPayloadInBodyDSL},
		{"server-multiple-methods", testdata.MultipleMethodsDSL},
		{"server-payload-extend-validate", testdata.PayloadExtendedValidateDSL},
		{"server-result-type-validate", testdata.ResultTypeValidateDSL},
		{"server-with-result-collection", testdata.ResultWithResultCollectionDSL},
		{"server-with-result-view", testdata.ResultWithResultViewDSL},
		{"server-empty-error-response-body", testdata.EmptyErrorResponseBodyDSL},
		{"server-with-error-custom-pkg", testdata.WithErrorCustomPkgDSL},
		{"server-body-custom-name", testdata.PayloadBodyCustomNameDSL},
		{"server-path-custom-name", testdata.PayloadPathCustomNameDSL},
		{"server-query-custom-name", testdata.PayloadQueryCustomNameDSL},
		{"server-header-custom-name", testdata.PayloadHeaderCustomNameDSL},
		{"server-cookie-custom-name", testdata.PayloadCookieCustomNameDSL},
		{"server-constructor-union", testdata.ConstructorUnionHTTPDSL},
		{"server-constructor-union-custom-keys", testdata.ConstructorUnionCustomKeysHTTPDSL},
		{"server-nested-top-level-constructor-union-custom-keys", testdata.NestedTopLevelConstructorUnionCustomKeysHTTPDSL},
		{"server-payload-with-validated-alias", testdata.PayloadWithValidatedAliasDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := RunHTTPDSL(t, c.DSL)
			services := CreateHTTPServices(root)
			fs := serverType(genpkg, root.API.HTTP.Services[0], services)
			var buf bytes.Buffer
			for _, s := range fs.SectionTemplates[1:] {
				require.NoError(t, s.Write(&buf))
			}
			code := codegen.FormatTestCode(t, "package foo\n"+buf.String())
			testutil.AssertGo(t, "testdata/golden/server_types_"+c.Name+".go.golden", code)
		})
	}
}

func TestServerTypesNormalizedConstructorUnionIdentifiers(t *testing.T) {
	root := RunHTTPDSL(t, testdata.ConstructorUnionNormalizedBranchNamesHTTPDSL)
	services := CreateHTTPServices(root)
	fs := serverType("gen", root.API.HTTP.Services[0], services)

	var buf bytes.Buffer
	for _, s := range fs.SectionTemplates[1:] {
		require.NoError(t, s.Write(&buf))
	}
	code := codegen.FormatTestCode(t, "package foo\n"+buf.String())

	require.Contains(t, code, ` = "FirstNormalizedBranch"`)
	require.Contains(t, code, ` = "SecondNormalizedBranch"`)
	require.Contains(t, code, "FooBar ")
	require.Contains(t, code, "FooBar2 ")
	require.Contains(t, code, "AsFirstNormalizedBranch()")
	require.Contains(t, code, "AsSecondNormalizedBranch()")
}
