// This file checks that an OpenAPI plan keeps the documents it built.
package codegen

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/stretchr/testify/require"

	goacodegen "goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestOpenAPIPlanKeepsBuiltFiles(t *testing.T) {
	root := expr.RunDSL(t, testdata.SimpleDSL)
	root.API.Contact = &expr.ContactExpr{Name: "before"}
	root.API.License = &expr.LicenseExpr{Name: "before"}
	root.API.HTTP.Consumes = []string{"application/before"}
	root.API.HTTP.Produces = []string{"application/before"}
	plan, err := NewOpenAPIPlan(root, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	files := plan.Files()
	before := renderOpenAPIFiles(t, files)

	root.API.Title = "changed after planning"
	root.API.HTTP.Services = nil
	root.API.Contact.Name = "after"
	root.API.License.Name = "after"
	root.API.HTTP.Consumes[0] = "application/after"
	root.API.HTTP.Produces[0] = "application/after"

	filesAgain := plan.Files()
	require.Len(t, filesAgain, len(files))
	for i := range files {
		require.Same(t, files[i], filesAgain[i])
	}
	require.Equal(t, before, renderOpenAPIFiles(t, filesAgain))
}

func TestOpenAPIPlanWithValuesDoesNotChangeDesign(t *testing.T) {
	root := expr.RunDSL(t, testdata.SimpleDSL)
	values := (openapi.Values{}).WithTitle(root.API, "Localized API")

	plan, err := NewOpenAPIPlanWithValues(
		root,
		expr.NewExampleGenerator(root.API.RandomizerFactory),
		values,
	)
	require.NoError(t, err)
	rendered := renderOpenAPIFiles(t, plan.Files())
	for _, document := range rendered {
		require.Contains(t, document, "Localized API")
	}
	require.NotEqual(t, "Localized API", root.API.Title)
}

func TestNewOpenAPIPlanFromSpecsUsesExactVersionsAndPaths(t *testing.T) {
	root := expr.RunDSL(t, testdata.SimpleDSL)
	plan, err := NewOpenAPIPlanFromSpecs(
		root,
		expr.NewExampleGenerator(root.API.RandomizerFactory),
		[]openapi.Spec{
			{Version: openapi.Version20, Path: "docs/api.v2"},
			{Version: openapi.Version32, Path: "reference/api"},
		},
		openapi.Values{},
	)
	require.NoError(t, err)
	paths := make([]string, len(plan.Files()))
	for index, file := range plan.Files() {
		paths[index] = file.Path
	}
	require.Equal(t, []string{
		"gen/docs/api.v2.json",
		"gen/docs/api.v2.yaml",
		"gen/reference/api.json",
		"gen/reference/api.yaml",
	}, paths)
}

func TestNewOpenAPIPlanFromSpecsRejectsInvalidSpecs(t *testing.T) {
	root := expr.RunDSL(t, testdata.SimpleDSL)
	tests := []struct {
		name  string
		specs []openapi.Spec
		err   string
	}{
		{name: "unknown version", specs: []openapi.Spec{{Version: "4.0", Path: "http/api"}}, err: `unsupported OpenAPI version "4.0"`},
		{name: "empty path", specs: []openapi.Spec{{Version: openapi.Version30}}, err: "path cannot be empty"},
		{name: "absolute path", specs: []openapi.Spec{{Version: openapi.Version30, Path: "/api"}}, err: "path must be relative"},
		{name: "escaping path", specs: []openapi.Spec{{Version: openapi.Version30, Path: "../api"}}, err: "path must not escape"},
		{name: "json extension", specs: []openapi.Spec{{Version: openapi.Version30, Path: "api.json"}}, err: "path must not include an extension"},
		{name: "same version", specs: []openapi.Spec{{Version: openapi.Version30, Path: "api"}, {Version: openapi.Version30, Path: "other"}}, err: `version "3.0" appears more than once`},
		{name: "same path", specs: []openapi.Spec{{Version: openapi.Version20, Path: "api"}, {Version: openapi.Version30, Path: "api"}}, err: `same output path "api"`},
		{name: "portable path collision", specs: []openapi.Spec{{Version: openapi.Version20, Path: "API"}, {Version: openapi.Version30, Path: "api"}}, err: "case-insensitive filesystem"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewOpenAPIPlanFromSpecs(
				root,
				expr.NewExampleGenerator(root.API.RandomizerFactory),
				test.specs,
				openapi.Values{},
			)
			require.ErrorContains(t, err, test.err)
		})
	}
}

func TestNewOpenAPIPlanWrappersKeepOrdinaryOutput(t *testing.T) {
	root := expr.RunDSL(t, testdata.SimpleDSL)
	specs, err := openapi.Specs(root.API.Meta)
	require.NoError(t, err)
	ordinary, err := NewOpenAPIPlan(root, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	explicit, err := NewOpenAPIPlanFromSpecs(
		root,
		expr.NewExampleGenerator(root.API.RandomizerFactory),
		specs,
		openapi.Values{},
	)
	require.NoError(t, err)
	require.Equal(t, renderOpenAPIFiles(t, ordinary.Files()), renderOpenAPIFiles(t, explicit.Files()))
}

// renderOpenAPIFiles renders each planned file and returns its text by path.
func renderOpenAPIFiles(t *testing.T, files []*goacodegen.File) map[string]string {
	t.Helper()
	rendered := make(map[string]string, len(files))
	for _, file := range files {
		require.Len(t, file.SectionTemplates, 1)
		section := file.SectionTemplates[0]
		var output bytes.Buffer
		tmpl := template.Must(template.New(section.Name).Funcs(section.FuncMap).Parse(section.Source))
		require.NoError(t, tmpl.Execute(&output, section.Data))
		rendered[file.Path] = output.String()
	}
	return rendered
}
