// These tests compile the CLI's temporary generator from a freshly tidied
// design module and from a prepared vendor directory. Module builds must add
// the generator's dependencies; vendor builds must use the supplied packages.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGeneratorCompileDesignModule(t *testing.T) {
	root, err := filepath.Abs("../..")
	require.NoError(t, err)
	for _, vendored := range []bool{false, true} {
		t.Run(fmt.Sprintf("vendored=%t", vendored), func(t *testing.T) {
			dir := t.TempDir()
			t.Chdir(dir)
			t.Setenv("GOWORK", "off")
			t.Setenv("GOFLAGS", "-mod=readonly")
			module := fmt.Sprintf("module example.com/generator-test\n\ngo 1.26.0\n\nrequire goa.design/goa/v3 v3.0.0\n\nreplace goa.design/goa/v3 => %q\n", filepath.ToSlash(root))
			require.NoError(t, os.WriteFile("go.mod", []byte(module), 0o600))
			require.NoError(t, os.Mkdir("design", 0o700))
			require.NoError(t, os.WriteFile(filepath.Join("design", "design.go"),
				[]byte("package design\n\nimport . \"goa.design/goa/v3/dsl\"\n\nvar _ = API(\"compile-test\", func() {})\n"), 0o600))
			runCompileTestGo(t, "mod", "tidy")

			generator := NewGenerator("gen", "example.com/generator-test/design", dir, false)
			t.Cleanup(generator.Remove)
			require.NoError(t, generator.Write(false))
			if vendored {
				// Include the generator while preparing vendor so compilation
				// receives every required package without changing the module.
				runCompileTestGo(t, "mod", "tidy")
				runCompileTestGo(t, "mod", "vendor")
				generator = NewGenerator("gen", "example.com/generator-test/design", dir, false)
				t.Cleanup(generator.Remove)
				require.True(t, generator.hasVendorDirectory)
				require.NoError(t, generator.Write(false))
				t.Setenv("GOFLAGS", "-mod=vendor")
				t.Setenv("GOPROXY", "off")
			}
			before, err := os.ReadFile("go.mod")
			require.NoError(t, err)
			require.NoError(t, generator.Compile(false))
			_, err = os.Stat(filepath.Join(generator.tmpDir, generator.bin))
			require.NoError(t, err)
			if vendored {
				after, err := os.ReadFile("go.mod")
				require.NoError(t, err)
				require.Equal(t, before, after)
			}
		})
	}
}

// runCompileTestGo prepares the temporary module with the Go command and
// reports its output if preparation fails, so the test can stop before compiling.
func runCompileTestGo(t *testing.T, args ...string) {
	t.Helper()
	output, err := exec.Command("go", args...).CombinedOutput()
	require.NoError(t, err, "%s", output)
}
