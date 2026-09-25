// This fixture compiles conversions and required checks against service-owned
// named unions. Declarations, branch types, and import aliases come from generation.
package generator

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	d "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestGenerateNamedUnionConversions(t *testing.T) {
	root := codegen.RunDSL(t, namedUnionConversionDSL)
	base := &expr.AttributeExpr{Type: root.UserType("Base")}
	derived := &expr.AttributeExpr{Type: root.UserType("Derived")}
	holder := &expr.AttributeExpr{Type: root.UserType("Holder")}
	outer := &expr.AttributeExpr{Type: root.UserType("Outer")}
	presence := &expr.AttributeExpr{Type: root.UserType("PresenceHolder")}
	type operation struct {
		name                   string
		source, target         *expr.AttributeExpr
		sourcePtr, targetPtr   bool
		ordinary, newVar       bool
		plan                   *codegen.TransformPlan
		sourceType, targetType *codegen.GoTypePlan
	}
	operations := []*operation{
		{name: "OrdinaryToBase", source: derived, target: base, ordinary: true},
		{name: "ToBase", source: derived, target: base},
		{name: "ToDerived", source: base, target: derived, newVar: true},
		{name: "ToDirect", source: derived, target: base.Type.(expr.UserType).Attribute()},
		{name: "FromDirect", source: base.Type.(expr.UserType).Attribute(), target: derived},
		{name: "OrdinaryHolder", source: holder, target: holder, ordinary: true, newVar: true},
		{name: "CopyHolder", source: holder, target: holder, newVar: true},
		{name: "ToOptional", source: holder, target: root.UserType("Holder").Attribute(), targetPtr: true},
		{name: "FromOptional", source: root.UserType("Holder").Attribute(), target: holder, sourcePtr: true},
		{name: "CopyOuter", source: outer, target: outer},
	}
	var pkg *codegen.GeneratedPackage
	var imports *codegen.GeneratedImportPlan
	var validation *codegen.ValidationPlan
	var holderValidation, presenceValidation, pointerValidation *codegen.ValidationPlan
	var presenceLayout, pointerLayout *codegen.GoTypePlan
	functions := make([]map[string]string, 0, len(operations))
	var optionalDefinition string
	registry := newDefaultRegistry()
	registry.registerPlugin("named-union-conversions", "gen", pluginNormal, func() Plugin {
		return Plugin{
			Plan: func(plan *Plan) error {
				generation := plan.Generation()
				var err error
				pkg, err = generation.ClaimPackage("generated.local/gen/checks")
				if err != nil {
					return err
				}
				imports = codegen.NewGeneratedImportPlan(pkg)
				bind := func(request codegen.GoTypeBindingRequest) (codegen.GoTypeBinding, error) {
					owner := request.InheritedOwner
					if location := codegen.UserTypeLocation(request.Attribute.Type); location != nil {
						owner = path.Join(generation.GenPkg(), location.RelImportPath)
					}
					binding := codegen.GoTypeBinding{Owner: owner}
					var err error
					if request.Kind == codegen.GoUnion {
						binding.Union, err = generation.Package(owner).Union(request.Attribute)
					} else {
						binding.Type, err = generation.Package(owner).Type(request.Attribute.Type.(expr.UserType))
					}
					return binding, err
				}
				for _, op := range operations {
					op.plan, err = codegen.NewTransformPlan(op.source, op.target, op.name, nil)
					if err != nil {
						return err
					}
					require.Empty(t, op.plan.Helpers())
					for index, attribute := range []*expr.AttributeExpr{op.source, op.target} {
						pointer := op.sourcePtr
						if index == 1 {
							pointer = op.targetPtr
						}
						layout, err := codegen.PlanGoType(attribute, codegen.GoTypePlanOptions{
							Owner: "generated.local/gen/left/types", Bind: bind, RetainNamedValue: true,
							Policy: codegen.GoLayoutPolicy{UseDefault: true, SumType: true, UnionPointer: pointer},
						})
						if err != nil {
							return err
						}
						if index == 0 {
							op.sourceType = layout
						} else {
							op.targetType = layout
						}
						for _, spec := range layout.CompleteImportPreferences() {
							if err := imports.AddGenerated(codegen.NewImport("gentypes", spec.Path)); err != nil {
								return err
							}
						}
					}
					if err := pkg.DeclareName(codegen.NewExactName(codegen.NameFunction, op.name)); err != nil {
						return err
					}
				}
				validation, err = codegen.NewValidationPlan(derived, operations[0].sourceType, codegen.ValidationPlanOptions{Required: true})
				if err != nil {
					return err
				}
				derivedValidator := codegen.NewExactName(codegen.NameFunction, "ValidateDerived")
				if err := pkg.DeclareName(derivedValidator); err != nil {
					return err
				}
				holderValidation, err = codegen.NewValidationPlan(holder, operations[5].sourceType, codegen.ValidationPlanOptions{
					Required: true,
					Bind: func(request codegen.ValidatorBindingRequest) (*codegen.NameDeclaration, error) {
						if request.Attribute.Type != derived.Type {
							return nil, fmt.Errorf("unexpected nested validator for %s", request.Attribute.Type.Name())
						}
						return derivedValidator, nil
					},
				})
				if err != nil {
					return err
				}
				options := codegen.GoTypePlanOptions{
					Owner: "generated.local/gen/values/types", Bind: bind, RetainNamedValue: true,
					Policy: codegen.GoLayoutPolicy{UseDefault: true, SumType: true},
				}
				presenceLayout, err = codegen.PlanGoType(presence, options)
				if err != nil {
					return err
				}
				presenceValidation, err = codegen.NewValidationPlan(presence, presenceLayout, codegen.ValidationPlanOptions{Required: true})
				if err != nil {
					return err
				}
				require.Empty(t, presenceValidation.ValidatorDeclarations())
				wantPresenceImports := []codegen.GoTypeImport{
					{Name: "goa", Path: codegen.GoaImport("").Path},
					{Path: "generated.local/gen/left/types"},
				}
				require.Equal(t, wantPresenceImports, presenceValidation.ImportPreferences(),
					"the required check alone needs the union method owner's import")
				presenceDefinition := presence.Type.(expr.UserType).Attribute()
				for _, collection := range []expr.DataType{
					&expr.Array{ElemType: presenceDefinition},
					&expr.Map{KeyType: &expr.AttributeExpr{Type: expr.String}, ElemType: presenceDefinition},
				} {
					attribute := &expr.AttributeExpr{Type: collection}
					layout, err := codegen.PlanGoType(attribute, options)
					if err != nil {
						return err
					}
					validation, err := codegen.NewValidationPlan(attribute, layout, codegen.ValidationPlanOptions{Required: true})
					if err != nil {
						return err
					}
					require.Equal(t, wantPresenceImports, validation.ImportPreferences(),
						"anonymous collection elements must retain required-check imports")
				}
				options.RetainNamedValue = false
				incomplete, err := codegen.PlanGoType(presenceDefinition, options)
				if err != nil {
					return err
				}
				_, err = codegen.NewValidationPlan(presenceDefinition, incomplete, codegen.ValidationPlanOptions{Required: true})
				require.ErrorContains(t, err, `required field "choice": named union layout does not retain its union definition`)
				options.RetainNamedValue = true
				options.Policy.Pointer, options.Policy.UnionPointer = true, true
				pointerLayout, err = codegen.PlanGoType(presenceDefinition, options)
				if err != nil {
					return err
				}
				pointerValidation, err = codegen.NewValidationPlan(presenceDefinition, pointerLayout, codegen.ValidationPlanOptions{Required: true})
				if err != nil {
					return err
				}
				require.Equal(t, wantPresenceImports[:1], pointerValidation.ImportPreferences(),
					"a nil presence check does not call union methods")
				for _, validation := range []*codegen.ValidationPlan{validation, holderValidation, presenceValidation, pointerValidation} {
					for _, spec := range validation.ImportPreferences() {
						if spec.Path == codegen.GoaImport("").Path {
							err = imports.Require(codegen.NewImport(spec.Name, spec.Path))
						} else {
							err = imports.AddGenerated(codegen.NewImport("gentypes", spec.Path))
						}
						if err != nil {
							return err
						}
					}
				}
				for _, declaration := range []*codegen.NameDeclaration{
					codegen.NewExactName(codegen.NameType, "Optional"),
					codegen.NewExactName(codegen.NameType, "PresencePointer"),
					codegen.NewExactName(codegen.NameFunction, "ValidateHolder"),
					codegen.NewExactName(codegen.NameFunction, "ValidatePresence"),
					codegen.NewExactName(codegen.NameFunction, "ValidatePresencePointer"),
					codegen.NewExactName(codegen.NameFunction, "ValidatePresenceLegacy"),
				} {
					if err := pkg.DeclareName(declaration); err != nil {
						return err
					}
				}
				return nil
			},
			Generate: func(plan *Plan, files []*codegen.File) ([]*codegen.File, error) {
				if err := imports.Link(); err != nil {
					return nil, err
				}
				for _, op := range operations {
					layouts := []codegen.LinkedGoType{
						op.sourceType.Link(pkg.ImportPath(), pkg.ImportName),
						op.targetType.Link(pkg.ImportPath(), pkg.ImportName),
					}
					contexts := make([]*codegen.AttributeContext, 2)
					for index, layout := range layouts {
						context := codegen.NewAttributeContext(false, false, true, "", codegen.NewNameScope())
						context.UnionPointer = op.sourcePtr
						if index == 1 {
							context.UnionPointer = op.targetPtr
						}
						if op.ordinary {
							context.Scope = plan.Service(root).Services().ServiceAttributor("catalog", pkg.ImportPath())
						} else {
							var err error
							context, err = context.WithGoTypeLayout(layout)
							if err != nil {
								return nil, err
							}
						}
						contexts[index] = context
					}
					if err := op.plan.BindContexts(contexts[0], contexts[1]); err != nil {
						return nil, err
					}
					body, helpers, err := op.plan.Render("input", "output", op.newVar)
					if err != nil {
						return nil, err
					}
					require.Empty(t, helpers)
					parameter, result := layouts[0].Ref(), layouts[1].Ref()
					if op.sourcePtr {
						parameter = "*Optional"
					}
					if op.targetPtr {
						result = "*Optional"
						optionalDefinition = layouts[1].Def()
					}
					if !op.newVar {
						body = "var output " + result + "\n" + body
					}
					functions = append(functions, map[string]string{
						"Name": op.name, "Parameter": parameter, "Result": result, "Body": body,
					})
				}
				linkedValidation, err := validation.Link(operations[0].sourceType.Link(pkg.ImportPath(), pkg.ImportName))
				if err != nil {
					return nil, err
				}
				validators := []map[string]string{{
					"Name":      "ValidateDerived",
					"Parameter": operations[0].sourceType.Link(pkg.ImportPath(), pkg.ImportName).Ref(),
					"Body":      linkedValidation.Render("value", "body"),
				}}
				for _, item := range []struct {
					name   string
					plan   *codegen.ValidationPlan
					layout *codegen.GoTypePlan
				}{
					{"ValidateHolder", holderValidation, operations[5].sourceType},
					{"ValidatePresence", presenceValidation, presenceLayout},
					{"ValidatePresencePointer", pointerValidation, pointerLayout},
				} {
					layout := item.layout.Link(pkg.ImportPath(), pkg.ImportName)
					linked, err := item.plan.Link(layout)
					if err != nil {
						return nil, err
					}
					parameter := layout.Ref()
					if item.plan == pointerValidation {
						parameter = "*PresencePointer"
					}
					if item.plan == presenceValidation {
						require.Contains(t, linked.Imports(), codegen.GoTypeImport{
							Name: pkg.ImportName("generated.local/gen/left/types"), Path: "generated.local/gen/left/types",
						})
					}
					validators = append(validators, map[string]string{
						"Name": item.name, "Parameter": parameter, "Body": linked.Render("value", "body"),
					})
				}
				legacy := codegen.NewAttributeContext(false, false, true, "", codegen.NewNameScope())
				legacy.Scope = plan.Service(root).Services().ServiceAttributor("catalog", pkg.ImportPath()).Enter(presence)
				validators = append(validators, map[string]string{
					"Name":      "ValidatePresenceLegacy",
					"Parameter": presenceLayout.Link(pkg.ImportPath(), pkg.ImportName).Ref(),
					"Body":      codegen.AttributeValidationCode(presence, nil, legacy, true, false, "value", "body"),
				})
				file := &codegen.File{
					Path: filepath.Join(pkg.OutputDirectory(), "conversions.go"),
					SectionTemplates: []*codegen.SectionTemplate{
						codegen.Header("Named union conversions", "checks", imports.Imports()),
						{Name: "optional", Source: "type Optional = {{.}}\n", Data: optionalDefinition},
						{Name: "presence-pointer", Source: "type PresencePointer = {{.}}\n", Data: pointerLayout.Link(pkg.ImportPath(), pkg.ImportName).Def()},
						{Name: "functions", Source: `{{range .}}
func {{.Name}}(input {{.Parameter}}) {{.Result}} {
	{{.Body}}
	return output
}
{{end}}`, Data: functions},
						{Name: "validation", Source: "{{range .}}func {{.Name}}(value {{.Parameter}}) (err error) {\n{{.Body}}\nreturn\n}\n{{end}}", Data: validators},
					},
				}
				return append(files, file), nil
			},
		}
	})
	run, err := newGenerationRun("gen", registry)
	require.NoError(t, err)
	result, err := run.execute("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	generation := result.plan.Generation()
	left := generation.Package("generated.local/gen/left/types")
	right := generation.Package("generated.local/gen/right/types")
	values := generation.Package("generated.local/gen/values/types")
	require.True(t, generation.Frozen())
	require.NotEqual(t, pkg.ImportName(left.ImportPath()), pkg.ImportName(right.ImportPath()))
	data := make(map[string]string)
	for _, name := range []string{"Base", "Derived", "Holder", "Outer", "PlainDerived", "PresenceHolder"} {
		owner := left
		if name == "Derived" || name == "PlainDerived" {
			owner = right
		} else if name == "Holder" || name == "Outer" || name == "PresenceHolder" {
			owner = values
		}
		declaration, err := owner.Type(root.UserType(name))
		require.NoError(t, err)
		data[name] = pkg.ImportName(owner.ImportPath()) + "." + declaration.Name()
	}
	for _, branchName := range []string{"text", "number"} {
		branch, err := left.UnionBranch(root.UserType("Base").Attribute(), branchName)
		require.NoError(t, err)
		data[branchName] = pkg.ImportName(left.ImportPath()) + "." + branch.Constructor()
	}
	branch, err := values.UnionBranch(root.UserType("Outer").Attribute(), "choice")
	require.NoError(t, err)
	data["choice"] = pkg.ImportName(values.ImportPath()) + "." + branch.Constructor()
	envelope, err := values.Union(root.UserType("Outer").Attribute())
	require.NoError(t, err)
	data["Envelope"] = pkg.ImportName(values.ImportPath()) + "." + envelope.Name()
	plain, err := left.UnionBranch(root.UserType("PlainBase").Attribute(), "text")
	require.NoError(t, err)
	data["plainText"] = pkg.ImportName(left.ImportPath()) + "." + plain.Constructor()
	directory := t.TempDir()
	writeGeneratedModule(t, directory, "generated.local")
	for _, file := range result.files {
		_, err := file.Render(directory)
		require.NoError(t, err)
	}
	testFile := &codegen.File{
		Path: filepath.Join(pkg.OutputDirectory(), "conversions_test.go"),
		SectionTemplates: []*codegen.SectionTemplate{
			codegen.Header("Named union conversion tests", "checks", []*codegen.ImportSpec{
				codegen.SimpleImport("testing"), codegen.SimpleImport("reflect"),
				pkg.Import(left.ImportPath()), pkg.Import(right.ImportPath()), pkg.Import(values.ImportPath()),
			}),
			{Name: "tests", Source: namedUnionConversionTests, Data: data},
		},
	}
	_, err = testFile.Render(directory)
	require.NoError(t, err)

	// A new render request after mutation must still read the retained named
	// chain. Repeating an identical request would only test the render cache.
	before := make(map[string]string)
	for _, op := range operations {
		if !op.ordinary {
			body, _, err := op.plan.Render("before", "output", op.newVar)
			require.NoError(t, err)
			before[op.name] = body
		}
	}
	retainedPresence, err := presenceValidation.Link(presenceLayout.Link(pkg.ImportPath(), pkg.ImportName))
	require.NoError(t, err)
	beforePresence := retainedPresence.Render("value", "body")
	beforeImports := retainedPresence.Imports()
	for _, userType := range root.Types {
		userType.SetAttribute(&expr.AttributeExpr{Type: expr.Boolean})
	}
	for _, op := range operations {
		if !op.ordinary {
			body, _, err := op.plan.Render("after", "output", op.newVar)
			require.NoError(t, err)
			require.Equal(t, strings.ReplaceAll(before[op.name], "before", "after"), body)
		}
	}
	require.Equal(t, beforePresence, retainedPresence.Render("value", "body"))
	require.Equal(t, beforeImports, retainedPresence.Imports())
	runGeneratedTests(t, directory)
}

func namedUnionConversionDSL() {
	d.Service("catalog", func() {})
	base := d.Type("Base", &expr.Union{TypeName: "Choice"}, func() {
		d.Meta("struct:pkg:path", "left/types")
		d.Meta("type:generate:force")
		d.Attribute("text", d.String)
		d.Attribute("number", d.Int, func() { d.Minimum(1) })
	})
	derived := d.Type("Derived", base, func() {
		d.Meta("struct:pkg:path", "right/types")
		d.Meta("type:generate:force")
	})
	d.Type("Holder", func() {
		d.Meta("struct:pkg:path", "values/types")
		d.Meta("type:generate:force")
		d.Attribute("single", derived)
		d.Attribute("optional", derived)
		d.Attribute("choices", d.ArrayOf(derived))
		d.Attribute("byName", d.MapOf(d.String, derived))
		d.Required("single", "choices", "byName")
	})
	d.Type("Outer", &expr.Union{TypeName: "Envelope"}, func() {
		d.Meta("struct:pkg:path", "values/types")
		d.Meta("type:generate:force")
		d.Attribute("choice", derived)
		d.Attribute("text", d.String)
	})
	plain := d.Type("PlainBase", &expr.Union{TypeName: "PlainChoice"}, func() {
		d.Meta("struct:pkg:path", "left/types")
		d.Meta("type:generate:force")
		d.Attribute("text", d.String)
	})
	plainDerived := d.Type("PlainDerived", plain, func() {
		d.Meta("struct:pkg:path", "right/types")
		d.Meta("type:generate:force")
	})
	d.Type("PresenceHolder", func() {
		d.Meta("struct:pkg:path", "values/types")
		d.Meta("type:generate:force")
		d.Attribute("choice", plainDerived)
		d.Required("choice")
	})
}

const namedUnionConversionTests = `
func TestNamedUnionConversion(t *testing.T) {
	for _, test := range []struct {
		name string
		value {{.Derived}}
		invalid bool
	}{
		{"text", {{.Derived}}({{.text}}("hello")), false},
		{"number", {{.Derived}}({{.number}}(1)), false},
		{"invalid", {{.Derived}}({{.number}}(0)), true},
	} {
		t.Run(test.name, func(t *testing.T) {
			value := test.value
			for _, convert := range []func(*{{.Derived}}) *{{.Base}}{ToBase, OrdinaryToBase} {
				output := ToDerived(convert(&value))
				if !reflect.DeepEqual(*output, value) || !reflect.DeepEqual(value, test.value) {
					t.Fatal("root conversion changed its value or caller")
				}
				if err := ValidateDerived(output); (err != nil) != test.invalid {
					t.Fatalf("validation = %v, invalid = %t", err, test.invalid)
				}
			}
			if output := FromDirect(ToDirect(&value)); !reflect.DeepEqual(*output, value) {
				t.Fatal("direct union round trip changed the value")
			}
			input := &{{.Holder}}{Single: value, Optional: value, Choices: []{{.Derived}}{value}, ByName: map[string]{{.Derived}}{"first": value}}
			if err := ValidateHolder(input); (err != nil) != test.invalid {
				t.Fatalf("holder validation = %v, invalid = %t", err, test.invalid)
			}
			for _, convert := range []func(*{{.Holder}}) *{{.Holder}}{CopyHolder, OrdinaryHolder} {
				output := convert(input)
				if !reflect.DeepEqual(output, input) { t.Fatal("collection conversion changed values") }
				mapped := output.ByName["first"]
				for _, selected := range []*{{.Derived}}{&output.Single, &output.Choices[0], &mapped} {
					if err := ValidateDerived(selected); (err != nil) != test.invalid {
						t.Fatalf("converted field/element validation = %v, invalid = %t", err, test.invalid)
					}
				}
				output.Choices[0] = {{.Derived}}({{.text}}("changed"))
				output.ByName["first"] = output.Choices[0]
				if !reflect.DeepEqual(input.Choices[0], value) || !reflect.DeepEqual(input.ByName["first"], value) {
					t.Fatal("collection conversion retained mutable storage")
				}
			}
			optional := ToOptional(input)
			if optional.Optional == nil || !reflect.DeepEqual(*optional.Optional, value) { t.Fatal("lost present optional field") }
			if output := FromOptional(optional); !reflect.DeepEqual(output, input) { t.Fatal("optional round trip changed value") }
			optional.Optional = nil
			absent := FromOptional(optional)
			if ToOptional(absent).Optional != nil { t.Fatal("absent optional field became present") }
			if !reflect.DeepEqual(input.Optional, value) { t.Fatal("optional conversion changed caller") }
		})
	}
}

func TestRequiredNamedUnionPresence(t *testing.T) {
	holder := &{{.Holder}}{Choices: []{{.Derived}}{}, ByName: map[string]{{.Derived}}{}}
	if err := ValidateHolder(holder); err == nil { t.Fatal("accepted missing required named union") }
	for _, validate := range []func(*{{.PresenceHolder}}) error{ValidatePresence, ValidatePresenceLegacy} {
		if err := validate(&{{.PresenceHolder}}{}); err == nil { t.Fatal("accepted missing unconstrained union") }
		value := &{{.PresenceHolder}}{Choice: {{.PlainDerived}}({{.plainText}}(""))}
		before := *value
		if err := validate(value); err != nil { t.Fatal(err) }
		if !reflect.DeepEqual(*value, before) { t.Fatal("required validation changed its caller") }
	}
	if err := ValidatePresencePointer(&PresencePointer{}); err == nil { t.Fatal("accepted nil required pointer") }
	selected := {{.PlainDerived}}({{.plainText}}(""))
	for _, pointer := range []*{{.PlainDerived}}{&selected, new({{.PlainDerived}})} {
		before := *pointer
		if err := ValidatePresencePointer(&PresencePointer{Choice: pointer}); err != nil {
			t.Fatalf("pointer presence check rejected a nonnil field: %v", err)
		}
		if !reflect.DeepEqual(*pointer, before) { t.Fatal("pointer presence validation changed its caller") }
	}
}

func TestNamedPointerUnionBranch(t *testing.T) {
	for _, selected := range []*{{.Derived}}{nil, new({{.Derived}})} {
		if selected != nil { *selected = {{.Derived}}({{.text}}("nested")) }
		input := {{.Outer}}({{.choice}}(selected))
		output := CopyOuter(&input)
		if !reflect.DeepEqual(*output, input) { t.Fatal("pointer branch conversion changed value") }
		converted, ok := (*{{.Envelope}})(output).AsChoice()
		if !ok { t.Fatal("lost selected pointer branch") }
		if selected != nil {
			before := *selected
			*converted = {{.Derived}}({{.text}}("changed"))
			if !reflect.DeepEqual(*selected, before) { t.Fatal("pointer branch conversion changed caller") }
		}
	}
}
`
