package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service/testdata"
	"goa.design/goa/v3/codegen/testutil"
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
			root := codegen.RunDSL(t, c.DSL)
			plan := mustServicePlan(t, root)
			require.Len(t, root.Services, 1)
			fs := endpointFile(plan, plan.facts.services[0])
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
	}{
		{"with-required-scopes", testdata.EndpointWithRequiredScopesDSL},
		{"with-optional-required-scopes", testdata.EndpointWithOptionalRequiredScopesDSL},
		{"with-api-key-override", testdata.EndpointWithAPIKeyOverrideDSL},
		{"with-bearer", testdata.EndpointWithBearerDSL},
		{"with-oauth2", testdata.EndpointWithOAuth2DSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := codegen.RunDSL(t, c.DSL)
			plan := mustServicePlan(t, root)
			require.Len(t, root.Services, 1)
			fs := endpointFile(plan, plan.facts.services[0])
			require.NotNil(t, fs)
			sections := fs.SectionTemplates
			code := codegen.SectionCode(t, sections[4])
			testutil.AssertGo(t, "testdata/golden/security_endpoint_"+c.Name+".go.golden", code)
		})
	}
}

func TestSecureWithSkipRequestBodyEncodeDecode(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"with-basicauth", testdata.EndpointWithBasicAuthAndSkipRequestBodyEncodeDecodeDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := codegen.RunDSL(t, c.DSL)
			plan := mustServicePlan(t, root)
			require.Len(t, root.Services, 1)
			fs := endpointFile(plan, plan.facts.services[0])
			require.NotNil(t, fs)
			sections := fs.SectionTemplates
			code := codegen.SectionCode(t, sections[5])
			testutil.AssertGo(t, "testdata/golden/security_endpoint_"+c.Name+".go.golden", code)
		})
	}
}
