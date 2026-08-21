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

func TestAppendSuffixDistinguishesEqualUIDOriginsAndTerminatesRecursion(t *testing.T) {
	first := &UserTypeExpr{TypeName: "First", UID: "shared"}
	firstObject := &Object{}
	first.AttributeExpr = &AttributeExpr{Type: firstObject}
	firstObject.Set("self", &AttributeExpr{Type: first})
	second := &UserTypeExpr{
		AttributeExpr: &AttributeExpr{Type: &Object{}},
		TypeName:      "Second",
		UID:           "shared",
	}
	root := &Object{
		{Name: "first", Attribute: &AttributeExpr{Type: first}},
		{Name: "second", Attribute: &AttributeExpr{Type: second}},
	}

	appendSuffix(root, "Body")
	require.Equal(t, "FirstBody", first.Name())
	require.Equal(t, "SecondBody", second.Name())
}

func TestAppendSuffixTerminatesAfterRecursiveCopyDetachesOrigin(t *testing.T) {
	original := &UserTypeExpr{TypeName: "Recursive"}
	object := &Object{}
	original.AttributeExpr = &AttributeExpr{Type: object}
	object.Set("self", &AttributeExpr{Type: original})
	copy := Dup(original).(UserType)

	appendSuffix(copy, "Body")
	require.Equal(t, "RecursiveBody", copy.Name())
}
