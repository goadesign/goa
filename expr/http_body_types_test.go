// This file verifies HTTP body graph rewrites visit independent declarations
// and terminate when recursive copies return to their authored origin.
package expr

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPStreamingBodyValidation(t *testing.T) {
	cases := []struct {
		Name             string
		StreamingPayload *AttributeExpr
		Required         []string
		NestedUserType   bool
	}{
		{
			Name: "inline object",
			StreamingPayload: &AttributeExpr{
				Type: &Object{
					&NamedAttributeExpr{Name: "required", Attribute: &AttributeExpr{Type: String}},
					&NamedAttributeExpr{Name: "optional", Attribute: &AttributeExpr{Type: String}},
				},
				Validation: &ValidationExpr{Required: []string{"required"}},
			},
			Required: []string{"required"},
		},
		{
			Name: "named user type",
			StreamingPayload: &AttributeExpr{
				Type: &UserTypeExpr{
					AttributeExpr: &AttributeExpr{
						Type: &Object{
							&NamedAttributeExpr{Name: "required", Attribute: &AttributeExpr{Type: String}},
							&NamedAttributeExpr{Name: "optional", Attribute: &AttributeExpr{Type: String}},
						},
						Validation: &ValidationExpr{Required: []string{"required"}},
					},
					TypeName: "StreamingCommand",
					UID:      "StreamingCommand",
				},
			},
			Required:       []string{"required"},
			NestedUserType: true,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			service := &ServiceExpr{Name: "Service"}
			method := &MethodExpr{
				Name:             "Method",
				Service:          service,
				Stream:           BidirectionalStreamKind,
				StreamingPayload: c.StreamingPayload,
			}
			endpoint := &HTTPEndpointExpr{
				MethodExpr: method,
				Service:    &HTTPServiceExpr{ServiceExpr: service},
			}

			body := httpStreamingBody(endpoint)
			assert.Equal(t, c.Required, body.Validation.Required)
			assert.True(t, NewMappedAttributeExpr(body).IsRequired("required"))
			assert.False(t, NewMappedAttributeExpr(body).IsRequired("optional"))
			_, nested := body.Type.(UserType).Attribute().Type.(UserType)
			assert.Equal(t, c.NestedUserType, nested)

			body.Validation.AddRequired("body-only")
			assert.False(t, c.StreamingPayload.IsRequired("body-only"))
		})
	}
}

func TestComputedBodyExamplesDistinguishHTTPAndJSONRPC(t *testing.T) {
	service := &ServiceExpr{Name: "Service"}
	method := &MethodExpr{Name: "Method", Service: service}
	httpEndpoint := &HTTPEndpointExpr{
		MethodExpr: method,
		Service:    &HTTPServiceExpr{ServiceExpr: service},
		Body: &AttributeExpr{Type: &UserTypeExpr{
			TypeName: "HTTPBody",
			AttributeExpr: &AttributeExpr{Type: &Object{
				{Name: "http", Attribute: &AttributeExpr{Type: String}},
			}},
		}},
	}
	jsonRPCEndpoint := &HTTPEndpointExpr{
		MethodExpr: method,
		Service:    &HTTPServiceExpr{ServiceExpr: service},
		Meta:       MetaExpr{"jsonrpc": {}},
		Body: &AttributeExpr{Type: &UserTypeExpr{
			TypeName: "JSONRPCBody",
			AttributeExpr: &AttributeExpr{Type: &Object{
				{Name: "jsonrpc", Attribute: &AttributeExpr{Type: String}},
			}},
		}},
	}
	httpBody := httpRequestBody(httpEndpoint)
	jsonRPCBody := httpRequestBody(jsonRPCEndpoint)
	require.NotEqual(t, httpBody.Type.(UserType).ID(), jsonRPCBody.Type.(UserType).ID())

	cases := []struct {
		Name       string
		First      *HTTPEndpointExpr
		FirstBody  *AttributeExpr
		Second     *HTTPEndpointExpr
		SecondBody *AttributeExpr
	}{
		{
			Name:       "HTTP then JSON-RPC",
			First:      httpEndpoint,
			FirstBody:  httpBody,
			Second:     jsonRPCEndpoint,
			SecondBody: jsonRPCBody,
		},
		{
			Name:       "JSON-RPC then HTTP",
			First:      jsonRPCEndpoint,
			FirstBody:  jsonRPCBody,
			Second:     httpEndpoint,
			SecondBody: httpBody,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			generator := NewExampleGenerator(NewFakerRandomizerFactory("test"))
			first := c.FirstBody.Example(generator.At(RequestBodyExampleIdentity(c.First))).(map[string]any)
			second := c.SecondBody.Example(generator.At(RequestBodyExampleIdentity(c.Second))).(map[string]any)

			require.Contains(t, first, transportBodyField(c.First))
			require.NotContains(t, first, transportBodyField(c.Second))
			require.Contains(t, second, transportBodyField(c.Second))
			require.NotContains(t, second, transportBodyField(c.First))
		})
	}
}

// transportBodyField returns the field unique to the endpoint's body mapping.
func transportBodyField(endpoint *HTTPEndpointExpr) string {
	if endpoint.IsJSONRPC() {
		return "jsonrpc"
	}
	return "http"
}

func TestRemovePkgPathDistinguishesEqualUIDOrigins(t *testing.T) {
	first := &UserTypeExpr{
		AttributeExpr: &AttributeExpr{
			Type: &Object{},
			Meta: MetaExpr{"struct:pkg:path": {"first/types"}},
		},
		TypeName: "First",
		UID:      "shared",
	}
	second := &UserTypeExpr{
		AttributeExpr: &AttributeExpr{
			Type: &Object{},
			Meta: MetaExpr{"struct:pkg:path": {"second/types"}},
		},
		TypeName: "Second",
		UID:      "shared",
	}
	root := &AttributeExpr{Type: &Object{
		{Name: "first", Attribute: &AttributeExpr{Type: first}},
		{Name: "second", Attribute: &AttributeExpr{Type: second}},
	}}

	RemovePkgPath(root)
	require.NotContains(t, first.Attribute().Meta, "struct:pkg:path")
	require.NotContains(t, second.Attribute().Meta, "struct:pkg:path")
}
