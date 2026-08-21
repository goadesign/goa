// This file verifies that disabling OpenAPI examples uses document-private
// disabled generators and never changes service or evaluated design state.
package codegen

import (
	"bytes"
	"maps"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/require"

	goacodegen "goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestOpenAPIDisabledExamplesDoNotConsumeServiceState(t *testing.T) {
	root := expr.RunDSL(t, testdata.SimpleDSL)
	root.API.Meta = expr.MetaExpr{"openapi:example": {"false"}}
	factory := root.API.RandomizerFactory
	meta := maps.Clone(root.API.Meta)
	examples := expr.NewExampleGenerator(factory)
	generation, err := goacodegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	require.NoError(t, service.Plan(root, generation))
	require.NoError(t, generation.Freeze())
	services, err := service.NewServicesData(root, generation, examples)
	require.NoError(t, err)
	method := services.Get("testService").Methods[0]
	payloadExample := method.PayloadEx
	require.NotNil(t, payloadExample)

	openapi.Definitions = make(map[string]*openapi.Schema)
	files, err := OpenAPIFiles(root, examples)
	require.NoError(t, err)
	require.Len(t, files, 6)
	for _, file := range files {
		require.Len(t, file.SectionTemplates, 1)
		section := file.SectionTemplates[0]
		var rendered bytes.Buffer
		tmpl := template.Must(template.New("openapi").Funcs(section.FuncMap).Parse(section.Source))
		require.NoError(t, tmpl.Execute(&rendered, section.Data))
		content := rendered.String()
		if strings.HasSuffix(file.Path, ".json") {
			require.NotContains(t, content, `"example"`)
		} else {
			require.NotContains(t, content, "\nexample:")
		}
	}

	require.Equal(t, payloadExample, method.PayloadEx)
	require.Equal(t, factory, root.API.RandomizerFactory)
	require.Equal(t, meta, root.API.Meta)
}
