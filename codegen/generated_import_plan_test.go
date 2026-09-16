package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

// TestGeneratedImportPlanLinkExactAliases verifies that one contribution records all
// package paths before freeze and reads the same selected names afterward.
func TestGeneratedImportPlanLinkExactAliases(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	pkg := mustClaimTestPackage(t, generation, "generated.local/gen/service")
	imports := NewGeneratedImportPlan(pkg)
	require.NoError(t, imports.Require(SimpleImport("strconv")))

	custom := &expr.AttributeExpr{
		Type: expr.String,
		Meta: expr.MetaExpr{
			"struct:field:type": {"strconv.Token", "generated.local/custom/strconv", "strconv"},
		},
	}
	relocated := &expr.UserTypeExpr{
		TypeName: "Record",
		AttributeExpr: &expr.AttributeExpr{
			Type: expr.String,
			Meta: expr.MetaExpr{"struct:pkg:path": {"types"}},
		},
	}
	references := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "custom", Attribute: custom},
		{Name: "relocated", Attribute: &expr.AttributeExpr{Type: relocated}},
	}}
	require.NoError(t, imports.AddRecursiveTypeReferences(references))
	require.Equal(t, []string{
		"generated.local/custom/strconv",
		"generated.local/gen/types",
		"strconv",
	}, imports.Paths())
	require.NoError(t, generation.Freeze())
	require.NoError(t, imports.Link())

	require.Equal(t, []*ImportSpec{
		{Name: "strconv2", Path: "generated.local/custom/strconv"},
		{Name: "types", Path: "generated.local/gen/types"},
		{Path: "strconv"},
	}, imports.Imports())
	require.Equal(t, "strconv2.Token", pkg.Scope().GoTypeName(custom))
	require.Equal(t, "types", pkg.ImportName("generated.local/gen/types"))

	copy := imports.Imports()
	copy[0].Path = "changed"
	require.Equal(t, "generated.local/custom/strconv", imports.Imports()[0].Path)
	paths := imports.Paths()
	paths[0] = "changed"
	require.Equal(t, "generated.local/custom/strconv", imports.Paths()[0])
}

// TestGeneratedImportPlanRejectsEarlyLink verifies that aliases cannot be
// read while another import can still change their names.
func TestGeneratedImportPlanRejectsEarlyLink(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	pkg := mustClaimTestPackage(t, generation, "generated.local/gen/service")
	imports := NewGeneratedImportPlan(pkg)
	require.NoError(t, imports.Require(SimpleImport("fmt")))
	require.EqualError(t, imports.Link(), "generated import plan for \"generated.local/gen/service\" linked before freeze")
}

// TestGeneratedImportPlanSeparatesWrittenTypesFromNestedFields verifies that
// signatures do not import packages used only inside a named type.
func TestGeneratedImportPlanSeparatesWrittenTypesFromNestedFields(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	pkg := mustClaimTestPackage(t, generation, "generated.local/gen/service")
	named := &expr.UserTypeExpr{
		TypeName: "Record",
		AttributeExpr: &expr.AttributeExpr{
			Type: &expr.Object{{
				Name: "token",
				Attribute: &expr.AttributeExpr{
					Type: expr.String,
					Meta: expr.MetaExpr{
						"struct:field:type": {"wire.Token", "generated.local/wire", "wire"},
					},
				},
			}},
			Meta: expr.MetaExpr{"struct:pkg:path": {"types"}},
		},
	}
	attribute := &expr.AttributeExpr{Type: named}
	shallow := NewGeneratedImportPlan(pkg)
	recursive := NewGeneratedImportPlan(pkg)
	require.NoError(t, shallow.AddTypeExpressions(attribute))
	require.NoError(t, recursive.AddRecursiveTypeReferences(attribute))
	require.Equal(t, []string{"generated.local/gen/types"}, shallow.Paths())
	require.Equal(t, []string{"generated.local/gen/types", "generated.local/wire"}, recursive.Paths())
}

// TestGeneratedImportPlanTraversesEveryTypeCopy verifies that copies of one
// authored type can name different packages in their nested generated fields.
func TestGeneratedImportPlanTraversesEveryTypeCopy(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	pkg := mustClaimTestPackage(t, generation, "generated.local/gen/service")
	origin := &expr.UserTypeExpr{
		TypeName: "Record",
		AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{{
			Name:      "token",
			Attribute: &expr.AttributeExpr{Type: expr.String},
		}}},
	}
	copyWithPackage := func(name, importPath string) expr.UserType {
		attribute := expr.DupAtt(origin.Attribute())
		expr.AsObject(attribute.Type).Attribute("token").AddMeta(
			"struct:field:type",
			name+".Token",
			importPath,
			name,
		)
		return origin.Dup(attribute)
	}
	attribute := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "first", Attribute: &expr.AttributeExpr{Type: copyWithPackage("first", "generated.local/first")}},
		{Name: "second", Attribute: &expr.AttributeExpr{Type: copyWithPackage("second", "generated.local/second")}},
	}}
	imports := NewGeneratedImportPlan(pkg)
	require.NoError(t, imports.AddRecursiveTypeReferences(attribute))
	require.Equal(t, []string{"generated.local/first", "generated.local/second"}, imports.Paths())
}

// TestGeneratedImportPlanIncludesBuiltInServiceError verifies that signatures
// using Goa's standard error reserve the package that defines ServiceError.
func TestGeneratedImportPlanIncludesBuiltInServiceError(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	pkg := mustClaimTestPackage(t, generation, "generated.local/gen/service")
	imports := NewGeneratedImportPlan(pkg)
	attribute := expr.DupAtt(&expr.AttributeExpr{Type: expr.ErrorResult})

	require.NoError(t, imports.AddTypeExpressions(attribute))
	require.Equal(t, []string{GoaImport("").Path}, imports.Paths())
}

// TestGeneratedImportPlanRejectsLateChangesAndRepeatedLink verifies that the
// planning and linking phases each run exactly once.
func TestGeneratedImportPlanRejectsLateChangesAndRepeatedLink(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	pkg := mustClaimTestPackage(t, generation, "generated.local/gen/service")
	imports := NewGeneratedImportPlan(pkg)
	require.NoError(t, imports.Require(SimpleImport("fmt")))
	require.NoError(t, generation.Freeze())
	require.EqualError(t, imports.Require(), "generated import plan for \"generated.local/gen/service\" changed after freeze")
	require.EqualError(t, imports.AddTypeExpressions(), "generated import plan for \"generated.local/gen/service\" changed after freeze")
	require.NoError(t, imports.Link())
	require.EqualError(t, imports.Link(), "generated import plan for \"generated.local/gen/service\" linked more than once")
}
