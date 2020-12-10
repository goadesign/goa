package openapiv3

import (
	"fmt"
	"testing"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
	"goa.design/goa/v3/http/codegen/openapi/v3/testdata/dsls"
)

func TestBuildInfo(t *testing.T) {
	const (
		title        = "test title"
		description  = "test description"
		terms        = "test terms of service"
		version      = "test version"
		contactName  = "test contact name"
		contactEmail = "test contact email"
		contactURL   = "test contact URL"
		licenseName  = "test license name"
		licenseURL   = "test license URL"
	)
	cases := []struct {
		Name           string
		Title          string
		Description    string
		TermsOfService string
		Version        string
		ContactName    string
		ContactEmail   string
		ContactURL     string
		LicenseName    string
		LicenseURL     string
	}{{
		Name:           "simple",
		Title:          title,
		Description:    description,
		TermsOfService: terms,
		Version:        version,
		ContactName:    contactName,
		ContactEmail:   contactEmail,
		ContactURL:     contactURL,
		LicenseName:    licenseName,
		LicenseURL:     licenseURL,
	}, {
		Name:  "empty version",
		Title: title,
	}, {
		Name:    "empty title",
		Version: version,
	}}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			api := &expr.APIExpr{
				Name:           c.Name,
				Title:          c.Title,
				Description:    c.Description,
				TermsOfService: c.TermsOfService,
				Version:        c.Version,
				Contact:        &expr.ContactExpr{Name: contactName, Email: contactEmail, URL: contactURL},
				License:        &expr.LicenseExpr{Name: licenseName, URL: licenseURL},
			}

			info := buildInfo(api)

			expected := c.Title
			if api.Title == "" {
				expected = "Goa API"
			}
			if info.Title != expected {
				t.Errorf("got API title %q, expected %q", info.Title, expected)
			}

			if info.Description != c.Description {
				t.Errorf("got API description %q, expected %q", info.Description, c.Description)
			}

			if info.TermsOfService != c.TermsOfService {
				t.Errorf("got API terms of service %q, expected %q", info.TermsOfService, c.TermsOfService)
			}

			expectedVer := c.Version
			if api.Version == "" {
				expectedVer = "1.0"
			}
			if info.Version != expectedVer {
				t.Errorf("got API version %q, expected %q", info.Version, expectedVer)
			}
		})
	}
}

type param struct {
	Name        string
	In          string
	Description string
	Style       string
	Required    bool
	Type        typ
}

type requestBody struct {
	Description string
	Type        typ
	Required    bool
}

type response struct {
	Description string
	Type        typ
	Headers     map[string]param
}

type responses map[string]response

func TestBuildOperation(t *testing.T) {
	const svcName = "test service"
	cases := []struct {
		Name string
		DSL  func()

		ExpectedDescription string
		ExpectedParameters  []param
		ExpectedRequestBody *requestBody
		ExpectedResponses   map[string]response
	}{{
		Name: "desc_only",
		DSL:  dsls.DescOnly(svcName, "desc_only", "desc"),

		ExpectedDescription: "desc",
		ExpectedResponses:   responses{"204": {Description: "No Content response."}},
	}, {
		Name: "request_string_body",
		DSL:  dsls.RequestStringBody(svcName, "request_string_body"),

		ExpectedRequestBody: &requestBody{"body", tstring, true},
		ExpectedResponses:   responses{"204": {Description: "No Content response."}},
	}, {
		Name: "request_object_body",
		DSL:  dsls.RequestObjectBody(svcName, "request_object_body"),

		ExpectedRequestBody: &requestBody{"", tobj("name", tstring), true},
		ExpectedResponses:   responses{"204": {Description: "No Content response."}},
	}, {
		Name: "request_streaming_string_body",
		DSL:  dsls.RequestObjectBody(svcName, "request_streaming_string_body"),

		ExpectedRequestBody: &requestBody{"", tobj("name", tstring), true},
		ExpectedResponses:   responses{"204": {Description: "No Content response."}},
	}, {
		Name: "response_array_of_string",
		DSL:  dsls.ResponseArrayOfString(svcName, "response_array_of_string"),

		ExpectedResponses: responses{"200": {"OK response.", tobj("result", tobj("children", tarray)), nil}},
	}, {
		Name: "response_recursive_user_type",
		DSL:  dsls.ResponseRecursiveUserType(svcName, "response_recursive_user_type"),

		ExpectedResponses: responses{"200": {"OK response.", tobj("recursive", tobj()), nil}},
	}, {
		Name: "response_recursive_array_user_type",
		DSL:  dsls.ResponseRecursiveArrayUserType(svcName, "response_recursive_array_user_type"),

		ExpectedResponses: responses{"200": {"OK response.", tobj("result", tobj("children", tarray)), nil}},
	}}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			api := codegen.RunDSL(t, c.DSL).API

			var bodies *EndpointBodies
			var types map[string]*openapi.Schema
			{
				var bds map[string]map[string]*EndpointBodies
				bds, types = buildBodyTypes(api)
				if svc, ok := bds[svcName]; ok {
					bodies, ok = svc[c.Name]
					if !ok {
						t.Error("bodies does not contain method details")
						return
					}
				}
			}

			var route *expr.RouteExpr
			{
				if len(api.HTTP.Services) == 0 {
					t.Error("no HTTP service created from DSL")
				}
				for _, s := range api.HTTP.Services {
					if s.Name() == svcName {
						for _, e := range s.HTTPEndpoints {
							if e.Name() == c.Name {
								route = e.Routes[0]
								break
							}
						}
					}
					if route != nil {
						break
					}
				}
				if route == nil {
					t.Error("could not find route")
					return
				}
			}

			op := buildOperation(c.Name, route, bodies, expr.NewRandom(c.Name))

			if op.Description != c.ExpectedDescription {
				t.Errorf("got description %q for method %q, expected %q", op.Description, c.Name, c.ExpectedDescription)
			}
			if len(op.Parameters) != len(c.ExpectedParameters) {
				t.Errorf("got %d parameters, expected %d", len(op.Parameters), len(c.ExpectedParameters))
				return
			}
			for i, p := range op.Parameters {
				matchesParameter(t, p, types, c.ExpectedParameters[i])
			}
			matchesRequestBody(t, op.RequestBody, types, c.ExpectedRequestBody)
			if len(op.Responses) != len(c.ExpectedResponses) {
				t.Errorf("got %d responses, expected %d", len(op.Responses), len(c.ExpectedResponses))
				return
			}
			for s, r := range op.Responses {
				matchesResponse(t, r, types, c.ExpectedResponses[s])
			}
		})
	}
}

func matchesParameter(t *testing.T, p *ParameterRef, types map[string]*openapi.Schema, expected param) {
	matchesParameterHeader(t, p, types, expected, "parameter")
}
func matchesParameterHeader(t *testing.T, p *ParameterRef, types map[string]*openapi.Schema, expected param, title string) {
	if p.Value == nil {
		t.Errorf("no value for %s", title)
		return
	}
	if p.Ref != "" {
		t.Errorf("got ref %q for %s %q, expected none", p.Ref, title, p.Value.Name)
	}
	v := p.Value
	if v.Name != expected.Name {
		t.Errorf("got %s name %q, expected %q", title, v.Name, expected.Name)
	}
	if v.In != expected.In {
		t.Errorf("got %s in %q, expected %q", title, v.In, expected.In)
	}
	if v.Description != expected.Description {
		t.Errorf("got %s description %q, expected %q", title, v.Description, expected.Description)
	}
	if v.Style != expected.Style {
		t.Errorf("got %s style %q, expected %q", title, v.Style, expected.Style)
	}
	if v.Required != expected.Required {
		t.Errorf("got %s required %v, expected %v", title, v.Required, expected.Required)
	}
	matchesSchema(t, fmt.Sprintf("%s %q", title, v.Name), v.Schema, types, expected.Type)
	if v.Content != nil {
		t.Errorf("got content %#v, expected none", v.Content)
	}
}

func matchesRequestBody(t *testing.T, b *RequestBodyRef, types map[string]*openapi.Schema, expected *requestBody) {
	if b == nil {
		if expected != nil {
			t.Error("request body is nil")
		}
		return
	}
	if b.Value == nil {
		t.Error("no value for request body")
		return
	}
	if b.Ref != "" {
		t.Errorf("got ref %q for request body, expected none", b.Ref)
	}
	v := b.Value
	if v.Description != expected.Description {
		t.Errorf("got request body description %q, expected %q", v.Description, expected.Description)
	}
	if v.Required != expected.Required {
		t.Errorf("got request body required %v, expected %v", v.Required, expected.Required)
	}
	ct, ok := v.Content["application/json"]
	if !ok {
		t.Error("missing request content, expected application/json")
		return
	}
	matchesSchema(t, "request body", ct.Schema, types, expected.Type)
}

func matchesResponse(t *testing.T, r *ResponseRef, types map[string]*openapi.Schema, expected response) {
	if r.Value == nil {
		t.Error("no value for response")
		return
	}
	if r.Ref != "" {
		t.Errorf("got ref %q for response, expected none", r.Ref)
	}
	v := r.Value
	if v.Description == nil && expected.Description != "" {
		t.Errorf("got no response description, expected %q", expected.Description)
	} else if *v.Description != expected.Description {
		t.Errorf("got response description %q, expected %q", *v.Description, expected.Description)
	}
	if len(v.Headers) != len(expected.Headers) {
		t.Errorf("got %d response header(s), expected %d", len(v.Headers), len(expected.Headers))
		return
	}
	for n, h := range v.Headers {
		exp, ok := expected.Headers[n]
		if !ok {
			t.Errorf("response header %q not expected", n)
		}
		matchesHeader(t, h, types, exp)
	}
	if expected.Type.Type != "" {
		ct, ok := v.Content["application/json"]
		if !ok {
			t.Error("missing response content, expected application/json")
			return
		}
		matchesSchema(t, "response body", ct.Schema, types, expected.Type)
	}
}

func matchesHeader(t *testing.T, h *HeaderRef, types map[string]*openapi.Schema, expected param) {
	if h.Value == nil {
		t.Error("no value for header")
		return
	}
	if h.Ref != "" {
		t.Errorf("got ref %q for header, expected none", h.Ref)
	}
	v := h.Value
	par := &ParameterRef{Value: &Parameter{
		Description:     v.Description,
		Style:           v.Style,
		Explode:         v.Explode,
		AllowEmptyValue: v.AllowEmptyValue,
		AllowReserved:   v.AllowReserved,
		Deprecated:      v.Deprecated,
		Required:        v.Required,
		Schema:          v.Schema,
		Example:         v.Example,
		Examples:        v.Examples,
		Content:         v.Content,
		Extensions:      v.Extensions,
		In:              "header",
	}}
	matchesParameterHeader(t, par, types, expected, "header")
}
