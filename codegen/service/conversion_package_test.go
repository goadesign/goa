// These tests compile external conversions beside the generated types that own
// them, including service names and explicit package paths with several words.
package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service/testdata/external"
	"goa.design/goa/v3/dsl"
)

func TestExternalConversionPackageDeclarations(t *testing.T) {
	for _, tc := range []struct {
		name      string
		service   string
		placement string
	}{
		{"simple service", "catalog", ""},
		{"underscore service", "demo_catalog", ""},
		{"acronym service", "HTTP API", ""},
		{"explicit package", "catalog", "shared/demo_models"},
		{"explicit acronym package", "catalog", "shared/HTTPModels"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := codegen.RunDSL(t, func() {
				model := dsl.Type("Model", func() {
					if tc.placement != "" {
						dsl.Meta("struct:pkg:path", tc.placement)
					}
					dsl.Attribute("foo", dsl.String)
					dsl.Required("foo")
					dsl.ConvertTo(external.ConvertModel{})
					dsl.CreateFrom(external.ConvertModel{})
				})
				dsl.Service(tc.service, func() {
					dsl.Method("exchange", func() {
						dsl.Payload(model)
						dsl.Result(model)
					})
				})
			})
			plan := retainedServicePlanForPackage(t, root)
			files, err := Files(plan)
			require.NoError(t, err)
			compileGeneratedServiceFiles(t, files)
		})
	}
}
