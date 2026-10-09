// These checks validate raw authored defaults without generating an application.
// Selecting an outer JSON kind must leave nested fields to normal validation.
package expr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUntaggedUnionDefaultValidation(t *testing.T) {
	file := &AttributeExpr{Type: &Object{{Name: "uri", Attribute: &AttributeExpr{Type: String}}},
		Validation: &ValidationExpr{Required: []string{"uri"}}}
	for _, test := range []struct {
		name    string
		value   any
		failure string
	}{
		{name: "dynamic", value: "dynamic"},
		{name: "empty manifest", value: []any{}},
		{name: "manifest", value: []any{map[string]any{"uri": "file://record"}}},
		{name: "null file", value: []any{nil}, failure: "must not be nil"},
		{name: "missing address", value: []any{map[string]any{}}, failure: "missing required field"},
		{name: "extra field", value: []any{map[string]any{"uri": "file://record", "other": true}}, failure: "unknown field"},
	} {
		t.Run(test.name, func(t *testing.T) {
			attribute := &AttributeExpr{Type: &Union{TypeName: "Resources", Untagged: true, Values: []*NamedAttributeExpr{
				{Name: "manifest", Attribute: &AttributeExpr{Type: &Array{ElemType: file, NonNullableElems: true}}},
				{Name: "dynamic", Attribute: &AttributeExpr{Type: String, Validation: &ValidationExpr{Values: []any{"dynamic"}}}},
			}}, DefaultValue: test.value}
			failures := attribute.Validate("", attribute)
			if test.failure != "" {
				require.Contains(t, failures.Error(), test.failure)
				return
			}
			require.Empty(t, failures.Errors)
		})
	}
}

func TestUntaggedBranchUsesOuterJSONKind(t *testing.T) {
	union := &Union{TypeName: "Value", Untagged: true, Values: []*NamedAttributeExpr{
		{Name: "bytes", Attribute: &AttributeExpr{Type: Bytes}},
		{Name: "number", Attribute: &AttributeExpr{Type: Int64}},
		{Name: "enabled", Attribute: &AttributeExpr{Type: Boolean}},
		{Name: "files", Attribute: &AttributeExpr{Type: &Array{ElemType: &AttributeExpr{Type: String}}}},
		{Name: "record", Attribute: &AttributeExpr{Type: &Object{}}},
	}}
	for _, test := range []struct {
		name   string
		value  any
		branch string
	}{
		{name: "bytes", value: []byte{1, 2}, branch: "bytes"},
		{name: "text", value: "encoded", branch: "bytes"},
		{name: "integer precision", value: int64(9007199254740993), branch: "number"},
		{name: "boolean", value: false, branch: "enabled"},
		{name: "empty array", value: []any{}, branch: "files"},
		{name: "nested null", value: []any{nil}, branch: "files"},
		{name: "byte array", value: [2]byte{1, 2}, branch: "files"},
		{name: "object", value: map[string]any{"value": nil}, branch: "record"},
		{name: "null", value: nil},
		{name: "unsupported", value: make(chan int)},
	} {
		t.Run(test.name, func(t *testing.T) {
			branch := union.UntaggedBranch(test.value)
			if test.branch == "" {
				require.Nil(t, branch)
				return
			}
			require.NotNil(t, branch)
			require.Equal(t, test.branch, branch.Name)
		})
	}
	union.Values = append(union.Values, &NamedAttributeExpr{Name: "text", Attribute: &AttributeExpr{Type: String}})
	require.Nil(t, union.UntaggedBranch("ambiguous"))
}
