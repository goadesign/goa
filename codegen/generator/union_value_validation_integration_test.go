// These tests generate ordinary union declarations and execute planned and
// legacy validators against selected branches, optional fields, and aliases.
package generator

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"goa.design/goa/v3/codegen"
	d "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestGenerateUnionValueValidation(t *testing.T) {
	root := codegen.RunDSL(t, unionValueValidationDSL)
	type operation struct {
		name        string
		attribute   *expr.AttributeExpr
		declaration *codegen.NameDeclaration
		layout      *codegen.GoTypePlan
		validation  *codegen.ValidationPlan
	}
	operations := []*operation{
		{name: "CheckChoicePlanned", attribute: &expr.AttributeExpr{Type: root.UserType("Choice")}},
		{name: "CheckDerivedPlanned", attribute: &expr.AttributeExpr{Type: root.UserType("Derived")}},
		{name: "CheckHolderPlanned", attribute: &expr.AttributeExpr{Type: root.UserType("Holder")}},
	}
	var pkg *codegen.GeneratedPackage
	var imports *codegen.GeneratedImportPlan
	registry := newDefaultRegistry()
	registry.registerPlugin("union-value-validation", "gen", pluginNormal, func() Plugin {
		return Plugin{
			Plan: func(plan *Plan) error {
				pkg = plan.Generation().Package("generated.local/gen/catalog")
				imports = codegen.NewGeneratedImportPlan(pkg)
				for _, op := range operations {
					op.declaration = codegen.NewExactName(codegen.NameFunction, op.name)
					if err := pkg.DeclareName(op.declaration); err != nil {
						return err
					}
				}
				for _, name := range []string{"CheckChoiceLegacy", "CheckDerivedLegacy"} {
					if err := pkg.DeclareName(codegen.NewExactName(codegen.NameFunction, name)); err != nil {
						return err
					}
				}
				bind := func(request codegen.GoTypeBindingRequest) (codegen.GoTypeBinding, error) {
					binding := codegen.GoTypeBinding{Owner: pkg.ImportPath()}
					var err error
					if request.Kind == codegen.GoUnion {
						binding.Union, err = pkg.Union(request.Attribute)
					} else {
						binding.Type, err = pkg.Type(request.Attribute.Type.(expr.UserType))
					}
					return binding, err
				}
				for _, op := range operations {
					var err error
					op.layout, err = codegen.PlanGoType(op.attribute, codegen.GoTypePlanOptions{
						Owner:            pkg.ImportPath(),
						Bind:             bind,
						RetainNamedValue: true,
						Policy:           codegen.GoLayoutPolicy{UseDefault: true, SumType: true},
					})
					if err != nil {
						return err
					}
					op.validation, err = codegen.NewValidationPlan(op.attribute, op.layout, codegen.ValidationPlanOptions{
						Required: true,
						Bind: func(request codegen.ValidatorBindingRequest) (*codegen.NameDeclaration, error) {
							for _, nested := range operations {
								if nested.attribute.Type == request.Attribute.Type {
									return nested.declaration, nil
								}
							}
							return nil, fmt.Errorf("unexpected nested validator for %s", request.Attribute.Type.Name())
						},
					})
					if err != nil {
						return err
					}
					for _, spec := range op.validation.ImportPreferences() {
						if spec.Path == pkg.ImportPath() {
							continue
						}
						if err := imports.Require(codegen.NewImport(spec.Name, spec.Path)); err != nil {
							return err
						}
					}
				}
				return nil
			},
			Generate: func(_ *Plan, files []*codegen.File) ([]*codegen.File, error) {
				if err := imports.Link(); err != nil {
					return nil, err
				}
				functions := make([]map[string]string, 0, len(operations)+1)
				for _, op := range operations {
					layout := op.layout.Link(pkg.ImportPath(), pkg.ImportName)
					validation, err := op.validation.Link(layout)
					if err != nil {
						return nil, err
					}
					functions = append(functions, map[string]string{
						"Name":      op.name,
						"Parameter": layout.Ref(),
						"Body":      validation.Render("value", "body"),
					})
				}
				for index, name := range []string{"CheckChoiceLegacy", "CheckDerivedLegacy"} {
					op := operations[index]
					layout := op.layout.Link(pkg.ImportPath(), pkg.ImportName)
					context, err := codegen.NewAttributeContext(false, false, true, "", codegen.NewNameScope()).WithGoTypeLayout(layout)
					if err != nil {
						return nil, err
					}
					functions = append(functions, map[string]string{
						"Name":      name,
						"Parameter": layout.Ref(),
						"Body":      codegen.ValidationCode(op.attribute, nil, context, true, false, false, "value"),
					})
				}
				return append(files, &codegen.File{
					Path: filepath.Join(pkg.OutputDirectory(), "value_validation.go"),
					SectionTemplates: []*codegen.SectionTemplate{
						codegen.Header("Union value validation", "catalog", imports.Imports()),
						{Name: "validators", Source: "{{range .}}func {{.Name}}(value {{.Parameter}}) (err error) {\n{{.Body}}\nreturn err\n}\n{{end}}", Data: functions},
					},
				}), nil
			},
		}
	})
	run, err := newGenerationRun("gen", registry)
	require.NoError(t, err)
	result, err := run.execute("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	data := make(map[string]string)
	for _, name := range []string{"Choice", "Derived", "Holder", "Empty"} {
		declaration, err := pkg.Type(root.UserType(name))
		require.NoError(t, err)
		data[name] = declaration.Name()
	}
	union, err := pkg.Union(root.UserType("Choice").Attribute())
	require.NoError(t, err)
	data["Union"] = union.Name()
	for _, name := range []string{"empty", "words", "counts", "text"} {
		branch, err := pkg.UnionBranch(root.UserType("Choice").Attribute(), name)
		require.NoError(t, err)
		data[name] = branch.Constructor()
	}
	directory := t.TempDir()
	writeGeneratedModule(t, directory, "generated.local")
	for _, file := range result.files {
		_, err := file.Render(directory)
		require.NoError(t, err)
	}
	testFile := &codegen.File{
		Path: filepath.Join(pkg.OutputDirectory(), "value_validation_test.go"),
		SectionTemplates: []*codegen.SectionTemplate{
			codegen.Header("Union value validation tests", "catalog", []*codegen.ImportSpec{codegen.SimpleImport("testing"), codegen.SimpleImport("reflect")}),
			{Name: "tests", Source: unionValueValidationTests, Data: data},
		},
	}
	_, err = testFile.Render(directory)
	require.NoError(t, err)
	runGeneratedTests(t, directory)
}

func unionValueValidationDSL() {
	empty := d.Type("Empty", func() {
	})
	choice := d.Type("Choice", &expr.Union{TypeName: "ChoiceValue"}, func() {
		d.Attribute("empty", empty)
		d.Attribute("words", d.ArrayOf(d.String))
		d.Attribute("counts", d.MapOf(d.String, d.Int))
		d.Attribute("text", d.String, func() {
			d.MinLength(1)
		})
	})
	derived := d.Type("Derived", choice)
	holder := d.Type("Holder", func() {
		d.Attribute("required", choice)
		d.Attribute("optional", choice)
		d.Attribute("aliased", derived)
		d.Attribute("choices", d.ArrayOf(choice))
		d.Required("required", "choices")
	})
	d.Service("catalog", func() {
		d.Method("inspect", func() {
			d.Payload(holder)
		})
	})
}

const unionValueValidationTests = `
func TestSelectedUnionValidation(t *testing.T) {
 for _, test := range []struct {
  name string
  value {{.Choice}}
  invalid bool
 }{
  {"missing kind", {{.Choice}}{}, true},
  {"unknown kind", {{.Choice}}({{.Union}}{kind: "other"}), true},
  {"nil object", {{.Choice}}({{.empty}}(nil)), true},
  {"empty object", {{.Choice}}({{.empty}}(&{{.Empty}}{})), false},
  {"empty array", {{.Choice}}({{.words}}([]string{})), false},
  {"empty map", {{.Choice}}({{.counts}}(map[string]int{})), false},
  {"text", {{.Choice}}({{.text}}("hello")), false},
  {"constrained text", {{.Choice}}({{.text}}("")), true},
 } {
  t.Run(test.name, func(t *testing.T) {
   before := test.value
   for _, validate := range []func(*{{.Choice}}) error{CheckChoicePlanned, CheckChoiceLegacy} {
    if err := validate(&test.value); (err != nil) != test.invalid {
     t.Errorf("invalid=%t want=%t: %v", err != nil, test.invalid, err)
    }
   }
   derived := {{.Derived}}(test.value)
   for _, validate := range []func(*{{.Derived}}) error{CheckDerivedPlanned, CheckDerivedLegacy} {
    if err := validate(&derived); (err != nil) != test.invalid {
     t.Errorf("derived invalid=%t want=%t: %v", err != nil, test.invalid, err)
    }
   }
   if !reflect.DeepEqual(before, test.value) {
    t.Error("validation changed caller value")
   }
  })
 }
}

func TestUnionCollectionNilSemantics(t *testing.T) {
 for _, selected := range []{{.Union}}{ {{.words}}(nil), {{.counts}}(nil) } {
  value := {{.Choice}}(selected)
  wantInvalid := selected.Validate() != nil
  for _, validate := range []func(*{{.Choice}}) error{CheckChoicePlanned, CheckChoiceLegacy} {
   if err := validate(&value); (err != nil) != wantInvalid {
    t.Errorf("collection invalid=%t want=%t: %v", err != nil, wantInvalid, err)
   }
  }
 }
}

func TestOptionalUnionAndCollections(t *testing.T) {
 selected := {{.Choice}}({{.text}}("hello"))
 value := &{{.Holder}}{Required: selected, Choices: []{{.Choice}}{}}
 if err := CheckHolderPlanned(value); err != nil {
  t.Errorf("optional absent unions or empty collection rejected: %v", err)
 }
 value.Optional = {{.Choice}}({{.empty}}(nil))
 if err := CheckHolderPlanned(value); err == nil {
  t.Error("accepted present optional nil branch")
 }
 value.Optional = selected
 value.Aliased = {{.Derived}}({{.empty}}(nil))
 if err := CheckHolderPlanned(value); err == nil {
  t.Error("accepted present aliased nil branch")
 }
 value.Aliased = {{.Derived}}(selected)
 value.Choices = []{{.Choice}}{ {{.Choice}}({{.empty}}(nil)) }
 if err := CheckHolderPlanned(value); err == nil {
  t.Error("accepted collection element nil branch")
 }
}
`
