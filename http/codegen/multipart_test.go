package codegen

import (
	"testing"

	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/expr"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestServerMultipartFuncType(t *testing.T) {
	const genpkg = "gen"
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"multipart-body-primitive", testdata.PayloadMultipartPrimitiveDSL},
		{"multipart-body-user-type", testdata.PayloadMultipartUserTypeDSL},
		{"multipart-body-array-type", testdata.PayloadMultipartArrayTypeDSL},
		{"multipart-body-map-type", testdata.PayloadMultipartMapTypeDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := expr.RunDSL(t, c.DSL)
			plan := linkedHTTPPlanForRoot(t, root)
			fs := plan.ServerFiles()
			require.Len(t, fs, 2)
			sections := fs[0].SectionTemplates
			require.Greater(t, len(sections), 5)
			code := codegen.SectionCode(t, sections[3])
			testutil.AssertGo(t, "testdata/golden/server_multipart_"+c.Name+".go.golden", code)
		})
	}
}

func TestClientMultipartFuncType(t *testing.T) {
	const genpkg = "gen"
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"multipart-body-primitive", testdata.PayloadMultipartPrimitiveDSL},
		{"multipart-body-user-type", testdata.PayloadMultipartUserTypeDSL},
		{"multipart-body-array-type", testdata.PayloadMultipartArrayTypeDSL},
		{"multipart-body-map-type", testdata.PayloadMultipartMapTypeDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := expr.RunDSL(t, c.DSL)
			plan := linkedHTTPPlanForRoot(t, root)
			fs := plan.ClientFiles()
			require.Len(t, fs, 2)
			sections := fs[0].SectionTemplates
			require.Greater(t, len(sections), 4)
			code := codegen.SectionCode(t, sections[2])
			testutil.AssertGo(t, "testdata/golden/client_multipart_"+c.Name+".go.golden", code)
		})
	}
}

func TestServerMultipartNewFunc(t *testing.T) {
	const genpkg = "gen"
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"server-multipart-body-primitive", testdata.PayloadMultipartPrimitiveDSL},
		{"server-multipart-body-user-type", testdata.PayloadMultipartUserTypeDSL},
		{"server-multipart-body-array-type", testdata.PayloadMultipartArrayTypeDSL},
		{"server-multipart-body-map-type", testdata.PayloadMultipartMapTypeDSL},
		{"server-multipart-with-param", testdata.PayloadMultipartWithParamDSL},
		{"server-multipart-with-params-and-headers", testdata.PayloadMultipartWithParamsAndHeadersDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := expr.RunDSL(t, c.DSL)
			plan := linkedHTTPPlanForRoot(t, root)
			fs := plan.ServerFiles()
			require.Len(t, fs, 2)
			sections := fs[1].SectionTemplates
			require.Greater(t, len(sections), 3)
			code := codegen.SectionCode(t, sections[3])
			testutil.AssertGo(t, "testdata/golden/server_multipart_"+c.Name+".go.golden", code)
		})
	}
}

func TestClientMultipartNewFunc(t *testing.T) {
	const genpkg = "gen"
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"client-multipart-body-primitive", testdata.PayloadMultipartPrimitiveDSL},
		{"client-multipart-body-user-type", testdata.PayloadMultipartUserTypeDSL},
		{"client-multipart-body-array-type", testdata.PayloadMultipartArrayTypeDSL},
		{"client-multipart-body-map-type", testdata.PayloadMultipartMapTypeDSL},
		{"client-multipart-with-param", testdata.PayloadMultipartWithParamDSL},
		{"client-multipart-with-params-and-headers", testdata.PayloadMultipartWithParamsAndHeadersDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := expr.RunDSL(t, c.DSL)
			plan := linkedHTTPPlanForRoot(t, root)
			fs := plan.ClientFiles()
			require.Len(t, fs, 2)
			sections := fs[1].SectionTemplates
			require.Greater(t, len(sections), 3)
			code := codegen.SectionCode(t, sections[3])
			testutil.AssertGo(t, "testdata/golden/client_multipart_"+c.Name+".go.golden", code)
		})
	}
}
