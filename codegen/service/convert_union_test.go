// These tests exercise the existing external-conversion planner with private
// union storage, exact public method types and shared generated receivers.
package service

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	externalunion "goa.design/goa/v3/codegen/service/testdata/external-union"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestExternalUnionConversions(t *testing.T) {
	var baseline string
	for _, test := range []struct {
		name    string
		reverse bool
		mutate  bool
	}{
		{name: "forward roots"},
		{name: "reverse roots", reverse: true},
		{name: "retained plan", mutate: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			roots := externalUnionRoots(t)
			if test.reverse {
				slices.Reverse(roots)
			}
			evaluated := make([]eval.Root, len(roots))
			inputs := make([]PlanInput, len(roots))
			for i, root := range roots {
				evaluated[i] = root
				inputs[i] = PlanInput{Root: root, Examples: expr.NewExampleGenerator(root.API.RandomizerFactory)}
			}
			generation := mustTestGeneration(t, "generated.local/gen", evaluated)
			plans, err := NewPlans(generation, inputs...)
			require.NoError(t, err)
			if test.mutate {
				packet := roots[0].Conversions[0].User
				entry := expr.AsArray(packet.Attribute().Find("entries").Type).ElemType.Type
				choice := expr.AsUnion(expr.AsObject(entry).Attribute("choice").Type)
				choice.Values[0].Name = "changed"
				choice.Values[0].Attribute.Type = expr.Int
				for _, root := range roots {
					for _, mapping := range append(root.Conversions, root.Creations...) {
						mapping.User = nil
						mapping.External = struct{}{}
					}
					root.Conversions = nil
					root.Creations = nil
				}
			}
			require.NoError(t, generation.Freeze())
			for _, plan := range plans {
				require.NoError(t, plan.Link())
			}
			files, err := Files(plans...)
			require.NoError(t, err)
			path := filepath.Join(codegen.Gendir, "shared", "types", "convert.go")
			require.Len(t, filesAtPath(files, path), 1)
			source := renderSingleFileAtPath(t, files, path)
			require.Contains(t, source, "Text_Value")
			require.Contains(t, source, "Text_List")
			require.Contains(t, source, "Text_Table")
			require.Contains(t, source, "Byte_Data")
			require.NotContains(t, source, ".TextValue")
			require.NotContains(t, source, "ChoiceBranchText")
			require.Contains(t, source, "codegen/service/testdata/external-union")
			require.Contains(t, source, "codegen/service/testdata/nested-alpha")
			require.Contains(t, source, "ConvertToEnvelope")
			require.Contains(t, source, "CreateFromEnvelope")
			if baseline == "" {
				baseline = source
			} else {
				require.Equal(t, baseline, source)
			}
			compileGeneratedServiceFilesWith(t, files, map[string]string{
				filepath.Join(codegen.Gendir, "shared", "types", "conversion_test.go"): externalUnionRoundTripTest,
			})
		})
	}
}

func TestExternalUnionPlanningRejectsIncompatibleMethods(t *testing.T) {
	tests := []struct {
		name     string
		external any
		textName string
		message  string
	}{
		{"kind", externalunion.KindEnvelope{}, "text", "requires Kind()"},
		{"getter", externalunion.GetterEnvelope{}, "text", "requires AsText()"},
		{"setter", externalunion.SetterEnvelope{}, "text", ".SetText("},
		{"value setter", externalunion.ValueSetterEnvelope{}, "text", "pointer receiver for SetText"},
		{"extra branch", externalunion.ExtraEnvelope{}, "text", "has no authored branch"},
		{"missing branch", externalunion.ChoiceEnvelope{}, "renamed", "requires AsRenamed()"},
		{"scalar pointer", externalunion.PointerTextEnvelope{}, "text", "incompatible pointer form"},
		{"object value", externalunion.ValueChildEnvelope{}, "text", "incompatible pointer form"},
	}
	for _, test := range tests {
		for _, create := range []bool{false, true} {
			direction := "ConvertTo"
			if create {
				direction = "CreateFrom"
			}
			t.Run(test.name+"/"+direction, func(t *testing.T) {
				root := codegen.RunDSL(t, func() {
					leaf := dsl.Type("Leaf", func() {
						dsl.Attribute("value", dsl.String)
						dsl.Required("value")
					})
					mapped := dsl.Type("Mapped", func() {
						if create {
							dsl.CreateFrom(test.external)
						} else {
							dsl.ConvertTo(test.external)
						}
						dsl.OneOf("choice", func() {
							dsl.TypeName("Selection")
							dsl.Attribute(test.textName, dsl.String)
							dsl.Attribute("child", leaf)
							dsl.Attribute("words", dsl.ArrayOf(dsl.String))
							dsl.Attribute("table", dsl.MapOf(dsl.String, dsl.ArrayOf(dsl.String)))
						})
						dsl.Required("choice")
					})
					dsl.Service("Values", func() {
						dsl.Method("Use", func() {
							dsl.Payload(mapped)
						})
					})
				})
				generation := mustTestGeneration(t, "generated.local/gen", []eval.Root{root})
				_, err := NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
				require.ErrorContains(t, err, test.message)
				require.ErrorContains(t, err, ".Choice")
			})
		}
	}
}

func TestExternalUnionPlanningRejectsPointerLocations(t *testing.T) {
	tests := []struct {
		name     string
		external any
		location string
		path     string
	}{
		{"double pointer", externalunion.DoublePointerEnvelope{}, "field", "<value>.Choice"},
		{"triple pointer", externalunion.TriplePointerEnvelope{}, "field", "<value>.Choice"},
		{"slice element", externalunion.SlicePointerEnvelope{}, "slice", "<value>.Choices[0]"},
		{"map value", externalunion.MapPointerEnvelope{}, "map", "<value>.Choices.value"},
		{"map key", externalunion.MapKeyPointerEnvelope{}, "key", "<value>.Choices.key"},
		{"nested collection", externalunion.NestedPointerEnvelope{}, "nested", "<value>.Choices[0].value"},
	}
	for _, test := range tests {
		for _, create := range []bool{false, true} {
			direction := "ConvertTo"
			if create {
				direction = "CreateFrom"
			}
			t.Run(test.name+"/"+direction, func(t *testing.T) {
				root := codegen.RunDSL(t, func() {
					leaf := dsl.Type("Leaf", func() {
						dsl.Attribute("value", dsl.String)
						dsl.Required("value")
					})
					branches := func() {
						externalUnionBranchDSL(leaf)
					}
					mapped := dsl.Type("Mapped", func() {
						if create {
							dsl.CreateFrom(test.external)
						} else {
							dsl.ConvertTo(test.external)
						}
						// Keep this rejection test independent of random map-key examples.
						dsl.Example(map[string]any{})
						choice := &expr.Union{TypeName: "Selection"}
						switch test.location {
						case "field":
							dsl.OneOf("choice", branches)
						case "slice":
							dsl.Attribute("choices", dsl.ArrayOf(choice, branches))
						case "map":
							dsl.Attribute("choices", dsl.MapOf(dsl.String, choice, func() {
								dsl.Elem(branches)
							}))
						case "key":
							dsl.Attribute("choices", dsl.MapOf(choice, dsl.String, func() {
								dsl.Key(branches)
							}))
						case "nested":
							dsl.Attribute("choices", dsl.ArrayOf(dsl.MapOf(dsl.String, choice, func() {
								dsl.Elem(branches)
							})))
						}
					})
					dsl.Service("Values", func() {
						dsl.Method("Use", func() {
							dsl.Payload(mapped)
						})
					})
				})
				generation := mustTestGeneration(t, "generated.local/gen", []eval.Root{root})
				_, err := NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
				require.ErrorContains(t, err, test.path)
				require.ErrorContains(t, err, "pointer to union")
			})
		}
	}
}

// externalUnionBranchDSL gives each fixture occurrence the public Choice
// branches, including named external collections with named keys and elements.
func externalUnionBranchDSL(leaf expr.UserType) {
	dsl.Attribute("text", dsl.String)
	dsl.Attribute("child", leaf)
	dsl.Attribute("words", dsl.ArrayOf(dsl.String))
	dsl.Attribute("table", dsl.MapOf(dsl.String, dsl.ArrayOf(dsl.String)))
}

// externalUnionRoots assigns opposite conversions for one receiver to two
// roots. The nested leaf and union come from packages with the same Go name.
func externalUnionRoots(t *testing.T) []*expr.RootExpr {
	t.Helper()
	var packet expr.UserType
	first := codegen.RunDSL(t, func() {
		text := dsl.Type("TextValue", dsl.String, func() {
			dsl.Meta("struct:pkg:path", "shared/types")
		})
		nested := dsl.Type("NestedLeaf", func() {
			dsl.Meta("struct:pkg:path", "shared/types")
			dsl.Attribute("value", dsl.String)
			dsl.Required("value")
		})
		leaf := dsl.Type("Leaf", func() {
			dsl.Meta("struct:pkg:path", "shared/types")
			dsl.Attribute("value", text)
			dsl.Attribute("optional", text)
			dsl.Attribute("next", "Leaf")
			dsl.Attribute("leaf", nested)
			dsl.Required("value")
		})
		branches := func() {
			externalUnionBranchDSL(leaf)
		}
		entry := dsl.Type("Entry", func() {
			dsl.Meta("struct:pkg:path", "shared/types")
			dsl.OneOf("choice", func() {
				dsl.TypeName("Selection")
				branches()
			})
			dsl.OneOf("optional", func() {
				dsl.TypeName("OptionalSelection")
				dsl.Attribute("table", dsl.MapOf(dsl.String, dsl.ArrayOf(dsl.String)))
				dsl.Attribute("words", dsl.ArrayOf(dsl.String))
				dsl.Attribute("child", leaf)
				dsl.Attribute("text", dsl.String)
			})
			dsl.Attribute("choices", dsl.ArrayOf(&expr.Union{TypeName: "ListSelection"}, branches))
			dsl.Attribute("by_name", dsl.MapOf(dsl.String, &expr.Union{TypeName: "MapSelection"}, func() {
				dsl.Elem(branches)
			}))
			dsl.OneOf("wrapped", func() {
				dsl.TypeName("WrappedSelection")
				dsl.Attribute("choice", &expr.Union{TypeName: "NestedSelection"}, branches)
			})
			dsl.Attribute("blob", dsl.Bytes)
			dsl.Required("choice")
		})
		packet = dsl.Type("Packet", func() {
			dsl.Meta("struct:pkg:path", "shared/types")
			dsl.ConvertTo(externalunion.Envelope{})
			dsl.Attribute("name", dsl.String)
			dsl.Attribute("label", text)
			dsl.Attribute("words", dsl.ArrayOf(dsl.String))
			dsl.Attribute("table", dsl.MapOf(dsl.String, dsl.ArrayOf(dsl.String)))
			dsl.Attribute("entries", dsl.ArrayOf(entry))
			dsl.Required("name", "entries")
		})
		dsl.Service("Alpha", func() {
			dsl.Method("Use", func() {
				dsl.Payload(packet)
			})
		})
	})
	second := codegen.RunDSL(t, func() {
		dsl.Service("Beta", func() {
			dsl.Method("Use", func() {
				dsl.Payload(packet)
			})
		})
	})
	second.Creations = append(second.Creations, &expr.TypeMap{
		User: packet, External: externalunion.Envelope{},
	})
	return []*expr.RootExpr{first, second}
}

const externalUnionRoundTripTest = `package types

import (
	"reflect"
	"testing"

	external "goa.design/goa/v3/codegen/service/testdata/external-union"
	leaf "goa.design/goa/v3/codegen/service/testdata/nested-alpha"
)

func TestExternalUnionRoundTrip(t *testing.T) {
	for _, text := range []string{"", "a.b/[key]"} {
		t.Run(text, func(t *testing.T) {
			empty := external.Text_Value("")
			label := external.Text_Value("named label")
			for _, optionalLabel := range []*external.Text_Value{nil, &empty, &label} {
				var first, second, optional, nilChild, words, nilWords, emptyWords, table, nilTable, emptyTable external.Choice
				first.SetText(external.Text_Value(text))
				second.SetChild(&external.Detail{
					Value:    "nested",
					Optional: &empty,
					Next:     &external.Detail{Value: "next", Optional: &label},
					Leaf:     &leaf.Child{Value: "other package"},
				})
				optional.SetText(external.Text_Value("optional"))
				nilChild.SetChild(nil)
				words.SetWords(external.Text_List{"", "one", external.Text_Value(text)})
				nilWords.SetWords(nil)
				emptyWords.SetWords(external.Text_List{})
				table.SetTable(external.Text_Table{
					"":      nil,
					"empty": {},
					"words": {"first", external.Text_Value(text)},
				})
				nilTable.SetTable(nil)
				emptyTable.SetTable(external.Text_Table{})
				var wrapped, nilWrapped external.ChoiceBox
				wrapped.SetChoice(&words)
				nilWrapped.SetChoice(nil)
				input := &external.Envelope{
					Name:  "sample",
					Label: optionalLabel,
					Words: external.Text_List{"", "ordinary"},
					Table: external.Text_Table{"nil": nil, "empty": {}, "words": {"ordinary"}},
					Entries: []*external.Entry{
						{Choice: first, Blob: external.Byte_Data{0, 255, 34, 92}},
						{Choice: second, Optional: &optional},
						{Choice: nilChild, Optional: &second},
						{
							Choice:  words,
							Wrapped: wrapped,
							Choices: []external.Choice{first, second, words, table},
							ByName: map[external.Text_Value]external.Choice{
								"":      table,
								"nil":   nilWords,
								"empty": emptyTable,
							},
						},
						{Choice: nilWords, Wrapped: nilWrapped},
						{Choice: emptyWords},
						{Choice: table},
						{Choice: nilTable},
						{Choice: emptyTable},
					},
				}
				var converted Packet
				converted.CreateFromEnvelope(input)
				actual, selected := converted.Entries[0].Choice.AsText()
				if !selected || string(actual) != text {
					t.Errorf("text branch = %q, selected=%v", actual, selected)
				}
				child, selected := converted.Entries[1].Choice.AsChild()
				if !selected || child == nil || child.Value != "nested" {
					t.Errorf("child branch = %#v, selected=%v", child, selected)
				}
				if converted.Entries[0].Optional.Kind() != "" {
					t.Error("nil external optional union became a selected local branch")
				}
				if (converted.Label == nil) != (optionalLabel == nil) {
					t.Error("named scalar conversion changed optional presence")
				}
				if optionalLabel != nil && converted.Label != nil && string(*converted.Label) != string(*optionalLabel) {
					t.Error("named scalar conversion changed its value")
				}
				output := converted.ConvertToEnvelope()
				if !reflect.DeepEqual(input, output) {
					t.Errorf("external round trip changed values:\ninput: %#v\noutput: %#v", input, output)
				}
				var repeated Packet
				repeated.CreateFromEnvelope(output)
				if !reflect.DeepEqual(converted, repeated) {
					t.Error("local round trip changed values")
				}
			}
		})
	}
}
`
