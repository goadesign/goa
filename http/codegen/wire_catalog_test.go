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
	catalog.collect(firstBody, wireRequestBody, request)
	catalog.collect(makeHTTPType(&expr.AttributeExpr{Type: first}), wireRequestBody, request)
	catalog.collect(makeHTTPType(&expr.AttributeExpr{Type: second}), wireRequestBody, request)
	catalog.collect(makeHTTPType(&expr.AttributeExpr{Type: first}), wireResponseBody, response)
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
	catalog.collect(body, wireRequestBody, policy)
	linkTestWireTypeCatalog(t, generation, catalog)
	record := catalog.lookupUser(body, wireRequestBody, policy)

	require.Equal(t, "Node", record.name)
	require.Len(t, catalog.records, 1)
}

func TestOptionalWireTypeRefUsesLinkedLayout(t *testing.T) {
	catalog, generation := testWireTypeCatalog(t)
	linkTestWireTypeCatalog(t, generation, catalog)
	resolver := catalog.resolver(catalog.scope, wireTypePolicy{}).(*wireAttributeScope)

	tests := []struct {
		name        string
		attribute   *expr.AttributeExpr
		want        string
		preserve    bool
		dereference bool
	}{
		{
			name:      "anonymous object",
			attribute: &expr.AttributeExpr{Type: &expr.Object{}},
			want:      "*struct {\n}",
			preserve:  true,
		},
		{
			name:        "primitive",
			attribute:   &expr.AttributeExpr{Type: expr.String},
			want:        "*string",
			preserve:    true,
			dereference: true,
		},
		{
			name:      "array already preserves absence",
			attribute: &expr.AttributeExpr{Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.String}}},
			want:      "[]string",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			layout, err := resolver.GoTypeLayout(test.attribute, wireGoLayoutPolicy(wireTypePolicy{}))
			require.NoError(t, err)
			ref, preserve, dereference := optionalWireTypeRef(layout, true)
			require.Equal(t, test.want, ref)
			require.Equal(t, test.preserve, preserve)
			require.Equal(t, test.dereference, dereference)
		})
	}
}

func TestWireTypeCatalogPreservesReleasedNestedNames(t *testing.T) {
	cases := []struct {
		name string
		role wireTypeRole
		body *expr.AttributeExpr
		want string
	}{
		{
			name: "request body",
			role: wireRequestBody,
			body: wireCatalogContainer(wireCatalogType("Child", "request-child", "value", true)),
			want: "ChildRequestBody",
		},
		{
			name: "streaming body",
			role: wireStreamPayload,
			body: wireCatalogContainer(wireCatalogType("Child", "stream-child", "value", true)),
			want: "ChildStreamingBody",
		},
		{
			name: "object response body",
			role: wireResponseBody,
			body: wireCatalogContainer(wireCatalogType("Child", "response-child", "value", true)),
			want: "ChildResponseBody",
		},
		{
			name: "collection response body",
			role: wireResponseBody,
			body: &expr.AttributeExpr{Type: &expr.Array{ElemType: &expr.AttributeExpr{
				Type: wireCatalogType("Child", "response-element", "value", true),
			}}},
			want: "ChildResponse",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			catalog, generation := testWireTypeCatalog(t)
			catalog.collect(test.body, test.role, wireTypePolicy{})
			linkTestWireTypeCatalog(t, generation, catalog)

			child := firstWireUserType(test.body)
			record := catalog.lookupUser(child, wireAttribute, wireTypePolicy{})
			require.Equal(t, test.want, record.name)
		})
	}
}

func TestWireTypeCatalogKeepsCurrentNameForSharedReleasedDeclarations(t *testing.T) {
	child := wireCatalogType("Shared", "shared", "value", true)
	request := wireCatalogContainer(child)
	stream := wireCatalogContainer(child)
	catalog, generation := testWireTypeCatalog(t)

	catalog.collect(request, wireRequestBody, wireTypePolicy{})
	catalog.collect(stream, wireStreamPayload, wireTypePolicy{})
	linkTestWireTypeCatalog(t, generation, catalog)

	record := catalog.lookupUser(firstWireUserType(request), wireAttribute, wireTypePolicy{})
	require.Equal(t, "Shared", record.name)
}

func TestWireTypeCatalogSuffixesReleasedNameAfterPackageCollision(t *testing.T) {
	body := wireCatalogContainer(wireCatalogType("Child", "child", "value", true))
	catalog, generation := testWireTypeCatalog(t, "ChildRequestBody")
	catalog.collect(body, wireRequestBody, wireTypePolicy{})
	linkTestWireTypeCatalog(t, generation, catalog)

	record := catalog.lookupUser(firstWireUserType(body), wireAttribute, wireTypePolicy{})
	require.Equal(t, "ChildRequestBody2", record.name)
}

func TestWireTypeCatalogSeparatesDeclarationIdentityFromValidatorPlacement(t *testing.T) {
	typeAttribute := &expr.AttributeExpr{Type: wireCatalogType("Shared", "shared", "value", true)}
	withoutValidator := wireTypePolicy{pointer: true}
	withValidator := wireTypePolicy{pointer: true, validate: true}
	catalog, generation := testWireTypeCatalog(t)

	first := catalog.collect(typeAttribute, wireResponseBody, withoutValidator)
	second := catalog.collect(expr.DupAtt(typeAttribute), wireAttribute, withValidator)
	catalog.addValidationRoot(&expr.AttributeExpr{Type: &expr.Object{
		{Name: "shared", Attribute: expr.DupAtt(typeAttribute)},
	}}, withValidator)

	require.Same(t, first, second)
	linkTestWireTypeCatalog(t, generation, catalog)
	catalog.bind(first, &TypeData{Def: "struct { Value string }"})
	catalog.bind(second, &TypeData{
		Def:               "struct { Value string }",
		ValidateDef:       "validate shared from body",
		NestedValidateDef: "validate shared from parent path",
	})
	require.Equal(t, "Shared", first.name)
	require.Equal(t, "validate shared from body", first.data.ValidateDef)
	require.Equal(t, "validate shared from parent path", first.data.NestedValidateDef)
	require.Equal(t, "ValidateShared", first.data.ValidatorName)
	require.Equal(t, "validateShared", first.data.NestedValidatorName)
}

func TestWireTypeCatalogErrorDescriptionUsesAllPlannedErrors(t *testing.T) {
	cases := []struct {
		name string
		uses []wireErrorUse
		want string
	}{
		{
			name: "same endpoint",
			uses: []wireErrorUse{
				{service: "Calc", method: "Add", name: "underflow"},
				{service: "Calc", method: "Add", name: "overflow"},
			},
			want: "Shared is the type of the \"Calc\" service \"Add\" endpoint HTTP response body for the \"overflow\" and \"underflow\" errors.",
		},
		{
			name: "several endpoints",
			uses: []wireErrorUse{
				{service: "Beta", method: "Write", name: "conflict"},
				{service: "Alpha", method: "Read", name: "missing"},
			},
			want: "Shared is the HTTP response body type for these service errors:\n" +
				"- \"Alpha\" service \"Read\" endpoint: \"missing\" error\n" +
				"- \"Beta\" service \"Write\" endpoint: \"conflict\" error",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			attribute := &expr.AttributeExpr{Type: wireCatalogType("Shared", "shared", "value", true)}
			policy := wireTypePolicy{pointer: true}
			catalog, generation := testWireTypeCatalog(t)
			record := catalog.collect(attribute, wireResponseBody, policy)
			for _, use := range test.uses {
				record.addErrorUse(use)
			}
			record.addErrorUse(test.uses[0])

			linkTestWireTypeCatalog(t, generation, catalog)
			catalog.bind(record, &TypeData{
				Description: "Shared is an HTTP response body.",
				Def:         "struct { Value string }",
			})

			require.Equal(t, test.want, record.data.Description)
		})
	}
}

func TestWireTypeCatalogPlansNestedValidatorNameWithPackageNames(t *testing.T) {
	attribute := &expr.AttributeExpr{Type: wireCatalogType("Shared", "shared", "value", true)}
	policy := wireTypePolicy{pointer: true, validate: true}
	catalog, generation := testWireTypeCatalog(t, "validateShared")
	record := catalog.collect(attribute, wireAttribute, policy)
	catalog.addValidationRoot(&expr.AttributeExpr{Type: &expr.Object{
		{Name: "shared", Attribute: expr.DupAtt(attribute)},
	}}, policy)

	linkTestWireTypeCatalog(t, generation, catalog)
	catalog.bind(record, &TypeData{
		ValidateDef:       "validate shared from body",
		NestedValidateDef: "validate shared from parent path",
	})

	require.Equal(t, "validateShared2", record.data.NestedValidatorName)
}

func TestWireTypeCatalogDoesNotRewriteValidationCalls(t *testing.T) {
	attribute := &expr.AttributeExpr{Type: wireCatalogType("Shared", "shared", "value", true)}
	policy := wireTypePolicy{pointer: true, validate: true}
	catalog, generation := testWireTypeCatalog(t, "ValidateShared")
	record := catalog.collect(attribute, wireAttribute, policy)
	linkTestWireTypeCatalog(t, generation, catalog)

	catalog.bind(record, &TypeData{
		ValidateDef: "validate shared from body",
		ValidateRef: "err = ValidateSharedCopy(v)",
	})

	require.Equal(t, "ValidateShared2", record.data.ValidatorName)
	require.Equal(t, "err = ValidateSharedCopy(v)", record.data.ValidateRef)
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

	catalog.collect(attribute, wireAttribute, wireTypePolicy{})
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
	catalog.collect(attribute, wireAttribute, policy)
	linkTestWireTypeCatalog(t, generation, catalog)

	first := catalog.lookupUser(attribute, wireAttribute, policy)
	second := catalog.lookupUser(attribute, wireAttribute, policy)

	require.Same(t, first, second)
	require.Equal(t, "Shared2", first.name)
}

func TestWireTypeCatalogBindingUsesCurrentLayoutPolicy(t *testing.T) {
	attribute := &expr.AttributeExpr{Type: wireCatalogType("Shared", "shared", "value", true)}
	valuePolicy := wireTypePolicy{}
	pointerPolicy := wireTypePolicy{pointer: true}
	catalog, generation := testWireTypeCatalog(t)
	catalog.collect(attribute, wireAttribute, valuePolicy)
	catalog.collect(attribute, wireAttribute, pointerPolicy)
	linkTestWireTypeCatalog(t, generation, catalog)

	valueRecord := catalog.lookupUser(attribute, wireAttribute, valuePolicy)
	pointerRecord := catalog.lookupUser(attribute, wireAttribute, pointerPolicy)
	require.NotSame(t, valueRecord, pointerRecord)

	valueScope := catalog.resolver(catalog.scope, valuePolicy)
	pointerScope := catalog.resolver(catalog.scope, pointerPolicy)
	require.Equal(t, valueRecord.name, valueScope.Name(attribute, "", false, false))
	require.Equal(t, pointerRecord.name, pointerScope.Name(attribute, "", true, false))
}

func TestWireTypeCatalogDoesNotNameTheSharedEmptySentinel(t *testing.T) {
	originalNil := expr.Empty.Attribute().Meta == nil
	original := expr.Empty.Attribute().Meta.Dup()
	attribute := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "empty", Attribute: &expr.AttributeExpr{Type: expr.Empty}},
	}}
	catalog, generation := testWireTypeCatalog(t)

	catalog.collect(attribute, wireAttribute, wireTypePolicy{})
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
	catalog.collect(typeAttribute, wireRequestBody, policy)
	linkTestWireTypeCatalog(t, generation, catalog)

	require.Panics(t, func() {
		catalog.collect(&expr.AttributeExpr{Type: wireCatalogType("Late", "late", "value", true)}, wireRequestBody, policy)
	})
	require.Panics(t, func() {
		catalog.lookupUser(&expr.AttributeExpr{Type: wireCatalogType("Unknown", "unknown", "value", true)}, wireRequestBody, policy)
	})
	require.Panics(t, func() {
		catalog.scope.Unique("late")
	})
}

func TestBindCopiedUnionOccurrencesRejectsDifferentGraphs(t *testing.T) {
	tests := []struct {
		name     string
		planned  *expr.AttributeExpr
		rendered *expr.AttributeExpr
		message  string
	}{
		{
			name:     "different type",
			planned:  &expr.AttributeExpr{Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.String}}},
			rendered: &expr.AttributeExpr{Type: &expr.Map{KeyType: &expr.AttributeExpr{Type: expr.String}, ElemType: &expr.AttributeExpr{Type: expr.String}}},
			message:  "planned array does not match rendered map at body",
		},
		{
			name: "missing field",
			planned: &expr.AttributeExpr{Type: &expr.Object{{
				Name:      "value",
				Attribute: &expr.AttributeExpr{Type: expr.String},
			}}},
			rendered: &expr.AttributeExpr{Type: &expr.Object{}},
			message:  `rendered object has no field "value" at body.value`,
		},
		{
			name:     "extra field",
			planned:  &expr.AttributeExpr{Type: &expr.Object{}},
			rendered: &expr.AttributeExpr{Type: &expr.Object{{Name: "value", Attribute: &expr.AttributeExpr{Type: expr.String}}}},
			message:  `planned object has no field "value" at body.value`,
		},
		{
			name: "missing union branch",
			planned: &expr.AttributeExpr{Type: &expr.Union{TypeName: "Choice", Values: []*expr.NamedAttributeExpr{{
				Name:      "value",
				Attribute: &expr.AttributeExpr{Type: expr.String},
			}}}},
			rendered: &expr.AttributeExpr{Type: &expr.Union{TypeName: "Choice"}},
			message:  `rendered OneOf "Choice" has no branch "value" at body.value`,
		},
		{
			name:     "extra union branch",
			planned:  &expr.AttributeExpr{Type: &expr.Union{TypeName: "Choice"}},
			rendered: &expr.AttributeExpr{Type: &expr.Union{TypeName: "Choice", Values: []*expr.NamedAttributeExpr{{Name: "value", Attribute: &expr.AttributeExpr{Type: expr.String}}}}},
			message:  `planned OneOf "Choice" has no branch "value" at body.value`,
		},
		{
			name:     "different union envelope",
			planned:  &expr.AttributeExpr{Type: &expr.Union{TypeName: "Choice", TypeKey: "type", ValueKey: "value"}},
			rendered: &expr.AttributeExpr{Type: &expr.Union{TypeName: "Choice", TypeKey: "kind", ValueKey: "body"}},
			message:  `planned OneOf "Choice" does not match rendered OneOf "Choice" at body`,
		},
		{
			name:     "different primitive",
			planned:  &expr.AttributeExpr{Type: expr.String},
			rendered: &expr.AttributeExpr{Type: expr.Int},
			message:  "planned string does not match rendered int at body",
		},
		{
			name: "different user type",
			planned: &expr.AttributeExpr{Type: &expr.UserTypeExpr{
				TypeName:      "First",
				AttributeExpr: &expr.AttributeExpr{Type: expr.String},
			}},
			rendered: &expr.AttributeExpr{Type: &expr.UserTypeExpr{
				TypeName:      "Second",
				AttributeExpr: &expr.AttributeExpr{Type: expr.String},
			}},
			message: `planned user type "First" does not match rendered user type "Second" at body`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			catalog := newWireTypeCatalog()

			err := catalog.bindCopiedUnionOccurrences(
				test.planned,
				test.rendered,
				"body",
				make(map[wireAttributePair]struct{}),
			)

			require.EqualError(t, err, test.message)
		})
	}
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

// wireCatalogContainer places a named type inside an object body.
func wireCatalogContainer(child expr.UserType) *expr.AttributeExpr {
	return &expr.AttributeExpr{Type: &expr.Object{{
		Name:      "child",
		Attribute: &expr.AttributeExpr{Type: child},
	}}}
}

// firstWireUserType returns the first named value inside an object or array body.
func firstWireUserType(body *expr.AttributeExpr) *expr.AttributeExpr {
	switch actual := body.Type.(type) {
	case *expr.Object:
		return (*actual)[0].Attribute
	case *expr.Array:
		return actual.ElemType
	default:
		panic("test body does not contain a named type")
	}
}
