package codegen

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// RunDSL returns the DSL root resulting from running the given DSL.
func RunDSL(t *testing.T, dsl func()) *expr.RootExpr {
	t.Helper()
	eval.Reset()
	expr.Root = new(expr.RootExpr)
	expr.GeneratedResultTypes = new(expr.ResultTypesRoot)
	require.NoError(t, eval.Register(expr.Root))
	require.NoError(t, eval.Register(expr.GeneratedResultTypes))
	expr.Root.API = expr.NewAPIExpr("test api", func() {})
	expr.Root.API.Servers = []*expr.ServerExpr{expr.Root.API.DefaultServer()}
	require.True(t, eval.Execute(dsl, nil), eval.Context.Error())
	require.NoError(t, eval.RunDSL())
	return expr.Root
}

// SectionCode generates and formats the code for the given section.
func SectionCode(t *testing.T, section *SectionTemplate) string {
	return sectionCodeWithPrefix(t, section, "package foo\n")
}

// SectionsCode generates and formats the code for the given sections.
func SectionsCode(t *testing.T, sections []*SectionTemplate) string {
	codes := make([]string, len(sections))
	for i, section := range sections {
		codes[i] = sectionCodeWithPrefix(t, section, "package foo\n")
	}
	return strings.Join(codes, "\n")
}

// SectionCodeFromImportsAndMethods generates and formats the code for given import and method definition sections.
func SectionCodeFromImportsAndMethods(t *testing.T, importSection, methodSection *SectionTemplate) string {
	t.Helper()
	var code bytes.Buffer
	require.NoError(t, importSection.Write(&code))
	return sectionCodeWithPrefix(t, methodSection, code.String())
}

func sectionCodeWithPrefix(t *testing.T, section *SectionTemplate, prefix string) string {
	var code bytes.Buffer
	require.NoError(t, section.Write(&code))
	codestr := code.String()

	if len(prefix) > 0 {
		codestr = fmt.Sprintf("%s\n%s", prefix, codestr)
	}

	return FormatTestCode(t, codestr)
}

// FormatTestCode formats the given Go code. The code must correspond to the
// content of a valid Go source file (i.e. start with "package")
func FormatTestCode(t *testing.T, code string) string {
	t.Helper()
	tmp := CreateTempFile(t, code)
	defer os.Remove(tmp)
	require.NoError(t, finalizeGoSource(tmp))
	content, err := os.ReadFile(tmp)
	require.NoError(t, err)
	return strings.Join(strings.Split(string(content), "\n")[2:], "\n")
}

// CreateTempFile creates a temporary file and writes the given content.
// It is used only for testing.
func CreateTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "")
	require.NoError(t, err)
	_, err = f.WriteString(content)
	if err != nil {
		os.Remove(f.Name())
		t.Fatal(err)
	}
	require.NoError(t, f.Close())
	return f.Name()
}
