package rest

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/go-openapi/loads"

	"goa.design/goa.v2/design"
	"goa.design/goa.v2/design/rest"
)

func newDesign(t *testing.T, resources ...*rest.ResourceExpr) *rest.RootExpr {
	a := &design.APIExpr{
		Name:    "test",
		Servers: []*design.ServerExpr{{URL: "https://goa.design"}},
	}
	services := make([]*design.ServiceExpr, len(resources))
	for i, r := range resources {
		services[i] = r.ServiceExpr
	}
	d := &design.RootExpr{API: a, Services: services}
	return &rest.RootExpr{
		Design:    d,
		Resources: resources,
	}
}

func newResource(t *testing.T) *rest.ResourceExpr {
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
	route := &rest.RouteExpr{Method: "GET", Path: "/"}
	action := &rest.ActionExpr{
		MethodExpr: ep,
		Routes:     []*rest.RouteExpr{route},
	}
	route.Action = action
	res := &rest.ResourceExpr{
		ServiceExpr: s,
		Path:        "/",
		Actions:     []*rest.ActionExpr{action},
	}
	action.Resource = res
	return res
}

func TestOpenAPI(t *testing.T) {
	const (
		invalidURL = "http://[::1]:namedport"
	)
	var (
		simple  = newDesign(t, newResource(t))
		empty   = newDesign(t)
		invalid = newDesign(t)
	)
	invalid.Design.API.Servers[0].URL = invalidURL
	cases := map[string]struct {
		Root  *rest.RootExpr
		Error bool
	}{
		"nil":     {Root: nil, Error: false},
		"empty":   {Root: empty, Error: false},
		"valid":   {Root: simple, Error: false},
		"invalid": {Root: invalid, Error: true},
	}
	for k, c := range cases {
		_, err := OpenAPI(c.Root)
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
		simple = newDesign(t, newResource(t))
	)
	o, err := OpenAPI(simple)
	if err != nil {
		t.Fatalf("OpenAPI failed with %s", err)
	}
	if o.OutputPath() != filepath.Join("openapi", "swagger.json") {
		t.Errorf("invalid output path %#v", o.OutputPath())
	}
}

func TestSections(t *testing.T) {
	const (
		genPkg = "goa.design/goa.v2"
	)
	var (
		empty  = newDesign(t)
		simple = newDesign(t, newResource(t))
	)
	cases := map[string]struct {
		Root *rest.RootExpr
	}{
		"empty": {Root: empty},
		"valid": {Root: simple},
	}
	for k, c := range cases {
		o, err := OpenAPI(c.Root)
		if err != nil {
			t.Fatalf("%s: OpenAPI failed with %s", k, err)
		}
		s := o.Sections(genPkg)
		if len(s) != 1 {
			t.Fatalf("%s: expected 1 section, got %d", k, len(s))
		}
		if s[0].Template == nil {
			t.Fatalf("%s: nil section template", k)
		}
		if s[0].Data == nil {
			t.Fatalf("%s: nil data", k)
		}
		var buf bytes.Buffer
		err = s[0].Template.Execute(&buf, s[0].Data)
		if err != nil {
			t.Fatalf("%s: failed to render template: %s", k, err)
		}
		validateSwagger(t, k, buf.Bytes())
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
