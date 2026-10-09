// These tests translate flattened union defaults through the existing HTTP
// field mapping. The branch keeps its designed wire name instead of the
// internal service JSON name, and ordinary tagged unions keep their envelope.
package codegen

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func TestFlattenedUnionHTTPDefault(t *testing.T) {
	branch := &expr.AttributeExpr{Type: &expr.Object{{Name: "reference:reference_id", Attribute: &expr.AttributeExpr{
		Type: expr.String, Meta: expr.MetaExpr{"struct:tag:json:name": {"stored_reference"}},
	}}}}
	union := &expr.Union{TypeName: "Choice", TypeKey: "resultType", Flatten: true, Values: []*expr.NamedAttributeExpr{{Name: "complete", Attribute: branch}}}
	actual := projectHTTPUnionDefault(union, reflect.ValueOf(map[string]any{"resultType": "complete", "reference": "done"}))
	require.Equal(t, map[string]any{"resultType": "complete", "reference_id": "done"}, actual)
	union.Flatten = false
	actual = projectHTTPUnionDefault(union, reflect.ValueOf(map[string]any{"resultType": "complete", "value": map[string]any{"reference": "done"}}))
	require.Equal(t, map[string]any{"resultType": "complete", "value": map[string]any{"reference_id": "done"}}, actual)
}

func TestUntaggedUnionHTTPDefault(t *testing.T) {
	branch := &expr.AttributeExpr{Type: &expr.Object{{Name: "reference:reference_id", Attribute: &expr.AttributeExpr{Type: expr.String}}}}
	union := &expr.Union{TypeName: "Choice", Untagged: true, Values: []*expr.NamedAttributeExpr{
		{Name: "complete", Attribute: branch},
		{Name: "dynamic", Attribute: &expr.AttributeExpr{Type: expr.String}},
	}}
	require.Equal(t, "dynamic", projectHTTPUnionDefault(union, reflect.ValueOf("dynamic")))
	require.Equal(t, map[string]any{"reference_id": "done"}, projectHTTPUnionDefault(union, reflect.ValueOf(map[string]any{"reference": "done"})))
}
