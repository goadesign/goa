package dsl_test

import (
	"testing"

	"goa.design/goa.v2/design"
	. "goa.design/goa.v2/design/dsl"
	"goa.design/goa.v2/eval"
)

func TestEndpoint(t *testing.T) {
	cases := map[string]struct {
		DSL    func()
		Assert func(t *testing.T, s []*design.EndpointExpr)
	}{
		"basic_usage": {
			func() {
				Endpoint("basic", func() {
					Description("Optional description")
					Docs(func() {
						Description("Optional description")
						URL("https://goa.design")
					})
					Request(func() {
						Description("Optional description")
						Attribute("required", design.String)
						Required("required")
					})
					Response(func() {
						Description("Optional description")
						Attribute("required", design.String)
						Required("required")
					})
					Error("basic_error")
					Error("basic_media_error", design.ErrorMedia)
					Metadata("name", "some value", "some other value")
				})
				Endpoint("another", func() {
					Docs(func() {
						Description("Optional description")
						URL("https://goa.design")
					})
					Request(func() {
						Attribute("name", design.String)
					})
					Response(func() {
						Attribute("name", design.String)
					})
					Error("basic_media_error", design.ErrorMedia)
				})
			},
			func(t *testing.T, endpoints []*design.EndpointExpr) {
				if len(endpoints) != 2 {
					t.Fatalf("expected 2 endpoints but got %d ", len(endpoints))
				}
				//assert on first endpoint
				endpoint := endpoints[0]
				assertEndpointDocs(t, endpoint.Docs, "https://goa.design", "Optional description")
				if len(endpoint.Errors) != 2 {
					t.Errorf("expected %d error definitions but got %d ", 2, len(endpoint.Errors))
				}
				assertDescription(t, "Optional description", endpoint.Description)
				assertEndpointError(t, endpoint.Errors[0], "basic_error", design.ErrorMedia)
				assertEndpointError(t, endpoint.Errors[1], "basic_media_error", design.ErrorMedia)
				expectedMeta := design.MetadataExpr{
					"name": []string{"some value", "some other value"},
				}
				assertEndpointMetaData(t, endpoint.Metadata, expectedMeta)
				expectedReq := &design.UserTypeExpr{
					TypeName:      "BasicRequest",
					AttributeExpr: &design.AttributeExpr{Description: "Optional description", Type: &design.Object{}}}
				assertUserType(t, "Request", endpoint.Request, expectedReq)
				expectedRes := &design.UserTypeExpr{
					TypeName:      "BasicResponse",
					AttributeExpr: &design.AttributeExpr{Description: "Optional description", Type: &design.Object{}}}
				assertUserType(t, "Response", endpoint.Response, expectedRes)

				//assert on second endpoint
				endpoint = endpoints[1]
				if len(endpoint.Errors) != 1 {
					t.Errorf("expected %d error definitions but got %d ", 1, len(endpoint.Errors))
				}
				if endpoint.Description != "" {
					t.Errorf("no endpoint Description was defined expected an empty Description but got %s", endpoint.Description)
				}
				if len(endpoint.Metadata) != 0 {
					t.Errorf("no endpoint Metadata defined expected an empty Metadata but got %v ", endpoint.Metadata)
				}
				assertEndpointError(t, endpoint.Errors[0], "basic_media_error", design.ErrorMedia)
				assertPrimitive(t, "Request", endpoint.Request, design.String)
				assertPrimitive(t, "Response", endpoint.Response, design.String)
			},
		},
	}
	//Run our tests
	for k, tc := range cases {
		t.Run(k, func(t *testing.T) {
			eval.Context = &eval.DSLContext{}
			serviceExpr := &design.ServiceExpr{}
			eval.Execute(tc.DSL, serviceExpr)
			if eval.Context.Errors != nil {
				t.Errorf("%s: Endpoint failed unexpectedly with %s", k, eval.Context.Errors)
			}
			tc.Assert(t, serviceExpr.Endpoints)
		})
	}
}

//helper funcs
func assertEndpointDocs(t *testing.T, doc *design.DocsExpr, url, desc string) {
	if doc.Description != desc {
		t.Errorf("expected docs description '%s' to match '%s' ", desc, doc.Description)
	}
	if doc.URL != url {
		t.Errorf("expected docs url '%s' to match '%s' ", url, doc.URL)
	}
}

func assertDescription(t *testing.T, expectedDesc, actualDesc string) {
	if expectedDesc != actualDesc {
		t.Errorf("expected description '%s' to match '%s' ", actualDesc, expectedDesc)
	}
}

func assertEndpointError(t *testing.T, actual *design.ErrorExpr, name string, dt design.DataType) {
	if actual.Name != name {
		t.Errorf("expected error to have name %s but got %s ", name, actual.Name)
	}

	if actual.AttributeExpr.Type != dt {
		t.Errorf("expected the error DataType to be %v but got %v ", dt, actual.AttributeExpr.Type)
	}
}

func assertEndpointMetaData(t *testing.T, actual design.MetadataExpr, expected design.MetadataExpr) {
	for key, val := range actual {
		vals, ok := expected[key]
		if !ok {
			t.Errorf("metaData was missing expected key %s ", key)
			continue
		}
		for _, metaVal := range val {
			if !hasValue(vals, metaVal) {
				t.Errorf("metaData was missing expected value %s ", metaVal)
			}
		}
	}
}

func assertPrimitive(t *testing.T, context string, actual design.DataType, expected design.Primitive) {
	v, ok := actual.(design.Primitive)
	if !ok {
		t.Errorf("expected %s to be a Primitive but got %s ", context, actual.Name())
		return
	}
	if v != expected {
		t.Errorf("expected %s to be a %s but got %s ", context, expected.Name(), actual.Name())
	}
}

func assertUserType(t *testing.T, context string, actual design.DataType, expected *design.UserTypeExpr) {

	ut, ok := actual.(*design.UserTypeExpr)
	if !ok {
		t.Errorf("expected %s to be a user type but got %s", context, actual.Name())
		return
	}
	if ut == nil {
		t.Errorf("did not expect user type %s to be nil ", context)
		return
	}
	if ut.Name() != expected.Name() {
		t.Errorf("expected user type %s name %s to match %s", context, ut.Name(), expected.Name())
	}
	if ut.AttributeExpr.Type.Name() != expected.Type.Name() {
		t.Errorf("expected user type %s TypeName %s to match %s ", context, ut.Type.Name(), expected.Type.Name())
	}
	if ut.Description != expected.Description {
		t.Errorf("expected user type %s description %s to match %s", context, ut.Description, expected.Description)
	}

}
