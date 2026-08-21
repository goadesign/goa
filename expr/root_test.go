// This file verifies root validation, including exact-origin dependency
// traversal for explicitly relocated user types.
package expr

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"goa.design/goa/v3/eval"
)

func TestRelocatedDependenciesUseDeclarationOrigin(t *testing.T) {
	dependency := &UserTypeExpr{
		TypeName:      "Dependency",
		UID:           "shared-semantic-id",
		AttributeExpr: &AttributeExpr{Type: String},
	}
	relocated := &UserTypeExpr{
		TypeName: "Relocated",
		UID:      "shared-semantic-id",
		AttributeExpr: &AttributeExpr{
			Meta: MetaExpr{"struct:pkg:path": {"types"}},
			Type: &Object{&NamedAttributeExpr{
				Name:      "dependency",
				Attribute: &AttributeExpr{Type: dependency},
			}},
		},
	}
	root := &RootExpr{Types: []UserType{relocated, dependency}}

	errors := root.validateRelocatedUserTypes()
	if len(errors.Errors) != 1 {
		t.Fatalf("expected one relocated dependency error, got %d", len(errors.Errors))
	}
	if message := errors.Errors[0].Error(); !strings.Contains(message, "Dependency") {
		t.Errorf("expected dependency name in error, got %q", message)
	}
}

func TestRelocatedDependencyWalkStopsAtExactOriginCopy(t *testing.T) {
	relocated := &UserTypeExpr{
		TypeName: "Relocated",
		UID:      "relocated",
		AttributeExpr: &AttributeExpr{
			Meta: MetaExpr{"struct:pkg:path": {"types"}},
			Type: String,
		},
	}
	copy := relocated.Dup(DupAtt(relocated.Attribute()))
	relocated.AttributeExpr.Type = &Object{&NamedAttributeExpr{
		Name:      "self",
		Attribute: &AttributeExpr{Type: copy},
	}}
	root := &RootExpr{Types: []UserType{relocated}}

	errors := root.validateRelocatedUserTypes()
	if len(errors.Errors) != 0 {
		t.Errorf("expected exact origin copy to be treated as recursion, got %v", errors)
	}
}

func TestRootExprValidate(t *testing.T) {
	cases := map[string]struct {
		api      *APIExpr
		expected *eval.ValidationErrors
	}{
		"no error": {
			api: &APIExpr{
				Name: "foo",
			},
			expected: &eval.ValidationErrors{
				Errors: []error{},
			},
		},
		"missing api declaration": {
			api: nil,
			expected: &eval.ValidationErrors{
				Errors: []error{fmt.Errorf("Missing API declaration")},
			},
		},
	}

	for k, tc := range cases {
		e := RootExpr{
			API: tc.api,
		}
		var actual *eval.ValidationErrors
		if errors.As(e.Validate(), &actual); len(tc.expected.Errors) != len(actual.Errors) {
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

func TestMetaExpr_Last(t *testing.T) {
	tt := map[string]struct {
		meta  MetaExpr
		value string
		ok    bool
	}{
		"no-key": {
			MetaExpr{},
			"",
			false,
		},
		"key-no-values": {
			MetaExpr{
				"test:key": []string{},
			},
			"",
			false,
		},
		"key-with-one-value": {
			MetaExpr{
				"test:key": []string{
					"value-one",
				},
			},
			"value-one",
			true,
		},
		"key-with-multiple-values": {
			MetaExpr{
				"test:key": []string{
					"value-one",
					"value-two",
					"value-n",
				},
			},
			"value-n",
			true,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			value, ok := tc.meta.Last("test:key")
			if tc.ok != ok {
				t.Errorf("expected ok to be %v, got %v", tc.ok, ok)
			}
			if tc.value != value {
				t.Errorf("expected value to be %s, got %s", value, value)
			}
		})
	}
}
