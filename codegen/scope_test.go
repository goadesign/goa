package codegen

import (
	"strings"
	"testing"

	"goa.design/goa/v3/expr"
)

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

func TestNameScope_GoTypeDef_UsesPointersOnlyForOptionalUnions(t *testing.T) {
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

	typeDef := NewNameScope().GoTypeDef(attribute, false, false)
	if !strings.Contains(typeDef, "Optional *Scope") {
		t.Fatalf("expected optional union field to be a pointer, got:\n%s", typeDef)
	}
	if !strings.Contains(typeDef, "Required Scope") {
		t.Fatalf("expected required union field to remain a value, got:\n%s", typeDef)
	}
}
