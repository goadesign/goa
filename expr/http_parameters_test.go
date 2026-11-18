// These tests check the parameter validation stage used by HTTP and plugins.
// Prepared string bindings must resolve to the actual payload type without
// changing the original mapping, including when payload and URL names differ.
package expr

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/eval"
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

func TestHTTPEndpointMappedDefaults(t *testing.T) {
	cases := []struct {
		name           string
		kind           string
		mappedDefault  any
		payloadDefault any
		want           string
	}{
		{name: "query without defaults", kind: "query"},
		{name: "query default from payload only", kind: "query", payloadDefault: "a"},
		{name: "query same default", kind: "query", mappedDefault: "a", payloadDefault: "a"},
		{name: "query default only", kind: "query", mappedDefault: "a", want: `query parameter "value" defines a default value but payload attribute "value" does not`},
		{name: "query different default", kind: "query", mappedDefault: "a", payloadDefault: "b", want: `query parameter "value" default value "a" differs from payload attribute "value" default value "b"`},
		{name: "path default only", kind: "path", mappedDefault: "a", want: `path parameter "value" defines a default value but payload attribute "value" does not`},
		{name: "header default only", kind: "header", mappedDefault: "a", want: `header "value" defines a default value but payload attribute "value" does not`},
		{name: "header different default", kind: "header", mappedDefault: "a", payloadDefault: "b", want: `header "value" default value "a" differs from payload attribute "value" default value "b"`},
		{name: "cookie default only", kind: "cookie", mappedDefault: "a", want: `cookie "value" defines a default value but payload attribute "value" does not`},
		{name: "cookie same default", kind: "cookie", mappedDefault: "a", payloadDefault: "a"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			payload := &Object{{Name: "value", Attribute: &AttributeExpr{Type: String, DefaultValue: test.payloadDefault}}}
			mapped := NewMappedAttributeExpr(&AttributeExpr{Type: &Object{
				{Name: "value", Attribute: &AttributeExpr{Type: String, DefaultValue: test.mappedDefault}},
			}})
			empty := func() *MappedAttributeExpr {
				return NewMappedAttributeExpr(&AttributeExpr{Type: &Object{}})
			}
			endpoint := &HTTPEndpointExpr{
				MethodExpr: &MethodExpr{Name: "read", Payload: &AttributeExpr{Type: payload}},
				Params:     empty(),
				Headers:    empty(),
				Cookies:    empty(),
			}
			path := "//records"
			switch test.kind {
			case "path":
				endpoint.Params = mapped
				path += "/{value}"
			case "query":
				endpoint.Params = mapped
			case "header":
				endpoint.Headers = mapped
			case "cookie":
				endpoint.Cookies = mapped
			}
			endpoint.Routes = []*RouteExpr{{Method: "GET", Path: path, Endpoint: endpoint}}
			verr := new(eval.ValidationErrors)
			verr.Merge(endpoint.ValidateParams())
			verr.Merge(endpoint.validateHeadersAndCookies())
			if test.want == "" {
				require.Empty(t, verr.Errors)
				return
			}
			require.Contains(t, verr.Error(), test.want)
		})
	}
}
