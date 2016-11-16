package dsl_test

import (
	"testing"

	"reflect"

	"github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/dsl"
	"github.com/goadesign/goa/eval"
)

func TestEndpoint(t *testing.T) {
	cases := map[string]struct {
		Expr   eval.Expression
		DSL    func()
		Assert map[string]func(t *testing.T, s *design.EndpointExpr)
	}{
		"basic": {
			&design.ServiceExpr{},
			func() {
				Endpoint("basic", func() {
					Description("Optional description")
					// Docs allows linking to external documentation.
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
					// Docs allows linking to external documentation.
					Docs(func() {
						Description("Optional description")
						URL("https://goa.design")
					})
					Request(design.String)
					Response(design.String)
					Error("basic_media_error", design.ErrorMedia)
				})
			},
			map[string]func(t *testing.T, s *design.EndpointExpr){
				"basic": func(t *testing.T, e *design.EndpointExpr) {
					assertEndpointDescription(t, "Optional description", e.Description)
					assertEndpointDocs(t, e.Docs, "https://goa.design", "Optional description")
					if len(e.Errors) != 2 {
						t.Errorf("expected %d error definitions but got %d ", 2, len(e.Errors))
					}
					assertEndpointError(t, e.Errors[0], "basic_error", design.ErrorMedia)
					assertEndpointError(t, e.Errors[1], "basic_media_error", design.ErrorMedia)
					expectedMeta := design.MetadataExpr{
						"name": []string{"some value", "some other value"},
					}
					assertEndpointMetaData(t, e.Metadata, expectedMeta)
					expectedReq := &design.UserTypeExpr{
						TypeName:      "BasicRequest",
						AttributeExpr: &design.AttributeExpr{Description: "Optional description", Type: &design.Object{}}}
					assertEndpointRequestResponse(t, "Request", e.Request, expectedReq)
					expectedRes := &design.UserTypeExpr{
						TypeName:      "BasicResponse",
						AttributeExpr: &design.AttributeExpr{Description: "Optional description", Type: &design.Object{}}}
					assertEndpointRequestResponse(t, "Response", e.Response, expectedRes)
				},
				"another": func(t *testing.T, e *design.EndpointExpr) {
					assertEndpointDocs(t, e.Docs, "https://goa.design", "Optional description")
					if len(e.Errors) != 1 {
						t.Errorf("expected %d error definitions but got %d ", 1, len(e.Errors))
					}
					assertEndpointError(t, e.Errors[0], "basic_media_error", design.ErrorMedia)
					expectedReq := &design.UserTypeExpr{TypeName: "AnotherRequest", AttributeExpr: &design.AttributeExpr{Type: design.String}}
					assertEndpointRequestResponse(t, "Request", e.Request, expectedReq)
					expectedRes := &design.UserTypeExpr{TypeName: "AnotherResponse", AttributeExpr: &design.AttributeExpr{Type: design.String}}
					assertEndpointRequestResponse(t, "Response", e.Response, expectedRes)
				},
			},
		},
	}
	//Run our tests
	for k, tc := range cases {
		t.Run(k, func(t *testing.T) {
			eval.Context = &eval.DSLContext{}
			eval.Execute(tc.DSL, tc.Expr)
			evalService := tc.Expr.(*design.ServiceExpr)
			if eval.Context.Errors != nil {
				t.Errorf("%s: Endpoint failed unexpectedly with %s", k, eval.Context.Errors)
			}
			for _, e := range evalService.Endpoints {
				if _, ok := tc.Assert[e.Name]; !ok {
					t.Errorf("no assert found for endpoint %s ", e.Name)
					break
				}
				tc.Assert[e.Name](t, e)
			}
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

func assertEndpointDescription(t *testing.T, expectedDesc, actualDesc string) {
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

func assertEndpointRequestResponse(t *testing.T, assertType string, actual design.DataType, expected *design.UserTypeExpr) {
	ut, ok := actual.(*design.UserTypeExpr)
	if !ok {
		t.Errorf("expected endpoint %s to be a *UserTypeExpr but got %v", assertType, reflect.TypeOf(ut))
	}
	if ut.Name() != expected.Name() {
		t.Errorf("expected endpoint %s name %s to match %s", assertType, ut.Name(), expected.Name())
	}
	if ut.AttributeExpr.Type.Name() != expected.Type.Name() {
		t.Errorf("expected endpoint %s TypeName %s to match %s ", assertType, ut.Type.Name(), expected.Type.Name())
	}
	if ut.Description != expected.Description {
		t.Errorf("expected endpoint %s description %s to match %s", assertType, ut.Description, expected.Description)
	}

}
