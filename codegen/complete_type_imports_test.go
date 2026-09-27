// These tests keep complete-layout import registration inside the compiler.
// A path can carry both generated and custom requests with different priorities.
package codegen

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
	"goa.design/goa/v3/expr"
)

func TestAddCompleteTypePreservesRequestPriority(t *testing.T) {
	for _, customName := range []string{"", "custom"} {
		for _, reverse := range []bool{false, true} {
			generation := mustTestGeneration(t, "generated.local/gen", nil)
			output := mustClaimTestPackage(t, generation, "generated.local/gen/output")
			owner := "generated.local/gen/shared"
			value := goTypeTestUserType("Value", expr.String)
			declaration := declareGoTypeTestUserType(t, generation, owner, value)
			qualifier := customName
			if qualifier == "" {
				qualifier = "shared"
			}
			fields := expr.Object{
				{Name: "generated", Attribute: &expr.AttributeExpr{Type: value}},
				{Name: "custom", Attribute: &expr.AttributeExpr{Type: expr.String,
					Meta: expr.MetaExpr{"struct:field:type": {qualifier + ".Other", owner, customName}}}},
			}
			if reverse {
				slices.Reverse(fields)
			}
			layout, err := PlanGoType(&expr.AttributeExpr{Type: &fields}, GoTypePlanOptions{
				Owner: output.ImportPath(), RetainNamedValue: true,
				Bind: goTypeTestBinder(map[expr.DataType]GoTypeBinding{
					value: {Owner: owner, Type: declaration, PreferredImportName: "shared"},
				}),
			})
			require.NoError(t, err)
			imports := NewGeneratedImportPlan(output)
			require.NoError(t, imports.AddCompleteType(layout))
			require.Equal(t, []string{owner}, imports.Paths())
			require.NoError(t, generation.Freeze())
			require.NoError(t, imports.Link())
			require.Equal(t, "shared", output.ImportName(owner))
			for _, field := range layout.Fields() {
				require.Contains(t, layout.Link(output.ImportPath(), output.ImportName).Enter(field).Def(), "shared.")
			}
		}
	}
}

func TestAddCompleteTypeKeepsAliasesPerOutput(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		generation := mustTestGeneration(t, "generated.local/gen", nil)
		first := mustClaimTestPackage(t, generation, "generated.local/gen/first")
		second := mustClaimTestPackage(t, generation, "generated.local/gen/second")
		require.NoError(t, first.RequireImport(NewImport("types", "encoding/json")))
		require.NoError(t, first.DeclareName(NewExactName(NameType, "types2")))
		left, right := goTypeTestUserType("Left", expr.String), goTypeTestUserType("Right", expr.String)
		leftPath, rightPath := "generated.local/gen/left/types", "generated.local/gen/right/types"
		leftDecl := declareGoTypeTestUserType(t, generation, leftPath, left)
		rightDecl := declareGoTypeTestUserType(t, generation, rightPath, right)
		fields := expr.Object{
			{Name: "left", Attribute: &expr.AttributeExpr{Type: left}},
			{Name: "right", Attribute: &expr.AttributeExpr{Type: right}},
		}
		if reverse {
			slices.Reverse(fields)
		}
		layout, err := PlanGoType(&expr.AttributeExpr{Type: &fields}, GoTypePlanOptions{
			Owner: first.ImportPath(), RetainNamedValue: true,
			Bind: goTypeTestBinder(map[expr.DataType]GoTypeBinding{
				left:  {Owner: leftPath, Type: leftDecl, PreferredImportName: "types"},
				right: {Owner: rightPath, Type: rightDecl, PreferredImportName: "types"},
			}),
		})
		require.NoError(t, err)
		firstImports, secondImports := NewGeneratedImportPlan(first), NewGeneratedImportPlan(second)
		require.NoError(t, firstImports.AddCompleteType(layout))
		require.NoError(t, secondImports.AddCompleteType(layout))
		require.NoError(t, generation.Freeze())
		require.Equal(t, "types", first.ImportName("encoding/json"))
		require.NotEqual(t, first.ImportName(leftPath), first.ImportName(rightPath))
		for _, path := range []string{leftPath, rightPath} {
			require.NotContains(t, []string{"types", "types2"}, first.ImportName(path))
		}
		require.Equal(t, "types", second.ImportName(leftPath))
		require.NotEqual(t, first.ImportName(leftPath), second.ImportName(leftPath))
	}
}

func TestAddCompleteTypeRequiresImportedNameBeforeFreeze(t *testing.T) {
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
	local := NewGeneratedImportPlan(owner)
	require.NoError(t, local.AddCompleteType(layout), "local representation bindings need no imported name")
	require.Empty(t, local.Paths())
	imported := NewGeneratedImportPlan(output)
	require.ErrorContains(t, imported.AddCompleteType(layout), "requires a preferred generated package name")
	require.Empty(t, imported.Paths())
	require.NoError(t, generation.Freeze())
	require.ErrorContains(t, local.AddCompleteType(layout), "changed after freeze")
}

func TestAddCompleteTypeTraversesNamedDefinitions(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	output := mustClaimTestPackage(t, generation, "generated.local/gen/isolated")
	childPath := "generated.local/gen/APIKeyService"
	child := goTypeTestUserType("Child", expr.String)
	childDecl := declareGoTypeTestUserType(t, generation, childPath, child)
	parentPath := "generated.local/gen/parent"
	parent := goTypeTestUserType("Parent", &expr.Object{
		{Name: "child", Attribute: &expr.AttributeExpr{Type: child}},
		{Name: "children", Attribute: &expr.AttributeExpr{Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: child}}}},
	})
	parentDecl := declareGoTypeTestUserType(t, generation, parentPath, parent)
	layout, err := PlanGoType(&expr.AttributeExpr{Type: parent}, GoTypePlanOptions{
		Owner: parentPath, RetainNamedValue: true,
		Bind: goTypeTestBinder(map[expr.DataType]GoTypeBinding{
			parent: {Owner: parentPath, Type: parentDecl, PreferredImportName: "parent"},
			child:  {Owner: childPath, Type: childDecl, PreferredImportName: "apikeyservice"},
		}),
	})
	require.NoError(t, err)
	imports := NewGeneratedImportPlan(output)
	require.NoError(t, imports.AddCompleteType(layout))
	require.Equal(t, []string{childPath, parentPath}, imports.Paths())
	require.Equal(t, []GoTypeImport{{Path: parentPath}}, layout.ImportPreferences(), "informational API stays compatible")
	require.NoError(t, generation.Freeze())
	require.Equal(t, "apikeyservice", output.ImportName(childPath))
}
