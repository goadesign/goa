// These tests verify copied Reference fields share equivalent HTTP conversions.
package codegen

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestTransformHelpersIgnoreGenerationMetadataAfterReference(t *testing.T) {
	var baseline []string
	for _, marked := range []bool{false, true} {
		name := "normal_reference"
		if marked {
			name = "plugin_generation_metadata"
		}
		t.Run(name, func(t *testing.T) {
			var identifier expr.UserType
			root := expr.RunDSL(t, func() {
				identifier = dsl.Type("Identifier", dsl.String, func() {
					dsl.Meta("struct:pkg:path", "types")
				})
				branch := dsl.Type("Selection", func() {
					dsl.Field(1, "id", identifier)
					dsl.Required("id")
					dsl.Meta("struct:pkg:path", "types")
				})
				scope := dsl.Type("Scope", func() {
					dsl.OneOf("choice", func() {
						dsl.Field(1, "selected", branch)
					})
					dsl.Required("choice")
					dsl.Meta("struct:pkg:path", "types")
				})
				input := dsl.Type("Input", func() {
					dsl.Field(1, "scope", scope)
					dsl.Field(2, "internal", dsl.String)
					dsl.Required("scope", "internal")
				})
				dsl.Service("catalog", func() {
					dsl.Method("lookup", func() {
						dsl.Payload(func() {
							dsl.Field(1, "scope", scope)
							dsl.Required("scope")
						})
						dsl.HTTP(func() {
							dsl.POST("/lookup")
						})
					})
					dsl.Method("create", func() {
						dsl.Payload(func() {
							dsl.Reference(input)
							dsl.Field(1, "scope")
							dsl.Required("scope")
						})
						dsl.HTTP(func() {
							dsl.POST("/create")
						})
					})
				})
			})
			if marked {
				identifier.Attribute().AddMeta("type:generate:force")
				identifier.Attribute().AddMeta("openapi:generate", "false")
			}
			plan := linkedHTTPPlanForRoot(t, root)
			service := plan.services.Get("catalog")
			require.Len(t, service.clientWireTypes.transformHelpers, 4)
			require.Len(t, service.serverWireTypes.transformHelpers, 2)
			var names []string
			for _, catalog := range []*wireTypeCatalog{service.clientWireTypes, service.serverWireTypes} {
				for _, helper := range catalog.transformHelpers {
					names = append(names, helper.declaration.Name())
				}
			}
			slices.Sort(names)
			if marked {
				require.Equal(t, baseline, names)
			} else {
				baseline = names
			}

		})
	}
}
