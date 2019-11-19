package expr_test

import (
	"strings"
	"testing"

	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/expr/testdata"
)

func TestFilesDSL(t *testing.T) {
	cases := []struct {
		Name  string
		DSL   func()
		Error string
	}{
		{Name: "valid", DSL: testdata.FilesValidDSL},
		{Name: "incompatible", DSL: testdata.FilesIncompatibleDSL, Error: "invalid use of Files in API files-incompatile"},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if c.Error == "" {
				expr.RunDSL(t, c.DSL)
			} else {
				err := expr.RunInvalidDSL(t, c.DSL)
				if !strings.HasSuffix(err.Error(), c.Error) {
					t.Errorf("got error %q, expected has suffix %q", err.Error(), c.Error)
				}
			}
		})
	}
}
