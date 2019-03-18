package expr

import (
	"testing"
)

func TestResultTypeExprView(t *testing.T) {
	var (
		viewFoo = &ViewExpr{
			Name: "foo",
		}
		viewBar = &ViewExpr{
			Name: "bar",
		}
	)
	cases := map[string]struct {
		name     string
		expected *ViewExpr
	}{
		"exist": {
			name:     "foo",
			expected: viewFoo,
		},
		"not exist": {
			name:     "baz",
			expected: nil,
		},
	}

	for k, tc := range cases {
		r := ResultTypeExpr{
			Views: []*ViewExpr{
				viewFoo,
				viewBar,
			},
		}
		if actual := r.View(tc.name); actual != tc.expected {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}

func TestResultTypeExprHasMultipleViews(t *testing.T) {
	var (
		viewFoo = &ViewExpr{
			Name: "foo",
		}
		viewBar = &ViewExpr{
			Name: "bar",
		}
	)
	cases := map[string]struct {
		views    []*ViewExpr
		expected bool
	}{
		"multiple views": {
			views: []*ViewExpr{
				viewFoo,
				viewBar,
			},
			expected: true,
		},
		"one view": {
			views: []*ViewExpr{
				viewFoo,
			},
			expected: false,
		},
		"no view": {
			views:    []*ViewExpr{},
			expected: false,
		},
	}

	for k, tc := range cases {
		t.Run(k, func(t *testing.T) {
			r := ResultTypeExpr{
				Views: tc.views,
			}
			if actual := r.HasMultipleViews(); actual != tc.expected {
				t.Errorf("got %#v, expected %#v", actual, tc.expected)
			}
		})
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

func TestViewExprEvalName(t *testing.T) {
	var (
		result = &ResultTypeExpr{
			UserTypeExpr: &UserTypeExpr{
				AttributeExpr: &AttributeExpr{},
			},
		}
	)
	cases := map[string]struct {
		name     string
		parent   *ResultTypeExpr
		expected string
	}{
		"empty name and empty parent": {
			name:     "",
			parent:   nil,
			expected: "unnamed view",
		},
		"name only": {
			name:     "foo",
			parent:   nil,
			expected: `view "foo"`,
		},
		"parent only": {
			name:     "",
			parent:   result,
			expected: "unnamed view of attribute",
		},
		"both name and parent": {
			name:     "foo",
			parent:   result,
			expected: `view "foo" of attribute`,
		},
	}

	for k, tc := range cases {
		view := ViewExpr{
			Name:   tc.name,
			Parent: tc.parent,
		}
		if actual := view.EvalName(); actual != tc.expected {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}
