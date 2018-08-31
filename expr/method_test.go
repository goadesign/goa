package expr

import (
	"fmt"
	"testing"

	"goa.design/goa/eval"
)

func TestMethodExprValidate(t *testing.T) {
	const (
		identifier = "result"
	)
	var (
		object = &UserTypeExpr{
			TypeName: "Object",
			AttributeExpr: &AttributeExpr{
				Description: "Object represents object values",
				Type:        &Object{},
			},
		}
		requirements = func(schemes ...*SchemeExpr) []*SecurityExpr {
			if len(schemes) > 0 {
				return []*SecurityExpr{{Schemes: schemes}}
			}
			return nil
		}
		service = func(schemes ...*SchemeExpr) *ServiceExpr {
			return &ServiceExpr{
				Name:         "test",
				Requirements: requirements(schemes...),
			}
		}
		basicScheme = func() *SchemeExpr {
			return &SchemeExpr{
				Kind:       BasicAuthKind,
				SchemeName: "Basic",
			}
		}
		apiKeyScheme = func() *SchemeExpr {
			return &SchemeExpr{
				Kind:       APIKeyKind,
				SchemeName: "APIKey",
			}
		}
		jwtScheme = func() *SchemeExpr {
			return &SchemeExpr{
				Kind:       JWTKind,
				SchemeName: "JWT",
			}
		}
		oauth2Scheme = func() *SchemeExpr {
			return &SchemeExpr{
				Kind:       OAuth2Kind,
				SchemeName: "OAuth2",
			}
		}
		attributeTypeObject = func() *AttributeExpr {
			return &AttributeExpr{
				Type: object,
			}
		}
		attributeTypeEmpty = func() *AttributeExpr {
			return &AttributeExpr{
				Type: Empty,
			}
		}
		attributeTypeNil = func() *AttributeExpr {
			return &AttributeExpr{
				Type: nil,
			}
		}
		meta = MetaExpr{
			"struct:error:name": []string{"error1"},
		}
		errorDuplicatedMeta = func() *AttributeExpr {
			return &AttributeExpr{
				Type: &ResultTypeExpr{
					UserTypeExpr: &UserTypeExpr{
						AttributeExpr: &AttributeExpr{
							Type: &Object{
								&NamedAttributeExpr{
									Name: "foo",
									Attribute: &AttributeExpr{
										Meta: meta,
									},
								},
								&NamedAttributeExpr{
									Name: "bar",
									Attribute: &AttributeExpr{
										Meta: meta,
									},
								},
							},
						},
					},
					Identifier: identifier,
				},
			}
		}
		errorMissingMeta = func() *AttributeExpr {
			return &AttributeExpr{
				Type: &ResultTypeExpr{
					UserTypeExpr: &UserTypeExpr{
						AttributeExpr: &AttributeExpr{
							Type: &Object{
								&NamedAttributeExpr{
									Name: "foo",
									Attribute: &AttributeExpr{
										Meta: MetaExpr{},
									},
								},
							},
						},
					},
					Identifier: identifier,
				},
			}
		}
		errAttributeTypeNil       = fmt.Errorf("attribute type is nil")
		errDuplicatedMeta         = fmt.Errorf("meta 'struct:error:name' already set for attribute %q of result type %q", "foo", identifier)
		errMissingMeta            = fmt.Errorf("meta 'struct:error:name' is missing in result type %q", identifier)
		errMissingUsernameAttr    = fmt.Errorf("payload of method \"test\" of service \"test\" does not define a username attribute, use Username to define one")
		errMissingPasswordAttr    = fmt.Errorf("payload of method \"test\" of service \"test\" does not define a password attribute, use Password to define one")
		errMissingAPIKeyAttr      = fmt.Errorf("payload of method \"test\" of service \"test\" does not define an API key attribute, use APIKey to define one")
		errMissingJWTAttr         = fmt.Errorf("payload of method \"test\" of service \"test\" does not define a JWT attribute, use Token to define one")
		errMissingAccessTokenAttr = fmt.Errorf("payload of method \"test\" of service \"test\" does not define a OAuth2 access token attribute, use AccessToken to define one")
	)

	cases := map[string]struct {
		service      *ServiceExpr
		requirements []*SecurityExpr
		payload      *AttributeExpr
		result       *AttributeExpr
		errors       []*ErrorExpr
		expected     *eval.ValidationErrors
	}{
		"no error": {
			payload:  attributeTypeEmpty(),
			result:   attributeTypeEmpty(),
			expected: &eval.ValidationErrors{},
		},
		"error only in payload": {
			payload:  attributeTypeNil(),
			result:   attributeTypeEmpty(),
			expected: &eval.ValidationErrors{Errors: []error{errAttributeTypeNil}},
		},
		"error only in result": {
			payload:  attributeTypeEmpty(),
			result:   attributeTypeNil(),
			expected: &eval.ValidationErrors{Errors: []error{errAttributeTypeNil}},
		},
		"errors only in errors": {
			payload: attributeTypeEmpty(),
			result:  attributeTypeEmpty(),
			errors: []*ErrorExpr{
				{
					AttributeExpr: errorDuplicatedMeta(),
					Name:          "foo",
				},
				{
					AttributeExpr: errorMissingMeta(),
					Name:          "bar",
				},
			},
			expected: &eval.ValidationErrors{Errors: []error{
				errDuplicatedMeta,
				errMissingMeta,
			}},
		},
		"error only in schemes": {
			requirements: requirements(basicScheme(), apiKeyScheme(), jwtScheme(), oauth2Scheme()),
			payload:      attributeTypeObject(),
			result:       attributeTypeEmpty(),
			expected: &eval.ValidationErrors{Errors: []error{
				errMissingUsernameAttr,
				errMissingPasswordAttr,
				errMissingAPIKeyAttr,
				errMissingJWTAttr,
				errMissingAccessTokenAttr,
			}},
		},
		"error only in inherited schemes": {
			service: service(basicScheme(), apiKeyScheme(), jwtScheme(), oauth2Scheme()),
			payload: attributeTypeObject(),
			result:  attributeTypeEmpty(),
			expected: &eval.ValidationErrors{Errors: []error{
				errMissingUsernameAttr,
				errMissingPasswordAttr,
				errMissingAPIKeyAttr,
				errMissingJWTAttr,
				errMissingAccessTokenAttr,
			}},
		},
		"errors in payload, schemes, result and errors": {
			requirements: requirements(basicScheme(), apiKeyScheme(), jwtScheme(), oauth2Scheme()),
			payload:      attributeTypeNil(),
			result:       attributeTypeNil(),
			errors: []*ErrorExpr{
				{
					AttributeExpr: errorDuplicatedMeta(),
					Name:          "foo",
				},
				{
					AttributeExpr: errorMissingMeta(),
					Name:          "bar",
				},
			},
			expected: &eval.ValidationErrors{Errors: []error{
				errAttributeTypeNil,
				errMissingUsernameAttr,
				errMissingPasswordAttr,
				errMissingAPIKeyAttr,
				errMissingJWTAttr,
				errMissingAccessTokenAttr,
				errAttributeTypeNil,
				errDuplicatedMeta,
				errMissingMeta,
			}},
		},
		"errors in payload, inherited schemes, result and errors": {
			service: service(basicScheme(), apiKeyScheme(), jwtScheme(), oauth2Scheme()),
			payload: attributeTypeNil(),
			result:  attributeTypeNil(),
			errors: []*ErrorExpr{
				{
					AttributeExpr: errorDuplicatedMeta(),
					Name:          "foo",
				},
				{
					AttributeExpr: errorMissingMeta(),
					Name:          "bar",
				},
			},
			expected: &eval.ValidationErrors{Errors: []error{
				errAttributeTypeNil,
				errMissingUsernameAttr,
				errMissingPasswordAttr,
				errMissingAPIKeyAttr,
				errMissingJWTAttr,
				errMissingAccessTokenAttr,
				errAttributeTypeNil,
				errDuplicatedMeta,
				errMissingMeta,
			}},
		},
	}

	for k, tc := range cases {
		var s *ServiceExpr
		if tc.service == nil {
			s = service()
		} else {
			s = tc.service
		}
		m := MethodExpr{
			Name:         "test",
			Service:      s,
			Requirements: tc.requirements,
			Payload:      tc.payload,
			Result:       tc.result,
			Errors:       tc.errors,

			StreamingPayload: &AttributeExpr{Type: Empty},
		}
		if actual := m.Validate().(*eval.ValidationErrors); len(tc.expected.Errors) != len(actual.Errors) {
			t.Errorf("%s: expected the number of error values to match %d got %d ", k, len(tc.expected.Errors), len(actual.Errors))
		} else {
			for i, err := range actual.Errors {
				if err.Error() != tc.expected.Errors[i].Error() {
					t.Errorf("%s: got %#v, expected %#v at index %d", k, err, tc.expected.Errors[i], i)
				}
			}
		}
	}
}

func TestMethodExpr_EvalName(t *testing.T) {
	cases := map[string]struct {
		name     string
		service  *ServiceExpr
		expected string
	}{
		"unnamed": {name: "", service: nil, expected: "unnamed method"},
		"foo":     {name: "foo", service: nil, expected: fmt.Sprintf("method %#v", "foo")},
		"bar":     {name: "bar", service: &ServiceExpr{Name: ""}, expected: fmt.Sprintf("unnamed service method %#v", "bar")},
		"baz":     {name: "baz", service: &ServiceExpr{Name: "baz service"}, expected: fmt.Sprintf("service %#v method %#v", "baz service", "baz")},
	}
	for k, tc := range cases {
		m := MethodExpr{Name: tc.name, Service: tc.service}
		if actual := m.EvalName(); actual != tc.expected {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}
