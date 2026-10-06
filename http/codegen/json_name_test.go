// These tests give HTTP fields internal JSON names and check that generated
// body tags and command-line defaults still use the designed HTTP names.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func TestHTTPFieldTagsIgnoreNonTransportJSONName(t *testing.T) {
	for _, test := range []struct {
		name     string
		meta     expr.MetaExpr
		optional bool
		want     string
	}{
		{
			name: "required",
			meta: expr.MetaExpr{"struct:tag:json:name": {"stored_name"}},
			want: " `form:\"publicName\" json:\"publicName\" xml:\"publicName\"`",
		},
		{
			name:     "optional",
			meta:     expr.MetaExpr{"struct:tag:json:name": {"stored_name"}},
			optional: true,
			want:     " `form:\"publicName,omitempty\" json:\"publicName,omitempty\" xml:\"publicName,omitempty\"`",
		},
		{
			name: "complete tag still overrides",
			meta: expr.MetaExpr{
				"struct:tag:json:name": {"stored_name"},
				"struct:tag:json":      {"explicit_name,omitempty"},
			},
			want: " `json:\"explicit_name,omitempty\"`",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			attribute := &expr.AttributeExpr{Type: expr.String, Meta: test.meta}
			require.Equal(t, test.want, attributeTags(attribute, "publicName", test.optional))
		})
	}
}

func TestHTTPDefaultNestedFieldsIgnoreNonTransportJSONName(t *testing.T) {
	attribute := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "items", Attribute: &expr.AttributeExpr{Type: &expr.Array{
			ElemType: &expr.AttributeExpr{Type: &expr.Object{
				{Name: "name:publicName", Attribute: &expr.AttributeExpr{
					Type: expr.String,
					Meta: expr.MetaExpr{"struct:tag:json:name": {"stored_name"}},
				}},
			}},
		}}},
	}}
	value := map[string]any{"items": []any{map[string]any{"name": "sample"}}}
	want := map[string]any{"items": []any{map[string]any{"publicName": "sample"}}}
	require.Equal(t, want, clientBodyDefault(attribute, value))
}
