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

	"goa.design/goa/design"
	httpdesign "goa.design/goa/http/design"
)

var update = flag.Bool("update", false, "update .golden files")

func newDesign(httpSvcs ...*httpdesign.ServiceExpr) *httpdesign.RootExpr {
	openapi.Definitions = make(map[string]*openapi.Schema)
	a := &design.APIExpr{Name: "test"}
	a.Servers = []*design.ServerExpr{a.DefaultServer()}
	a.Servers[0].Hosts[0].URIs = []design.URIExpr{design.URIExpr("https://goa.design")}
	services := make([]*design.ServiceExpr, len(httpSvcs))
	for i, r := range httpSvcs {
		services[i] = r.ServiceExpr
	}
	d := &design.RootExpr{API: a, Services: services}
	return &httpdesign.RootExpr{
		Design:       d,
		HTTPServices: httpSvcs,
	}
}

func newService(endpoints ...*httpdesign.EndpointExpr) *httpdesign.ServiceExpr {
	s := &design.ServiceExpr{
		Name: "testService",
	}
	res := &httpdesign.ServiceExpr{
		ServiceExpr:   s,
		Paths:         []string{"/"},
		HTTPEndpoints: endpoints,
	}
	for _, ep := range endpoints {
		ep.MethodExpr.Service = s
		ep.Service = res
		ep.Prepare()
		ep.Finalize()
		s.Methods = append(s.Methods, ep.MethodExpr)
	}
	return res
}

func newSimpleEndpoint() *httpdesign.EndpointExpr {
	method := &design.MethodExpr{
		Name: "testEndpoint",
		Payload: &design.AttributeExpr{
			Type: &design.UserTypeExpr{
				AttributeExpr: &design.AttributeExpr{Type: design.String},
			}},
		Result: &design.AttributeExpr{
			Type: &design.UserTypeExpr{
				AttributeExpr: &design.AttributeExpr{Type: design.String},
			}},
	}
	route := &httpdesign.RouteExpr{Method: "GET", Path: "/"}
	ep := &httpdesign.EndpointExpr{
		MethodExpr: method,
		Routes:     []*httpdesign.RouteExpr{route},
		Headers:    design.NewEmptyMappedAttributeExpr(),
		Params:     design.NewEmptyMappedAttributeExpr(),
	}
	route.Endpoint = ep
	return ep
}

func TestOpenAPI(t *testing.T) {
	const (
		invalidURL = "http://[::1]:namedport"
	)
	var (
		simple  = newDesign(newService(newSimpleEndpoint()))
		empty   = newDesign()
		invalid = newDesign()
	)
	invalid.Design.API.Servers[0].Hosts[0].URIs[0] = invalidURL
	cases := map[string]struct {
		Root  *httpdesign.RootExpr
		Error bool
	}{
		"empty":   {Root: empty, Error: false},
		"valid":   {Root: simple, Error: false},
		"invalid": {Root: invalid, Error: true},
	}
	for k, c := range cases {
		_, err := OpenAPIFiles(c.Root)
		if err != nil && !c.Error {
			t.Errorf("%s: unexpected error %s", k, err)
		}
		if err == nil && c.Error {
			t.Errorf("%s: expected error", k)
		}
	}
}

func TestOutputPath(t *testing.T) {
	var (
		simple = newDesign(newService(newSimpleEndpoint()))
	)
	o, err := OpenAPIFiles(simple)
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
	const (
		genPkg = "goa.design/goa"
	)
	var (
		empty  = newDesign()
		simple = newDesign(newService(newSimpleEndpoint()))
	)
	cases := []struct {
		Name string
		Root *httpdesign.RootExpr
	}{
		{"empty", empty},
		{"valid", simple},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			oFiles, err := OpenAPIFiles(c.Root)
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

func TestValidations(t *testing.T) {
	var (
		goldenPath = filepath.Join("testdata", "openapi_v2", t.Name())
		newInt     = func(v int) *int { return &v }
		newFloat64 = func(v float64) *float64 { return &v }
	)
	cases := []struct {
		Name     string
		Endpoint *httpdesign.EndpointExpr
	}{
		{"string", newEndpointSimpleValidation(design.String,
			&design.ValidationExpr{
				MinLength: newInt(0),
				MaxLength: newInt(42),
			}),
		},
		{"integer", newEndpointSimpleValidation(design.Int,
			&design.ValidationExpr{
				Minimum: newFloat64(0),
				Maximum: newFloat64(42),
			}),
		},
		{"array", newEndpointComplexValidation(
			&design.ValidationExpr{
				MinLength: newInt(0),
				MaxLength: newInt(42),
			}),
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := newDesign(newService(c.Endpoint))
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

func newEndpointSimpleValidation(typ design.Primitive, validation *design.ValidationExpr) *httpdesign.EndpointExpr {
	route := &httpdesign.RouteExpr{Method: "POST", Path: "/"}
	ep := &httpdesign.EndpointExpr{
		MethodExpr: &design.MethodExpr{
			Name:    "testEndpoint",
			Payload: &design.AttributeExpr{},
			Result: &design.AttributeExpr{
				Type:         typ,
				UserExamples: []*design.ExampleExpr{{}},
				Validation:   validation,
			},
		},
		Body: &design.AttributeExpr{
			Type:         typ,
			UserExamples: []*design.ExampleExpr{{}},
			Validation:   validation,
		},
		Routes:    []*httpdesign.RouteExpr{route},
		Responses: []*httpdesign.HTTPResponseExpr{},
	}
	route.Endpoint = ep
	return ep
}

func newEndpointComplexValidation(validation *design.ValidationExpr) *httpdesign.EndpointExpr {
	route := &httpdesign.RouteExpr{Method: "POST", Path: "/"}
	ep := &httpdesign.EndpointExpr{
		MethodExpr: &design.MethodExpr{
			Name:    "testEndpoint",
			Payload: &design.AttributeExpr{},
			Result: &design.AttributeExpr{
				Type:         design.String,
				UserExamples: []*design.ExampleExpr{{}},
				Validation:   validation,
			},
		},
		Body: &design.AttributeExpr{
			Type: &design.Array{
				ElemType: &design.AttributeExpr{
					Type: &design.Object{
						{
							Name: "foo",
							Attribute: &design.AttributeExpr{
								Type: &design.Array{
									ElemType: &design.AttributeExpr{
										Type:         design.String,
										Validation:   validation,
										UserExamples: []*design.ExampleExpr{{}},
									},
								},
								UserExamples: []*design.ExampleExpr{{}},
								Validation:   validation,
							},
						},
						{
							Name: "bar",
							Attribute: &design.AttributeExpr{
								Type: &design.Array{
									ElemType: &design.AttributeExpr{
										Type: &design.UserTypeExpr{
											TypeName: "bar",
											AttributeExpr: &design.AttributeExpr{
												Type:         design.String,
												UserExamples: []*design.ExampleExpr{{}},
												Validation:   validation,
											},
										},
										UserExamples: []*design.ExampleExpr{{}},
									},
								},
								UserExamples: []*design.ExampleExpr{{}},
								Validation:   validation,
							},
						},
					},
					UserExamples: []*design.ExampleExpr{{}},
				},
			},
			Validation: validation,
		},
		Routes:    []*httpdesign.RouteExpr{route},
		Responses: []*httpdesign.HTTPResponseExpr{},
	}
	route.Endpoint = ep
	return ep
}

func TestExtensions(t *testing.T) {
	var (
		goldenPath = filepath.Join("testdata", "openapi_v2", t.Name())
	)
	cases := []struct {
		Name     string
		Endpoint *httpdesign.EndpointExpr
	}{
		{"endpoint", newEndpointExtensions(design.String,
			design.MetadataExpr{
				"swagger:extension:x-test-foo": []string{"bar"},
			},
		)},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := newDesign(newService(c.Endpoint))
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

func newEndpointExtensions(typ design.Primitive, metadata design.MetadataExpr) *httpdesign.EndpointExpr {
	route := &httpdesign.RouteExpr{Method: "POST", Path: "/"}
	ep := &httpdesign.EndpointExpr{
		MethodExpr: &design.MethodExpr{
			Name:    "testEndpoint",
			Payload: &design.AttributeExpr{Type: design.Empty},
			Result: &design.AttributeExpr{
				Type: design.Empty,
			},
		},
		Routes:    []*httpdesign.RouteExpr{route},
		Responses: []*httpdesign.HTTPResponseExpr{},
		Metadata:  metadata,
	}
	route.Endpoint = ep
	return ep
}
