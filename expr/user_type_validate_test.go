package expr_test

import (
    "strings"
    "testing"

    "goa.design/goa/v3/expr"
    "goa.design/goa/v3/expr/testdata"
)

func TestDuplicateUserTypeNames(t *testing.T) {
    err := expr.RunInvalidDSL(t, testdata.DuplicateUserTypeNamesDSL)
    if err == nil {
        t.Fatal("expected error, got none")
    }
    if !strings.Contains(err.Error(), `type "P" defined twice`) {
        t.Errorf("unexpected error:\n%s", err)
    }
}

