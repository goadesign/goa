// This file verifies that HTTP generation keeps request- and response-shaped
// unions separate when their nested branch types have different Go names.
package generator

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"goa.design/goa/v3/codegen"
	d "goa.design/goa/v3/dsl"
)

func TestGenerateHTTPUnionUsedByRequestAndResponseCompiles(t *testing.T) {
	registry := testRegistry(
		"gen",
		testGenerator(planServiceData, Service),
		testGenerator(planTransportData, Transport),
	)

	dsl := func() {
		d.API("test", func() {})

		siteSet := d.Type("SiteSet", func() {
			d.Attribute("site_ids", d.ArrayOf(d.String))
			d.Required("site_ids")
		})
		allSites := d.Type("AllSites", func() {
			d.Attribute("include_current", d.Boolean)
			d.Required("include_current")
		})
		setup := d.Type("Setup", func() {
			d.OneOf("scope", func() {
				d.Attribute("site_set", siteSet)
				d.Attribute("all_sites", allSites)
			})
			d.Required("scope")
		})

		d.Service("front", func() {
			d.Method("configure", func() {
				d.Payload(setup)
				d.Result(setup)
				d.HTTP(func() {
					d.POST("/configure")
					d.Response(200)
				})
			})
			d.Method("reconfigure", func() {
				d.Payload(setup)
				d.Result(d.String)
				d.HTTP(func() {
					d.POST("/reconfigure")
					d.Response(200)
				})
			})
		})
	}

	_ = codegen.RunDSL(t, dsl)
	dir := t.TempDir()
	genDir := filepath.Join(dir, codegen.Gendir)
	writeGeneratedModule(t, genDir, "gen")
	if _, err := generate(dir, "gen", false, registry); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	assertGeneratedUnionDeclarations(t, genDir)
	runGeneratedTests(t, genDir)
}

// assertGeneratedUnionDeclarations proves identical request derivations reuse
// one union while the differently shaped response receives another.
func assertGeneratedUnionDeclarations(t *testing.T, genDir string) {
	t.Helper()
	path := filepath.Join(genDir, "http", "front", "server", "types.go")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read generated server types: %v", err)
	}
	code := string(content)
	if strings.Count(code, "type Scope struct {") != 1 {
		t.Fatalf("expected one request Scope declaration:\n%s", code)
	}
	if strings.Count(code, "type Scope2 struct {") != 1 {
		t.Fatalf("expected one response Scope2 declaration:\n%s", code)
	}
	if strings.Contains(code, "type Scope3 struct {") {
		t.Fatalf("identical request derivation produced a third union declaration:\n%s", code)
	}
	if strings.Contains(code, "SiteSetRequestBody") {
		t.Fatalf("identical request derivation produced a second branch declaration:\n%s", code)
	}
	if strings.Count(code, "\tSiteSet  *SiteSet\n") != 1 {
		t.Fatalf("request union does not reference its canonical branch declaration:\n%s", code)
	}
}

// writeGeneratedModule creates a temporary module that resolves this Goa
// checkout explicitly instead of downloading a released generator runtime.
func writeGeneratedModule(t *testing.T, dir, modulePath string) {
	t.Helper()
	goaRoot := moduleDirectory(t, "goa.design/goa/v3")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatalf("create generated module directory: %v", err)
	}
	module := "module " + modulePath + "\n\ngo 1.24\n\nrequire goa.design/goa/v3 v3.0.0\n\n" +
		"replace goa.design/goa/v3 => " + filepath.ToSlash(goaRoot) + "\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(module), 0o600); err != nil {
		t.Fatalf("write generated go.mod: %v", err)
	}
}

// moduleDirectory returns the checked-out directory for module from the outer
// test environment.
func moduleDirectory(t *testing.T, module string) string {
	t.Helper()
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", module)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("resolve module %s: %v\n%s", module, err, output)
	}
	dir := strings.TrimSpace(string(output))
	if dir == "" {
		t.Fatalf("resolve module %s: empty directory", module)
	}
	return dir
}

// runGeneratedTests compiles every generated service and HTTP transport
// package in the isolated module.
func runGeneratedTests(t *testing.T, dir string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "test", "-mod=mod", "./...")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("compile generated packages: %v\n%s", err, output)
	}
}
