package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixReservedGo(t *testing.T) {
	cases := map[string]struct {
		w    string
		want string
	}{
		"predeclared type":           {w: "bool", want: "bool_"},
		"predeclared constant":       {w: "true", want: "true_"},
		"predeclared zero value":     {w: "nil", want: "nil_"},
		"predeclared function":       {w: "append", want: "append_"},
		"non predeclared identifier": {w: "foo", want: "foo"},
		"package":                    {w: "fmt", want: "fmt_"},
	}
	for k, tc := range cases {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, tc.want, fixReservedGo(tc.w))
		})
	}
}
