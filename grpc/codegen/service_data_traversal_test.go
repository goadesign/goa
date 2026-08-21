// This file verifies that gRPC validation discovery distinguishes unrelated
// declarations while stopping recursion through copied declarations.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

func TestCollectValidationsDistinguishesEqualUIDOrigins(t *testing.T) {
	minimumLength := 3
	minimum := 5.0
	first := grpcValidationTraversalType("First", "shared", &expr.AttributeExpr{
		Type:       expr.String,
		Validation: &expr.ValidationExpr{MinLength: &minimumLength},
	})
	second := grpcValidationTraversalType("Second", "shared", &expr.AttributeExpr{
		Type:       expr.Int,
		Validation: &expr.ValidationExpr{Minimum: &minimum},
	})
	root := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "first", Attribute: &expr.AttributeExpr{Type: first}},
		{Name: "second", Attribute: &expr.AttributeExpr{Type: second}},
	}}
	sd := &ServiceData{PkgName: "pb", Scope: codegen.NewNameScope()}

	collectValidations(root, "root", true, sd)
	var names []string
	for _, validation := range sd.validations {
		names = append(names, validation.SrcName)
	}
	require.ElementsMatch(t, []string{"First", "Second"}, names)
}

// grpcValidationTraversalType builds an authored message declaration with one
// constrained field so validation discovery must emit a helper for it.
func grpcValidationTraversalType(name, uid string, field *expr.AttributeExpr) *expr.UserTypeExpr {
	return &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
			{Name: "value", Attribute: field},
		}},
		TypeName: name,
		UID:      uid,
	}
}
