package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func TestHTTPUnionRejectsChangedNamedDescendant(t *testing.T) {
	leaf := wireCatalogType("Leaf", "leaf", "label", false)
	branch := wireCatalogType("Branch", "branch", "leaf", false)
	branch.Attribute().Find("leaf").Type = leaf
	original := &expr.AttributeExpr{Type: &expr.Union{
		TypeName: "Choice",
		Values: []*expr.NamedAttributeExpr{
			{Name: "branch", Attribute: &expr.AttributeExpr{Type: branch}},
		},
	}}
	copied := expr.DupAtt(original)
	copiedBranch := expr.AsUnion(copied.Type).Values[0].Attribute.Type.(expr.UserType)
	copiedLeaf := copiedBranch.Attribute().Find("leaf").Type.(expr.UserType)
	require.Same(t, leaf.Origin(), copiedLeaf.Origin())
	expr.AsObject(copiedLeaf).Set("count", &expr.AttributeExpr{Type: expr.Int})

	catalog, _ := testWireTypeCatalog(t)
	catalog.collect(original, wireResponseBody, wireTypePolicy{})
	catalog.collect(copied, wireResponseBody, wireTypePolicy{})
	require.ErrorContains(t, catalog.Declare(), `OneOf "Choice" produces different ResponseBody definitions`)
}
