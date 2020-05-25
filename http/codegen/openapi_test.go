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
	"goa.design/goa/v3/codegen"
	openapi "goa.design/goa/v3/http/codegen/openapi"
	"goa.design/goa/v3/http/codegen/testdata"
)

var update = flag.Bool("update", false, "update .golden files")

func TestOpenAPI(t *testing.T) {
	cases := map[string]struct {
		DSL     func()
		NilSpec bool
	}{
		"empty": {DSL: testdata.EmptyDSL, NilSpec: true},
		"valid": {DSL: testdata.SimpleDSL, NilSpec: false},
	}
	for k, c := range cases {
		// Reset global variables
		openapi.Definitions = make(map[string]*openapi.Schema)
		root := RunHTTPDSL(t, c.DSL)
		spec, err := OpenAPIFiles(root)
		if err != nil {
			t.Fatalf("OpenAPI failed with %s", err)
		}
		if spec == nil && !c.NilSpec {
			t.Errorf("%s: unexpected specs: got nil, expected non-nil", k)
		}
		if spec != nil && c.NilSpec {
			t.Errorf("%s: unexpected specs: got non-nil, expected nil", k)
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
	c := 4 // number of files we expect
	if len(o) != c {
		t.Fatalf("unexpected number of OpenAPI files %d instead of %d", len(o), c)
	}
	if o[0].Path != filepath.Join("gen", "http", "openapi.json") {
		t.Errorf("invalid output path %#v", o[0].Path)
	}
	if o[1].Path != filepath.Join("gen", "http", "openapi.yaml") {
		t.Errorf("invalid output path %#v", o[1].Path)
	}
	if o[2].Path != filepath.Join("gen", "http", "openapi_v3.json") {
		t.Errorf("invalid output path %#v", o[2].Path)
	}
	if o[3].Path != filepath.Join("gen", "http", "openapi_v3.yaml") {
		t.Errorf("invalid output path %#v", o[3].Path)
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
			root := RunHTTPDSL(t, c.DSL)
			oFiles, err := OpenAPIFiles(root)
			if err != nil {
				t.Fatalf("OpenAPI failed with %s", err)
			}
			for i, o := range v2Files(oFiles) {
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
			for i, o := range v2Files(oFiles) {
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
			for i, o := range v2Files(oFiles) {
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

func v2Files(files []*codegen.File) []*codegen.File {
	var v2 []*codegen.File
	for _, f := range files {
		if filepath.Base(f.Path) == "openapi.go" {
			v2 = append(v2, f)
		}
	}
	return v2
}
