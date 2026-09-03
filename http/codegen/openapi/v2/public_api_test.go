// This file protects the released OpenAPI v2 function signatures and checks
// that their default files match files produced with the root's example
// generator and no replacement values.
package openapiv2_test

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
	openapiv2 "goa.design/goa/v3/http/codegen/openapi/v2"
)

var (
	_ func(*expr.RootExpr, *expr.HostExpr) (*openapiv2.V2, error) = openapiv2.NewV2
	_ func(*expr.RootExpr, string) ([]*codegen.File, error)       = openapiv2.Files

	facadeDSL = func() {
		dsl.API("facade", func() {
			dsl.Server("facade", func() {
				dsl.Host("localhost", func() {
					dsl.URI("https://goa.design")
				})
			})
		})
		dsl.Service("facade", func() {
			dsl.Method("show", func() {
				dsl.Payload(func() {
					dsl.Attribute("message", dsl.String)
				})
				dsl.Result(func() {
					dsl.Attribute("answer", dsl.String)
				})
				dsl.HTTP(func() {
					dsl.POST("/items")
				})
			})
		})
	}
)

func TestDefaultFacadeMatchesWithValues(t *testing.T) {
	root := expr.RunDSL(t, facadeDSL)
	root.API.RandomizerFactory = expr.NewFakerRandomizerFactory("released facade")
	host := root.API.Servers[0].Hosts[0]

	gotSpec, err := openapiv2.NewV2(root, host)
	require.NoError(t, err)
	wantSpec, err := openapiv2.NewV2WithValues(
		root,
		host,
		expr.NewExampleGenerator(root.API.RandomizerFactory),
		openapi.Values{},
	)
	require.NoError(t, err)
	require.Equal(t, wantSpec, gotSpec)
	otherSpec, err := openapiv2.NewV2WithValues(
		root,
		host,
		expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("other")),
		openapi.Values{},
	)
	require.NoError(t, err)
	require.NotEqual(t, otherSpec, gotSpec)

	gotFiles, err := openapiv2.Files(root, openapi.DefaultPath20)
	require.NoError(t, err)
	wantFiles, err := openapiv2.FilesWithValues(
		root,
		openapi.DefaultPath20,
		expr.NewExampleGenerator(root.API.RandomizerFactory),
		openapi.Values{},
	)
	require.NoError(t, err)
	require.Equal(t, renderFiles(t, wantFiles), renderFiles(t, gotFiles))
}

// renderFiles runs each file template so the test compares the documents that
// users receive instead of comparing template implementation details.
func renderFiles(t *testing.T, files []*codegen.File) map[string]string {
	t.Helper()

	rendered := make(map[string]string, len(files))
	for _, file := range files {
		var buf bytes.Buffer
		for _, section := range file.SectionTemplates {
			tmpl, err := template.New("openapi").Funcs(section.FuncMap).Parse(section.Source)
			require.NoError(t, err)
			require.NoError(t, tmpl.Execute(&buf, section.Data))
		}
		rendered[file.Path] = buf.String()
	}
	return rendered
}
