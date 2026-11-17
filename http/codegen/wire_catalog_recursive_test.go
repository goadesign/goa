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

func TestHTTPUnionSharedGraphKeepsRoles(t *testing.T) {
	for _, requestFirst := range []bool{false, true} {
		t.Run(map[bool]string{false: "response first", true: "request first"}[requestFirst], func(t *testing.T) {
			branch := wireCatalogType("Branch", "branch", "label", false)
			choice := &expr.AttributeExpr{Type: &expr.Union{
				TypeName: "Choice",
				Values: []*expr.NamedAttributeExpr{
					{Name: "branch", Attribute: &expr.AttributeExpr{Type: branch}},
				},
			}}
			catalog, generation := testWireTypeCatalog(t)
			request := wireTypePolicy{request: true, pointer: true}
			response := wireTypePolicy{}
			collectRequest := func() { catalog.collect(choice, wireRequestBody, request) }
			collectResponse := func() { catalog.collect(choice, wireResponseBody, response) }
			if requestFirst {
				collectRequest()
				collectResponse()
			} else {
				collectResponse()
				collectRequest()
			}
			detailed := wireTypePolicy{view: "detailed"}
			catalog.collect(choice, wireResponseBody, detailed)
			linkTestWireTypeCatalog(t, generation, catalog)
			branchAttribute := expr.AsUnion(choice.Type).Values[0].Attribute
			requestBranch := catalog.find(newWireTypeIdentity(branchAttribute, wireAttribute, request, wireTypePreferredName(branch, request)))
			responseBranch := catalog.find(newWireTypeIdentity(branchAttribute, wireAttribute, response, wireTypePreferredName(branch, response)))
			require.NotSame(t, requestBranch, responseBranch)
			for _, union := range catalog.unionTypes() {
				switch union.Name {
				case "ChoiceRequestBody":
					require.Equal(t, requestBranch.ref, union.Fields[0].FieldType)
				case "ChoiceResponseBody", "ChoiceDetailedResponseBody":
					require.Equal(t, responseBranch.ref, union.Fields[0].FieldType)
				default:
					t.Errorf("unexpected union %q", union.Name)
				}
			}
			for _, test := range []struct {
				use    wireUnionUse
				policy wireTypePolicy
				name   string
			}{
				{wireUnionUse{role: wireRequestBody}, request, "ChoiceRequestBody"},
				{wireUnionUse{role: wireResponseBody}, response, "ChoiceResponseBody"},
				{wireUnionUse{role: wireResponseBody, view: "detailed"}, detailed, "ChoiceDetailedResponseBody"},
				{wireUnionUse{role: wireRequestBody}, request, "ChoiceRequestBody"},
			} {
				scope := catalog.resolverForUse(catalog.scope, test.policy, test.use)
				require.Equal(t, test.name, scope.Name(choice, "", false, false))
				copied := expr.DupAtt(choice)
				require.NoError(t, catalog.bindCopiedUnionOccurrences(copied, choice, test.use, "body", make(map[wireAttributePair]struct{})))
				require.Equal(t, test.name, scope.Name(copied, "", false, false))
			}
		})
	}
}
