package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
)

// TestGenerateMergesSamePathFiles verifies that when two generators emit content
// targeting the same output path, Generate merges the sections into a single
// file rather than overwriting earlier content. This is a regression test for
// an issue where only a later section (e.g., a union value method) remained and
// the earlier struct definition was lost.
func TestGenerateMergesSamePathFiles(t *testing.T) {
	t.Cleanup(func() { Generators = generators })

	// Fake generators emit two files with identical Path, one containing a
	// type definition and the other containing a method. Without merging, the
	// second write would overwrite the first.
	Generators = func(cmd string) ([]Genfunc, error) {
		return []Genfunc{
			func(genpkg string, roots []eval.Root) ([]*codegen.File, error) {
				f := &codegen.File{Path: filepath.Join(codegen.Gendir, "types", "merge_test.go")}
				f.SectionTemplates = []*codegen.SectionTemplate{
					codegen.Header("User types", "types", nil),
					{ // struct definition
						Name:   "struct-type",
						Source: "type MergeTest struct{}\n",
					},
				}
				return []*codegen.File{f}, nil
			},
			func(genpkg string, roots []eval.Root) ([]*codegen.File, error) {
				f := &codegen.File{Path: filepath.Join(codegen.Gendir, "types", "merge_test.go")}
				f.SectionTemplates = []*codegen.SectionTemplate{
					codegen.Header("User types", "types", nil),
					{ // method on MergeTest
						Name:   "method",
						Source: "func (*MergeTest) Marker() {}\n",
					},
				}
				return []*codegen.File{f}, nil
			},
		}, nil
	}

	dir := t.TempDir()
	_, err := Generate(dir, "gen")
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
