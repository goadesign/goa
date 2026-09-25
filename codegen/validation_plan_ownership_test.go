// These tests bind validation functions separately from their argument types.
// They check declaration errors and output-relative access before code is rendered.
package codegen

import (
	"cmp"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

type validationOwnershipOrder int

func TestValidationPlanFunctionBinding(t *testing.T) {
	for _, test := range []struct {
		name  string
		kind  PackageNameKind
		owned bool
		want  string
	}{
		{name: "function", kind: NameFunction, owned: true},
		{name: "type", kind: NameType, owned: true, want: "must be a function, got type"},
		{name: "variable", kind: NameVariable, owned: true, want: "must be a function, got variable"},
		{name: "constant", kind: NameConstant, owned: true, want: "must be a function, got constant"},
		{name: "unowned", kind: NameFunction, want: "is not owned"},
		{name: "nil", want: "must not be nil"},
	} {
		t.Run(test.name, func(t *testing.T) {
			generation, attribute, layout := validationOwnershipLayout(t)
			var declaration *NameDeclaration
			if test.kind != 0 {
				declaration = NewExactName(test.kind, "validateChild")
				if test.owned {
					pkg, err := generation.ClaimPackage("generated.local/gen/checks")
					require.NoError(t, err)
					require.NoError(t, pkg.DeclareName(declaration))
				}
			}
			plan, err := NewValidationPlan(attribute, layout, ValidationPlanOptions{
				Bind: func(request ValidatorBindingRequest) (*NameDeclaration, error) {
					require.True(t, request.Layout.MatchesOccurrence(request.Attribute))
					require.Equal(t, "generated.local/gen/types", request.Layout.Owner())
					require.Same(t, layout.Fields()[0].TypeDeclaration(), request.Layout.TypeDeclaration())
					return declaration, nil
				},
			})
			if test.want != "" {
				require.ErrorContains(t, err, test.want)
				require.Nil(t, plan)
				return
			}
			require.NoError(t, err)
			require.Equal(t, []*NameDeclaration{declaration, declaration}, plan.ValidatorDeclarations())
		})
	}
}

func TestValidationPlanFunctionAccess(t *testing.T) {
	for _, test := range []struct {
		name      string
		spelling  string
		form      string
		local     bool
		typeOwner bool
		collision bool
		wantError bool
	}{
		{name: "local exact private", spelling: "validateChild", local: true},
		{name: "local exact exported", spelling: "ValidateChild", local: true},
		{name: "foreign exact exported", spelling: "ValidateChild"},
		{name: "foreign exact private", spelling: "validateChild", wantError: true},
		{name: "foreign type-owned private", spelling: "validateChild", typeOwner: true, wantError: true},
		{name: "foreign type-owned exported", spelling: "ValidateChild", typeOwner: true},
		{name: "local preferred collision", spelling: "validateChild", form: "preferred", local: true, collision: true},
		{name: "foreign preferred collision", spelling: "ValidateChild", form: "preferred", collision: true},
		{name: "foreign preferred private", spelling: "validateChild", form: "preferred", wantError: true},
		{name: "local dependent private", spelling: "validateChild", form: "dependent", local: true},
		{name: "foreign dependent private", spelling: "validateChild", form: "dependent", wantError: true},
		{name: "foreign dependent exported collision", spelling: "ValidateChild", form: "dependent", collision: true},
		{name: "foreign unicode exported", spelling: "Évaluer"},
		{name: "foreign unicode private", spelling: "évaluer", wantError: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			generation, attribute, layout := validationOwnershipLayout(t)
			helperOwner := "generated.local/gen/checks"
			if test.typeOwner {
				helperOwner = layout.Owner()
			}
			outputPath := "generated.local/gen/output"
			if test.local {
				outputPath = helperOwner
			}
			helperPackage, err := generation.ClaimPackage(helperOwner)
			require.NoError(t, err)
			if test.collision {
				require.NoError(t, helperPackage.DeclareName(NewExactName(NameVariable, test.spelling)))
			}
			var declaration *NameDeclaration
			switch test.form {
			case "preferred":
				visibility := ExportedName
				if test.spelling == "validateChild" {
					visibility = UnexportedName
				}
				declaration = NewPreferredName(NameFunction, test.spelling, visibility, validationOwnershipOrder(0))
				require.NoError(t, helperPackage.DeclareName(declaration))
			case "dependent":
				base := layout.Fields()[0].TypeDeclaration().Declaration()
				prefix := "Validate"
				if test.spelling == "validateChild" {
					prefix = "validate"
				}
				var err error
				declaration, err = helperPackage.DeclareDependentName(NameFunction, base, prefix, "", validationOwnershipOrder(0))
				require.NoError(t, err)
			default:
				declaration = NewExactName(NameFunction, test.spelling)
				require.NoError(t, helperPackage.DeclareName(declaration))
			}
			plan, err := NewValidationPlan(attribute, layout, ValidationPlanOptions{
				Bind: func(ValidatorBindingRequest) (*NameDeclaration, error) {
					return declaration, nil
				},
			})
			require.NoError(t, err)
			require.NoError(t, generation.Freeze())
			linked, err := plan.Link(layout.Link(outputPath, func(importPath string) string {
				require.NotEqual(t, outputPath, importPath, "a local call must not request an import")
				if importPath == GoaImport("").Path {
					return "goa"
				}
				require.Equal(t, helperOwner, importPath)
				return "checks"
			}))
			if test.wantError {
				require.ErrorContains(t, err, "is not exported to output package")
				require.Equal(t, LinkedValidationPlan{}, linked)
				return
			}
			require.NoError(t, err)
			name := test.spelling
			if test.collision {
				name += "2"
			}
			require.Equal(t, name, declaration.Name())
			if !test.local {
				name = "checks." + name
			}
			require.Contains(t, linked.Render("value", "body"), name+"(value.First)")
			imports := linked.Imports()
			if test.local {
				require.Len(t, imports, 1)
			} else {
				require.Len(t, imports, 2)
			}
			for _, spec := range imports {
				require.NotEqual(t, outputPath, spec.Path)
			}
		})
	}
}

func TestValidationPlanLinkKeepsItsExactLayout(t *testing.T) {
	generation, attribute, layout := validationOwnershipLayout(t)
	declaration := NewExactName(NameFunction, "validateChild")
	require.NoError(t, generation.Package(layout.Owner()).DeclareName(declaration))
	plan, err := NewValidationPlan(attribute, layout, ValidationPlanOptions{
		Bind: func(ValidatorBindingRequest) (*NameDeclaration, error) {
			return declaration, nil
		},
	})
	require.NoError(t, err)
	linked, err := plan.Link(layout.Link(layout.Owner(), validationPlanTestQualifier))
	require.ErrorContains(t, err, "is not frozen")
	require.Equal(t, LinkedValidationPlan{}, linked)
	require.NoError(t, generation.Freeze())
	other := *layout
	linked, err = plan.Link(other.Link(layout.Owner(), validationPlanTestQualifier))
	require.ErrorContains(t, err, "does not belong to this validation plan")
	require.Equal(t, LinkedValidationPlan{}, linked)
	_, err = plan.Link(layout.Link(layout.Owner(), validationPlanTestQualifier))
	require.NoError(t, err)
}

func TestValidationPlanImportsFollowOutput(t *testing.T) {
	for _, outputPath := range []string{
		"generated.local/gen/types",
		"generated.local/gen/one/checks",
		"generated.local/gen/output",
	} {
		t.Run(outputPath, func(t *testing.T) {
			generation, attribute, layout := validationOwnershipLayout(t)
			owners := []string{layout.Owner(), "generated.local/gen/one/checks", "generated.local/gen/two/checks"}
			declarations := make([]*NameDeclaration, len(owners))
			for index, owner := range owners {
				declarations[index] = NewExactName(NameFunction, "ValidateChild")
				pkg, err := generation.ClaimPackage(owner)
				require.NoError(t, err)
				require.NoError(t, pkg.DeclareName(declarations[index]))
			}
			output, err := generation.ClaimPackage(outputPath)
			require.NoError(t, err)
			require.NoError(t, output.DeclareName(NewExactName(NameVariable, "checks")))
			require.NoError(t, output.DeclareName(NewExactName(NameVariable, "goa")))
			plans := make([]*ValidationPlan, 0, len(declarations))
			fileImports := NewGeneratedImportPlan(output)
			require.NoError(t, fileImports.AddGenerated(NewImport("genTypes", layout.Owner())))
			for _, declaration := range declarations {
				plan, err := NewValidationPlan(attribute, layout, ValidationPlanOptions{
					Bind: func(ValidatorBindingRequest) (*NameDeclaration, error) {
						return declaration, nil
					},
				})
				require.NoError(t, err)
				require.Len(t, plan.ImportPreferences(), 2, "two calls retain one function package, even the root type owner")
				for _, preference := range plan.ImportPreferences() {
					require.NoError(t, fileImports.AddGenerated(NewImport(preference.Name, preference.Path)))
				}
				plans = append(plans, plan)
			}
			beforeFreeze := fileImports.Paths()
			require.NoError(t, generation.Freeze())
			require.NoError(t, fileImports.Link())
			require.Equal(t, beforeFreeze, fileImports.Paths())
			require.NotEqual(t, "goa", output.ImportName(GoaImport("").Path))
			for index, plan := range plans {
				linked, err := plan.Link(layout.Link(outputPath, output.ImportName))
				require.NoError(t, err)
				name := declarations[index].Name()
				if owners[index] != outputPath {
					name = output.ImportName(owners[index]) + "." + name
				}
				require.Contains(t, linked.Render("value", "body"), name+"(value.First)")
				for _, spec := range linked.Imports() {
					require.NotEqual(t, outputPath, spec.Path)
					require.Equal(t, output.ImportName(spec.Path), spec.Name)
					require.Contains(t, beforeFreeze, spec.Path)
				}
			}
			if outputPath != owners[1] {
				require.NotEqual(t, output.ImportName(owners[1]), output.ImportName(owners[2]))
			}
		})
	}
}

func TestValidationPlanKeepsCopiedLayoutsAndDeclarations(t *testing.T) {
	original := goTypeTestUserType("Record", &expr.Object{})
	first := original.Dup(nil)
	second := original.Dup(nil)
	minimum := 1.0
	first.SetAttribute(&expr.AttributeExpr{
		Type: &expr.Object{{Name: "left", Attribute: &expr.AttributeExpr{
			Type: expr.String, Validation: &expr.ValidationExpr{Pattern: "^left$"},
		}}},
		Validation: &expr.ValidationExpr{Required: []string{"left"}},
	})
	second.SetAttribute(&expr.AttributeExpr{
		Type: &expr.Object{{Name: "right", Attribute: &expr.AttributeExpr{
			Type: expr.Int, Validation: &expr.ValidationExpr{Minimum: &minimum},
		}}},
		Validation: &expr.ValidationExpr{Required: []string{"right"}},
	})
	require.Same(t, first.Origin(), second.Origin())
	attribute := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "first", Attribute: &expr.AttributeExpr{Type: first}},
		{Name: "second", Attribute: &expr.AttributeExpr{Type: second}},
		{Name: "repeated", Attribute: &expr.AttributeExpr{Type: first}},
	}}
	generation, err := NewGeneration("generated.local/gen", nil)
	require.NoError(t, err)
	firstType := declareGoTypeTestUserType(t, generation, "generated.local/gen/one", first)
	secondType := declareGoTypeTestUserType(t, generation, "generated.local/gen/two", second)
	output, err := generation.ClaimPackage("generated.local/gen/checks")
	require.NoError(t, err)
	firstCall, err := output.DeclareDependentName(NameFunction, firstType.Declaration(), "validate", "", validationOwnershipOrder(1))
	require.NoError(t, err)
	secondCall, err := output.DeclareDependentName(NameFunction, secondType.Declaration(), "validate", "", validationOwnershipOrder(2))
	require.NoError(t, err)
	bind := goTypeTestBinder(map[expr.DataType]GoTypeBinding{
		first:  {Owner: firstType.PackagePath(), Type: firstType},
		second: {Owner: secondType.PackagePath(), Type: secondType},
	})
	policy := GoLayoutPolicy{UseDefault: true, SumType: true}
	layout, err := PlanGoType(attribute, GoTypePlanOptions{Owner: output.ImportPath(), Policy: policy, Bind: bind})
	require.NoError(t, err)
	plan, err := NewValidationPlan(attribute, layout, ValidationPlanOptions{
		Bind: func(request ValidatorBindingRequest) (*NameDeclaration, error) {
			switch request.Layout.TypeDeclaration() {
			case firstType:
				require.Same(t, first, request.Attribute.Type)
				return firstCall, nil
			case secondType:
				require.Same(t, second, request.Attribute.Type)
				return secondCall, nil
			default:
				return nil, fmt.Errorf("validation requested another original declaration")
			}
		},
	})
	require.NoError(t, err)
	require.Equal(t, []*NameDeclaration{firstCall, secondCall, firstCall}, plan.ValidatorDeclarations())

	bodies := make([]*ValidationPlan, 0, 2)
	for index, userType := range []expr.UserType{first, second} {
		owner := []*TypeDeclaration{firstType, secondType}[index].PackagePath()
		definition, err := PlanGoType(userType.Attribute(), GoTypePlanOptions{
			Owner: owner, Policy: policy, Bind: bind,
		})
		require.NoError(t, err)
		body, err := NewValidationPlan(userType.Attribute(), definition, ValidationPlanOptions{Required: true})
		require.NoError(t, err)
		bodies = append(bodies, body)
	}
	require.NoError(t, generation.Freeze())
	first.SetAttribute(&expr.AttributeExpr{Type: expr.String})
	second.SetAttribute(&expr.AttributeExpr{Type: expr.String})
	minimum = 99
	linked, err := plan.Link(layout.Link(output.ImportPath(), validationPlanTestQualifier))
	require.NoError(t, err)
	require.Contains(t, linked.Render("value", "body"), firstCall.Name()+"(value.First)")
	require.Contains(t, linked.Render("value", "body"), secondCall.Name()+"(value.Second)")
	require.NotEqual(t, firstCall.Name(), secondCall.Name())
	for index, body := range bodies {
		linked, err := body.Link(body.layout.Link(output.ImportPath(), validationPlanTestQualifier))
		require.NoError(t, err)
		if index == 0 {
			require.Contains(t, linked.Render("value", "body"), `value.Left, "^left$"`)
		} else {
			require.Contains(t, linked.Render("value", "body"), "value.Right < 1")
			require.NotContains(t, linked.Render("value", "body"), "99")
		}
	}
}

// validationOwnershipLayout keeps one original Child declaration behind two
// field occurrences, so every test checks repeated calls without changing types.
func validationOwnershipLayout(t *testing.T) (*Generation, *expr.AttributeExpr, *GoTypePlan) {
	t.Helper()
	minimum := 1
	child := goTypeTestUserType("Child", &expr.Object{
		{Name: "label", Attribute: &expr.AttributeExpr{
			Type: expr.String, Validation: &expr.ValidationExpr{MinLength: &minimum},
		}},
	})
	attribute := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "first", Attribute: &expr.AttributeExpr{Type: child}},
		{Name: "second", Attribute: &expr.AttributeExpr{Type: child}},
	}}
	generation, err := NewGeneration("generated.local/gen", nil)
	require.NoError(t, err)
	const owner = "generated.local/gen/types"
	declaration := declareGoTypeTestUserType(t, generation, owner, child)
	layout, err := PlanGoType(attribute, GoTypePlanOptions{
		Owner: owner, Policy: GoLayoutPolicy{UseDefault: true, SumType: true},
		Bind: goTypeTestBinder(map[expr.DataType]GoTypeBinding{
			child: {Owner: owner, Type: declaration},
		}),
	})
	require.NoError(t, err)
	return generation, attribute, layout
}

// ComparePackageName orders test declarations independently of registration order.
func (o validationOwnershipOrder) ComparePackageName(other PackageNameOrder) int {
	return cmp.Compare(o, other.(validationOwnershipOrder))
}
