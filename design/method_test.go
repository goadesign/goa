package design

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
		metadata = MetadataExpr{
			"struct:error:name": []string{"error1"},
		}
		errorDuplicatedMetadata = func() *AttributeExpr {
			return &AttributeExpr{
				Type: &ResultTypeExpr{
					UserTypeExpr: &UserTypeExpr{
						AttributeExpr: &AttributeExpr{
							Type: &Object{
								&NamedAttributeExpr{
									Name: "foo",
									Attribute: &AttributeExpr{
										Metadata: metadata,
									},
								},
								&NamedAttributeExpr{
									Name: "bar",
									Attribute: &AttributeExpr{
										Metadata: metadata,
									},
								},
							},
						},
					},
					Identifier: identifier,
				},
			}
		}
		errorMissingMetadata = func() *AttributeExpr {
			return &AttributeExpr{
				Type: &ResultTypeExpr{
					UserTypeExpr: &UserTypeExpr{
						AttributeExpr: &AttributeExpr{
							Type: &Object{
								&NamedAttributeExpr{
									Name: "foo",
									Attribute: &AttributeExpr{
										Metadata: MetadataExpr{},
									},
								},
							},
						},
					},
					Identifier: identifier,
				},
			}
		}
		errAttributeTypeNil   = fmt.Errorf("attribute type is nil")
		errDuplicatedMetadata = fmt.Errorf("metadata 'struct:error:name' already set for attribute %q of result type %q", "foo", identifier)
		errMissingMetadata    = fmt.Errorf("metadata 'struct:error:name' is missing in result type %q", identifier)
	)

	cases := map[string]struct {
		payload  *AttributeExpr
		result   *AttributeExpr
		errors   []*ErrorExpr
		expected *eval.ValidationErrors
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
					AttributeExpr: errorDuplicatedMetadata(),
					Name:          "foo",
				},
				{
					AttributeExpr: errorMissingMetadata(),
					Name:          "bar",
				},
			},
			expected: &eval.ValidationErrors{Errors: []error{
				errDuplicatedMetadata,
				errMissingMetadata,
			}},
		},
		"errors in all": {
			payload: attributeTypeNil(),
			result:  attributeTypeNil(),
			errors: []*ErrorExpr{
				{
					AttributeExpr: errorDuplicatedMetadata(),
					Name:          "foo",
				},
				{
					AttributeExpr: errorMissingMetadata(),
					Name:          "bar",
				},
			},
			expected: &eval.ValidationErrors{Errors: []error{
				errAttributeTypeNil,
				errAttributeTypeNil,
				errDuplicatedMetadata,
				errMissingMetadata,
			}},
		},
	}

	for k, tc := range cases {
		m := MethodExpr{
			Payload: tc.payload,
			Result:  tc.result,
			Errors:  tc.errors,

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
