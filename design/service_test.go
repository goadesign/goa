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

func TestFlowExpr_EvalName(t *testing.T) {
	const tokenURL = "http://domain/token"
	const refreshURL = "http://domain/refresh"

	cases := map[string]struct {
		tokenURL   string
		refreshURL string
		expected   string
	}{
		"tokenURL test":   {tokenURL: tokenURL, refreshURL: "", expected: fmt.Sprintf("flow with token URL %q", tokenURL)},
		"refreshURL test": {tokenURL: "", refreshURL: refreshURL, expected: fmt.Sprintf("flow with refresh URL %q", refreshURL)},
	}

	for k, tc := range cases {
		fe := &FlowExpr{TokenURL: tc.tokenURL, RefreshURL: tc.refreshURL}
		if actual := fe.EvalName(); actual != tc.expected {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}

func TestSchemeExpr_EvalName(t *testing.T) {
	cases := map[string]struct {
		kind     SchemeKind
		expected string
	}{
		"OAuth2Kind":    {kind: OAuth2Kind, expected: "OAuth2Security"},
		"BasicAuthKind": {kind: BasicAuthKind, expected: "BasicAuthSecurity"},
		"APIKeyKind":    {kind: APIKeyKind, expected: "APIKeySecurity"},
		"JWTKind":       {kind: JWTKind, expected: "JWTSecurity"},
		"NoKind":        {kind: NoKind, expected: "This case is panic"},
	}

	for k, tc := range cases {
		func() {
			// panic recover
			defer func() {
				if k != "NoKind" {
					return
				}

				if recover() == nil {
					t.Errorf("should have panicked!")
				}
			}()

			se := &SchemeExpr{Kind: tc.kind}
			if actual := se.EvalName(); actual != tc.expected {
				t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
			}
		}()
	}
}
