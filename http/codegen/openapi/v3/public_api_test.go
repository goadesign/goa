// This file protects the released OpenAPI v3 function signatures and checks
// that their default files match files produced with the root's example
// generator and no replacement values.
package openapiv3_test

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
	openapiv3 "goa.design/goa/v3/http/codegen/openapi/v3"
)

var (
	_ func(*expr.RootExpr, openapi.Version) *openapiv3.OpenAPI      = openapiv3.New
	_ func(*expr.RootExpr, openapi.Version, string) []*codegen.File = openapiv3.Files

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

	gotSpec := openapiv3.New(root, openapi.Version30)
	wantSpec := openapiv3.NewWithValues(
		root,
		openapi.Version30,
		expr.NewExampleGenerator(root.API.RandomizerFactory),
		openapi.Values{},
	)
	require.Equal(t, wantSpec, gotSpec)
	otherSpec := openapiv3.NewWithValues(
		root,
		openapi.Version30,
		expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("other")),
		openapi.Values{},
	)
	require.NotEqual(t, otherSpec, gotSpec)

	gotFiles := openapiv3.Files(root, openapi.Version30, openapi.DefaultPath30)
	wantFiles := openapiv3.FilesWithValues(
		root,
		openapi.Version30,
		openapi.DefaultPath30,
		expr.NewExampleGenerator(root.API.RandomizerFactory),
		openapi.Values{},
	)
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
