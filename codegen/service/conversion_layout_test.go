// These tests require external conversion helpers to use complete receiver
// layouts when a union creates unlocated named collection branches.
package service

import (
	"fmt"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	external "goa.design/goa/v3/codegen/service/testdata/external-union"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestExternalConversionRetainsHelperLayouts(t *testing.T) {
	for _, placement := range []string{"local", "inherited", "explicit child"} {
		for _, reverse := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/reverse=%t", placement, reverse), func(t *testing.T) {
				var packet expr.UserType
				first := codegen.RunDSL(t, func() {
					dsl.API("first", func() {})
					child := dsl.Type("Child", func() {
						dsl.Attribute("value", dsl.String)
						dsl.Required("value")
					})
					entry := dsl.Type("Entry", func() {
						if placement == "explicit child" {
							dsl.Meta("struct:pkg:path", "inner/types")
						}
						dsl.OneOf("choice", func() {
							dsl.TypeName("Selection")
							dsl.Attribute("text", dsl.String)
							dsl.Attribute("child", child)
							dsl.Attribute("words", dsl.ArrayOf(dsl.String))
							dsl.Attribute("table", dsl.MapOf(dsl.String, dsl.ArrayOf(dsl.String)))
						})
						dsl.Required("choice")
					})
					packet = dsl.Type("Packet", func() {
						if placement != "local" {
							dsl.Meta("struct:pkg:path", "outer/types")
						}
						dsl.Attribute("entries", dsl.ArrayOf(entry))
						dsl.Required("entries")
						dsl.ConvertTo(external.Envelope{})
						dsl.CreateFrom(external.Envelope{})
					})
					dsl.Service("alpha", func() {
						dsl.Method("exchange", func() {
							dsl.Payload(packet)
							dsl.Result(packet)
						})
					})
				})
				second := codegen.RunDSL(t, func() {
					dsl.API("second", func() {})
					dsl.Service("beta", func() {
						dsl.Method("exchange", func() {
							dsl.Payload(packet)
							dsl.Result(packet)
						})
					})
				})
				roots := []*expr.RootExpr{first, second}
				if reverse {
					slices.Reverse(roots)
				}
				generation, plans := inheritedPlacementPlans(t, roots)
				proofs := make(map[string]string)
				var operationCount int
				for _, plan := range plans {
					for _, file := range plan.facts.externalConversions {
						owner := file.owner.ImportPath()
						if placement != "local" {
							require.Equal(t, "generated.local/gen/outer/types", owner)
							require.NotContains(t, file.imports.Paths(), "generated.local/gen/alpha")
							require.NotContains(t, file.imports.Paths(), "generated.local/gen/beta")
						}
						for _, operation := range file.operations {
							operationCount++
							require.Same(t, operation.receiverType, operation.receiverLayout.TypeDeclaration())
							entry := expr.AsArray(operation.receiverAttribute.Find("entries").Type).ElemType.Type
							choice := expr.AsUnion(expr.AsObject(entry).Attribute("choice").Type)
							for _, branch := range choice.Values {
								if branch.Name != "words" && branch.Name != "table" {
									continue
								}
								wrapper := branch.Attribute.Type.(expr.UserType)
								require.False(t, plan.facts.rootTypes.contains(wrapper))
								require.Nil(t, plan.facts.rootTypes.location(wrapper))
								require.NotContains(t, wrapper.Attribute().Meta, "struct:pkg:path")
								branchOwner := owner
								if placement == "explicit child" {
									branchOwner = "generated.local/gen/inner/types"
								}
								declaration, err := generation.Package(branchOwner).Type(wrapper)
								require.NoError(t, err)
								layouts := operation.receiverLayout.PlansForOccurrence(branch.Attribute)
								require.NotEmpty(t, layouts)
								for _, layout := range layouts {
									require.Same(t, declaration, layout.TypeDeclaration())
									require.Equal(t, branchOwner, layout.Owner())
								}
								t.Logf("receiver=%s service=%s branch=%s owner=%s exact_occurrences=%d", owner, operation.servicePath, wrapper.Name(), branchOwner, len(layouts))
							}
						}
						packageName := "types"
						if placement == "local" {
							packageName = filepath.Base(owner)
						}
						proofs[filepath.Join(codegen.Gendir, owner[len(generation.GenPkg())+1:], "conversion_layout_test.go")] = "package " + packageName + conversionLayoutRoundTrip
					}
				}
				if placement == "local" {
					require.Equal(t, 4, operationCount, "both directions retain both service-local receivers")
				} else {
					require.Equal(t, 2, operationCount, "shared receivers have one operation per direction")
				}
				require.NoError(t, generation.Freeze())
				for _, plan := range plans {
					require.NoError(t, plan.Link())
				}
				files, err := Files(plans...)
				require.NoError(t, err)
				compileGeneratedServiceFilesWith(t, files, proofs)
			})
		}
	}
}

const conversionLayoutRoundTrip = `
import (
	"reflect"
	"testing"
	external "goa.design/goa/v3/codegen/service/testdata/external-union"
)
func TestConversionHelperValues(t *testing.T) {
	choices := make([]external.Choice, 4)
	choices[0].SetText(external.Text_Value("text"))
	choices[1].SetChild(&external.Detail{Value: external.Text_Value("child")})
	choices[2].SetWords(external.Text_List{"first", "second"})
	choices[3].SetTable(external.Text_Table{"key": external.Text_List{"value"}})
	for _, choice := range choices {
		original := &external.Envelope{Entries: []*external.Entry{{Choice: choice}}}
		var value Packet
		value.CreateFromEnvelope(original)
		actual := value.ConvertToEnvelope()
		if !reflect.DeepEqual(original, actual) {
			t.Fatalf("branch %s changed: %#v", choice.Kind(), actual)
		}
	}
}
`
