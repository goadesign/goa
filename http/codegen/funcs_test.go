package codegen

import (
	"testing"
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
		if actual != tc.Expected {
			t.Errorf("%s: got `%s`, expected `%s`", k, actual, tc.Expected)
		}
	}
}
