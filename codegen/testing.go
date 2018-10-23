package codegen

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"

	"goa.design/goa/design"
	"goa.design/goa/eval"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// RunDSL returns the DSL root resulting from running the given DSL.
func RunDSL(t *testing.T, dsl func()) *design.RootExpr {
	eval.Reset()
	design.Root = new(design.RootExpr)
	design.Root.GeneratedTypes = &design.GeneratedRoot{}
	eval.Register(design.Root)
	eval.Register(design.Root.GeneratedTypes)
	design.Root.API = &design.APIExpr{Name: "test api"}
	design.Root.API.Servers = []*design.ServerExpr{design.Root.API.DefaultServer()}
	if !eval.Execute(dsl, nil) {
		t.Fatal(eval.Context.Error())
	}
	if err := eval.RunDSL(); err != nil {
		t.Fatal(err)
	}
	return design.Root
}

// RunDSLWithFunc returns the DSL root resulting from running the given DSL.
// It executes a function to add any top-level types to the design Root before
// running the DSL.
func RunDSLWithFunc(t *testing.T, dsl func(), fn func()) *design.RootExpr {
	eval.Reset()
	design.Root = new(design.RootExpr)
	design.Root.GeneratedTypes = &design.GeneratedRoot{}
	eval.Register(design.Root)
	eval.Register(design.Root.GeneratedTypes)
	design.Root.API = &design.APIExpr{Name: "test api"}
	design.Root.API.Servers = []*design.ServerExpr{design.Root.API.DefaultServer()}
	fn()
	if !eval.Execute(dsl, nil) {
		t.Fatal(eval.Context.Error())
	}
	if err := eval.RunDSL(); err != nil {
		t.Fatal(err)
	}
	return design.Root
}

// SectionCode generates and formats the code for the given section.
func SectionCode(t *testing.T, section *SectionTemplate) string {
	return sectionCodeWithPrefix(t, section, "package foo\n")
}

// SectionCodeFromImportsAndMethods generates and formats the code for given import and method definition sections.
func SectionCodeFromImportsAndMethods(t *testing.T, importSection *SectionTemplate, methodSection *SectionTemplate) string {
	var code bytes.Buffer
	if err := importSection.Write(&code); err != nil {
		t.Fatal(err)
	}

	return sectionCodeWithPrefix(t, methodSection, code.String())
}

func sectionCodeWithPrefix(t *testing.T, section *SectionTemplate, prefix string) string {
	var code bytes.Buffer
	if err := section.Write(&code); err != nil {
		t.Fatal(err)
	}

	codestr := code.String()

	if len(prefix) > 0 {
		codestr = fmt.Sprintf("%s\n%s", prefix, codestr)
	}

	return FormatTestCode(t, codestr)
}

// FormatTestCode formats the given Go code. The code must correspond to the
// content of a valid Go source file (i.e. start with "package")
func FormatTestCode(t *testing.T, code string) string {
	tmp := createTempFile(t, code)
	defer os.Remove(tmp)
	if err := finalizeGoSource(tmp); err != nil {
		t.Fatal(err)
	}
	content, err := ioutil.ReadFile(tmp)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Join(strings.Split(string(content), "\n")[2:], "\n")
}

// Diff returns a diff between s1 and s2. It uses the diff tool if installed
// otherwise degrades to using the dmp package.
func Diff(t *testing.T, s1, s2 string) string {
	_, err := exec.LookPath("diff")
	supportsDiff := err == nil
	if !supportsDiff {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(s1, s2, false)
		return dmp.DiffPrettyText(diffs)
	}
	left := createTempFile(t, s1)
	right := createTempFile(t, s2)
	defer os.Remove(left)
	defer os.Remove(right)
	cmd := exec.Command("diff", left, right)
	diffb, _ := cmd.CombinedOutput()
	return strings.Replace(string(diffb), "\t", " ␉ ", -1)
}

func createTempFile(t *testing.T, content string) string {
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
