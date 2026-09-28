// These tests check names actually needed by external conversions. Readable
// private values remain usable when the generated code need not name their type.
package service

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	external "goa.design/goa/v3/codegen/service/testdata/external-accessibility"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestExternalConversionRejectsInaccessibleTargetNames(t *testing.T) {
	for _, test := range []struct {
		field string
		path  string
	}{
		{"scalar", ":Scalar"},
		{"words", ":Words[0]"},
		{"keys", ":Keys.key"},
		{"values", ":Values.value"},
		{"choice", ".words[0]"},
	} {
		t.Run(test.field, func(t *testing.T) {
			root := externalAccessibilityRoot(t, test.field, external.Envelope{}, nil)
			generation := mustTestGeneration(t, "generated.local/gen", []eval.Root{root})
			_, err := NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
			require.ErrorContains(t, err, "ConvertTo cannot name unexported external type")
			require.ErrorContains(t, err, "external-accessibility.word at <value>.")
			require.ErrorContains(t, err, test.path)
			require.ErrorContains(t, err, `from generated package "generated.local/gen/values"`)
		})
	}
}

func TestExternalConversionChecksSignatureNames(t *testing.T) {
	for _, test := range []struct {
		name     string
		external any
		root     bool
		typeName string
	}{
		{"root", external.PrivateRoot(), true, "privateRoot"},
		{"helper", external.HiddenEnvelope{}, false, "hiddenRecord"},
	} {
		for _, create := range []bool{false, true} {
			direction := "ConvertTo"
			if create {
				direction = "CreateFrom"
			}
			t.Run(test.name+"/"+direction, func(t *testing.T) {
				root := codegen.RunDSL(t, func() {
					fields := func() {
						dsl.Attribute("scalar", dsl.String)
						dsl.Required("scalar")
					}
					nested := dsl.Type("Nested", fields)
					mapped := dsl.Type("Mapped", func() {
						if create {
							dsl.CreateFrom(test.external)
						} else {
							dsl.ConvertTo(test.external)
						}
						if test.root {
							fields()
						} else {
							dsl.Attribute("nested", nested)
							dsl.Required("nested")
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
				require.ErrorContains(t, err, direction+" cannot name unexported external type")
				require.ErrorContains(t, err, "external-accessibility."+test.typeName)
				if create && !test.root {
					require.ErrorContains(t, err, "<helper parameter>")
				}
			})
		}
	}
}

func TestExternalConversionReadsPrivateValues(t *testing.T) {
	for _, field := range []string{"scalar", "words", "keys", "values", "choice"} {
		for _, source := range []struct {
			name     string
			external any
		}{
			{"Envelope", external.Envelope{}},
			{"PublicEnvelope", external.PublicEnvelope{}},
		} {
			t.Run(field+"/"+source.name, func(t *testing.T) {
				root := externalAccessibilityRoot(t, field, external.PublicEnvelope{}, source.external)
				generation := mustTestGeneration(t, "generated.local/gen", []eval.Root{root})
				plan, err := NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
				require.NoError(t, err)
				require.NoError(t, generation.Freeze())
				require.NoError(t, plan.Link())
				files, err := Files(plan)
				require.NoError(t, err)
				comparison := fmt.Sprintf(
					"if !reflect.DeepEqual(want.Nested.%s, got.Nested.%s) { t.Fatalf(\"changed value or presence: %%#v\", got.Nested) }",
					codegen.Goify(field, true), codegen.Goify(field, true),
				)
				if field == "choice" {
					comparison = `wantWords, wantSelected := want.Choice.AsWords()
	gotWords, gotSelected := got.Choice.AsWords()
	if wantSelected != gotSelected || !reflect.DeepEqual(wantWords, gotWords) {
		t.Fatalf("changed union value or presence: %#v, selected=%v", gotWords, gotSelected)
	}`
				}
				proof := fmt.Sprintf(externalAccessibilityRuntimeTest, source.name, source.name, comparison)
				compileGeneratedServiceFilesWith(t, files, map[string]string{
					filepath.Join(codegen.Gendir, "values", "accessibility_test.go"): proof,
				})
			})
		}
	}
}

// externalAccessibilityRoot selects one nested field or union branch. Other
// external fields deliberately remain undesigned and must not affect access.
func externalAccessibilityRoot(t *testing.T, field string, convertTo, createFrom any) *expr.RootExpr {
	t.Helper()
	return codegen.RunDSL(t, func() {
		var nested expr.UserType
		if field != "choice" {
			var fieldType expr.DataType = dsl.String
			switch field {
			case "words":
				fieldType = dsl.Type("Words", dsl.ArrayOf(dsl.String))
			case "keys", "values":
				fieldType = dsl.Type("Table", dsl.MapOf(dsl.String, dsl.String))
			}
			nested = dsl.Type("Nested", func() {
				dsl.Attribute(field, fieldType)
				if field == "scalar" {
					dsl.Required(field)
				}
			})
		}
		mapped := dsl.Type("Mapped", func() {
			if convertTo != nil {
				dsl.ConvertTo(convertTo)
			}
			if createFrom != nil {
				dsl.CreateFrom(createFrom)
			}
			if field == "choice" {
				dsl.OneOf("choice", func() {
					dsl.TypeName("Selection")
					dsl.Attribute("words", dsl.ArrayOf(dsl.String))
				})
				dsl.Required("choice")
			} else {
				dsl.Attribute("nested", nested)
				dsl.Required("nested")
			}
		})
		dsl.Service("Values", func() {
			dsl.Method("Use", func() {
				dsl.Payload(mapped)
			})
		})
	})
}

const externalAccessibilityRuntimeTest = `package values

import (
	"reflect"
	"testing"

	external "goa.design/goa/v3/codegen/service/testdata/external-accessibility"
)

func TestExternalValueRoundTrip(t *testing.T) {
	for _, state := range []int{0, 1, 2} {
		input := external.New%s(state)
		want := external.NewPublicEnvelope(state)
		var value Mapped
		value.CreateFrom%s(input)
		got := value.ConvertToPublicEnvelope()
		%s
	}
}
`
