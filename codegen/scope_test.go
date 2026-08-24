// This file verifies package-level name allocation and generated type identity.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

type exactHasher string

// Hash returns the exact map identity supplied by the test.
func (h exactHasher) Hash() string {
	return string(h)
}

func TestNameScope_Freeze(t *testing.T) {
	scope := NewNameScope()
	existing := &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{Type: expr.String},
		TypeName:      "Existing",
		UID:           "existing",
	}
	require.Equal(t, "Existing", scope.GoTypeName(&expr.AttributeExpr{Type: existing}))

	scope.Freeze()
	scope.Freeze()
	require.Equal(t, "Existing", scope.GoTypeName(&expr.AttributeExpr{Type: existing}))
	require.Equal(t, "Next", scope.PeekUnique("Next"))
	require.Equal(t, "Next", scope.Name("Next"))
	require.Panics(t, func() {
		scope.Unique("Next")
	})
	require.Panics(t, func() {
		scope.HashedUnique(&expr.UserTypeExpr{
			AttributeExpr: &expr.AttributeExpr{Type: expr.String},
			TypeName:      "Next",
			UID:           "next",
		}, "Next")
	})
	require.Panics(t, func() {
		scope.GoTypeName(&expr.AttributeExpr{Type: &expr.UserTypeExpr{
			AttributeExpr: &expr.AttributeExpr{Type: expr.String},
			TypeName:      "Indirect",
			UID:           "indirect",
		}})
	})
}

func TestNameScope_Unique(t *testing.T) {
	sequence := []struct {
		Input    string
		Suffix   []string
		Expected string
	}{
		{Input: "a", Expected: "a"},
		{Input: "a", Expected: "a2"},
		{Input: "a", Expected: "a3"},
		{Input: "a", Expected: "a4"},
		{Input: "b", Expected: "b"},
		{Input: "c", Expected: "c"},
		{Input: "hel", Expected: "hel"},
		{Input: "hel", Suffix: []string{"lo"}, Expected: "hello"},
		{Input: "hello", Expected: "hello2"},
		{Input: "hello", Suffix: []string{"1"}, Expected: "hello1"},
		{Input: "hello", Suffix: []string{"1"}, Expected: "hello12"},
		{Input: "hello", Suffix: []string{"2"}, Expected: "hello22"},
		{Input: "hello", Suffix: []string{"2"}, Expected: "hello23"},
		{Input: "hello,world", Expected: "hello,world"},
		{Input: "hello,world1", Expected: "hello,world1"},
		{Input: "hello,world2", Expected: "hello,world2"},
		{Input: "hello", Suffix: []string{",world"}, Expected: "hello,world3"},
	}

	scope := NewNameScope()
	for i, v := range sequence {
		if got := scope.Unique(v.Input, v.Suffix...); v.Expected != got {
			t.Errorf("#%v, expected %v, got %v", i, v.Expected, got)
		}
	}
}

func TestNameScope_HashedUniqueUsesExactHash(t *testing.T) {
	scope := NewNameScope()
	require.Equal(t, "First", scope.HashedUnique(exactHasher("shared"), "First"))
	require.Equal(t, "First", scope.HashedUnique(exactHasher("shared"), "Ignored"))
	require.Equal(t, "First2", scope.HashedUnique(exactHasher("distinct"), "First"))
}

func TestNameScope_GoFullTypeName_UsesScopedNameWhenQualified(t *testing.T) {
	scope := NewNameScope()

	// Simulate the service generator reserving/using "Request" for a different
	// identifier before naming a user type that also wants to be "Request".
	scope.Unique("Request")

	ut := &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{Type: expr.String},
		TypeName:      "Request",
		UID:           "t",
	}
	att := &expr.AttributeExpr{Type: ut}

	if got, want := scope.GoTypeName(att), "Request2"; got != want {
		t.Fatalf("expected scoped type name %q, got %q", want, got)
	}
	if got, want := scope.GoFullTypeName(att, "svc"), "svc.Request2"; got != want {
		t.Fatalf("expected qualified scoped name %q, got %q", want, got)
	}

	fresh := NewNameScope()
	if got, want := fresh.GoFullTypeName(att, "svc"), "svc.Request"; got != want {
		t.Fatalf("expected qualified base name %q with fresh scope, got %q", want, got)
	}
}

func TestNameScope_GoFullTypeNameUsesAuthoredUnionNameBeforePlanning(t *testing.T) {
	scope := NewNameScope()
	first := &expr.Union{TypeName: "Value"}
	second := &expr.Union{TypeName: "Value"}
	scope.GoTypeName(&expr.AttributeExpr{Type: first})
	secondAtt := &expr.AttributeExpr{Type: second}
	if got, want := scope.GoTypeName(secondAtt), "Value"; got != want {
		t.Errorf("GoTypeName() = %q, want %q", got, want)
	}
	if got, want := scope.GoFullTypeName(secondAtt, "types"), "types.Value"; got != want {
		t.Errorf("GoFullTypeName() = %q, want %q", got, want)
	}
}

func TestNameScope_GoFullTypeNameRejectsUnplannedUnionAfterFreeze(t *testing.T) {
	scope := NewNameScope()
	attribute := &expr.AttributeExpr{Type: &expr.Union{TypeName: "Value"}}
	require.Equal(t, "Value", scope.GoFullTypeName(attribute, ""))

	scope.Freeze()
	require.PanicsWithValue(t,
		`union "Value" was not declared before the package name scope was frozen`,
		func() { scope.GoFullTypeName(attribute, "") },
	)
}

func TestNameScope_GoTypeNameDoesNotPlanUnionDeclarations(t *testing.T) {
	scope := NewNameScope()
	first := &expr.Union{TypeName: "Value", TypeKey: "type", ValueKey: "value"}
	second := &expr.Union{TypeName: "Value", TypeKey: "kind", ValueKey: "data"}
	assert.Equal(t, first.Hash(), second.Hash(), "compatibility hash should remain unchanged")
	assert.NotEqual(t, NewUnionTypeID(first), NewUnionTypeID(second))
	assert.Equal(t, "Value", scope.GoTypeName(&expr.AttributeExpr{Type: first}))
	assert.Equal(t, "Value", scope.GoTypeName(&expr.AttributeExpr{Type: second}))
}

func TestNameScope_GoTypeNameDistinguishesUnionBranchPackages(t *testing.T) {
	branch := func(path string) expr.UserType {
		return &expr.UserTypeExpr{
			TypeName: "Entry",
			UID:      path,
			AttributeExpr: &expr.AttributeExpr{
				Type: expr.String,
				Meta: expr.MetaExpr{"struct:pkg:path": {path}},
			},
		}
	}
	first := &expr.Union{
		TypeName: "Value",
		Values: []*expr.NamedAttributeExpr{
			{Name: "entry", Attribute: &expr.AttributeExpr{Type: branch("types/first")}},
		},
	}
	second := &expr.Union{
		TypeName: "Value",
		Values: []*expr.NamedAttributeExpr{
			{Name: "entry", Attribute: &expr.AttributeExpr{Type: branch("types/second")}},
		},
	}
	assert.Equal(t, first.Hash(), second.Hash(), "compatibility hash should remain unchanged")
	assert.NotEqual(t, NewUnionTypeID(first), NewUnionTypeID(second))
	scope := NewNameScope()
	assert.Equal(t, "Value", scope.GoTypeName(&expr.AttributeExpr{Type: first}))
	assert.Equal(t, "Value", scope.GoTypeName(&expr.AttributeExpr{Type: second}))
}

func TestNameScope_GoTypeNameDistinguishesUnionBranchOrder(t *testing.T) {
	branch := func(name string) *expr.NamedAttributeExpr {
		return &expr.NamedAttributeExpr{Name: name, Attribute: &expr.AttributeExpr{Type: expr.String}}
	}
	first := &expr.Union{TypeName: "Value", Values: []*expr.NamedAttributeExpr{branch("left"), branch("right")}}
	second := &expr.Union{TypeName: "Value", Values: []*expr.NamedAttributeExpr{branch("right"), branch("left")}}
	assert.Equal(t, first.Hash(), second.Hash(), "compatibility hash should remain unchanged")
	assert.NotEqual(t, NewUnionTypeID(first), NewUnionTypeID(second))
	scope := NewNameScope()
	assert.Equal(t, "Value", scope.GoTypeName(&expr.AttributeExpr{Type: first}))
	assert.Equal(t, "Value", scope.GoTypeName(&expr.AttributeExpr{Type: second}))
}

func TestNameScope_GoTypeNameDistinguishesInlineObjectFieldOrder(t *testing.T) {
	object := func(names ...string) *expr.Object {
		fields := make(expr.Object, len(names))
		for i, name := range names {
			fields[i] = &expr.NamedAttributeExpr{Name: name, Attribute: &expr.AttributeExpr{Type: expr.String}}
		}
		return &fields
	}
	union := func(fields *expr.Object) *expr.Union {
		return &expr.Union{
			TypeName: "Value",
			Values: []*expr.NamedAttributeExpr{
				{Name: "object", Attribute: &expr.AttributeExpr{Type: fields}},
			},
		}
	}
	first := union(object("left", "right"))
	second := union(object("right", "left"))
	assert.Equal(t, first.Hash(), second.Hash(), "compatibility hash should remain unchanged")
	assert.NotEqual(t, NewUnionTypeID(first), NewUnionTypeID(second))
	scope := NewNameScope()
	assert.Equal(t, "Value", scope.GoTypeName(&expr.AttributeExpr{Type: first}))
	assert.Equal(t, "Value", scope.GoTypeName(&expr.AttributeExpr{Type: second}))
}

// TestNameScope_GoTypeNameUsesAuthoredNameForEqualUnionDefinitions verifies
// reference formatting does not decide whether two declarations can share a
// package name. The generated package catalog makes that decision before the
// scope is frozen.
func TestNameScope_GoTypeNameUsesAuthoredNameForEqualUnionDefinitions(t *testing.T) {
	branch := func(name, id string) expr.UserType {
		return &expr.UserTypeExpr{
			TypeName: name,
			UID:      id,
			AttributeExpr: &expr.AttributeExpr{
				Type: expr.String,
				Meta: expr.MetaExpr{"struct:pkg:path": {"types"}},
			},
		}
	}
	union := func(user expr.UserType) *expr.Union {
		return &expr.Union{
			TypeName: "Value",
			Values: []*expr.NamedAttributeExpr{
				{Name: "entry", Attribute: &expr.AttributeExpr{Type: user}},
			},
		}
	}
	first := union(branch("foo-bar", "first"))
	second := union(branch("foo_bar", "second"))
	assert.Equal(t, NewUnionTypeID(first), NewUnionTypeID(second))
	scope := NewNameScope()
	assert.Equal(t, "Value", scope.GoTypeName(&expr.AttributeExpr{Type: first}))
	assert.Equal(t, "Value", scope.GoTypeName(&expr.AttributeExpr{Type: second}))
}

// TestNameScopeForkPreservesBindingsAndAcceptsHelperNames verifies a frozen
// declaration scope can seed a separate mutable helper namespace.
func TestNameScopeForkPreservesBindingsAndAcceptsHelperNames(t *testing.T) {
	typeExpr := &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{Type: expr.String},
		TypeName:      "Value",
	}
	scope := NewNameScope()
	assert.Equal(t, "Value", scope.GoTypeName(&expr.AttributeExpr{Type: typeExpr}))
	scope.Freeze()

	fork := scope.Fork()
	assert.Equal(t, "Value", fork.GoTypeName(&expr.AttributeExpr{Type: typeExpr}))
	assert.Equal(t, "Value2", fork.Unique("Value"))
	assert.Equal(t, "helper", fork.Unique("helper"))
}

func TestUnionTypeID(t *testing.T) {
	branch := func(name string, dataType expr.DataType) *expr.NamedAttributeExpr {
		return &expr.NamedAttributeExpr{Name: name, Attribute: &expr.AttributeExpr{Type: dataType}}
	}
	userType := func(path string) expr.UserType {
		return &expr.UserTypeExpr{
			AttributeExpr: &expr.AttributeExpr{
				Type: expr.String,
				Meta: expr.MetaExpr{"struct:pkg:path": {path}},
			},
			TypeName: "Entry",
			UID:      "entry",
		}
	}
	tests := []struct {
		name   string
		first  *expr.Union
		second *expr.Union
	}{
		{
			name:   "wire keys",
			first:  &expr.Union{TypeName: "Value", TypeKey: "type", ValueKey: "value"},
			second: &expr.Union{TypeName: "Value", TypeKey: "kind", ValueKey: "data"},
		},
		{
			name: "branch order",
			first: &expr.Union{TypeName: "Value", Values: []*expr.NamedAttributeExpr{
				branch("left", expr.String),
				branch("right", expr.Int),
			}},
			second: &expr.Union{TypeName: "Value", Values: []*expr.NamedAttributeExpr{
				branch("right", expr.Int),
				branch("left", expr.String),
			}},
		},
		{
			name: "branch Go shape",
			first: &expr.Union{TypeName: "Value", Values: []*expr.NamedAttributeExpr{
				branch("entry", expr.String),
			}},
			second: &expr.Union{TypeName: "Value", Values: []*expr.NamedAttributeExpr{
				{Name: "entry", Attribute: &expr.AttributeExpr{
					Type: expr.String,
					Meta: expr.MetaExpr{"struct:field:type": {"CustomString"}},
				}},
			}},
		},
		{
			name: "relocated branch package",
			first: &expr.Union{TypeName: "Value", Values: []*expr.NamedAttributeExpr{
				branch("entry", userType("types/first")),
			}},
			second: &expr.Union{TypeName: "Value", Values: []*expr.NamedAttributeExpr{
				branch("entry", userType("types/second")),
			}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.NotEqual(t, NewUnionTypeID(test.first), NewUnionTypeID(test.second))
		})
	}
}

// TestUnionTypeIDIncludesDefaultDrivenPointerShape verifies that a default
// which changes an inline service field from a pointer to a value also changes
// the union definition identity.
func TestUnionTypeIDIncludesDefaultDrivenPointerShape(t *testing.T) {
	union := func(withDefault bool) (*expr.Union, *expr.AttributeExpr) {
		field := &expr.AttributeExpr{Type: expr.String}
		if withDefault {
			field.DefaultValue = "ready"
		}
		branch := &expr.AttributeExpr{Type: &expr.Object{
			{Name: "label", Attribute: field},
		}}
		return &expr.Union{
			TypeName: "Value",
			Values: []*expr.NamedAttributeExpr{
				{Name: "entry", Attribute: branch},
			},
		}, branch
	}
	withoutDefault, pointerBranch := union(false)
	withDefault, valueBranch := union(true)
	policy := GoLayoutPolicy{UseDefault: true, SumType: true}
	pointerLayout, err := PlanGoType(pointerBranch, GoTypePlanOptions{
		Owner:  "generated.local/gen/service",
		Policy: policy,
	})
	require.NoError(t, err)
	valueLayout, err := PlanGoType(valueBranch, GoTypePlanOptions{
		Owner:  "generated.local/gen/service",
		Policy: policy,
	})
	require.NoError(t, err)

	require.True(t, pointerLayout.Fields()[0].IsPointer())
	require.False(t, valueLayout.Fields()[0].IsPointer())
	require.NotEqual(t, NewUnionTypeID(withoutDefault), NewUnionTypeID(withDefault))
}

func TestUnionTypeIDIgnoresNonEmittedPointerSharing(t *testing.T) {
	object := func() *expr.Object {
		fields := expr.Object{
			{Name: "text", Attribute: &expr.AttributeExpr{Type: expr.String}},
		}
		return &fields
	}
	innerUnion := func() *expr.Union {
		return &expr.Union{
			TypeName: "Inner",
			Values: []*expr.NamedAttributeExpr{
				{Name: "text", Attribute: &expr.AttributeExpr{Type: expr.String}},
			},
		}
	}
	outerUnion := func(left, right expr.DataType) *expr.Union {
		return &expr.Union{
			TypeName: "Outer",
			Values: []*expr.NamedAttributeExpr{
				{Name: "left", Attribute: &expr.AttributeExpr{Type: left}},
				{Name: "right", Attribute: &expr.AttributeExpr{Type: right}},
			},
		}
	}

	t.Run("inline object", func(t *testing.T) {
		shared := object()
		assert.Equal(t, NewUnionTypeID(outerUnion(shared, shared)), NewUnionTypeID(outerUnion(object(), object())))
	})
	t.Run("nested union", func(t *testing.T) {
		shared := innerUnion()
		assert.Equal(t, NewUnionTypeID(outerUnion(shared, shared)), NewUnionTypeID(outerUnion(innerUnion(), innerUnion())))
	})
}

// TestUnionTypeIDIncludesGeneratedUserTypeShape verifies generated aliases
// with one name and distinct definitions produce distinct union identities.
func TestUnionTypeIDIncludesGeneratedUserTypeShape(t *testing.T) {
	generatedBranch := func(dataType expr.DataType) *expr.Union {
		alias := &expr.UserTypeExpr{
			AttributeExpr: &expr.AttributeExpr{Type: dataType},
			TypeName:      "ValueText",
		}
		return &expr.Union{
			TypeName: "Value",
			Values: []*expr.NamedAttributeExpr{{
				Name:      "text",
				Attribute: &expr.AttributeExpr{Type: alias},
			}},
		}
	}

	require.NotEqual(t,
		NewUnionTypeID(generatedBranch(expr.String)),
		NewUnionTypeID(generatedBranch(expr.Int)),
	)
}

// TestUnionTypeIDEncodesRecursiveGeneratedUserTypeShape verifies recursive
// generated aliases terminate and retain their distinct field definitions.
func TestUnionTypeIDEncodesRecursiveGeneratedUserTypeShape(t *testing.T) {
	recursiveUnion := func(fieldType expr.DataType) *expr.Union {
		alias := &expr.UserTypeExpr{TypeName: "ValueNode"}
		alias.AttributeExpr = &expr.AttributeExpr{Type: &expr.Object{
			{Name: "value", Attribute: &expr.AttributeExpr{Type: fieldType}},
			{Name: "next", Attribute: &expr.AttributeExpr{Type: alias}},
		}}
		return &expr.Union{
			TypeName: "Value",
			Values: []*expr.NamedAttributeExpr{
				{Name: "node", Attribute: &expr.AttributeExpr{Type: alias}},
			},
		}
	}

	require.NotEqual(t,
		NewUnionTypeID(recursiveUnion(expr.String)),
		NewUnionTypeID(recursiveUnion(expr.Int)),
	)
}

func TestNameScope_GoFullTypeName_UsesScopedRelocatedUserTypeNameWhenQualified(t *testing.T) {
	scope := NewNameScope()
	first := &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{
			Type: expr.String,
			Meta: expr.MetaExpr{"struct:pkg:path": {"types"}},
		},
		TypeName: "foo-bar",
		UID:      "first",
	}
	second := &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{
			Type: expr.String,
			Meta: expr.MetaExpr{"struct:pkg:path": {"types"}},
		},
		TypeName: "foo_bar",
		UID:      "second",
	}
	scope.GoTypeName(&expr.AttributeExpr{Type: first})
	secondAtt := &expr.AttributeExpr{Type: second}
	if got, want := scope.GoTypeName(secondAtt), "FooBar2"; got != want {
		t.Fatalf("GoTypeName() = %q, want %q", got, want)
	}
	if got, want := scope.GoFullTypeName(secondAtt, "types"), "types.FooBar2"; got != want {
		t.Errorf("GoFullTypeName() = %q, want %q", got, want)
	}
}

func TestNameScope_PeekUnique_MatchesUniqueWithoutMutation(t *testing.T) {
	seed := func(scope *NameScope) {
		scope.Unique("a")
		scope.Unique("a")
		scope.Unique("a2")
		scope.Unique("hello")
		scope.Unique("hello1")
	}

	peek := NewNameScope()
	seed(peek)

	mutating := NewNameScope()
	seed(mutating)

	if got, want := peek.PeekUnique("a"), mutating.Unique("a"); got != want {
		t.Fatalf("expected peek %q, got %q", want, got)
	}
	if got, want := peek.PeekUnique("hel", "lo"), mutating.Unique("hel", "lo"); got != want {
		t.Fatalf("expected peek %q, got %q", want, got)
	}
	if got, want := peek.PeekUnique("hello", "1"), mutating.Unique("hello", "1"); got != want {
		t.Fatalf("expected peek %q, got %q", want, got)
	}

	// PeekUnique must not mutate the scope.
	if got, want := peek.Unique("a"), "a3"; got != want {
		t.Fatalf("expected scope unchanged, got %q", got)
	}
}

func TestNameScopeGoTypeDefUsesValueUnions(t *testing.T) {
	union := &expr.Union{
		TypeName: "Scope",
		Values: []*expr.NamedAttributeExpr{
			{Name: "description", Attribute: &expr.AttributeExpr{Type: expr.String}},
			{Name: "aliases", Attribute: &expr.AttributeExpr{Type: &expr.Array{
				ElemType: &expr.AttributeExpr{Type: expr.String},
			}}},
		},
	}
	attribute := &expr.AttributeExpr{
		Type: &expr.Object{
			{Name: "optional", Attribute: &expr.AttributeExpr{Type: union}},
			{Name: "required", Attribute: &expr.AttributeExpr{Type: union}},
		},
		Validation: &expr.ValidationExpr{Required: []string{"required"}},
	}

	for _, pointer := range []bool{false, true} {
		typeDef := NewNameScope().GoTypeDef(attribute, pointer, false)
		assert.Contains(t, typeDef, "Optional Scope")
		assert.Contains(t, typeDef, "Required Scope")
	}
}
