package codegen

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/go-openapi/loads"

	"goa.design/goa/design"
	httpdesign "goa.design/goa/http/design"
)

func newDesign(t *testing.T, httpSvcs ...*httpdesign.ServiceExpr) *httpdesign.RootExpr {
	a := &design.APIExpr{
		Name:    "test",
		Servers: []*design.ServerExpr{{URL: "https://goa.design"}},
	}
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

func newService(t *testing.T) *httpdesign.ServiceExpr {
	ep := &design.MethodExpr{
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
	s := &design.ServiceExpr{
		Name:    "testService",
		Methods: []*design.MethodExpr{ep},
	}
	ep.Service = s
	route := &httpdesign.RouteExpr{Method: "GET", Path: "/"}
	endpoint := &httpdesign.EndpointExpr{
		MethodExpr: ep,
		Routes:     []*httpdesign.RouteExpr{route},
		Service:    &httpdesign.ServiceExpr{ServiceExpr: s},
	}
	endpoint.Finalize()
	route.Endpoint = endpoint
	res := &httpdesign.ServiceExpr{
		ServiceExpr:   s,
		Paths:         []string{"/"},
		HTTPEndpoints: []*httpdesign.EndpointExpr{endpoint},
	}
	endpoint.Service = res
	return res
}

func TestOpenAPI(t *testing.T) {
	const (
		invalidURL = "http://[::1]:namedport"
	)
	var (
		simple  = newDesign(t, newService(t))
		empty   = newDesign(t)
		invalid = newDesign(t)
	)
	invalid.Design.API.Servers[0].URL = invalidURL
	cases := map[string]struct {
		Root  *httpdesign.RootExpr
		Error bool
	}{
		"nil":     {Root: nil, Error: false},
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
		simple = newDesign(t, newService(t))
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
		empty  = newDesign(t)
		simple = newDesign(t, newService(t))
	)
	cases := map[string]struct {
		Root *httpdesign.RootExpr
	}{
		"empty": {Root: empty},
		"valid": {Root: simple},
	}
	for k, c := range cases {
		o, err := OpenAPIFiles(c.Root)
		if err != nil {
			t.Fatalf("%s: OpenAPI failed with %s", k, err)
		}
		for i := 0; i < len(o); i++ {
			s := o[i].SectionTemplates
			if len(s) != 1 {
				t.Fatalf("%s: expected 1 section, got %d", k, len(s))
			}
			if s[0].Source == "" {
				t.Fatalf("%s: empty section template", k)
			}
			if s[0].Data == nil {
				t.Fatalf("%s: nil data", k)
			}
			var buf bytes.Buffer
			tmpl := template.Must(template.New("openapi").Funcs(s[0].FuncMap).Parse(s[0].Source))
			err = tmpl.Execute(&buf, s[0].Data)
			if err != nil {
				t.Fatalf("%s: failed to render template: %s", k, err)
			}
			validateSwagger(t, k, buf.Bytes())
		}
	}
}

// validateSwagger asserts that the given bytes contain a valid Swagger spec.
func validateSwagger(t *testing.T, title string, b []byte) {
	doc, err := loads.Analyzed(json.RawMessage(b), "")
	if err != nil {
		t.Errorf("%s: invalid swagger: %s", title, err)
	}
	if doc == nil {
		t.Errorf("%s: nil swagger", title)
	}
}
