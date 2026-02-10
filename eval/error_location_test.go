package eval_test

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"goa.design/goa/v3/eval"
)

func TestReportErrorRecordsLocation(t *testing.T) {
	t.Skip("computeErrorLocation intentionally skips eval package frames; this is tested from another package")
}

func TestValidationErrorsIncludeLocation(t *testing.T) {
	var verr eval.ValidationErrors

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dsl := func() {}

	pc := reflect.ValueOf(dsl).Pointer()
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		t.Fatal("runtime.FuncForPC returned nil")
	}
	_, expectedLine := fn.FileLine(pc)
	expectedFile := relativeToWorkdir(t, file)

	expr := &testExpr{name: "expr", dsl: dsl}
	verr.AddError(expr, errors.New("bad"))

	expected := "[" + expectedFile + ":" + itoa(expectedLine) + "] expr: bad"
	if got := verr.Error(); got != expected {
		t.Fatalf("unexpected error message:\n got: %q\nwant: %q", got, expected)
	}
}

func TestValidationErrorsWithoutLocation(t *testing.T) {
	var verr eval.ValidationErrors
	verr.AddError(eval.Top, errors.New("bad"))
	if got, expected := verr.Error(), "top-level: bad"; got != expected {
		t.Fatalf("unexpected error message:\n got: %q\nwant: %q", got, expected)
	}
}

func TestValidationErrorsNilDSL(t *testing.T) {
	var verr eval.ValidationErrors
	expr := &testExpr{name: "expr", dsl: nil}
	verr.AddError(expr, errors.New("bad"))
	if got, expected := verr.Error(), "expr: bad"; got != expected {
		t.Fatalf("unexpected error message:\n got: %q\nwant: %q", got, expected)
	}
}

type testExpr struct {
	name string
	dsl  func()
}

func (e *testExpr) EvalName() string { return e.name }
func (e *testExpr) DSL() func()      { return e.dsl }

func relativeToWorkdir(t *testing.T, file string) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd failed: %v", err)
	}
	wd, err = filepath.Abs(wd)
	if err != nil {
		t.Fatalf("filepath.Abs failed: %v", err)
	}
	rel, err := filepath.Rel(wd, file)
	if err != nil {
		t.Fatalf("filepath.Rel failed: %v", err)
	}
	return rel
}

func itoa(v int) string {
	// Avoid pulling fmt/strconv into this focused test.
	if v == 0 {
		return "0"
	}
	var buf [32]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}
