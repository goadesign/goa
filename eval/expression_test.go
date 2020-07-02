package eval

import (
	"testing"
)

type Expr int

func (e Expr) EvalName() string { return "test expression" }

func TestToExpressionSet(t *testing.T) {
	cases := []struct {
		Name        string
		Slice       []interface{}
		ExpectPanic bool
	}{
		{"simple", []interface{}{Expr(42)}, false},
		{"nil", nil, false},
		{"invalid", []interface{}{42}, true},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil && c.ExpectPanic {
					t.Errorf("test did not panic")
				}
			}()
			set := ToExpressionSet(c.Slice)
			if len(set) != len(c.Slice) {
				t.Errorf("got set of length %d, expected %d.", len(set), len(c.Slice))
			} else {
				for i, e := range set {
					if e != c.Slice[i] {
						t.Errorf("got value %v at index %d, expected %v", e, i, c.Slice[i])
					}
				}
			}
		})
	}
}
