// This file verifies HTTP output packages distinguish declarations by source
// provenance and wire policy while reusing identical emitted shapes.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func TestWireTypeCatalogIdentity(t *testing.T) {
	request := wireTypePolicy{request: true, pointer: true}
	response := wireTypePolicy{useDefault: true}
	first := wireCatalogType("Shared", "same", "first", true)
	second := wireCatalogType("Shared", "same", "second", false)

	catalog := newWireTypeCatalog()
	firstBody := makeHTTPType(&expr.AttributeExpr{Type: first})
	catalog.collect(firstBody, wireRequestBody, request, "")
	catalog.collect(makeHTTPType(&expr.AttributeExpr{Type: first}), wireRequestBody, request, "")
	catalog.collect(makeHTTPType(&expr.AttributeExpr{Type: second}), wireRequestBody, request, "")
	catalog.collect(makeHTTPType(&expr.AttributeExpr{Type: first}), wireResponseBody, response, "")
	catalog.Freeze()
	firstRecord := catalog.lookupUser(firstBody, wireRequestBody, request)
	reusedRecord := catalog.lookupUser(makeHTTPType(&expr.AttributeExpr{Type: first}), wireRequestBody, request)
	secondRecord := catalog.lookupUser(makeHTTPType(&expr.AttributeExpr{Type: second}), wireRequestBody, request)
	responseRecord := catalog.lookupUser(makeHTTPType(&expr.AttributeExpr{Type: first}), wireResponseBody, response)

	require.Same(t, firstRecord, reusedRecord)
	require.Equal(t, "Shared", firstRecord.name)
	require.Equal(t, "Shared2", secondRecord.name)
	require.Equal(t, "Shared3", responseRecord.name)
}

func TestWireTypeCatalogRecursiveIdentityTerminates(t *testing.T) {
	recursive := &expr.UserTypeExpr{TypeName: "Node", UID: "node"}
	object := &expr.Object{}
	recursive.AttributeExpr = &expr.AttributeExpr{Type: object}
	object.Set("next", &expr.AttributeExpr{Type: recursive})

	catalog := newWireTypeCatalog()
	body := makeHTTPType(&expr.AttributeExpr{Type: recursive})
	policy := wireTypePolicy{request: true, pointer: true}
	catalog.collect(body, wireRequestBody, policy, "")
	catalog.Freeze()
	record := catalog.lookupUser(body, wireRequestBody, policy)

	require.Equal(t, "Node", record.name)
	require.Len(t, catalog.records, 1)
}

func TestWireTypeCatalogSeparatesDeclarationIdentityFromValidatorPlacement(t *testing.T) {
	typeAttribute := &expr.AttributeExpr{Type: wireCatalogType("Shared", "shared", "value", true)}
	withoutValidator := wireTypePolicy{pointer: true}
	withValidator := wireTypePolicy{pointer: true, validate: true}
	catalog := newWireTypeCatalog()

	first := catalog.collect(typeAttribute, wireResponseBody, withoutValidator, "")
	second := catalog.collect(expr.DupAtt(typeAttribute), wireAttribute, withValidator, "")

	require.Same(t, first, second)
	catalog.Freeze()
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
	catalog := newWireTypeCatalog()

	catalog.collect(attribute, wireAttribute, wireTypePolicy{}, "")
	require.Len(t, catalog.unionOccurrences, 1)
	catalog.Freeze()
	catalog.applyNames(attribute, wireAttribute, wireTypePolicy{})
	require.Len(t, catalog.unions, 1)
	require.NotNil(t, catalog.unions[0].data)
}

func TestWireTypeCatalogLookupDoesNotDeriveIdentityFromAssignedName(t *testing.T) {
	attribute := &expr.AttributeExpr{Type: wireCatalogType("Shared", "shared", "value", true)}
	policy := wireTypePolicy{pointer: true}
	catalog := newWireTypeCatalog("Shared")
	catalog.collect(attribute, wireAttribute, policy, "")
	catalog.Freeze()

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
	catalog := newWireTypeCatalog()

	catalog.collect(attribute, wireAttribute, wireTypePolicy{}, "")
	catalog.Freeze()
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
	catalog := newWireTypeCatalog()
	catalog.collect(typeAttribute, wireRequestBody, policy, "")
	catalog.Freeze()

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

// wireCatalogType builds an independent authored declaration. Equal UIDs are
// intentional: wire identity follows Origin rather than the example ID.
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
