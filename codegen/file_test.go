// These tests compare complete source finalization with the previous formatter
// and verify which bytes File.Render writes before invoking a file callback.
package codegen

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/scanner"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/imports"
)

func TestFinalizeGoSourceMatchesPreviousFormatter(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		localPrefix string
		invalid     bool
	}{
		{
			name: "grouped imports",
			source: `package sample
import (
 "strings"
 "fmt"
)
func Label(s string) string { return fmt.Sprint(strings.TrimSpace(s)) }
`,
		},
		{
			name: "single import",
			source: `package sample
import "fmt"
func Label(v int) string { return fmt.Sprint(v) }
`,
		},
		{
			name:   "no imports",
			source: "package sample\nfunc Value() int { return 3 }\n",
		},
		{
			name: "aliases blank and dot imports",
			source: `package sample
import (
 text "strings"
 _ "net/http/pprof"
 . "math"
 output "fmt"
)
func Label(v string) string { return output.Sprint(text.ToUpper(v), Pi) }
`,
		},
		{
			name: "unused import comments",
			source: `package sample
import (
 // Discarded package comment.
 "bytes" // Discarded package suffix.
 // Text operations.
 "strings" // Retained package suffix.
)
// Label formats text.
func Label(s string) string { return strings.TrimSpace(s) }
`,
		},
		{
			name: "all imports removed",
			source: `package sample
// Import declaration comment.
import (
 "bytes" // Unused package.
)
const Value = 1
`,
		},
		{
			name: "separate declarations and import groups",
			source: `package sample
import "strings"
// Output helpers.
import (
 "fmt"

 // External helpers.
 helper "example.com/helper"
)
func Label(s string) string { return fmt.Sprint(strings.TrimSpace(s), helper.Value) }
`,
		},
		{
			name:        "local import prefix",
			localPrefix: "example.com/local",
			source: `package sample
import (
 local "example.com/local/value"
 "fmt"
 external "example.net/helper"
)
var Label = fmt.Sprint(local.Value, external.Value)
`,
		},
		{
			name: "numeric literals",
			source: `package sample
import (
 "fmt"
 "strings"
)
const ( Hex = 0XAB; Binary = 0B101; Octal = 0O77; Float = 1E+10; HexFloat = 0X1.FP+2 )
var Label = strings.TrimSpace(fmt.Sprint(Hex, Binary, Octal, Float, HexFloat))
`,
		},
		{
			name: "build and line directives",
			source: `//go:build linux && amd64
// +build linux,amd64

package sample
import (
 "fmt"
 "strings"
)
//line generated.go:40
func Label(s string) string { return fmt.Sprint(strings.TrimSpace(s)) }
`,
		},
		{
			name: "raw strings and documentation",
			source: "package sample\nimport (\"fmt\"; \"strings\")\n" +
				"// Label formats text.\n//\n// Example:\n//\n//     Label(\"hello\")\n" +
				"func Label(s string) string { return fmt.Sprint(strings.TrimSpace(s), `first\n\tsecond`) }\n",
		},
		{
			name:    "malformed source",
			source:  "package sample\nfunc Broken( {\n",
			invalid: true,
		},
		{
			name:   "complete synthetic file",
			source: string(finalizeBenchmarkSource(64)),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			previousPrefix := imports.LocalPrefix
			imports.LocalPrefix = test.localPrefix
			t.Cleanup(func() { imports.LocalPrefix = previousPrefix })

			source := []byte(test.source)
			want, wantErr := finalizeGoSourcePrevious("sample.go", source)
			got, gotErr := finalizeGoSource("sample.go", source)
			if test.invalid {
				require.Error(t, wantErr)
				require.EqualError(t, gotErr, wantErr.Error())
			} else {
				require.NoError(t, wantErr)
				require.NoError(t, gotErr)
			}
			require.Equal(t, want, got)
			require.Equal(t, test.source, string(source), "source bytes must remain unchanged")
		})
	}
}

func TestFileRenderPreservesWriteAndCallbackBehavior(t *testing.T) {
	callbackError := errors.New("callback failed")
	tests := []struct {
		name        string
		filename    string
		source      string
		existing    string
		skip        bool
		callback    bool
		callbackErr error
		want        string
		wantError   error
		invalid     bool
	}{
		{
			name:     "formatted Go before callback",
			filename: "sample.go",
			source:   "package sample\nvar Value=1\n",
			callback: true,
			want:     "package sample\n\nvar Value = 1\n",
		},
		{
			name:     "skip existing before parsing",
			filename: "sample.go",
			source:   "not Go",
			existing: "existing bytes",
			skip:     true,
			callback: true,
			want:     "existing bytes",
		},
		{
			name:     "syntax failure preserves existing file",
			filename: "sample.go",
			source:   "package sample\nfunc Broken( {\n",
			existing: "existing bytes",
			callback: true,
			want:     "existing bytes",
			invalid:  true,
		},
		{
			name:     "syntax failure creates no file",
			filename: "sample.go",
			source:   "package sample\nfunc Broken( {\n",
			callback: true,
			invalid:  true,
		},
		{
			name:        "callback failure retains written bytes",
			filename:    "sample.go",
			source:      "package sample\nvar Value=1\n",
			callback:    true,
			callbackErr: callbackError,
			wantError:   callbackError,
			want:        "package sample\n\nvar Value = 1\n",
		},
		{
			name:     "non Go bytes",
			filename: "sample.txt",
			source:   "  arbitrary\x00bytes\r\n",
			callback: true,
			want:     "  arbitrary\x00bytes\r\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			directory := t.TempDir()
			filename := filepath.Join(directory, test.filename)
			if test.existing != "" {
				require.NoError(t, os.WriteFile(filename, []byte(test.existing), 0600))
			}
			file := &File{
				Path:             test.filename,
				SkipExist:        test.skip,
				SectionTemplates: []*SectionTemplate{{Name: "source", Source: test.source}},
			}
			called := false
			if test.callback {
				file.FinalizeFunc = func(path string) error {
					called = true
					require.Equal(t, filename, path)
					written, err := os.ReadFile(path)
					require.NoError(t, err)
					require.Equal(t, test.want, string(written))
					return test.callbackErr
				}
			}
			path, err := file.Render(directory)
			switch {
			case test.invalid:
				require.Error(t, err)
				require.Contains(t, err.Error(), filename+":2:")
				require.Contains(t, err.Error(), "\n========\nContent:\n"+test.source)
			case test.wantError != nil:
				require.ErrorIs(t, err, test.wantError)
			default:
				require.NoError(t, err)
			}
			if test.skip || test.invalid || test.wantError != nil {
				require.Empty(t, path)
			} else {
				require.Equal(t, filename, path)
			}
			require.Equal(t, test.callback && !test.skip && !test.invalid, called)
			written, readErr := os.ReadFile(filename)
			if test.invalid && test.existing == "" {
				require.ErrorIs(t, readErr, os.ErrNotExist)
			} else {
				require.NoError(t, readErr)
				require.Equal(t, test.want, string(written))
			}
		})
	}
}

func BenchmarkFinalizeGoSource(b *testing.B) {
	for _, declarations := range []int{64, 512} {
		source := finalizeBenchmarkSource(declarations)
		b.Run(fmt.Sprintf("declarations_%d", declarations), func(b *testing.B) {
			for _, finalizer := range []struct {
				name string
				run  func(string, []byte) ([]byte, error)
			}{
				{name: "previous", run: finalizeGoSourcePrevious},
				{name: "direct_printer", run: finalizeGoSource},
			} {
				b.Run(finalizer.name, func(b *testing.B) {
					b.ReportAllocs()
					b.SetBytes(int64(len(source)))
					for b.Loop() {
						_, err := finalizer.run("sample.go", source)
						if err != nil {
							b.Fatal(err)
						}
					}
				})
			}
		})
	}
}

// finalizeGoSourcePrevious keeps the old complete finalization flow independent
// of the current printer so comparisons detect changes in output and errors.
func finalizeGoSourcePrevious(path string, content []byte) ([]byte, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, content, parser.ParseComments)
	if err != nil {
		var buf bytes.Buffer
		scanner.PrintError(&buf, err)
		return nil, fmt.Errorf("%s\n========\nContent:\n%s", buf.String(), content)
	}
	impMap := buildImportMap(file)
	detectUsedImports(file, impMap)
	removeUnusedImports(fset, file, impMap)
	ast.SortImports(fset, file)
	var formatted bytes.Buffer
	if err := format.Node(&formatted, fset, file); err != nil {
		return nil, err
	}
	opt := imports.Options{Comments: true, FormatOnly: true}
	result, err := imports.Process(path, formatted.Bytes(), &opt)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// finalizeBenchmarkSource builds a complete synthetic client-style file with
// repeated request types and methods; both finalizers receive these same bytes.
func finalizeBenchmarkSource(declarations int) []byte {
	var source strings.Builder
	source.WriteString("package sample\nimport (\"strings\"; \"fmt\"; \"context\")\n")
	for index := range declarations {
		fmt.Fprintf(&source, `
// Request%d carries one request's input.
type Request%d struct {
 Name string
 Labels map[string]string
 Values []int
}
// Call%d validates input and returns a formatted result.
func Call%d(ctx context.Context, request *Request%d) (string, error) {
 if err := ctx.Err(); err != nil { return "", err }
 name := strings.TrimSpace(request.Name)
 if name == "" { return "", fmt.Errorf("name is required") }
 for key, value := range request.Labels {
  if strings.TrimSpace(key) == "" { return "", fmt.Errorf("invalid label: %%s", value) }
 }
 return fmt.Sprintf("%%s:%%d", name, len(request.Values)), nil
}
`, index, index, index, index, index)
	}
	return []byte(source.String())
}
