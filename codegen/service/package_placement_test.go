// These tests evaluate reusable schema DSL before planning complete generations.
// They verify shared identity, strict conflicts, immutable metadata and external
// conversions through the emitted Go packages.
package service

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	external "goa.design/goa/v3/codegen/service/testdata/external-union"
	"goa.design/goa/v3/codegen/service/testdata/inherited-placement/schema"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestInheritedPlacementAcrossRoots(t *testing.T) {
	var baseline map[string][]byte
	for _, reverseRoots := range []bool{false, true} {
		for _, reverseServices := range []bool{false, true} {
			roots, values := inheritedPlacementRoots(t, reverseServices)
			if reverseRoots {
				slices.Reverse(roots)
			}
			for _, value := range []expr.UserType{values.Child, values.Entry, values.Nested, values.Text} {
				require.NotContains(t, value.Attribute().Meta, "struct:pkg:path")
			}
			generation, plans := inheritedPlacementPlans(t, roots)
			var childDeclaration *codegen.TypeDeclaration
			for original, declaration := range generation.UserTypes() {
				if original.Origin() == values.Child.Origin() {
					require.Nil(t, childDeclaration, "child emitted in more than one package")
					childDeclaration = declaration
					require.Equal(t, "generated.local/gen/shared/types", declaration.PackagePath())
				}
				firstAttribute, first, err := plans[0].UserTypeLayout(original, declaration)
				require.NoError(t, err)
				require.Same(t, original, firstAttribute.Type)
				require.True(t, first.MatchesOccurrence(firstAttribute))
				for _, plan := range plans {
					otherAttribute, other, err := plan.UserTypeLayout(original, declaration)
					require.NoError(t, err)
					require.Same(t, firstAttribute, otherAttribute, "all roots must query the same retained attribute")
					require.Same(t, first, other, "all roots must query the same retained pair")
				}
			}
			require.NotNil(t, childDeclaration)
			for _, plan := range plans {
				for _, service := range plan.Root().Services {
					layout, err := plan.MethodTypeLayout(service.Method("child"), service.Method("child").Payload)
					require.NoError(t, err)
					require.Same(t, childDeclaration, layout.TypeDeclaration())
				}
			}
			require.NoError(t, generation.Freeze())
			for _, plan := range plans {
				require.NoError(t, plan.Link())
			}
			files, err := Files(plans...)
			require.NoError(t, err)
			require.Len(t, filesAtPath(files, filepath.Join(codegen.Gendir, "shared", "types", "child.go")), 1)
			require.Len(t, filesAtPath(files, filepath.Join(codegen.Gendir, "shared", "types", "convert.go")), 1)
			rendered := renderedServiceFiles(t, files)
			if baseline == nil {
				baseline = rendered
			} else {
				require.Equal(t, baseline, rendered)
			}
			compileGeneratedServiceFilesWith(t, files, map[string]string{
				filepath.Join(codegen.Gendir, "shared", "types", "placement_test.go"): inheritedPlacementRoundTrip,
			})
			for _, value := range []expr.UserType{values.Child, values.Entry, values.Nested, values.Text} {
				require.NotContains(t, value.Attribute().Meta, "struct:pkg:path")
			}
		}
	}
}

func TestInheritedPlacementConflicts(t *testing.T) {
	for _, edge := range []string{"direct", "array", "map key", "map value", "union"} {
		for _, reverse := range []bool{false, true} {
			var child expr.UserType
			first := codegen.RunDSL(t, func() {
				dsl.API("first", func() {})
				child = dsl.Type("Child", dsl.String)
				dsl.Type("FirstRoot", func() {
					dsl.Meta("struct:pkg:path", "first/types")
					switch edge {
					case "direct":
						dsl.Attribute("child", child)
					case "array":
						dsl.Attribute("child", dsl.ArrayOf(child))
					case "map key":
						dsl.Attribute("child", dsl.MapOf(child, dsl.String))
					case "map value":
						dsl.Attribute("child", dsl.MapOf(dsl.String, child))
					case "union":
						dsl.OneOf("child", func() {
							dsl.Attribute("value", child)
							dsl.Attribute("text", dsl.String)
						})
					}
				})
			})
			second := codegen.RunDSL(t, func() {
				dsl.API("second", func() {})
				dsl.Type("SecondRoot", func() {
					dsl.Meta("struct:pkg:path", "second/types")
					dsl.Attribute("child", child)
				})
			})
			roots := []*expr.RootExpr{first, second}
			if reverse {
				slices.Reverse(roots)
			}
			generation, inputs := inheritedPlacementInputs(t, roots)
			plans, err := NewPlans(generation, inputs...)
			require.ErrorContains(t, err, `type "Child" requires incompatible packages`)
			require.ErrorContains(t, err, `"first/types" via first.FirstRoot.child`)
			require.ErrorContains(t, err, `"second/types"`)
			require.Nil(t, plans)
			for range generation.UserTypes() {
				t.Error("conflict submitted an original declaration")
			}
			require.NotContains(t, child.Attribute().Meta, "struct:pkg:path")
		}
	}
}

func TestUserTypeLayoutExactPairsAndIndependentCommands(t *testing.T) {
	var shared, ordinary expr.UserType
	root := codegen.RunDSL(t, func() {
		shared = dsl.Type("Child", func() {
			dsl.Attribute("value", dsl.String)
		})
		dsl.Type("SharedRoot", func() {
			dsl.Meta("struct:pkg:path", "shared/types")
			dsl.Attribute("child", shared)
		})
		ordinary = dsl.Type("Local", func() {
			dsl.Attribute("value", dsl.String)
		})
		for _, name := range []string{"alpha", "beta"} {
			dsl.Service(name, func() {
				dsl.Method("child", func() { dsl.Payload(shared) })
				dsl.Method("local", func() { dsl.Payload(ordinary) })
			})
		}
	})
	generation, plans := inheritedPlacementPlans(t, []*expr.RootExpr{root})
	var sharedDeclaration *codegen.TypeDeclaration
	var localDeclarations []*codegen.TypeDeclaration
	for original, declaration := range generation.UserTypes() {
		switch original.Origin() {
		case shared.Origin():
			sharedDeclaration = declaration
		case ordinary.Origin():
			localDeclarations = append(localDeclarations, declaration)
		}
	}
	require.NotNil(t, sharedDeclaration)
	require.Len(t, localDeclarations, 2)
	localAttributes := make([]*expr.AttributeExpr, 0, len(localDeclarations))
	localLayouts := make([]*codegen.GoTypePlan, 0, len(localDeclarations))
	for _, declaration := range localDeclarations {
		attribute, layout, err := plans[0].UserTypeLayout(ordinary, declaration)
		require.NoError(t, err)
		require.Same(t, ordinary.Origin(), attribute.Type)
		require.True(t, layout.MatchesOccurrence(attribute))
		require.Equal(t, declaration.PackagePath(), layout.Owner())
		localAttributes = append(localAttributes, attribute)
		localLayouts = append(localLayouts, layout)
	}
	require.NotSame(t, localAttributes[0], localAttributes[1])
	require.NotSame(t, localLayouts[0], localLayouts[1])
	attribute, layout, err := plans[0].UserTypeLayout(shared, sharedDeclaration)
	require.NoError(t, err)
	require.Same(t, shared.Origin(), attribute.Type)
	require.True(t, layout.MatchesOccurrence(attribute))
	require.False(t, layout.MatchesOccurrence(&expr.AttributeExpr{Type: shared}))
	require.Same(t, sharedDeclaration, layout.TypeDeclaration())

	copy := expr.DupAtt(&expr.AttributeExpr{Type: shared}).Type.(expr.UserType)
	copyAttribute, copyLayout, err := plans[0].UserTypeLayout(copy, sharedDeclaration)
	require.NoError(t, err)
	require.NotSame(t, copy, copyAttribute.Type)
	require.Same(t, attribute, copyAttribute)
	require.Same(t, layout, copyLayout)
	for _, test := range []struct {
		name        string
		original    expr.UserType
		declaration *codegen.TypeDeclaration
	}{
		{"unrelated original", ordinary, sharedDeclaration},
		{"nil original", nil, sharedDeclaration},
		{"nil declaration", shared, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			rejectedAttribute, rejectedLayout, err := plans[0].UserTypeLayout(test.original, test.declaration)
			require.ErrorContains(t, err, "not a registered pair")
			require.Nil(t, rejectedAttribute)
			require.Nil(t, rejectedLayout)
		})
	}

	// The same imported declaration is local in another generation without the
	// shared root. This also gives the query a foreign declaration for one origin.
	independent := codegen.RunDSL(t, func() {
		dsl.Service("independent", func() {
			dsl.Method("child", func() { dsl.Payload(shared) })
		})
	})
	otherGeneration, otherPlans := inheritedPlacementPlans(t, []*expr.RootExpr{independent})
	for original, declaration := range otherGeneration.UserTypes() {
		if original.Origin() != shared.Origin() {
			continue
		}
		otherAttribute, other, err := otherPlans[0].UserTypeLayout(original, declaration)
		require.NoError(t, err)
		require.Same(t, shared.Origin(), otherAttribute.Type)
		require.True(t, other.MatchesOccurrence(otherAttribute))
		require.NotSame(t, attribute, otherAttribute)
		require.NotSame(t, layout, other)
		require.Equal(t, "generated.local/gen/independent", other.Owner())
		rejectedAttribute, rejectedLayout, err := plans[0].UserTypeLayout(shared, declaration)
		require.ErrorContains(t, err, "not a registered pair")
		require.Nil(t, rejectedAttribute)
		require.Nil(t, rejectedLayout)
	}
	rejectedAttribute, rejectedLayout, err := otherPlans[0].UserTypeLayout(shared, sharedDeclaration)
	require.ErrorContains(t, err, "not a registered pair")
	require.Nil(t, rejectedAttribute)
	require.Nil(t, rejectedLayout)
	require.NoError(t, generation.Freeze())
	frozenAttribute, frozenLayout, err := plans[0].UserTypeLayout(shared, sharedDeclaration)
	require.NoError(t, err)
	require.Same(t, attribute, frozenAttribute)
	require.Same(t, layout, frozenLayout)
	require.NotContains(t, shared.Attribute().Meta, "struct:pkg:path")
}

func TestInheritedPlacementDistinguishesSameNameOrigins(t *testing.T) {
	var located, local expr.UserType
	first := codegen.RunDSL(t, func() {
		dsl.API("first", func() {})
		located = dsl.Type("Child", func() { dsl.Attribute("value", dsl.String) })
		dsl.Type("SharedRoot", func() {
			dsl.Meta("struct:pkg:path", "shared/types")
			dsl.Attribute("child", located)
		})
		dsl.Service("alpha", func() {
			dsl.Method("child", func() { dsl.Payload(located) })
		})
	})
	second := codegen.RunDSL(t, func() {
		dsl.API("second", func() {})
		local = dsl.Type("Child", func() { dsl.Attribute("value", dsl.String) })
		dsl.Service("beta", func() {
			dsl.Method("child", func() { dsl.Payload(local) })
		})
	})
	require.NotSame(t, located.Origin(), local.Origin())
	generation, plans := inheritedPlacementPlans(t, []*expr.RootExpr{first, second})
	for original, declaration := range generation.UserTypes() {
		switch original.Origin() {
		case located.Origin():
			require.Equal(t, "generated.local/gen/shared/types", declaration.PackagePath())
			attribute, layout, err := plans[0].UserTypeLayout(local, declaration)
			require.ErrorContains(t, err, "not a registered pair")
			require.Nil(t, attribute)
			require.Nil(t, layout)
		case local.Origin():
			require.Equal(t, "generated.local/gen/beta", declaration.PackagePath())
		}
	}
}

func inheritedPlacementInputs(t *testing.T, roots []*expr.RootExpr) (*codegen.Generation, []PlanInput) {
	t.Helper()
	evaluated := make([]eval.Root, len(roots))
	inputs := make([]PlanInput, len(roots))
	for index, root := range roots {
		evaluated[index] = root
		inputs[index] = PlanInput{Root: root, Examples: expr.NewExampleGenerator(root.API.RandomizerFactory)}
	}
	return mustTestGeneration(t, "generated.local/gen", evaluated), inputs
}

func inheritedPlacementPlans(t *testing.T, roots []*expr.RootExpr) (*codegen.Generation, []*Plan) {
	t.Helper()
	generation, inputs := inheritedPlacementInputs(t, roots)
	plans, err := NewPlans(generation, inputs...)
	require.NoError(t, err)
	return generation, plans
}

func inheritedPlacementRoots(t *testing.T, reverseServices bool) ([]*expr.RootExpr, schema.Types) {
	t.Helper()
	var values schema.Types
	var packet, local expr.UserType
	first := codegen.RunDSL(t, func() {
		dsl.API("first", func() {})
		values = schema.Define()
		packet = dsl.Type("Packet", func() {
			dsl.Meta("struct:pkg:path", "shared/types")
			dsl.Field(1, "name", dsl.String)
			dsl.Field(2, "label", values.Text)
			dsl.Field(3, "words", dsl.ArrayOf(dsl.String))
			dsl.Field(4, "table", dsl.MapOf(dsl.String, dsl.ArrayOf(dsl.String)))
			dsl.Field(5, "entries", dsl.ArrayOf(values.Entry))
			dsl.Required("name", "entries")
			dsl.ConvertTo(external.Envelope{})
			dsl.CreateFrom(external.Envelope{})
		})
		local = dsl.Type("LocalValue", func() {
			dsl.Field(1, "child", values.Child)
			dsl.Field(2, "children", dsl.ArrayOf(values.Child))
			dsl.Field(3, "by_name", dsl.MapOf(values.Text, values.Child))
		})
		names := []string{"alpha", "gamma"}
		if reverseServices {
			slices.Reverse(names)
		}
		for _, name := range names {
			inheritedPlacementService(name, packet, local, values.Child)
		}
	})
	second := codegen.RunDSL(t, func() {
		dsl.API("second", func() {})
		dsl.Type("AnotherRoot", func() {
			dsl.Meta("struct:pkg:path", "shared/types")
			dsl.Attribute("entry", values.Entry)
		})
		inheritedPlacementService("beta", packet, local, values.Child)
	})
	return []*expr.RootExpr{first, second}, values
}

func inheritedPlacementService(name string, packet, local, child expr.UserType) {
	dsl.Service(name, func() {
		for _, value := range []struct {
			name string
			typ  expr.UserType
		}{{"exchange", packet}, {"local", local}, {"child", child}} {
			dsl.Method(value.name, func() {
				dsl.Payload(value.typ)
				dsl.Result(value.typ)
			})
		}
	})
}

const inheritedPlacementRoundTrip = `package types
import (
	"reflect"
	"testing"
	external "goa.design/goa/v3/codegen/service/testdata/external-union"
	nested "goa.design/goa/v3/codegen/service/testdata/nested-alpha"
)
func TestInheritedExternalRoundTrip(t *testing.T) {
	empty := external.Text_Value("")
	var text, child, words, table external.Choice
	text.SetText("")
	child.SetChild(&external.Detail{
		Value: "child", Optional: &empty,
		Next: &external.Detail{Value: "next"}, Leaf: &nested.Child{Value: "leaf"},
	})
	words.SetWords(external.Text_List{"", "word"})
	table.SetTable(external.Text_Table{"empty": {}, "nil": nil, "value": {"word"}})
	input := &external.Envelope{
		Name: "packet", Label: &empty,
		Words: external.Text_List{"word"},
		Table: external.Text_Table{"value": {"word"}},
		Entries: []*external.Entry{
			{Choice: text, Blob: external.Byte_Data{0, 255}},
			{Choice: child}, {Choice: words}, {Choice: table},
		},
	}
	var value Packet
	value.CreateFromEnvelope(input)
	output := value.ConvertToEnvelope()
	if !reflect.DeepEqual(input, output) {
		t.Fatalf("external graph changed: %#v", output)
	}
	var again Packet
	again.CreateFromEnvelope(output)
	if !reflect.DeepEqual(value, again) {
		t.Fatal("shared graph changed on a second round trip")
	}
}
`
