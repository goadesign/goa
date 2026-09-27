// These tests register only imports written in retained Go type references.
// Imports used by a named definition or a union branch belong to another file.
package codegen

import (
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
	"goa.design/goa/v3/expr"
)

func TestAddTypeReferenceStopsAtDeclarations(t *testing.T) {
	for _, shape := range []string{"named", "union", "array", "map", "struct"} {
		t.Run(shape, func(t *testing.T) {
			generation := mustTestGeneration(t, "generated.local/gen", nil)
			output := mustClaimTestPackage(t, generation, "generated.local/gen/output")
			hiddenPath := "generated.local/gen/a/types"
			parentPath := "generated.local/gen/z/types"
			unionPath := "generated.local/gen/unions"
			hidden := goTypeTestUserType("Hidden", expr.String)
			parent := goTypeTestUserType("Parent", &expr.Object{
				{Name: "hidden", Attribute: &expr.AttributeExpr{Type: hidden}},
			})
			choice := &expr.Union{
				TypeName: "Choice",
				Values: []*expr.NamedAttributeExpr{
					{Name: "hidden", Attribute: &expr.AttributeExpr{Type: hidden}},
					{Name: "text", Attribute: &expr.AttributeExpr{Type: expr.String}},
				},
			}
			hiddenDeclaration := declareGoTypeTestUserType(t, generation, hiddenPath, hidden)
			parentDeclaration := declareGoTypeTestUserType(t, generation, parentPath, parent)
			unionDeclaration := declareGoTypeTestUnion(t, generation, unionPath, choice)
			array := &expr.AttributeExpr{Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: parent}}}
			table := &expr.AttributeExpr{Type: &expr.Map{
				KeyType: &expr.AttributeExpr{
					Type: expr.Int64,
					Meta: expr.MetaExpr{"struct:field:type": {"clock.Duration", "time", "clock"}},
				},
				ElemType: &expr.AttributeExpr{Type: choice},
			}}
			var attribute *expr.AttributeExpr
			var want []string
			switch shape {
			case "named":
				attribute, want = &expr.AttributeExpr{Type: parent}, []string{parentPath}
			case "union":
				attribute, want = &expr.AttributeExpr{Type: choice}, []string{unionPath}
			case "array":
				attribute, want = array, []string{parentPath}
			case "map":
				attribute, want = table, []string{unionPath, "time"}
			case "struct":
				attribute = &expr.AttributeExpr{Type: &expr.Object{
					{Name: "parents", Attribute: array},
					{Name: "choices", Attribute: table},
				}}
				want = []string{unionPath, parentPath, "time"}
			}
			layout, err := PlanGoType(attribute, GoTypePlanOptions{
				Owner: output.ImportPath(), RetainNamedValue: true,
				Policy: GoLayoutPolicy{SumType: true},
				Bind: goTypeTestBinder(map[expr.DataType]GoTypeBinding{
					hidden: {Owner: hiddenPath, Type: hiddenDeclaration, PreferredImportName: "types"},
					parent: {Owner: parentPath, Type: parentDeclaration, PreferredImportName: "types"},
					choice: {Owner: unionPath, Union: unionDeclaration, PreferredImportName: "unions"},
				}),
			})
			require.NoError(t, err)
			imports := NewGeneratedImportPlan(output)
			require.NoError(t, imports.AddTypeReference(layout))
			require.Equal(t, want, imports.Paths())
			require.NoError(t, generation.Freeze())
			require.NoError(t, imports.Link())
			require.Len(t, imports.Imports(), len(want))
			require.NotEmpty(t, layout.Link(output.ImportPath(), output.ImportName).Ref())
			require.PanicsWithValue(t, fmt.Sprintf("import path %q has no planned alias", hiddenPath), func() {
				output.ImportName(hiddenPath)
			})
			if slices.Contains(want, parentPath) {
				require.Equal(t, "types", output.ImportName(parentPath), "hidden fields must not reserve the preferred alias")
			}
		})
	}
}

func TestAddTypeReferencePreservesPrioritiesAndAliases(t *testing.T) {
	for _, customName := range []string{"", "custom"} {
		for _, reverse := range []bool{false, true} {
			generation := mustTestGeneration(t, "generated.local/gen", nil)
			first := mustClaimTestPackage(t, generation, "generated.local/gen/first")
			second := mustClaimTestPackage(t, generation, "generated.local/gen/second")
			require.NoError(t, first.RequireImport(NewImport("types", "encoding/json")))
			require.NoError(t, first.DeclareName(NewExactName(NameType, "types2")))
			leftPath, rightPath := "generated.local/gen/APIKeyService", "generated.local/gen/right/types"
			left, right := goTypeTestUserType("Left", expr.String), goTypeTestUserType("Right", expr.String)
			leftDeclaration := declareGoTypeTestUserType(t, generation, leftPath, left)
			rightDeclaration := declareGoTypeTestUserType(t, generation, rightPath, right)
			qualifier := customName
			if qualifier == "" {
				qualifier = "APIKeyService"
			}
			fields := expr.Object{
				{Name: "left", Attribute: &expr.AttributeExpr{Type: left}},
				{Name: "right", Attribute: &expr.AttributeExpr{Type: right}},
				{Name: "custom", Attribute: &expr.AttributeExpr{
					Type: expr.String,
					Meta: expr.MetaExpr{"struct:field:type": {qualifier + ".Other", leftPath, customName}},
				}},
			}
			if reverse {
				slices.Reverse(fields)
			}
			layout, err := PlanGoType(&expr.AttributeExpr{Type: &fields}, GoTypePlanOptions{
				Owner: first.ImportPath(), RetainNamedValue: true,
				Bind: goTypeTestBinder(map[expr.DataType]GoTypeBinding{
					left:  {Owner: leftPath, Type: leftDeclaration, PreferredImportName: "types"},
					right: {Owner: rightPath, Type: rightDeclaration, PreferredImportName: "types"},
				}),
			})
			require.NoError(t, err)
			firstImports, secondImports := NewGeneratedImportPlan(first), NewGeneratedImportPlan(second)
			require.NoError(t, firstImports.AddTypeReference(layout))
			require.NoError(t, secondImports.AddTypeReference(layout))
			require.Equal(t, []string{leftPath, rightPath}, firstImports.Paths())
			require.Equal(t, firstImports.Paths(), secondImports.Paths())
			require.NoError(t, generation.Freeze())
			require.Equal(t, "types", first.ImportName("encoding/json"))
			require.NotContains(t, []string{"types", "types2"}, first.ImportName(leftPath))
			require.NotEqual(t, first.ImportName(leftPath), first.ImportName(rightPath))
			require.Equal(t, "types", second.ImportName(leftPath), "generated preference wins over custom spelling and path basename")
			require.NotEqual(t, first.ImportName(leftPath), second.ImportName(leftPath))
			for _, field := range layout.Fields() {
				importPath := leftPath
				if field.TypeDeclaration() == rightDeclaration {
					importPath = rightPath
				}
				require.Contains(t, layout.Link(first.ImportPath(), first.ImportName).Enter(field).Def(), first.ImportName(importPath)+".")
			}
		}
	}
}

func TestAddTypeReferenceRejectsInvalidRequests(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	owner := mustClaimTestPackage(t, generation, "generated.local/gen/local")
	output := mustClaimTestPackage(t, generation, "generated.local/gen/output")
	value := goTypeTestUserType("Value", expr.String)
	declaration, err := owner.DeclareUserType(value)
	require.NoError(t, err)
	layout, err := PlanGoType(&expr.AttributeExpr{Type: value}, GoTypePlanOptions{
		Owner: owner.ImportPath(), RetainNamedValue: true,
		Bind: goTypeTestBinder(map[expr.DataType]GoTypeBinding{
			value: {Owner: owner.ImportPath(), Type: declaration},
		}),
	})
	require.NoError(t, err)
	local, imported := NewGeneratedImportPlan(owner), NewGeneratedImportPlan(output)
	require.ErrorContains(t, imported.AddTypeReference(nil), "require a retained Go layout")
	require.NoError(t, local.AddTypeReference(layout))
	require.Empty(t, local.Paths(), "a local reference needs no imported name")
	require.ErrorContains(t, imported.AddTypeReference(layout), "requires a preferred generated package name")
	require.Empty(t, imported.Paths())
	custom, err := PlanGoType(&expr.AttributeExpr{
		Type: expr.String,
		Meta: expr.MetaExpr{"struct:field:type": {"local.Other", owner.ImportPath(), "local"}},
	}, GoTypePlanOptions{Owner: owner.ImportPath()})
	require.NoError(t, err)
	require.NoError(t, local.AddTypeReference(custom))
	require.Empty(t, local.Paths(), "custom self-imports are omitted too")
	require.NoError(t, generation.Freeze())
	require.ErrorContains(t, local.AddTypeReference(layout), "changed after freeze")
	require.ErrorContains(t, local.AddTypeReference(nil), "changed after freeze")
}
