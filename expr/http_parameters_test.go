// These tests check the parameter validation stage used by HTTP and plugins.
// Prepared string bindings must resolve to the actual payload type without
// changing the original mapping, including when payload and URL names differ.
package expr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHTTPEndpointValidateParams(t *testing.T) {
	object := &Object{{Name: "id", Attribute: &AttributeExpr{Type: String}}}
	cases := []struct {
		name    string
		typ     DataType
		path    bool
		missing bool
		want    string
	}{
		{name: "path scalar", typ: Int64, path: true},
		{name: "path primitive array", typ: &Array{ElemType: &AttributeExpr{Type: Int}}, path: true},
		{name: "path named primitive array", typ: &UserTypeExpr{TypeName: "Identifiers", AttributeExpr: &AttributeExpr{Type: &Array{ElemType: &AttributeExpr{Type: String}}}}, path: true},
		{name: "path object", typ: object, path: true, want: "path parameter domain_id cannot be an object"},
		{name: "path object array", typ: &Array{ElemType: &AttributeExpr{Type: object}}, path: true, want: `elements of array path parameter "domain_id" must be primitive`},
		{name: "path map", typ: &Map{KeyType: &AttributeExpr{Type: String}, ElemType: &AttributeExpr{Type: String}}, path: true, want: "path parameter domain_id cannot be an object"},
		{name: "query primitive array", typ: &Array{ElemType: &AttributeExpr{Type: Int}}},
		{name: "query object array", typ: &Array{ElemType: &AttributeExpr{Type: object}}, want: `elements of array query parameter "domain_id" must be primitive`},
		{name: "query map", typ: &Map{KeyType: &AttributeExpr{Type: String}, ElemType: &AttributeExpr{Type: String}}},
		{name: "missing path field", typ: String, path: true, missing: true, want: `Path parameter "domain_id" not found in payload.`},
		{name: "missing query field", typ: String, missing: true, want: `Query string parameter "domain_id" not found in payload.`},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			payload := &Object{}
			if !test.missing {
				payload.Set("domain_id", &AttributeExpr{Type: test.typ})
			}
			endpoint := &HTTPEndpointExpr{
				MethodExpr: &MethodExpr{Name: "read", Payload: &AttributeExpr{Type: payload}},
				Params: NewMappedAttributeExpr(&AttributeExpr{Type: &Object{
					{Name: "domain_id:url_id", Attribute: &AttributeExpr{Type: String}},
				}}),
			}
			url := "//records"
			if test.path {
				url += "/{url_id}"
			}
			endpoint.Routes = []*RouteExpr{{Method: "GET", Path: url, Endpoint: endpoint}}
			validation := endpoint.ValidateParams()
			if test.want == "" {
				require.Empty(t, validation.Errors)
			} else {
				require.Contains(t, validation.Error(), test.want)
			}
			require.Equal(t, String, endpoint.Params.Find("domain_id").Type)
			require.Equal(t, "url_id", endpoint.Params.ElemName("domain_id"))
		})
	}
}
