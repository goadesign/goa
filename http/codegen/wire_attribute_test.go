package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"
	"goa.design/goa/v3/expr"
)

func TestWireAttributeCustomRepresentation(t *testing.T) {
	raw := &expr.UserTypeExpr{TypeName: "Raw", AttributeExpr: &expr.AttributeExpr{
		Type: expr.Bytes,
		Meta: expr.MetaExpr{"struct:field:type": {"stdjson.RawMessage", "encoding/json", "stdjson"}},
	}}
	nested := &expr.UserTypeExpr{TypeName: "Nested", AttributeExpr: &expr.AttributeExpr{Type: raw}}
	for _, test := range []struct {
		name   string
		source *expr.AttributeExpr
		want   []string
	}{
		{"named", &expr.AttributeExpr{Type: raw}, []string{"stdjson.RawMessage", "encoding/json", "stdjson"}},
		{"nested", &expr.AttributeExpr{Type: nested}, []string{"stdjson.RawMessage", "encoding/json", "stdjson"}},
		{"override", &expr.AttributeExpr{Type: nested, Meta: expr.MetaExpr{"struct:field:type": {"custom.Bytes", "example.org/custom"}}}, []string{"custom.Bytes", "example.org/custom"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			copied := WireAttribute(test.source)
			require.Equal(t, expr.Bytes, copied.Type)
			require.Equal(t, test.want, copied.Meta["struct:field:type"])
			require.IsType(t, &expr.UserTypeExpr{}, test.source.Type)
			copied.Meta["struct:field:type"][0] = "changed"
			require.Equal(t, "stdjson.RawMessage", raw.Meta["struct:field:type"][0])
		})
	}
	require.Nil(t, nested.Meta)
}
