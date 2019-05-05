package expr

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
	"goa.design/goa/v3/eval"
)

// RunDSL returns the DSL root resulting from running the given DSL.
// Used only in tests.
func RunDSL(t *testing.T, dsl func()) *RootExpr {
	setupDSLRun()

	// run DSL (first pass)
	if !eval.Execute(dsl, nil) {
		t.Fatal(eval.Context.Error())
	}

	// run DSL (second pass)
	if err := eval.RunDSL(); err != nil {
		t.Fatal(err)
	}

	// return generated root
	return Root
}

// RunInvalidDSL returns the error resulting from running the given DSL.
// It is used only in tests.
func RunInvalidDSL(t *testing.T, dsl func()) error {
	setupDSLRun()

	// run DSL (first pass)
	if !eval.Execute(dsl, nil) {
		return eval.Context.Errors
	}

	// run DSL (second pass)
	if err := eval.RunDSL(); err != nil {
		return err
	}

	// expected an error - didn't get one
	t.Fatal("expected a DSL evaluation error - got none")

	return nil
}

// Diff returns a diff between s1 and s2. It uses the diff tool if installed
// otherwise degrades to using the dmp package.
func Diff(t *testing.T, s1, s2 string) string {
	_, err := exec.LookPath("diff")
	supportsDiff := (err == nil)
	if !supportsDiff {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(s1, s2, false)
		return dmp.DiffPrettyText(diffs)
	}
	left := CreateTempFile(t, s1)
	right := CreateTempFile(t, s2)
	defer os.Remove(left)
	defer os.Remove(right)
	cmd := exec.Command("diff", left, right)
	diffb, _ := cmd.CombinedOutput()
	return strings.Replace(string(diffb), "\t", " ␉ ", -1)
}

// CreateTempFile creates a temporary file and writes the given content.
// It is used only for testing.
func CreateTempFile(t *testing.T, content string) string {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.WriteString(content)
	if err != nil {
		os.Remove(f.Name())
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return f.Name()
}

func setupDSLRun() {
	// reset all roots and codegen data structures
	eval.Reset()
	Root = new(RootExpr)
	Root.GeneratedTypes = &GeneratedRoot{}
	eval.Register(Root)
	eval.Register(Root.GeneratedTypes)
	Root.API = NewAPIExpr("test api", func() {})
	Root.API.Servers = []*ServerExpr{Root.API.DefaultServer()}
}
