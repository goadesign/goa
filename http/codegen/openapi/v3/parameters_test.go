package openapiv3

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func TestParamForAllowEmptyValue(t *testing.T) {
	tests := []struct {
		name     string
		location string
		want     bool
	}{
		{name: "query", location: "query", want: true},
		{name: "path", location: "path", want: false},
		{name: "header", location: "header", want: false},
		{name: "cookie", location: "cookie", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			param := paramFor(
				&expr.AttributeExpr{Type: expr.String},
				"value",
				test.location,
				false,
				expr.NewRandom(test.name),
			)

			require.Equal(t, test.want, param.AllowEmptyValue)
		})
	}
}
