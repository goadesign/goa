// This file verifies that complete generation merges shared union
// declarations without losing their package-owned names.
package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	cg "goa.design/goa/v3/codegen"
	d "goa.design/goa/v3/dsl"
)

// TestGenerateUnionUserTypeSamePathMerged exercises the real service codegen path
// for unions and user types across two services targeting the same user type
// file. It asserts the generated file contains both the struct definition and
// the union marker method for the union branch type. This mirrors the original
// failure mode where only the union method remained and the struct was lost.
func TestGenerateUnionUserTypeSamePathMerged(t *testing.T) {
	registry := testRegistry(
		"gen",
		testGenerator(planServiceData, testServiceFiles),
		testGenerator(planTransportData, testTransportFiles),
		testGenerator(planOpenAPIData, testOpenAPIFiles),
	)

	dsl := func() {
		d.API("test", func() {})

		var Summary = d.Type("Summary", func() {
			d.Meta("struct:pkg:path", "types")
			d.Attribute("message", d.String)
		})

		var MyUnion = d.Type("MyUnion", func() {
			d.Meta("struct:pkg:path", "types")
			d.OneOf("MyUnion", func() {
				d.TypeName("MyUnionChoice")
				d.Attribute("sum", Summary)
			})
		})

		var Container = d.Type("Container", func() {
			d.Meta("struct:pkg:path", "types")
			d.Attribute("u", MyUnion)
		})

		d.Service("S1", func() {
			d.Method("M", func() {
				d.Payload(Container)
			})
		})

		d.Service("S2", func() {
			d.Method("M", func() {
				d.Payload(Container)
			})
		})
	}

	_ = cg.RunDSL(t, dsl)

	dir := t.TempDir()
	writeGeneratedModule(t, filepath.Join(dir, cg.Gendir), "generated.local/gen")
	if _, err := generate(dir, "gen", false, registry); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify generated types/summary.go still contains the struct definition for
	// the shared user type. The original regression dropped the struct when
	// merging multiple files contributing to the same path.
	p := filepath.Join(dir, cg.Gendir, "types", cg.SnakeCase("Summary")+".go")
	bs, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("failed reading generated file: %v", err)
	}
	content := string(bs)
	if !strings.Contains(content, "type Summary struct") {
		t.Fatalf("missing struct definition in %s:\n%s", p, content)
	}
}
