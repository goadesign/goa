package jsonrepresentation

import (
	"testing"

	"github.com/stretchr/testify/require"
	"goa.design/goa/v3/expr"
)

func TestRawJSONRepresentation(t *testing.T) {
	raw := &expr.AttributeExpr{Type: expr.Bytes, Meta: expr.MetaExpr{"struct:field:type": {"json.RawMessage", "encoding/json"}}}
	named := &expr.UserTypeExpr{TypeName: "Raw", AttributeExpr: raw}
	alias := &expr.UserTypeExpr{TypeName: "Alias", AttributeExpr: &expr.AttributeExpr{Type: named}}
	for _, test := range []struct {
		name  string
		value *expr.AttributeExpr
		want  bool
	}{
		{"inline", raw, true},
		{"named", &expr.AttributeExpr{Type: named}, true},
		{"alias", &expr.AttributeExpr{Type: alias}, true},
		{"ordinary", &expr.AttributeExpr{Type: expr.Bytes}, false},
		{"declared alias", &expr.AttributeExpr{Type: expr.Bytes, Meta: expr.MetaExpr{"struct:field:type": {"stdjson.RawMessage", "encoding/json", "stdjson"}}}, true},
		{"alias mismatch", &expr.AttributeExpr{Type: expr.Bytes, Meta: expr.MetaExpr{"struct:field:type": {"other.RawMessage", "encoding/json", "stdjson"}}}, false},
		{"aliased lookalike", &expr.AttributeExpr{Type: expr.Bytes, Meta: expr.MetaExpr{"struct:field:type": {"stdjson.RawMessage", "example.test/json", "stdjson"}}}, false},
		{"pointer", &expr.AttributeExpr{Type: expr.Bytes, Meta: expr.MetaExpr{"struct:field:type": {"*json.RawMessage", "encoding/json"}}}, false},
		{"collection", &expr.AttributeExpr{Type: expr.Bytes, Meta: expr.MetaExpr{"struct:field:type": {"[]json.RawMessage", "encoding/json"}}}, false},

		{"lookalike", &expr.AttributeExpr{Type: expr.Bytes, Meta: expr.MetaExpr{"struct:field:type": {"json.RawMessage", "example.test/json"}}}, false},
		{"other representation", &expr.AttributeExpr{Type: named, Meta: expr.MetaExpr{"struct:field:type": {"custom.Bytes", "example.test/custom"}}}, false},
		{"wrong primitive", &expr.AttributeExpr{Type: expr.String, Meta: expr.MetaExpr{"struct:field:type": {"json.RawMessage", "encoding/json"}}}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, IsRawJSON(test.value))
		})
	}
}
