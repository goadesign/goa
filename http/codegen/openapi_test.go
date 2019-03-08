package codegen

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
	"goa.design/goa/http/codegen/openapi"
	"goa.design/goa/http/codegen/testdata"
)

var update = flag.Bool("update", false, "update .golden files")

func TestOpenAPI(t *testing.T) {
	cases := map[string]struct {
		DSL   func()
		Error bool
	}{
		"empty":   {DSL: testdata.EmptyDSL, Error: false},
		"valid":   {DSL: testdata.SimpleDSL, Error: false},
		"invalid": {DSL: testdata.InvalidDSL, Error: true},
	}
	for k, c := range cases {
		// Reset global variables
		openapi.Definitions = make(map[string]*openapi.Schema)
		root := RunHTTPDSL(t, c.DSL)
		_, err := OpenAPIFiles(root)
		if err != nil && !c.Error {
			t.Errorf("%s: unexpected error %s", k, err)
		}
		if err == nil && c.Error {
			t.Errorf("%s: expected error", k)
		}
	}
}

func TestOutputPath(t *testing.T) {
	// Reset global variables
	openapi.Definitions = make(map[string]*openapi.Schema)
	root := RunHTTPDSL(t, testdata.SimpleDSL)
	o, err := OpenAPIFiles(root)
	if err != nil {
		t.Fatalf("OpenAPI failed with %s", err)
	}
	c := 2 // number of files we expect
	if len(o) != c {
		t.Fatalf("unexpected number of OpenAPI files %d instead of %d", len(o), c)
	}
	if o[0].Path != filepath.Join("gen", "http", "openapi.json") {
		t.Errorf("invalid output path %#v", o[0].Path)
	}
	if o[1].Path != filepath.Join("gen", "http", "openapi.yaml") {
		t.Errorf("invalid output path %#v", o[1].Path)
	}
}

func TestSections(t *testing.T) {
	var (
		goldenPath = filepath.Join("testdata", "openapi_v2", t.Name())
	)
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"empty", testdata.EmptyDSL},
		{"valid", testdata.SimpleDSL},
		{"multiple-services", testdata.MultipleServicesDSL},
		{"multiple-views", testdata.MultipleViewsDSL},
		{"explicit-view", testdata.ExplicitViewDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			// Reset global variables
			openapi.Definitions = make(map[string]*openapi.Schema)
			root := RunHTTPDSL(t, c.DSL)
			oFiles, err := OpenAPIFiles(root)
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
					err = tmpl.Execute(&buf, s[0].Data)
					if err != nil {
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

func TestValidations(t *testing.T) {
	var (
		goldenPath = filepath.Join("testdata", "openapi_v2", t.Name())
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
			root := RunHTTPDSL(t, c.DSL)
			oFiles, err := OpenAPIFiles(root)
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
					err = tmpl.Execute(&buf, s[0].Data)
					if err != nil {
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
		goldenPath = filepath.Join("testdata", "openapi_v2", t.Name())
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
			root := RunHTTPDSL(t, c.DSL)
			oFiles, err := OpenAPIFiles(root)
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
					err = tmpl.Execute(&buf, s[0].Data)
					if err != nil {
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
