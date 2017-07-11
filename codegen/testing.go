package codegen

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// FormatTestCode formats the given Go code. The code must correspond to the
// content of a valid Go source file (i.e. start with "package")
func FormatTestCode(t *testing.T, code string) string {
	tmp := createTempFile(t, code)
	defer os.Remove(tmp)
	if err := Format(tmp); err != nil {
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
	supportsDiff := (err == nil)
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
