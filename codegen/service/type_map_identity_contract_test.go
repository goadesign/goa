// This file verifies external type mappings follow the exact retained user
// type origin rather than a display name shared by unrelated declarations.
package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

// TestTypeMapMatchesExactRetainedUserType catches a mapping selected only
// because an unrelated service type has the same display name.
func TestTypeMapMatchesExactRetainedUserType(t *testing.T) {
	serviceType := &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{Type: expr.String},
		TypeName:      "Shared",
		UID:           "service-shared",
	}
	mappedType := &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{Type: expr.String},
		TypeName:      "Shared",
		UID:           "mapped-shared",
	}
	facts := &serviceFacts{
		reachableTypes: map[expr.UserType]struct{}{serviceType.Origin(): {}},
	}

	require.NotSame(t, serviceType.Origin(), mappedType.Origin())
	require.True(t, typeMapMatchesFacts(&expr.TypeMap{User: serviceType}, facts))
	require.False(t, typeMapMatchesFacts(&expr.TypeMap{User: mappedType}, facts))
}
