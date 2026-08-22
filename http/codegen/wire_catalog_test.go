// This file verifies copied HTTP request and response types receive the right
// Go names when their source types, field rules, or generated functions differ.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

func TestWireTypeCatalogIdentity(t *testing.T) {
	request := wireTypePolicy{request: true, pointer: true}
	response := wireTypePolicy{useDefault: true}
	first := wireCatalogType("Shared", "same", "first", true)
	second := wireCatalogType("Shared", "same", "second", false)

	catalog, generation := testWireTypeCatalog(t)
	firstBody := makeHTTPType(&expr.AttributeExpr{Type: first})
	catalog.collect(firstBody, wireRequestBody, request, "")
	catalog.collect(makeHTTPType(&expr.AttributeExpr{Type: first}), wireRequestBody, request, "")
	catalog.collect(makeHTTPType(&expr.AttributeExpr{Type: second}), wireRequestBody, request, "")
	catalog.collect(makeHTTPType(&expr.AttributeExpr{Type: first}), wireResponseBody, response, "")
	linkTestWireTypeCatalog(t, generation, catalog)
	firstRecord := catalog.lookupUser(firstBody, wireRequestBody, request)
	reusedRecord := catalog.lookupUser(makeHTTPType(&expr.AttributeExpr{Type: first}), wireRequestBody, request)
	secondRecord := catalog.lookupUser(makeHTTPType(&expr.AttributeExpr{Type: second}), wireRequestBody, request)
	responseRecord := catalog.lookupUser(makeHTTPType(&expr.AttributeExpr{Type: first}), wireResponseBody, response)

	require.Same(t, firstRecord, reusedRecord)
	require.Len(t, map[string]struct{}{
		firstRecord.name: {}, secondRecord.name: {}, responseRecord.name: {},
	}, 3)
}

func TestWireTypeCatalogRecursiveIdentityTerminates(t *testing.T) {
	recursive := &expr.UserTypeExpr{TypeName: "Node", UID: "node"}
	object := &expr.Object{}
	recursive.AttributeExpr = &expr.AttributeExpr{Type: object}
	object.Set("next", &expr.AttributeExpr{Type: recursive})

	catalog, generation := testWireTypeCatalog(t)
	body := makeHTTPType(&expr.AttributeExpr{Type: recursive})
	policy := wireTypePolicy{request: true, pointer: true}
	catalog.collect(body, wireRequestBody, policy, "")
	linkTestWireTypeCatalog(t, generation, catalog)
	record := catalog.lookupUser(body, wireRequestBody, policy)

	require.Equal(t, "Node", record.name)
	require.Len(t, catalog.records, 1)
}

func TestWireTypeCatalogSeparatesDeclarationIdentityFromValidatorPlacement(t *testing.T) {
	typeAttribute := &expr.AttributeExpr{Type: wireCatalogType("Shared", "shared", "value", true)}
	withoutValidator := wireTypePolicy{pointer: true}
	withValidator := wireTypePolicy{pointer: true, validate: true}
	catalog, generation := testWireTypeCatalog(t)

	first := catalog.collect(typeAttribute, wireResponseBody, withoutValidator, "")
	second := catalog.collect(expr.DupAtt(typeAttribute), wireAttribute, withValidator, "")

	require.Same(t, first, second)
	linkTestWireTypeCatalog(t, generation, catalog)
	catalog.bind(first, &TypeData{Def: "struct { Value string }"})
	catalog.bind(second, &TypeData{Def: "struct { Value string }", ValidateDef: "validate shared"})
	require.Equal(t, "Shared", first.name)
	require.Equal(t, "validate shared", first.data.ValidateDef)
}

func TestWireTypeCatalogCollectsUnionsBeforeFreeze(t *testing.T) {
	union := &expr.Union{
		TypeName: "Choice",
		Values: []*expr.NamedAttributeExpr{
			{Name: "text", Attribute: &expr.AttributeExpr{Type: expr.String}},
			{Name: "count", Attribute: &expr.AttributeExpr{Type: expr.Int}},
		},
	}
	attribute := &expr.AttributeExpr{Type: union}
	catalog, generation := testWireTypeCatalog(t)

	catalog.collect(attribute, wireAttribute, wireTypePolicy{}, "")
	require.Len(t, catalog.unionOccurrences, 1)
	linkTestWireTypeCatalog(t, generation, catalog)
	catalog.applyNames(attribute, wireAttribute, wireTypePolicy{})
	require.Len(t, catalog.unions, 1)
	require.NotNil(t, catalog.unions[0].data)
}

func TestWireTypeCatalogLookupDoesNotDeriveIdentityFromAssignedName(t *testing.T) {
	attribute := &expr.AttributeExpr{Type: wireCatalogType("Shared", "shared", "value", true)}
	policy := wireTypePolicy{pointer: true}
	catalog, generation := testWireTypeCatalog(t, "Shared")
	catalog.collect(attribute, wireAttribute, policy, "")
	linkTestWireTypeCatalog(t, generation, catalog)

	first := catalog.lookupUser(attribute, wireAttribute, policy)
	second := catalog.lookupUser(attribute, wireAttribute, policy)

	require.Same(t, first, second)
	require.Equal(t, "Shared2", first.name)
}

func TestWireTypeCatalogDoesNotNameTheSharedEmptySentinel(t *testing.T) {
	originalNil := expr.Empty.Attribute().Meta == nil
	original := expr.Empty.Attribute().Meta.Dup()
	attribute := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "empty", Attribute: &expr.AttributeExpr{Type: expr.Empty}},
	}}
	catalog, generation := testWireTypeCatalog(t)

	catalog.collect(attribute, wireAttribute, wireTypePolicy{}, "")
	linkTestWireTypeCatalog(t, generation, catalog)
	catalog.applyNames(attribute, wireAttribute, wireTypePolicy{})

	if originalNil {
		require.Nil(t, expr.Empty.Attribute().Meta)
	} else {
		require.Equal(t, original, expr.Empty.Attribute().Meta)
	}
}

func TestWireTypeCatalogRejectsLateAndUnknownDeclarations(t *testing.T) {
	typeAttribute := &expr.AttributeExpr{Type: wireCatalogType("Known", "known", "value", true)}
	policy := wireTypePolicy{request: true, pointer: true}
	catalog, generation := testWireTypeCatalog(t)
	catalog.collect(typeAttribute, wireRequestBody, policy, "")
	linkTestWireTypeCatalog(t, generation, catalog)

	require.Panics(t, func() {
		catalog.collect(&expr.AttributeExpr{Type: wireCatalogType("Late", "late", "value", true)}, wireRequestBody, policy, "")
	})
	require.Panics(t, func() {
		catalog.lookupUser(&expr.AttributeExpr{Type: wireCatalogType("Unknown", "unknown", "value", true)}, wireRequestBody, policy)
	})
	require.Panics(t, func() {
		catalog.scope.Unique("late")
	})
}

// testWireTypeCatalog creates the generated package that assigns names for a test.
// Reserved names simulate declarations contributed by another generator.
func testWireTypeCatalog(t *testing.T, reserved ...string) (*wireTypeCatalog, *codegen.Generation) {
	t.Helper()
	generation, err := codegen.NewGeneration("generated.local/gen", nil)
	require.NoError(t, err)
	pkg, err := generation.ClaimPackage("generated.local/gen/http/test/client")
	require.NoError(t, err)
	for _, name := range reserved {
		require.NoError(t, pkg.DeclareName(codegen.NewExactName(codegen.NameVariable, name)))
	}
	return newWireTypeCatalog(pkg), generation
}

// linkTestWireTypeCatalog submits the collected declarations, asks the
// generation to assign all package names, and makes those names available to
// the test.
func linkTestWireTypeCatalog(t *testing.T, generation *codegen.Generation, catalog *wireTypeCatalog) {
	t.Helper()
	require.NoError(t, catalog.Declare())
	require.NoError(t, generation.Freeze())
	catalog.Link()
}

// wireCatalogType builds an independent declared type. Equal UIDs are
// intentional because the original declared type, not the example ID, selects
// the copied HTTP type.
func wireCatalogType(name, uid, field string, required bool) *expr.UserTypeExpr {
	attribute := &expr.AttributeExpr{Type: expr.String}
	attribute.Validation = &expr.ValidationExpr{Pattern: field}
	object := &expr.Object{{Name: field, Attribute: attribute}}
	if required {
		objectAttribute := &expr.AttributeExpr{Type: object, Validation: &expr.ValidationExpr{Required: []string{field}}}
		return &expr.UserTypeExpr{AttributeExpr: objectAttribute, TypeName: name, UID: uid}
	}
	return &expr.UserTypeExpr{AttributeExpr: &expr.AttributeExpr{Type: object}, TypeName: name, UID: uid}
}
