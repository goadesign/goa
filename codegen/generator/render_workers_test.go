// These tests check that rendering respects the process CPU budget and still
// completes every file before returning the first error in file order.
package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
)

func TestRenderWorkersRespectCPUBudget(t *testing.T) {
	for _, test := range []struct {
		budget int
		fail   bool
	}{{1, false}, {2, false}, {8, false}, {2, true}} {
		t.Run(fmt.Sprintf("budget%d/fail%t", test.budget, test.fail), func(t *testing.T) {
			previous := runtime.GOMAXPROCS(test.budget)
			defer runtime.GOMAXPROCS(previous)
			finalizeMergeTestRoots(t)
			var calls atomic.Int32
			registry := testRegistryFromGenfuncs([]testGenfunc{
				testRenderOnly(func(_ string, _ []eval.Root) ([]*codegen.File, error) {
					files := make([]*codegen.File, 0, 5)
					for index := range 5 {
						files = append(files, &codegen.File{
							Path:             filepath.Join(codegen.Gendir, fmt.Sprintf("file%d.txt", index)),
							SectionTemplates: []*codegen.SectionTemplate{{Name: "body", Source: "complete\n"}},
							FinalizeFunc: func(string) error {
								calls.Add(1)
								if test.fail && index < 2 {
									return fmt.Errorf("file %d failed", index)
								}
								return nil
							},
						})
					}
					return files, nil
				}),
			})
			dir := t.TempDir()
			writeGeneratedModule(t, filepath.Join(dir, codegen.Gendir), "generated.local/gen")
			diagnostics, err := os.Create(filepath.Join(dir, "debug.txt"))
			require.NoError(t, err)
			originalStderr := os.Stderr
			os.Stderr = diagnostics
			defer func() {
				os.Stderr = originalStderr
			}()
			outputs, generateErr := generate(dir, "gen", true, registry)
			os.Stderr = originalStderr
			require.NoError(t, diagnostics.Close())
			debug, err := os.ReadFile(diagnostics.Name())
			require.NoError(t, err)
			require.Contains(t, string(debug), fmt.Sprintf("parallel file writing with %d workers", min(test.budget, 6)))
			require.Equal(t, int32(5), calls.Load())
			for index := range 5 {
				content, err := os.ReadFile(filepath.Join(dir, codegen.Gendir, fmt.Sprintf("file%d.txt", index)))
				require.NoError(t, err)
				require.Equal(t, "complete\n", string(content))
			}
			if test.fail {
				require.ErrorContains(t, generateErr, "file 0 failed")
				return
			}
			require.NoError(t, generateErr)
			require.Len(t, outputs, 6)
			require.True(t, strings.HasSuffix(outputs[len(outputs)-1], "goa.json"))
		})
	}
}
