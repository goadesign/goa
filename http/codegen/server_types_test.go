package codegen

import (
	"goa.design/goa/v3/codegen/testutil"
	"bytes"
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











