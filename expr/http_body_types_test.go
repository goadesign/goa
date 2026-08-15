package expr

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
