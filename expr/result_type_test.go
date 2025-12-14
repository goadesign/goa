package expr_test

import (
	"strings"
	"testing"

	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/expr/testdata"
)

func TestDuplicateResultTypeNames(t *testing.T) {
	err := expr.RunInvalidDSL(t, testdata.DuplicateResultTypeNamesDSL)
	if err == nil {
		t.Fatal("expected error, got none")
	}
	// Root validation prefixes with the expression EvalName ("design").
	if !strings.Contains(err.Error(), `result type "A" defined twice`) {
		t.Errorf("unexpected error:\n%s", err)
	}
}
