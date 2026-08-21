// This file verifies file aggregation across core generators and plugins,
// including the separately assigned same-label section regression.
package generator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
	goa "goa.design/goa/v3/pkg"
)

// TestMergeFilesPreservesSameLabelSections verifies that diagnostic section
// labels do not cause the merger to discard different generated bodies.
func TestMergeFilesPreservesSameLabelSections(t *testing.T) {
	registry := testRegistryFromGenfuncs("gen", []testGenfunc{
		testRenderOnly(func(_ string, _ []eval.Root) ([]*codegen.File, error) {
			return []*codegen.File{{
				Path: filepath.Join(codegen.Gendir, "types", "same_label.go"),
				SectionTemplates: []*codegen.SectionTemplate{
					codegen.Header("Types", "types", nil),
					{Name: "type-def", Source: "type First struct{}\n"},
				},
			}}, nil
		}),
		testRenderOnly(func(_ string, _ []eval.Root) ([]*codegen.File, error) {
			return []*codegen.File{{
				Path: filepath.Join(codegen.Gendir, "types", "same_label.go"),
				SectionTemplates: []*codegen.SectionTemplate{
					codegen.Header("Types", "types", nil),
					{Name: "type-def", Source: "type Second struct{}\n"},
				},
			}}, nil
		}),
	})

	dir := t.TempDir()
	_, err := generate(dir, "gen", false, registry)
	require.NoError(t, err)
	content, err := os.ReadFile(filepath.Join(dir, codegen.Gendir, "types", "same_label.go"))
	require.NoError(t, err)
	require.Contains(t, string(content), "type First struct{}")
	require.Contains(t, string(content), "type Second struct{}")
}

// TestGenerateMergesSamePathFiles verifies that when two generators emit content
// targeting the same output path, Generate merges the sections into a single
// file rather than overwriting earlier content. This is a regression test for
// an issue where only a later section (e.g., a union value method) remained and
// the earlier struct definition was lost.
func TestGenerateMergesSamePathFiles(t *testing.T) {

	// Fake generators emit two files with identical Path, one containing a
	// type definition and the other containing a method. Without merging, the
	// second write would overwrite the first.
	registry := testRegistryFromGenfuncs("gen", []testGenfunc{
		testRenderOnly(func(genpkg string, roots []eval.Root) ([]*codegen.File, error) {
			f := &codegen.File{Path: filepath.Join(codegen.Gendir, "types", "merge_test.go")}
			f.SectionTemplates = []*codegen.SectionTemplate{
				codegen.Header("User types", "types", nil),
				{ // struct definition
					Name:   "struct-type",
					Source: "type MergeTest struct{}\n",
				},
			}
			return []*codegen.File{f}, nil
		}),
		testRenderOnly(func(genpkg string, roots []eval.Root) ([]*codegen.File, error) {
			f := &codegen.File{Path: filepath.Join(codegen.Gendir, "types", "merge_test.go")}
			f.SectionTemplates = []*codegen.SectionTemplate{
				codegen.Header("User types", "types", nil),
				{ // method on MergeTest
					Name:   "method",
					Source: "func (*MergeTest) Marker() {}\n",
				},
			}
			return []*codegen.File{f}, nil
		}),
	})

	dir := t.TempDir()
	_, err := generate(dir, "gen", false, registry)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	// Read the merged output directly from the temp dir regardless of how
	// outputs are relativized by Generate.
	outpath := filepath.Join(dir, codegen.Gendir, "types", "merge_test.go")
	bs, err := os.ReadFile(outpath)
	if err != nil {
		t.Fatalf("failed reading merged file: %v", err)
	}
	content := string(bs)
	if !strings.Contains(content, "type MergeTest struct{}") {
		t.Fatalf("merged file missing struct definition:\n%s", content)
	}
	if !strings.Contains(content, "func (*MergeTest) Marker() {}") {
		t.Fatalf("merged file missing method definition:\n%s", content)
	}
}

// TestGenerateParallelManyFiles verifies that parallel file writing correctly
// handles multiple files (more than runtime.NumCPU()) to exercise the worker
// pool distribution. This ensures all workers process files and all files are
// written correctly.
func TestGenerateParallelManyFiles(t *testing.T) {

	// Generate 20 files to ensure we exceed typical CPU counts and exercise
	// the worker pool's work distribution.
	const numFiles = 20
	registry := testRegistryFromGenfuncs("gen", []testGenfunc{
		testRenderOnly(func(genpkg string, roots []eval.Root) ([]*codegen.File, error) {
			files := make([]*codegen.File, numFiles)
			for i := 0; i < numFiles; i++ {
				f := &codegen.File{
					Path: filepath.Join(codegen.Gendir, "types", filepath.Join("parallel", filepath.Join("file"+string(rune('a'+i%26)), "test"+string(rune('0'+i/26))+".go"))),
				}
				f.SectionTemplates = []*codegen.SectionTemplate{
					codegen.Header("Types", "types", nil),
					{
						Name:   "type-def",
						Source: "type Test" + string(rune('A'+i)) + " struct{ ID int }\n",
					},
				}
				files[i] = f
			}
			return files, nil
		}),
	})

	dir := t.TempDir()
	outputs, err := generate(dir, "gen", false, registry)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	outputs = assertVersionFile(t, dir, outputs)

	// Verify all generator files were written
	if len(outputs) != numFiles {
		t.Fatalf("expected %d output files, got %d", numFiles, len(outputs))
	}

	// Verify each file exists and has correct content
	for i := 0; i < numFiles; i++ {
		outpath := filepath.Join(dir, codegen.Gendir, "types", filepath.Join("parallel", filepath.Join("file"+string(rune('a'+i%26)), "test"+string(rune('0'+i/26))+".go")))
		bs, err := os.ReadFile(outpath)
		if err != nil {
			t.Fatalf("failed reading file %d: %v", i, err)
		}
		content := string(bs)
		expected := "type Test" + string(rune('A'+i)) + " struct"
		if !strings.Contains(content, expected) {
			t.Fatalf("file %d missing expected content %q:\n%s", i, expected, content)
		}
	}
}

// TestGenerateParallelWithMerge verifies that parallel file writing correctly
// handles file merging when multiple generators target the same path. This
// tests the interaction between mergeFilesByPath and parallel rendering.
func TestGenerateParallelWithMerge(t *testing.T) {

	// Three generators: first two merge to same path, third is separate.
	// This exercises both merging and parallel writing with NumCPU workers.
	registry := testRegistryFromGenfuncs("gen", []testGenfunc{
		testRenderOnly(func(genpkg string, roots []eval.Root) ([]*codegen.File, error) {
			f1 := &codegen.File{Path: filepath.Join(codegen.Gendir, "types", "merged.go")}
			f1.SectionTemplates = []*codegen.SectionTemplate{
				codegen.Header("Types", "types", nil),
				{Name: "type1", Source: "type Type1 struct{}\n"},
			}
			return []*codegen.File{f1}, nil
		}),
		testRenderOnly(func(genpkg string, roots []eval.Root) ([]*codegen.File, error) {
			f2 := &codegen.File{Path: filepath.Join(codegen.Gendir, "types", "merged.go")}
			f2.SectionTemplates = []*codegen.SectionTemplate{
				codegen.Header("Types", "types", nil),
				{Name: "type2", Source: "type Type2 struct{}\n"},
			}
			return []*codegen.File{f2}, nil
		}),
		testRenderOnly(func(genpkg string, roots []eval.Root) ([]*codegen.File, error) {
			f3 := &codegen.File{Path: filepath.Join(codegen.Gendir, "types", "separate.go")}
			f3.SectionTemplates = []*codegen.SectionTemplate{
				codegen.Header("Types", "types", nil),
				{Name: "type3", Source: "type Type3 struct{}\n"},
			}
			return []*codegen.File{f3}, nil
		}),
	})

	dir := t.TempDir()
	outputs, err := generate(dir, "gen", false, registry)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	outputs = assertVersionFile(t, dir, outputs)

	if len(outputs) != 2 {
		t.Fatalf("expected 2 output files, got %d", len(outputs))
	}

	// Verify merged file contains both types
	mergedPath := filepath.Join(dir, codegen.Gendir, "types", "merged.go")
	bs, err := os.ReadFile(mergedPath)
	if err != nil {
		t.Fatalf("failed reading merged file: %v", err)
	}
	content := string(bs)
	if !strings.Contains(content, "type Type1 struct{}") {
		t.Fatalf("merged file missing Type1:\n%s", content)
	}
	if !strings.Contains(content, "type Type2 struct{}") {
		t.Fatalf("merged file missing Type2:\n%s", content)
	}

	// Verify separate file
	separatePath := filepath.Join(dir, codegen.Gendir, "types", "separate.go")
	bs, err = os.ReadFile(separatePath)
	if err != nil {
		t.Fatalf("failed reading separate file: %v", err)
	}
	content = string(bs)
	if !strings.Contains(content, "type Type3 struct{}") {
		t.Fatalf("separate file missing Type3:\n%s", content)
	}
}

// TestGenerateParallelErrorHandling verifies that when file rendering fails
// in the parallel worker pool, the first error is captured and returned while
// other workers continue processing.
func TestGenerateParallelErrorHandling(t *testing.T) {

	// Create multiple files where some will fail to render due to invalid paths.
	// Worker pool should capture first error but continue processing other files.
	registry := testRegistryFromGenfuncs("gen", []testGenfunc{
		testRenderOnly(func(genpkg string, roots []eval.Root) ([]*codegen.File, error) {
			files := make([]*codegen.File, 5)
			for i := 0; i < 5; i++ {
				f := &codegen.File{
					Path: filepath.Join(codegen.Gendir, "types", "file"+string(rune('0'+i))+".go"),
				}
				f.SectionTemplates = []*codegen.SectionTemplate{
					codegen.Header("Types", "types", nil),
					{Name: "type", Source: "type T" + string(rune('0'+i)) + " struct{}\n"},
				}
				// Make file 2 fail by adding an invalid path character after writing starts
				if i == 2 {
					// Use a FinalizeFunc that returns an error
					f.FinalizeFunc = func(fp string) error {
						return os.ErrInvalid
					}
				}
				files[i] = f
			}
			return files, nil
		}),
	})

	dir := t.TempDir()
	_, err := generate(dir, "gen", false, registry)
	if err == nil {
		t.Fatal("expected error from parallel generation, got nil")
	}
	// Verify we got an error (the first one encountered)
	if !strings.Contains(err.Error(), "invalid") && !strings.Contains(err.Error(), "finalize") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

// TestGenerateParallelSingleFile verifies that parallel file writing works
// correctly with just a single file (minimal parallelism edge case).
func TestGenerateParallelSingleFile(t *testing.T) {

	registry := testRegistryFromGenfuncs("gen", []testGenfunc{
		testRenderOnly(func(genpkg string, roots []eval.Root) ([]*codegen.File, error) {
			f := &codegen.File{Path: filepath.Join(codegen.Gendir, "types", "single.go")}
			f.SectionTemplates = []*codegen.SectionTemplate{
				codegen.Header("Types", "types", nil),
				{Name: "type", Source: "type Single struct{}\n"},
			}
			return []*codegen.File{f}, nil
		}),
	})

	dir := t.TempDir()
	outputs, err := generate(dir, "gen", false, registry)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	outputs = assertVersionFile(t, dir, outputs)

	if len(outputs) != 1 {
		t.Fatalf("expected 1 output file, got %d", len(outputs))
	}

	outpath := filepath.Join(dir, codegen.Gendir, "types", "single.go")
	bs, err := os.ReadFile(outpath)
	if err != nil {
		t.Fatalf("failed reading file: %v", err)
	}
	content := string(bs)
	if !strings.Contains(content, "type Single struct{}") {
		t.Fatalf("file missing expected content:\n%s", content)
	}
}

// assertVersionFile checks that goa.json was emitted with the correct version
// and returns the remaining outputs (excluding goa.json) for further assertions.
func assertVersionFile(t *testing.T, dir string, outputs []string) []string {
	t.Helper()

	versionPath := filepath.Join(codegen.Gendir, "goa.json")

	// Read and validate goa.json content.
	bs, err := os.ReadFile(filepath.Join(dir, versionPath))
	if err != nil {
		t.Fatalf("failed reading goa.json: %v", err)
	}
	var data map[string]string
	if err := json.Unmarshal(bs, &data); err != nil {
		t.Fatalf("goa.json is not valid JSON: %v", err)
	}
	if v := data["goa_version"]; v != goa.Version() {
		t.Fatalf("goa.json version = %q, want %q", v, goa.Version())
	}

	// Filter goa.json out of outputs.
	var rest []string
	for _, o := range outputs {
		if filepath.Base(o) != "goa.json" {
			rest = append(rest, o)
		}
	}
	return rest
}
