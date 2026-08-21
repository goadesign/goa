// This file verifies that service transformations resolve every named type
// through the frozen package catalog as recursion crosses explicit package
// locations.
package service

import (
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

// TestDeclarationResolverTransformsRelocatedUnionBranches verifies both
// conversion directions use the frozen generated alias in the owning package.
func TestDeclarationResolverTransformsRelocatedUnionBranches(t *testing.T) {
	service := &expr.ServiceExpr{Name: "Convert"}
	generatedBranch := resolverUserType("ValueText", expr.String)
	union := &expr.Union{
		TypeName: "Value",
		Values: []*expr.NamedAttributeExpr{
			{Name: "text", Attribute: &expr.AttributeExpr{Type: generatedBranch}},
		},
	}
	relocated := resolverUserType("Record", &expr.Object{
		{Name: "value", Attribute: &expr.AttributeExpr{Type: union}},
	})
	relocated.Attribute().AddMeta("struct:pkg:path", "types")

	generation := codegen.NewGeneration("generated.local/gen", nil)
	types := generation.GeneratedPackage("generated.local/gen/types")
	_, err := types.DeclareUserType(relocated)
	require.NoError(t, err)
	_, err = types.DeclareUserType(resolverUserType("ValueText", expr.Int))
	require.NoError(t, err)
	_, err = types.DeclareUnion(union)
	require.NoError(t, err)
	branchDeclaration, err := types.DeclareUnionBranchType(union, "text", generatedBranch)
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.Equal(t, "ValueText2", branchDeclaration.Name())

	externalBranch := resolverUserType("ExternalValueText", expr.String)
	externalUnion := &expr.Union{
		TypeName: "ExternalValue",
		Values: []*expr.NamedAttributeExpr{
			{Name: "text", Attribute: &expr.AttributeExpr{Type: externalBranch}},
		},
	}
	external := resolverUserType("ExternalRecord", &expr.Object{
		{Name: "value", Attribute: &expr.AttributeExpr{Type: externalUnion}},
	})

	relocatedAttribute := &expr.AttributeExpr{Type: relocated}
	externalAttribute := &expr.AttributeExpr{Type: external}
	resolver := newServiceResolver(
		generation,
		aliasesForTest(t, "generated.local/gen/types"),
		service,
		"generated.local/gen/types",
	)
	relocatedContext := declarationContext(resolver.Enter(relocatedAttribute), false)
	externalContext := codegen.NewAttributeContext(false, false, true, "external", codegen.NewNameScope())

	_, _, err = codegen.GoTransform(
		relocatedAttribute,
		externalAttribute,
		"record",
		"externalRecord",
		relocatedContext,
		externalContext,
		"convert",
		true,
	)
	require.NoError(t, err)

	toRelocated, toRelocatedHelpers, err := codegen.GoTransform(
		externalAttribute,
		relocatedAttribute,
		"externalRecord",
		"record",
		externalContext,
		relocatedContext,
		"create",
		true,
	)
	require.NoError(t, err)
	require.Contains(t, transformSource(toRelocated, toRelocatedHelpers), "ValueText2")
}

// TestDeclarationResolverQualifiesRelocatedConsumersWithoutRenamingLocalType
// verifies errors and interceptor fields use their actual package owner.
func TestDeclarationResolverQualifiesRelocatedConsumersWithoutRenamingLocalType(t *testing.T) {
	service := &expr.ServiceExpr{Name: "Collisions"}
	local := resolverUserType("Fault", expr.String)
	relocated := resolverUserType("fault", expr.String)
	relocated.Attribute().AddMeta("struct:pkg:path", "errors")
	container := resolverUserType("Container", &expr.Object{
		{Name: "fault", Attribute: &expr.AttributeExpr{Type: relocated}},
	})
	container.Attribute().AddMeta("struct:pkg:path", "types")

	generation := codegen.NewGeneration("generated.local/gen", nil)
	servicePackage := generation.GeneratedPackage(servicePackagePath(generation.GenPkg(), service))
	localDeclaration, err := servicePackage.DeclareUserType(local)
	require.NoError(t, err)
	errorsPackage := generation.GeneratedPackage("generated.local/gen/errors")
	_, err = errorsPackage.DeclareUserType(relocated)
	require.NoError(t, err)
	typesPackage := generation.GeneratedPackage("generated.local/gen/types")
	_, err = typesPackage.DeclareUserType(container)
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())

	resolver := newServiceResolver(
		generation,
		aliasesForTest(
			t,
			servicePackagePath(generation.GenPkg(), service),
			"generated.local/gen/errors",
			"generated.local/gen/types",
		),
		service,
		servicePackagePath(generation.GenPkg(), service),
	)
	require.Equal(t, "Fault", localDeclaration.Name())
	require.Equal(t, "Fault", resolver.Ref(&expr.AttributeExpr{Type: local}, ""))

	errorData := buildErrorInitData(&expr.ErrorExpr{
		AttributeExpr: &expr.AttributeExpr{Type: relocated},
		Name:          "fault",
	}, resolver)
	require.Equal(t, "errors_.Fault", errorData.TypeName)
	require.Equal(t, "errors_.Fault", errorData.TypeRef)

	attributes := collectAttributes(
		&expr.AttributeExpr{Type: &expr.Object{
			{Name: "fault", Attribute: &expr.AttributeExpr{Type: expr.String}},
		}},
		&expr.AttributeExpr{Type: container},
		resolver,
	)
	require.Equal(t, "errors_.Fault", attributes[0].TypeRef)
}

// TestDeclarationResolverPanicsWhenPlanOmittedType verifies render analysis
// fails immediately instead of allocating a missing declaration.
func TestDeclarationResolverPanicsWhenPlanOmittedType(t *testing.T) {
	service := &expr.ServiceExpr{Name: "Missing"}
	generation := codegen.NewGeneration("generated.local/gen", nil)
	generation.GeneratedPackage(servicePackagePath(generation.GenPkg(), service))
	require.NoError(t, generation.Freeze())
	resolver := newServiceResolver(
		generation,
		aliasesForTest(t, servicePackagePath(generation.GenPkg(), service)),
		service,
		servicePackagePath(generation.GenPkg(), service),
	)
	missing := resolverUserType("Missing", expr.String)
	require.PanicsWithValue(
		t,
		"resolve user type \"Missing\" for service \"Missing\" in package \"generated.local/gen/missing\": user type \"Missing\" has no declaration in generated package \"generated.local/gen/missing\"",
		func() {
			resolver.Name(&expr.AttributeExpr{Type: missing}, "", false, true)
		},
	)
}

// TestServicesDataServiceAttributorUsesFrozenPackageDeclarations verifies
// transport generators can consume the same local, relocated, and nested
// declaration records used by service rendering without accessing resolver
// state.
func TestServicesDataServiceAttributorUsesFrozenPackageDeclarations(t *testing.T) {
	var record expr.UserType
	root := codegen.RunDSL(t, func() {
		record = dsl.Type("Record", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.OneOf("Value", func() {
				dsl.Attribute("text", dsl.String)
			})
			dsl.Attribute("external", dsl.String, func() {
				dsl.Meta("struct:field:type", "custom.Value", "example.com/custom", "custom")
			})
		})
		dsl.Service("Values", func() {
			dsl.Method("Read", func() {
				dsl.Payload(record)
			})
		})
	})
	services := mustServicesData(t, root)
	external := services.ServiceAttributor("Values", "example.com/consumer")
	recordAttribute := &expr.AttributeExpr{Type: record}
	recordResolver := external.Enter(recordAttribute)
	value := expr.AsObject(record.Attribute().Type).Attribute("Value")
	externalValue := expr.AsObject(record.Attribute().Type).Attribute("external")

	require.Equal(t, "*types.Record", external.Ref(recordAttribute, ""))
	require.Equal(t, "*types.Value", recordResolver.Ref(value, ""))
	require.Equal(t, "custom.Value", recordResolver.Ref(externalValue, ""))

	typesPackage := "goa.design/goa/example/types"
	local := services.ServiceAttributor("Values", typesPackage).Enter(recordAttribute)
	require.Equal(t, "*Record", local.Ref(recordAttribute, ""))
	require.Equal(t, "*Value", local.Ref(value, ""))
}

// aliasesForTest builds the same frozen full-path qualifier table used by
// service analysis for the package paths exercised by a focused resolver test.
func aliasesForTest(t *testing.T, paths ...string) *importAliases {
	t.Helper()
	generation := codegen.NewGeneration("generated.local/gen", nil)
	for _, importPath := range paths {
		require.NoError(t, generation.DeclareImport(codegen.NewImport(codegen.Goify(path.Base(importPath), false), importPath)))
	}
	require.NoError(t, generation.Freeze())
	return &importAliases{generation: generation}
}

// resolverUserType constructs one exact declaration for resolver tests.
func resolverUserType(name string, dataType expr.DataType) *expr.UserTypeExpr {
	return &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{Type: dataType},
		TypeName:      name,
		UID:           "resolver-test#" + name,
	}
}

// transformSource combines an inline transformation with every recursive
// helper so tests can assert the complete code emitted for one conversion.
func transformSource(code string, helpers []*codegen.TransformFunctionData) string {
	var source strings.Builder
	source.WriteString(code)
	for _, helper := range helpers {
		source.WriteString(helper.Code)
	}
	return source.String()
}
