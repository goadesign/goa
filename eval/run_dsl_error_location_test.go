package eval_test

import (
	"errors"
	"reflect"
	"runtime"
	"testing"

	"goa.design/goa/v3/eval"
)

func TestRunDSL_ReportErrorLocation(t *testing.T) {
	eval.Reset()

	var expectedFile string
	var expectedLine int
	expr := &runDSLExpr{
		name: "expr",
		dsl: func() {
			_, file, line, ok := runtime.Caller(0)
			eval.ReportError("boom")
			if !ok {
				t.Fatal("runtime.Caller failed")
			}
			expectedFile = relativeToWorkdir(t, file)
			expectedLine = line + 1
		},
	}

	if err := eval.Register(&runDSLRoot{expr: expr}); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	err := eval.RunDSL()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var merr eval.MultiError
	if !errors.As(err, &merr) {
		t.Fatalf("expected MultiError, got %T", err)
	}
	if len(merr) != 1 {
		t.Fatalf("expected 1 error, got %d", len(merr))
	}

	got := merr[0]
	if got.File != expectedFile {
		t.Fatalf("unexpected file: got %q, expected %q", got.File, expectedFile)
	}
	if got.Line != expectedLine {
		t.Fatalf("unexpected line: got %d, expected %d", got.Line, expectedLine)
	}
}

func TestRunDSL_ValidationErrorLocation(t *testing.T) {
	eval.Reset()

	dsl := func() {}
	expr := &runDSLExpr{
		name:     "expr",
		dsl:      dsl,
		validate: errors.New("bad"),
	}

	if err := eval.Register(&runDSLRoot{expr: expr}); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	err := eval.RunDSL()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var merr eval.MultiError
	if !errors.As(err, &merr) {
		t.Fatalf("expected MultiError, got %T", err)
	}
	if len(merr) != 1 {
		t.Fatalf("expected 1 error, got %d", len(merr))
	}

	got := merr[0]
	if got.File != "" {
		t.Fatalf("unexpected file: got %q, expected empty (validation errors are embedded)", got.File)
	}
	if got.Line != 0 {
		t.Fatalf("unexpected line: got %d, expected 0 (validation errors are embedded)", got.Line)
	}

	expectedFile, expectedLine := dslDeclLocation(t, dsl)
	expected := "[" + expectedFile + ":" + itoa(expectedLine) + "] expr: bad"
	if got := err.Error(); got != expected {
		t.Fatalf("unexpected error message:\n got: %q\nwant: %q", got, expected)
	}
}

type runDSLRoot struct {
	expr eval.Expression
}

func (*runDSLRoot) EvalName() string { return "test" }
func (*runDSLRoot) DependsOn() []eval.Root {
	return nil
}
func (*runDSLRoot) Packages() []string {
	return nil
}
func (r *runDSLRoot) WalkSets(walk eval.SetWalker) {
	walk(eval.ExpressionSet{r.expr})
}

type runDSLExpr struct {
	name     string
	dsl      func()
	validate error
}

func (e *runDSLExpr) EvalName() string { return e.name }
func (e *runDSLExpr) DSL() func()      { return e.dsl }
func (e *runDSLExpr) Validate() error  { return e.validate }

func dslDeclLocation(t *testing.T, fn func()) (file string, line int) {
	t.Helper()

	pc := reflect.ValueOf(fn).Pointer()
	f := runtime.FuncForPC(pc)
	if f == nil {
		t.Fatal("runtime.FuncForPC returned nil")
	}
	file, line = f.FileLine(pc)
	return relativeToWorkdir(t, file), line
}
