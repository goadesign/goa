package design

import (
	"testing"
)

func TestResultTypeExprIsError(t *testing.T) {
	cases := map[string]struct {
		identifier string
		expected   bool
	}{
		"error": {
			identifier: "application/vnd.goa.error",
			expected:   true,
		},
		"not error": {
			identifier: "application/vnd.goa.foo",
			expected:   false,
		},
	}

	for k, tc := range cases {
		r := ResultTypeExpr{
			Identifier: tc.identifier,
		}
		if actual := r.IsError(); tc.expected != actual {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}

func TestResultTypeExprComputeViews(t *testing.T) {
	var (
		foo = &ViewExpr{
			Name: "foo",
		}
		bar = &ViewExpr{
			Name: "bar",
		}
		baz = &ViewExpr{
			Name: "baz",
		}
		qux = &ViewExpr{
			Name: "qux",
		}
	)
	cases := map[string]struct {
		views    []*ViewExpr
		userType *UserTypeExpr
		expected []*ViewExpr
	}{
		"views": {
			views:    []*ViewExpr{foo, bar},
			expected: []*ViewExpr{foo, bar},
		},
		"views of result type array": {
			userType: &UserTypeExpr{
				AttributeExpr: &AttributeExpr{
					Type: &Array{
						ElemType: &AttributeExpr{
							Type: &ResultTypeExpr{
								Views: []*ViewExpr{baz, qux},
							},
						},
					},
				},
			},
			expected: []*ViewExpr{baz, qux},
		},
		"no view": {
			userType: &UserTypeExpr{
				AttributeExpr: &AttributeExpr{
					Type: Boolean,
				},
			},
			expected: nil,
		},
	}

	for k, tc := range cases {
		r := ResultTypeExpr{
			Views:        tc.views,
			UserTypeExpr: tc.userType,
		}
		if actual := r.ComputeViews(); len(tc.expected) != len(actual) {
			t.Errorf("%s: expected the number of views to match %d got %d ", k, len(tc.expected), len(actual))
		} else {
			for i, v := range actual {
				if v != tc.expected[i] {
					t.Errorf("%s: got %#v, expected %#v at index %d", k, v, tc.expected[i], i)
				}
			}
		}
	}
}

func TestCanonicalIdentifier(t *testing.T) {
	cases := map[string]struct {
		identifier string
		expected   string
	}{
		"standards": {
			identifier: "application/json",
			expected:   "application/json",
		},
		"standards with parameter": {
			identifier: "application/json; charset=utf-8",
			expected:   "application/json; charset=utf-8",
		},
		"vendor": {
			identifier: "application/vnd.goa.error+json",
			expected:   "application/vnd.goa.error",
		},
		"vendor with parameter": {
			identifier: "application/vnd.goa.error+json; type=collection",
			expected:   "application/vnd.goa.error; type=collection",
		},
	}

	for k, tc := range cases {
		if actual := CanonicalIdentifier(tc.identifier); tc.expected != actual {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}
