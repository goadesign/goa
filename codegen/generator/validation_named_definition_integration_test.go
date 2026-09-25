// These tests compile validation against ordinary Goa declarations. A named
// union keeps its parameter type while validation uses the union's own methods.
package generator

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	d "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

type (
	namedValidationFixture struct {
		pkg        *codegen.GeneratedPackage
		imports    *codegen.GeneratedImportPlan
		operations []namedValidationOperation
		optional   *codegen.GoTypePlan
	}

	namedValidationOperation struct {
		name       string
		parameter  *codegen.GoTypePlan
		pointer    bool
		layout     *codegen.GoTypePlan
		validation *codegen.ValidationPlan
	}

	namedValidationFunction struct {
		Name      string
		Parameter string
		Body      string
	}
)

func TestGenerateNamedDefinitionValidation(t *testing.T) {
	root := codegen.RunDSL(t, namedDefinitionValidationDSL)
	fixture := new(namedValidationFixture)
	registry := newDefaultRegistry()
	registry.registerPlugin("named-definition-validation", "gen", pluginNormal, func() Plugin {
		return Plugin{
			Plan: func(plan *Plan) error {
				return fixture.plan(plan.Generation(), root)
			},
			Generate: func(_ *Plan, files []*codegen.File) ([]*codegen.File, error) {
				if err := fixture.imports.Link(); err != nil {
					return nil, err
				}
				file, err := fixture.file()
				if err != nil {
					return nil, err
				}
				return append(files, file), nil
			},
		}
	})
	run, err := newGenerationRun("gen", registry)
	require.NoError(t, err)
	result, err := run.execute("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	left := result.plan.Generation().Package("generated.local/gen/left/types")
	right := result.plan.Generation().Package("generated.local/gen/right/types")
	require.NotEqual(t, fixture.pkg.ImportName(left.ImportPath()), fixture.pkg.ImportName(right.ImportPath()))
	require.Equal(t, []codegen.GoTypeImport{{Path: right.ImportPath()}},
		fixture.operations[0].parameter.ImportPreferences())
	require.Contains(t, fixture.operations[0].validation.ImportPreferences(), codegen.GoTypeImport{Path: left.ImportPath()})

	directory := t.TempDir()
	writeGeneratedModule(t, directory, "generated.local")
	for _, file := range result.files {
		_, err := file.Render(directory)
		require.NoError(t, err)
	}
	data := make(map[string]string)
	for _, name := range []string{"Base", "Derived", "LongBase", "LongLeaf", "ValueDerived", "Holder", "DefaultedHolder"} {
		owner := left
		if name == "Derived" || name == "LongLeaf" || name == "ValueDerived" {
			owner = right
		}
		declaration, err := owner.Type(root.UserType(name))
		require.NoError(t, err)
		data[name] = fixture.pkg.ImportName(owner.ImportPath()) + "." + declaration.Name()
	}
	for key, attribute := range map[string]*expr.AttributeExpr{
		"Choice":    root.UserType("Base").Attribute(),
		"Selection": expr.AsObject(root.UserType("Holder").Attribute().Type).Attribute("Selection"),
	} {
		for _, branchName := range []string{"text", "number"} {
			branch, err := left.UnionBranch(attribute, branchName)
			require.NoError(t, err)
			data[key+branchName] = fixture.pkg.ImportName(left.ImportPath()) + "." + branch.Constructor()
		}
	}
	testFile := &codegen.File{
		Path: filepath.Join(fixture.pkg.OutputDirectory(), "validation_test.go"),
		SectionTemplates: []*codegen.SectionTemplate{
			codegen.Header("Named definition validation tests", "checks", []*codegen.ImportSpec{
				codegen.SimpleImport("testing"), fixture.pkg.Import(left.ImportPath()), fixture.pkg.Import(right.ImportPath()),
			}),
			{Name: "tests", Source: namedDefinitionValidationTests, Data: data},
		},
	}
	_, err = testFile.Render(directory)
	require.NoError(t, err)

	// Rendering must use the retained declaration, fields, and constraints even
	// after the caller changes every original expression.
	before, err := os.ReadFile(filepath.Join(directory, fixture.pkg.OutputDirectory(), "validation.go"))
	require.NoError(t, err)
	for _, userType := range root.Types {
		userType.SetAttribute(&expr.AttributeExpr{Type: expr.Boolean})
	}
	file, err := fixture.file()
	require.NoError(t, err)
	repeatedDirectory := t.TempDir()
	_, err = file.Render(repeatedDirectory)
	require.NoError(t, err)
	after, err := os.ReadFile(filepath.Join(repeatedDirectory, file.Path))
	require.NoError(t, err)
	require.Equal(t, before, after)
	runGeneratedTests(t, directory)
}

// plan obtains the original declarations from the service generator and plans
// both named parameters and the exact layouts validated by their bodies.
func (f *namedValidationFixture) plan(generation *codegen.Generation, root *expr.RootExpr) error {
	var err error
	f.pkg, err = generation.ClaimPackage("generated.local/gen/checks")
	if err != nil {
		return err
	}
	f.imports = codegen.NewGeneratedImportPlan(f.pkg)
	bind := func(request codegen.GoTypeBindingRequest) (codegen.GoTypeBinding, error) {
		owner := request.InheritedOwner
		if location := codegen.UserTypeLocation(request.Attribute.Type); location != nil {
			owner = path.Join(generation.GenPkg(), location.RelImportPath)
		}
		pkg := generation.Package(owner)
		binding := codegen.GoTypeBinding{Owner: owner}
		var err error
		if request.Kind == codegen.GoUnion {
			binding.Union, err = pkg.Union(request.Attribute)
		} else {
			binding.Type, err = pkg.Type(request.Attribute.Type.(expr.UserType))
		}
		return binding, err
	}
	derived := &expr.AttributeExpr{Type: root.UserType("Derived")}
	holder := &expr.AttributeExpr{Type: root.UserType("Holder")}
	var derivedValidator *codegen.NameDeclaration
	for _, test := range []struct {
		name      string
		parameter *expr.AttributeExpr
		body      *expr.AttributeExpr
		pointer   bool
		nullable  bool
	}{
		{"ValidateDerived", derived, derived, true, false},
		{"ValidateDefinition", derived, root.UserType("Derived").Attribute(), true, false},
		{"ValidateBase", &expr.AttributeExpr{Type: root.UserType("Base")}, &expr.AttributeExpr{Type: root.UserType("Base")}, true, false},
		{"ValidateDirect", root.UserType("Base").Attribute(), root.UserType("Base").Attribute(), false, false},
		{"ValidateHolder", holder, root.UserType("Holder").Attribute(), true, false},
		{"ValidateOptional", root.UserType("Holder").Attribute(), root.UserType("Holder").Attribute(), true, true},
		{"ValidateLong", &expr.AttributeExpr{Type: root.UserType("LongLeaf")}, &expr.AttributeExpr{Type: root.UserType("LongLeaf")}, true, false},
		{"ValidateLongBase", &expr.AttributeExpr{Type: root.UserType("LongBase")}, &expr.AttributeExpr{Type: root.UserType("LongBase")}, true, false},
		{"ValidateValue", &expr.AttributeExpr{Type: root.UserType("ValueDerived")}, &expr.AttributeExpr{Type: root.UserType("ValueDerived")}, true, false},
		{"ValidateDefaults", &expr.AttributeExpr{Type: root.UserType("DefaultedHolder")}, root.UserType("DefaultedHolder").Attribute(), true, false},
		{"ValidateArray", &expr.AttributeExpr{Type: &expr.Array{ElemType: derived}}, &expr.AttributeExpr{Type: &expr.Array{ElemType: derived}}, false, false},
		{"ValidateMap", &expr.AttributeExpr{Type: &expr.Map{KeyType: &expr.AttributeExpr{Type: expr.String}, ElemType: derived}}, &expr.AttributeExpr{Type: &expr.Map{KeyType: &expr.AttributeExpr{Type: expr.String}, ElemType: derived}}, false, false},
	} {
		options := codegen.GoTypePlanOptions{
			Owner: "generated.local/gen/left/types", RetainNamedValue: true, Bind: bind,
			Policy: codegen.GoLayoutPolicy{UseDefault: true, SumType: true, Pointer: test.nullable, UnionPointer: test.nullable},
		}
		parameter, err := codegen.PlanGoType(test.parameter, options)
		if err != nil {
			return err
		}
		layout, err := codegen.PlanGoType(test.body, options)
		if err != nil {
			return err
		}
		validation, err := codegen.NewValidationPlan(test.body, layout, codegen.ValidationPlanOptions{
			Required: true,
			Bind: func(request codegen.ValidatorBindingRequest) (*codegen.NameDeclaration, error) {
				if request.Attribute.Type != derived.Type {
					return nil, fmt.Errorf("unexpected nested validation type %s", request.Attribute.Type.Name())
				}
				return derivedValidator, nil
			},
		})
		if err != nil {
			return err
		}
		f.operations = append(f.operations, namedValidationOperation{
			name: test.name, parameter: parameter, pointer: test.pointer, layout: layout, validation: validation,
		})
		if test.nullable {
			f.optional = parameter
		}
		declaration := codegen.NewExactName(codegen.NameFunction, test.name)
		if err := f.pkg.DeclareName(declaration); err != nil {
			return err
		}
		if test.name == "ValidateDerived" {
			derivedValidator = declaration
		}
		for _, spec := range append(parameter.ImportPreferences(), validation.ImportPreferences()...) {
			if spec.Path == "unicode/utf8" || spec.Path == codegen.GoaImport("").Path {
				err = f.imports.Require(codegen.NewImport(spec.Name, spec.Path))
			} else {
				// Both generated packages request the same alias. The output
				// package must choose distinct final names for every use.
				err = f.imports.AddGenerated(codegen.NewImport("gentypes", spec.Path))
			}
			if err != nil {
				return err
			}
		}
	}
	return f.pkg.DeclareName(codegen.NewExactName(codegen.NameType, "Optional"))
}

// file renders only planned types, validation operations, and linked imports.
func (f *namedValidationFixture) file() (*codegen.File, error) {
	functions := make([]namedValidationFunction, 0, len(f.operations))
	for _, operation := range f.operations {
		validation, err := operation.validation.Link(operation.layout.Link(f.pkg.ImportPath(), f.pkg.ImportName))
		if err != nil {
			return nil, err
		}
		parameter := operation.parameter.Link(f.pkg.ImportPath(), f.pkg.ImportName).RefWithPointer(operation.pointer)
		if operation.parameter.Kind() == codegen.GoArray || operation.parameter.Kind() == codegen.GoMap {
			parameter = operation.parameter.Link(f.pkg.ImportPath(), f.pkg.ImportName).Def()
		}
		if operation.name == "ValidateOptional" {
			parameter = "*Optional"
		}
		functions = append(functions, namedValidationFunction{
			Name: operation.name, Parameter: parameter, Body: validation.Render("value", "body"),
		})
	}
	return &codegen.File{
		Path: filepath.Join(f.pkg.OutputDirectory(), "validation.go"),
		SectionTemplates: []*codegen.SectionTemplate{
			codegen.Header("Named definition validation", "checks", f.imports.Imports()),
			{Name: "optional", Source: "type Optional {{.}}\n", Data: f.optional.Link(f.pkg.ImportPath(), f.pkg.ImportName).Def()},
			{Name: "functions", Source: `{{range .}}
func {{.Name}}(value {{.Parameter}}) (err error) {
	{{.Body}}
	return
}
{{end}}`, Data: functions},
		},
	}, nil
}

func namedDefinitionValidationDSL() {
	d.Service("catalog", func() {})
	base := d.Type("Base", &expr.Union{TypeName: "Choice"}, func() {
		d.Meta("struct:pkg:path", "left/types")
		d.Meta("type:generate:force")
		d.Attribute("text", d.String, func() {
			d.MinLength(1)
		})
		d.Attribute("number", d.Int, func() {
			d.Minimum(1)
		})
	})
	d.Type("Derived", base, func() {
		d.Meta("struct:pkg:path", "right/types")
		d.Meta("type:generate:force")
	})
	d.Type("Holder", func() {
		d.Meta("struct:pkg:path", "left/types")
		d.Meta("type:generate:force")
		d.OneOf("Selection", func() {
			d.Attribute("text", d.String, func() {
				d.MinLength(1)
			})
			d.Attribute("number", d.Int, func() {
				d.Minimum(1)
			})
		})
	})
	longBase := d.Type("LongBase", func() {
		d.Meta("struct:pkg:path", "left/types")
		d.Meta("type:generate:force")
		d.Attribute("Note", d.String)
	})
	middle := d.Type("LongMiddle", longBase, func() {
		d.Meta("struct:pkg:path", "left/types")
		d.Meta("type:generate:force")
	})
	d.Type("LongLeaf", middle, func() {
		d.Meta("struct:pkg:path", "right/types")
		d.Meta("type:generate:force")
		d.Required("Note")
	})
	valueBase := d.Type("ValueBase", func() {
		d.Meta("struct:pkg:path", "left/types")
		d.Meta("type:generate:force")
		d.Attribute("Note", d.String)
	})
	d.Type("ValueDerived", valueBase, func() {
		d.Meta("struct:pkg:path", "right/types")
		d.Meta("type:generate:force")
		d.Required("Note")
	})
	label := d.Type("Label", d.String, func() {
		d.Meta("struct:pkg:path", "left/types")
		d.Enum("brief", "detailed")
		d.Default("brief")
	})
	d.Type("DefaultedHolder", func() {
		d.Meta("struct:pkg:path", "left/types")
		d.Meta("type:generate:force")
		d.Attribute("label", label)
	})
}

const namedDefinitionValidationTests = `
func TestDefaultedAliasValidation(t *testing.T) {
	for _, test := range []struct {
		value {{.DefaultedHolder}}
		invalid bool
	}{
		{value: {{.DefaultedHolder}}{Label: "brief"}},
		{value: {{.DefaultedHolder}}{Label: "detailed"}},
		{value: {{.DefaultedHolder}}{Label: ""}, invalid: true},
		{value: {{.DefaultedHolder}}{Label: "other"}, invalid: true},
	} {
		before := test.value
		if err := ValidateDefaults(&test.value); (err != nil) != test.invalid {
			t.Errorf("validation = %v, want invalid %t", err, test.invalid)
		}
		if test.value != before {
			t.Error("validation changed the caller's value")
		}
	}
}

func TestNamedUnionCollections(t *testing.T) {
	for _, invalid := range []bool{false, true} {
		value := {{.Derived}}({{.Choicetext}}("valid"))
		if invalid {
			value = {{.Derived}}({{.Choicetext}}(""))
		}
		array := []{{.Derived}}{value}
		mapped := map[string]{{.Derived}}{"first": value}
		if err := ValidateArray(array); (err != nil) != invalid {
			t.Errorf("array validation = %v, want invalid %t", err, invalid)
		}
		if err := ValidateMap(mapped); (err != nil) != invalid {
			t.Errorf("map validation = %v, want invalid %t", err, invalid)
		}
		if array[0] != value || mapped["first"] != value {
			t.Error("validation changed a collection element")
		}
	}
}

func TestNamedUnionBranches(t *testing.T) {
	for _, test := range []struct {
		name string
		value {{.Derived}}
		invalid bool
	}{
		{"valid text", {{.Derived}}({{.Choicetext}}("valid")), false},
		{"invalid text", {{.Derived}}({{.Choicetext}}("")), true},
		{"valid number", {{.Derived}}({{.Choicenumber}}(1)), false},
		{"invalid number", {{.Derived}}({{.Choicenumber}}(0)), true},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, validate := range []func(*{{.Derived}}) error{ValidateDerived, ValidateDefinition} {
				value := test.value
				if err := validate(&value); (err != nil) != test.invalid {
					t.Errorf("validation = %v, want invalid %t", err, test.invalid)
				}
				if value != test.value {
					t.Error("validation changed its caller's value")
				}
			}
			base := {{.Base}}(test.value)
			if err := ValidateBase(&base); (err != nil) != test.invalid {
				t.Errorf("base validation = %v, want invalid %t", err, test.invalid)
			}
		})
	}
}

func TestDirectUnionValueAndOptionalPointer(t *testing.T) {
	if err := ValidateDirect({{.Choicetext}}("")); err == nil {
		t.Error("direct union accepted invalid text")
	}
	if err := ValidateDirect({{.Choicetext}}("valid")); err != nil {
		t.Error(err)
	}
	if err := ValidateOptional(&Optional{}); err != nil {
		t.Errorf("absent optional union: %v", err)
	}
	for _, invalid := range []bool{false, true} {
		value := {{.Selectiontext}}("valid")
		if invalid {
			value = {{.Selectiontext}}("")
		}
		optional := &Optional{Selection: &value}
		before := value
		if err := ValidateOptional(optional); (err != nil) != invalid {
			t.Errorf("optional validation = %v, want invalid %t", err, invalid)
		}
		if value != before {
			t.Error("optional validation changed its union")
		}
		holder := &{{.Holder}}{Selection: value}
		if err := ValidateHolder(holder); (err != nil) != invalid {
			t.Errorf("value field validation = %v, want invalid %t", err, invalid)
		}
	}
}

func TestEvaluatedRequiredFieldRepresentations(t *testing.T) {
	if err := ValidateLong(&{{.LongLeaf}}{}); err == nil {
		t.Error("longer named chain accepted nil required Note")
	}
	empty := ""
	value := &{{.LongLeaf}}{Note: &empty}
	if err := ValidateLong(value); err != nil {
		t.Error(err)
	}
	if value.Note != &empty || *value.Note != "" {
		t.Error("validation changed required Note")
	}
	if err := ValidateLongBase(&{{.LongBase}}{}); err != nil {
		t.Errorf("base field remains optional: %v", err)
	}
	if err := ValidateValue(&{{.ValueDerived}}{Note: ""}); err != nil {
		t.Errorf("two-level required Note is an empty string value: %v", err)
	}
}
`
