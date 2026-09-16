// This file verifies that generation accepts only importable package identities
// returned by Go's package loader and never invents an import path.
package generator

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

func TestGeneratedPackageImportPath(t *testing.T) {
	t.Run("module package", func(t *testing.T) {
		t.Setenv("GO111MODULE", "on")
		t.Setenv("GOWORK", "off")
		t.Setenv("GOENV", "off")
		t.Setenv("GOPATH", t.TempDir())
		t.Setenv("GOPACKAGESDRIVER", "off")
		moduleDir := t.TempDir()
		writePackageFixture(t, moduleDir, "module generated.local\n\ngo 1.25\n")

		got, err := generatedPackageImportPath(filepath.Join(moduleDir, "gen"))
		require.NoError(t, err)
		require.Equal(t, "generated.local/gen", got)
	})

	t.Run("authored underscore module", func(t *testing.T) {
		t.Setenv("GO111MODULE", "on")
		t.Setenv("GOWORK", "off")
		t.Setenv("GOENV", "off")
		t.Setenv("GOPATH", t.TempDir())
		t.Setenv("GOPACKAGESDRIVER", "off")
		moduleDir := t.TempDir()
		writePackageFixture(t, moduleDir, "module _/authored\n\ngo 1.25\n")

		got, err := generatedPackageImportPath(filepath.Join(moduleDir, "gen"))
		require.NoError(t, err)
		require.Equal(t, "_/authored/gen", got)
	})

	t.Run("workspace module", func(t *testing.T) {
		t.Setenv("GO111MODULE", "on")
		t.Setenv("GOENV", "off")
		t.Setenv("GOPATH", t.TempDir())
		t.Setenv("GOPACKAGESDRIVER", "off")
		workspaceDir := t.TempDir()
		otherDir := filepath.Join(workspaceDir, "other")
		targetDir := filepath.Join(workspaceDir, "target")
		require.NoError(t, os.MkdirAll(otherDir, 0o750))
		require.NoError(t, os.MkdirAll(targetDir, 0o750))
		writePackageFixture(t, otherDir, "module workspace.local/other\n\ngo 1.25\n")
		writePackageFixture(t, targetDir, "module _/target\n\ngo 1.25\n")
		workFile := filepath.Join(workspaceDir, "go.work")
		require.NoError(t, os.WriteFile(workFile, []byte("go 1.25\n\nuse (\n\t./other\n\t./target\n)\n"), 0o600))
		t.Setenv("GOWORK", workFile)

		got, err := generatedPackageImportPath(filepath.Join(targetDir, "gen"))
		require.NoError(t, err)
		require.Equal(t, "_/target/gen", got)
	})

	t.Run("GOPATH package", func(t *testing.T) {
		gopath := t.TempDir()
		packageDir := filepath.Join(gopath, "src", "_", "foo")
		require.NoError(t, os.MkdirAll(packageDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(packageDir, "generated.go"), []byte("package foo\n"), 0o600))
		t.Setenv("GO111MODULE", "off")
		t.Setenv("GOWORK", "off")
		t.Setenv("GOENV", "off")
		t.Setenv("GOPATH", gopath)
		t.Setenv("GOPACKAGESDRIVER", "off")

		pkgs, err := packages.Load(&packages.Config{
			Mode: packages.NeedName | packages.NeedModule | packages.NeedFiles,
			Dir:  packageDir,
		}, ".")
		require.NoError(t, err)
		require.Len(t, pkgs, 1)
		require.Empty(t, pkgs[0].Errors)
		require.Nil(t, pkgs[0].Module)
		require.Equal(t, packageDir, pkgs[0].Dir)
		require.Equal(t, "_/foo", pkgs[0].PkgPath)

		got, err := generatedPackageImportPath(packageDir)
		require.NoError(t, err)
		require.Equal(t, "_/foo", got)
	})

	t.Run("GOPATH package reached through symlink", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Windows symlinks require privileges not available on every test host")
		}
		gopath := t.TempDir()
		realPackageDir := filepath.Join(gopath, "src", "_", "linked")
		require.NoError(t, os.MkdirAll(realPackageDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(realPackageDir, "generated.go"), []byte("package linked\n"), 0o600))
		linkedGOPATH := filepath.Join(t.TempDir(), "linked-gopath")
		require.NoError(t, os.Symlink(gopath, linkedGOPATH))
		linkedPackageDir := filepath.Join(linkedGOPATH, "src", "_", "linked")
		t.Setenv("GO111MODULE", "off")
		t.Setenv("GOWORK", "off")
		t.Setenv("GOENV", "off")
		t.Setenv("GOPATH", gopath)
		t.Setenv("GOPACKAGESDRIVER", "off")

		pkgs, err := packages.Load(&packages.Config{
			Mode: packages.NeedName | packages.NeedModule | packages.NeedFiles,
			Dir:  linkedPackageDir,
		}, ".")
		require.NoError(t, err)
		require.Len(t, pkgs, 1)
		require.Empty(t, pkgs[0].Errors)
		require.Nil(t, pkgs[0].Module)
		require.Equal(t, realPackageDir, pkgs[0].Dir)
		require.Equal(t, "_/linked", pkgs[0].PkgPath)

		owned, err := gopathOwnsImportPath(linkedPackageDir, "_/linked")
		require.NoError(t, err)
		require.True(t, owned)

		got, err := generatedPackageImportPath(linkedPackageDir)
		require.NoError(t, err)
		require.Equal(t, "_/linked", got)
	})

	t.Run("missing GOPATH package", func(t *testing.T) {
		gopath := filepath.Join(t.TempDir(), "missing")
		importPath := "_/missing"
		lexicalDir := filepath.Join(gopath, "src", filepath.FromSlash(importPath))
		t.Setenv("GO111MODULE", "off")
		t.Setenv("GOWORK", "off")
		t.Setenv("GOENV", "off")
		t.Setenv("GOPATH", gopath)
		t.Setenv("GOPACKAGESDRIVER", "off")

		owned, err := gopathOwnsImportPath(lexicalDir, importPath)
		require.NoError(t, err)
		require.True(t, owned)

		owned, err = gopathOwnsImportPath(t.TempDir(), importPath)
		require.NoError(t, err)
		require.False(t, owned)
	})

	t.Run("GOPATH symlink resolution error", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Windows symlinks require privileges not available on every test host")
		}
		gopath := t.TempDir()
		importPath := "_/loop"
		packagePath := filepath.Join(gopath, "src", filepath.FromSlash(importPath))
		require.NoError(t, os.MkdirAll(filepath.Dir(packagePath), 0o750))
		require.NoError(t, os.Symlink(packagePath, packagePath))
		t.Setenv("GO111MODULE", "off")
		t.Setenv("GOWORK", "off")
		t.Setenv("GOENV", "off")
		t.Setenv("GOPATH", gopath)
		t.Setenv("GOPACKAGESDRIVER", "off")

		owned, err := gopathOwnsImportPath(t.TempDir(), importPath)
		require.Error(t, err)
		require.False(t, owned)
		require.ErrorContains(t, err, "resolve GOPATH package path")
	})

	t.Run("non-first GOPATH package", func(t *testing.T) {
		first := t.TempDir()
		second := t.TempDir()
		packageDir := filepath.Join(second, "src", "_", "second")
		require.NoError(t, os.MkdirAll(packageDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(packageDir, "generated.go"), []byte("package second\n"), 0o600))
		t.Setenv("GO111MODULE", "off")
		t.Setenv("GOWORK", "off")
		t.Setenv("GOENV", "off")
		t.Setenv("GOPATH", strings.Join([]string{first, second}, string(os.PathListSeparator)))
		t.Setenv("GOPACKAGESDRIVER", "off")

		got, err := generatedPackageImportPath(packageDir)
		require.NoError(t, err)
		require.Equal(t, "_/second", got)
	})

	t.Run("GOPATH ending in space", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Windows paths cannot portably end in a space")
		}
		gopath := filepath.Join(t.TempDir(), "gopath ")
		packageDir := filepath.Join(gopath, "src", "_", "space")
		require.NoError(t, os.MkdirAll(packageDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(packageDir, "generated.go"), []byte("package space\n"), 0o600))
		t.Setenv("GO111MODULE", "off")
		t.Setenv("GOWORK", "off")
		t.Setenv("GOENV", "off")
		t.Setenv("GOPATH", gopath)
		t.Setenv("GOPACKAGESDRIVER", "off")

		got, err := generatedPackageImportPath(packageDir)
		require.NoError(t, err)
		require.Equal(t, "_/space", got)
	})

	t.Run("GOENV GOPATH package", func(t *testing.T) {
		gopath := t.TempDir()
		packageDir := filepath.Join(gopath, "src", "_", "goenv")
		require.NoError(t, os.MkdirAll(packageDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(packageDir, "generated.go"), []byte("package goenv\n"), 0o600))
		goenvFile := filepath.Join(t.TempDir(), "go.env")
		require.NoError(t, os.WriteFile(goenvFile, []byte("GOPATH="+gopath+"\n"), 0o600))
		t.Setenv("GO111MODULE", "off")
		t.Setenv("GOWORK", "off")
		t.Setenv("GOENV", goenvFile)
		t.Setenv("GOPACKAGESDRIVER", "off")
		unsetTestEnv(t, "GOPATH")

		got, err := generatedPackageImportPath(packageDir)
		require.NoError(t, err)
		require.Equal(t, "_/goenv", got)
	})

	t.Run("empty GOPATH uses default", func(t *testing.T) {
		home := t.TempDir()
		gopath := filepath.Join(home, "go")
		packageDir := filepath.Join(gopath, "src", "_", "default")
		require.NoError(t, os.MkdirAll(packageDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(packageDir, "generated.go"), []byte("package defaultpkg\n"), 0o600))
		t.Setenv("GO111MODULE", "off")
		t.Setenv("GOWORK", "off")
		t.Setenv("GOENV", "off")
		t.Setenv("GOPATH", "")
		t.Setenv("HOME", home)
		t.Setenv("USERPROFILE", home)
		t.Setenv("GOPACKAGESDRIVER", "off")

		got, err := generatedPackageImportPath(packageDir)
		require.NoError(t, err)
		require.Equal(t, "_/default", got)
	})

	t.Run("synthetic GOPATH package", func(t *testing.T) {
		gopath := filepath.Join(t.TempDir(), "gopath")
		require.NoError(t, os.MkdirAll(gopath, 0o750))
		packageDir := filepath.Join(t.TempDir(), "outside")
		require.NoError(t, os.MkdirAll(packageDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(packageDir, "generated.go"), []byte("package gen\n"), 0o600))
		t.Setenv("GO111MODULE", "off")
		t.Setenv("GOWORK", "off")
		t.Setenv("GOENV", "off")
		t.Setenv("GOPATH", gopath)
		t.Setenv("GOPACKAGESDRIVER", "off")

		pkgs, err := packages.Load(&packages.Config{
			Mode: packages.NeedName | packages.NeedModule | packages.NeedFiles,
			Dir:  packageDir,
		}, ".")
		require.NoError(t, err)
		require.Len(t, pkgs, 1)
		require.Empty(t, pkgs[0].Errors)
		require.Nil(t, pkgs[0].Module)
		require.Equal(t, packageDir, pkgs[0].Dir)
		require.True(t, strings.HasPrefix(pkgs[0].PkgPath, "_/"), pkgs[0].PkgPath)

		got, err := generatedPackageImportPath(packageDir)
		require.Error(t, err)
		require.Empty(t, got)
		require.ErrorContains(t, err, "synthetic import path")
	})

	t.Run("synthetic package with missing GOPATH roots", func(t *testing.T) {
		first := filepath.Join(t.TempDir(), "missing-first")
		second := filepath.Join(t.TempDir(), "missing-second")
		packageDir := filepath.Join(t.TempDir(), "outside")
		require.NoError(t, os.MkdirAll(packageDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(packageDir, "generated.go"), []byte("package gen\n"), 0o600))
		t.Setenv("GO111MODULE", "off")
		t.Setenv("GOWORK", "off")
		t.Setenv("GOENV", "off")
		t.Setenv("GOPATH", strings.Join([]string{first, second}, string(os.PathListSeparator)))
		t.Setenv("GOPACKAGESDRIVER", "off")

		got, err := generatedPackageImportPath(packageDir)
		require.Error(t, err)
		require.Empty(t, got)
		require.ErrorContains(t, err, "synthetic import path")
	})

	t.Run("invalid package", func(t *testing.T) {
		t.Setenv("GO111MODULE", "on")
		t.Setenv("GOWORK", "off")
		t.Setenv("GOENV", "off")
		t.Setenv("GOPATH", t.TempDir())
		t.Setenv("GOPACKAGESDRIVER", "off")
		moduleDir := t.TempDir()
		writePackageFixture(t, moduleDir, "module generated.local\n\ngo 1.25\n")
		genDir := filepath.Join(moduleDir, "gen")
		require.NoError(t, os.WriteFile(filepath.Join(genDir, "other.go"), []byte("package other\n"), 0o600))

		got, err := generatedPackageImportPath(genDir)
		require.Error(t, err)
		require.Empty(t, got)
		var packageError packages.Error
		require.ErrorAs(t, err, &packageError)
	})

	t.Run("invalid module", func(t *testing.T) {
		t.Setenv("GO111MODULE", "on")
		t.Setenv("GOWORK", "off")
		t.Setenv("GOENV", "off")
		t.Setenv("GOPATH", t.TempDir())
		t.Setenv("GOPACKAGESDRIVER", "off")
		moduleDir := t.TempDir()
		writePackageFixture(t, moduleDir, "module invalid path\n\ngo 1.25\n")

		got, err := generatedPackageImportPath(filepath.Join(moduleDir, "gen"))
		require.Error(t, err)
		require.Empty(t, got)
		require.ErrorContains(t, err, "errors parsing")
	})
}

// unsetTestEnv removes key for one subtest and restores its exact process
// state after the loader has observed the missing variable.
func unsetTestEnv(t *testing.T, key string) {
	t.Helper()
	value, exists := os.LookupEnv(key)
	require.NoError(t, os.Unsetenv(key))
	t.Cleanup(func() {
		if exists {
			require.NoError(t, os.Setenv(key, value))
			return
		}
		require.NoError(t, os.Unsetenv(key))
	})
}

// writePackageFixture creates the module and generated package consumed by the
// real package loader in each test case.
func writePackageFixture(t *testing.T, moduleDir, moduleFile string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(moduleDir, "go.mod"), []byte(moduleFile), 0o600))
	genDir := filepath.Join(moduleDir, "gen")
	require.NoError(t, os.MkdirAll(genDir, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(genDir, "generated.go"), []byte("package gen\n"), 0o600))
}
