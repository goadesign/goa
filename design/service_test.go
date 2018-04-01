package design

import (
	"fmt"
	"testing"

	"goa.design/goa/eval"
)

func TestErrorExprValidate(t *testing.T) {
	const (
		identifier = "result"
	)
	var (
		metadata = MetadataExpr{
			"struct:error:name": []string{"error1"},
		}
		foo = &NamedAttributeExpr{
			Name: "foo",
			Attribute: &AttributeExpr{
				Metadata: metadata,
			},
		}
		bar = &NamedAttributeExpr{
			Name: "bar",
			Attribute: &AttributeExpr{
				Metadata: metadata,
			},
		}
		baz = &NamedAttributeExpr{
			Name: "foo",
			Attribute: &AttributeExpr{
				Metadata: MetadataExpr{},
			},
		}
	)
	cases := map[string]struct {
		att      *AttributeExpr
		expected *eval.ValidationErrors
	}{
		"no error": {
			att: &AttributeExpr{
				Type: &ResultTypeExpr{
					UserTypeExpr: &UserTypeExpr{
						AttributeExpr: &AttributeExpr{
							Type: &Object{
								foo,
							},
						},
					},
					Identifier: identifier,
				},
			},
			expected: &eval.ValidationErrors{
				Errors: []error{},
			},
		},
		"not result type": {
			att:      &AttributeExpr{Type: Boolean},
			expected: &eval.ValidationErrors{},
		},
		"duplicated metadata": {
			att: &AttributeExpr{
				Type: &ResultTypeExpr{
					UserTypeExpr: &UserTypeExpr{
						AttributeExpr: &AttributeExpr{
							Type: &Object{
								foo,
								bar,
							},
						},
					},
					Identifier: identifier,
				},
			},
			expected: &eval.ValidationErrors{
				Errors: []error{fmt.Errorf("metadata 'struct:error:name' already set for attribute %q of result type %q", "foo", identifier)},
			},
		},
		"missing metadata": {
			att: &AttributeExpr{
				Type: &ResultTypeExpr{
					UserTypeExpr: &UserTypeExpr{
						AttributeExpr: &AttributeExpr{
							Type: &Object{
								baz,
							},
						},
					},
					Identifier: identifier,
				},
			},
			expected: &eval.ValidationErrors{
				Errors: []error{fmt.Errorf("metadata 'struct:error:name' is missing in result type %q", identifier)},
			},
		},
	}

	for k, tc := range cases {
		e := ErrorExpr{
			AttributeExpr: tc.att,
		}
		if actual := e.Validate().(*eval.ValidationErrors); len(tc.expected.Errors) != len(actual.Errors) {
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
