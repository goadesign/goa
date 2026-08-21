// This file verifies recursive HTTP and JSON-RPC expression inspection across
// unrelated declarations that share a semantic identifier.
package expr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHasJSONRPCIDFieldDistinguishesEqualUIDOrigins(t *testing.T) {
	first := &UserTypeExpr{
		AttributeExpr: &AttributeExpr{Type: String},
		TypeName:      "First",
		UID:           "shared",
	}
	second := &UserTypeExpr{
		AttributeExpr: &AttributeExpr{
			Type: String,
			Meta: MetaExpr{"jsonrpc:id": {}},
		},
		TypeName: "Second",
		UID:      "shared",
	}
	root := &AttributeExpr{Type: &Object{
		{Name: "first", Attribute: &AttributeExpr{Type: first}},
		{Name: "second", Attribute: &AttributeExpr{Type: second}},
	}}

	require.True(t, hasJSONRPCIDField(root))
}
