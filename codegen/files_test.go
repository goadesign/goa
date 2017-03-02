package codegen

import (
	"bytes"
	"testing"
)

func TestHeaderTmpl(t *testing.T) {
	cases := map[string]struct {
		data     map[string]interface{}
		expected string
	}{
		"a minimum header": {
			data: map[string]interface{}{
				"Pkg": "foo",
				"Imports": []*ImportSpec{
					&ImportSpec{Path: "context"},
					&ImportSpec{Path: "goa.design/goa.v2"},
				},
			},
			expected: `package foo

import (
	"context"
	"goa.design/goa.v2"
)

`,
		},
	}
	for k, tc := range cases {
		buf := new(bytes.Buffer)
		if err := headerTmpl.ExecuteTemplate(buf, "header", tc.data); err != nil {
			t.Fatalf("ExecuteTemplate returned %s", err)
		}
		actual := buf.String()
		if actual != tc.expected {
			t.Errorf("%s: got %v, expected %v", k, actual, tc.expected)
		}
	}
}
