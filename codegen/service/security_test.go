package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service/testdata"
	"goa.design/goa/v3/expr"
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
			require.Len(t, expr.Root.Services, 1)
			fs := EndpointFile("", expr.Root.Services[0])
			require.NotNil(t, fs)
			sections := fs.SectionTemplates
			require.Greater(t, len(sections), 1)
			code := codegen.SectionCode(t, sections[2])
			assert.Equal(t, c.Code, code)
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
		{"with-optional-required-scopes", testdata.EndpointWithOptionalRequiredScopesDSL, testdata.EndpointWithOptionalRequiredScopesCode},
		{"with-api-key-override", testdata.EndpointWithAPIKeyOverrideDSL, testdata.EndpointWithAPIKeyOverrideCode},
		{"with-oauth2", testdata.EndpointWithOAuth2DSL, testdata.EndpointWithOAuth2Code},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			codegen.RunDSL(t, c.DSL)
			require.Len(t, expr.Root.Services, 1)
			fs := EndpointFile("", expr.Root.Services[0])
			require.NotNil(t, fs)
			sections := fs.SectionTemplates
			code := codegen.SectionCode(t, sections[4])
			assert.Equal(t, c.Code, code)
		})
	}
}

func TestSecureWithSkipRequestBodyEncodeDecode(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"with-basicauth", testdata.EndpointWithBasicAuthAndSkipRequestBodyEncodeDecodeDSL, testdata.EndpointWithBasicAuthAndSkipRequestBodyEncodeDecodeCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			codegen.RunDSL(t, c.DSL)
			require.Len(t, expr.Root.Services, 1)
			fs := EndpointFile("", expr.Root.Services[0])
			require.NotNil(t, fs)
			sections := fs.SectionTemplates
			code := codegen.SectionCode(t, sections[5])
			assert.Equal(t, c.Code, code)
		})
	}
}
