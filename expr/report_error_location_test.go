package expr_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"goa.design/goa/v3/eval"
)

func TestReportErrorRecordsLocation(t *testing.T) {
	eval.Reset()

	file, line := reportErrorHere(t)

	if eval.Context.Errors == nil || len(eval.Context.Errors) != 1 {
		t.Fatalf("expected exactly one error, got %#v", eval.Context.Errors)
	}

	got := eval.Context.Errors[0]
	expectedFile := relativeToWorkdir(t, file)
	if got.File != expectedFile {
		t.Fatalf("unexpected file: got %q, expected %q", got.File, expectedFile)
	}
	expectedLine := line
	if got.Line != expectedLine {
		t.Fatalf("unexpected line: got %d, expected %d", got.Line, expectedLine)
	}
}

//go:noinline
func reportErrorHere(t *testing.T) (file string, line int) {
	t.Helper()

	if _, file, line, ok := runtime.Caller(0); ok {
		eval.ReportError("boom")
		return file, line + 1
	}

	t.Fatal("runtime.Caller failed")
	return "", 0
}

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
