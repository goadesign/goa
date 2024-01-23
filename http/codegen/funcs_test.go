package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusCodeToHttpConst(t *testing.T) {
	cases := map[string]struct {
		Code     int
		Expected string
	}{
		"know-status-code":   {Code: 200, Expected: "http.StatusOK"},
		"unknow-status-code": {Code: 700, Expected: "700"},
	}
	for k, tc := range cases {
		actual := statusCodeToHTTPConst(tc.Code)
		assert.Equal(t, tc.Expected, actual, k)
	}
}
