package openapiv2_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/go-openapi/loads"
	"goa.design/goa/v3/codegen"
	httpgen "goa.design/goa/v3/http/codegen"
	openapi "goa.design/goa/v3/http/codegen/openapi"
	openapiv2 "goa.design/goa/v3/http/codegen/openapi/v2"
	"goa.design/goa/v3/http/codegen/testdata"
)

var update = flag.Bool("update", false, "update .golden files")

func TestSections(t *testing.T) {
	var (
		goldenPath = filepath.Join("testdata", t.Name())
	)
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"empty", testdata.EmptyDSL},
		{"file-service", testdata.FileServiceDSL},
		{"valid", testdata.SimpleDSL},
		{"multiple-services", testdata.MultipleServicesDSL},
		{"multiple-views", testdata.MultipleViewsDSL},
		{"explicit-view", testdata.ExplicitViewDSL},
		{"security", testdata.SecurityDSL},
		{"server-host-with-variables", testdata.ServerHostWithVariablesDSL},
		{"with-spaces", testdata.WithSpacesDSL},
		{"with-map", testdata.WithMapDSL},
		{"path-with-wildcards", testdata.PathWithWildcardDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			// Reset global variables
			openapi.Definitions = make(map[string]*openapi.Schema)
			root := httpgen.RunHTTPDSL(t, c.DSL)
			oFiles, err := openapiv2.Files(root)
			if err != nil {
				t.Fatalf("OpenAPI failed with %s", err)
			}
			for i, o := range oFiles {
				tname := fmt.Sprintf("file%d", i)
				s := o.SectionTemplates
				t.Run(tname, func(t *testing.T) {
					if len(s) != 1 {
						t.Fatalf("expected 1 section, got %d", len(s))
					}
					if s[0].Source == "" {
						t.Fatalf("empty section template")
					}
					if s[0].Data == nil {
						t.Fatalf("nil data")
					}
					var buf bytes.Buffer
					tmpl := template.Must(template.New("openapi").Funcs(s[0].FuncMap).Parse(s[0].Source))
					if err := tmpl.Execute(&buf, s[0].Data); err != nil {
						t.Fatalf("failed to render template: %s", err)
					}
					if err := validateSwagger(buf.Bytes()); err != nil {
						t.Errorf("invalid swagger: %s", err)
					}

					golden := filepath.Join(goldenPath, fmt.Sprintf("%s_%s.golden", c.Name, tname))
					if *update {
						if err := ioutil.WriteFile(golden, buf.Bytes(), 0644); err != nil {
							t.Fatalf("failed to update golden file: %s", err)
						}
					}

					want, err := ioutil.ReadFile(golden)
					want = bytes.Replace(want, []byte{'\r', '\n'}, []byte{'\n'}, -1)
					if err != nil {
						t.Fatalf("failed to read golden file: %s", err)
					}
					if !bytes.Equal(buf.Bytes(), want) {
						t.Errorf("result does not match the golden file, diff:\n%s\nGot bytes:\n%x\nExpected bytes:\n%x\n", codegen.Diff(t, buf.String(), string(want)), buf.Bytes(), want)
					}
				})
			}
		})
	}
}

func TestValidations(t *testing.T) {
	var (
		goldenPath = filepath.Join("testdata", t.Name())
	)
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"string", testdata.StringValidationDSL},
		{"integer", testdata.IntValidationDSL},
		{"array", testdata.ArrayValidationDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			// Reset global variables
			openapi.Definitions = make(map[string]*openapi.Schema)
			root := httpgen.RunHTTPDSL(t, c.DSL)
			oFiles, err := openapiv2.Files(root)
			if err != nil {
				t.Fatalf("OpenAPI failed with %s", err)
			}
			if len(oFiles) == 0 {
				t.Fatalf("No swagger files")
			}
			for i, o := range oFiles {
				tname := fmt.Sprintf("file%d", i)
				s := o.SectionTemplates
				t.Run(tname, func(t *testing.T) {
					if len(s) != 1 {
						t.Fatalf("expected 1 section, got %d", len(s))
					}
					if s[0].Source == "" {
						t.Fatalf("empty section template")
					}
					if s[0].Data == nil {
						t.Fatalf("nil data")
					}
					var buf bytes.Buffer
					tmpl := template.Must(template.New("openapi").Funcs(s[0].FuncMap).Parse(s[0].Source))
					if err := tmpl.Execute(&buf, s[0].Data); err != nil {
						t.Fatalf("failed to render template: %s", err)
					}
					if err := validateSwagger(buf.Bytes()); err != nil {
						t.Fatalf("invalid swagger: %s", err)
					}

					golden := filepath.Join(goldenPath, fmt.Sprintf("%s_%s.golden", c.Name, tname))
					if *update {
						if err := ioutil.WriteFile(golden, buf.Bytes(), 0644); err != nil {
							t.Fatalf("failed to update golden file: %s", err)
						}
					}

					want, err := ioutil.ReadFile(golden)
					want = bytes.Replace(want, []byte{'\r', '\n'}, []byte{'\n'}, -1)
					if err != nil {
						t.Fatalf("failed to read golden file: %s", err)
					}
					if !bytes.Equal(buf.Bytes(), want) {
						t.Errorf("result do not match the golden file:\n--BEGIN--\n%s\n--END--\n", buf.Bytes())
					}
				})
			}
		})
	}
}

func TestExtensions(t *testing.T) {
	var (
		goldenPath = filepath.Join("testdata", t.Name())
	)
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"endpoint", testdata.ExtensionDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			// Reset global variables
			openapi.Definitions = make(map[string]*openapi.Schema)
			root := httpgen.RunHTTPDSL(t, c.DSL)
			oFiles, err := openapiv2.Files(root)
			if err != nil {
				t.Fatalf("OpenAPI failed with %s", err)
			}
			if len(oFiles) == 0 {
				t.Fatalf("No swagger files")
			}
			for i, o := range oFiles {
				tname := fmt.Sprintf("file%d", i)
				s := o.SectionTemplates
				t.Run(tname, func(t *testing.T) {
					if len(s) != 1 {
						t.Fatalf("expected 1 section, got %d", len(s))
					}
					if s[0].Source == "" {
						t.Fatalf("empty section template")
					}
					if s[0].Data == nil {
						t.Fatalf("nil data")
					}
					var buf bytes.Buffer
					tmpl := template.Must(template.New("openapi").Funcs(s[0].FuncMap).Parse(s[0].Source))
					if err := tmpl.Execute(&buf, s[0].Data); err != nil {
						t.Fatalf("failed to render template: %s", err)
					}
					if err := validateSwagger(buf.Bytes()); err != nil {
						t.Fatalf("invalid swagger: %s", err)
					}

					golden := filepath.Join(goldenPath, fmt.Sprintf("%s_%s.golden", c.Name, tname))
					if *update {
						if err := ioutil.WriteFile(golden, buf.Bytes(), 0644); err != nil {
							t.Fatalf("failed to update golden file: %s", err)
						}
					}

					want, err := ioutil.ReadFile(golden)
					want = bytes.Replace(want, []byte{'\r', '\n'}, []byte{'\n'}, -1)
					if err != nil {
						t.Fatalf("failed to read golden file: %s", err)
					}
					if !bytes.Equal(buf.Bytes(), want) {
						t.Errorf("result do not match the golden file:\n--BEGIN--\n%s\n--END--\n", buf.Bytes())
					}
				})
			}
		})
	}
}

// validateSwagger asserts that the given bytes contain a valid Swagger spec.
func validateSwagger(b []byte) error {
	doc, err := loads.Analyzed(json.RawMessage(b), "")
	if err != nil {
		return err
	}
	if doc == nil {
		return errors.New("nil swagger")
	}
	return nil
}
