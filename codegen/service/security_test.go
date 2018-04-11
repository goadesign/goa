package service

import (
	"testing"

	"goa.design/goa/codegen"
	"goa.design/goa/codegen/service/testdata"
	"goa.design/goa/design"
)

func TestSecureEndpointInit(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"endpoint-without-requirement", testdata.EndpointWithoutRequirementDSL, testdata.EndpointInitWithoutRequirementCode},
		{"endpoints-with-requirements", testdata.EndpointsWithRequirementsDSL, testdata.EndpointInitWithRequirementsCode},
		{"endpoints-with-service-requirements", testdata.EndpointsWithServiceRequirementsDSL, testdata.EndpointInitWithServiceRequirementsCode},
		{"endpoints-no-security", testdata.EndpointNoSecurityDSL, testdata.EndpointInitNoSecurityCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			codegen.RunDSL(t, c.DSL)
			if len(design.Root.Services) != 1 {
				t.Fatalf("got %d services, expected 1", len(design.Root.Services))
			}
			fs := EndpointFile(design.Root.Services[0])
			if fs == nil {
				t.Fatalf("got nil file, expected not nil")
			}
			sections := fs.SectionTemplates
			if len(sections) < 2 {
				t.Fatalf("got %d sections, expected at least 2", len(sections))
			}
			code := codegen.SectionCode(t, sections[2])
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}

func TestSecureEndpoint(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"with-required-scopes", testdata.EndpointWithRequiredScopesDSL, testdata.EndpointWithRequiredScopesCode},
		{"with-api-key-override", testdata.EndpointWithAPIKeyOverrideDSL, testdata.EndpointWithAPIKeyOverrideCode},
		{"with-oauth2", testdata.EndpointWithOAuth2DSL, testdata.EndpointWithOAuth2Code},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			codegen.RunDSL(t, c.DSL)
			if len(design.Root.Services) != 1 {
				t.Fatalf("got %d services, expected 1", len(design.Root.Services))
			}
			fs := EndpointFile(design.Root.Services[0])
			if fs == nil {
				t.Fatalf("got nil file, expected not nil")
			}
			sections := fs.SectionTemplates
			code := codegen.SectionCode(t, sections[4])
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}
