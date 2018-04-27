package service

import (
	"bytes"
	"strings"
	"testing"

	"goa.design/goa/codegen"
	"goa.design/goa/design"
)

func TestEndpoint(t *testing.T) {
	const (
		singleMethod = `
// Endpoints wraps the "Single" service endpoints.
type Endpoints struct {
	A goa.Endpoint
}
// NewEndpoints wraps the methods of the "Single" service with endpoints.
func NewEndpoints(s Service) *Endpoints {
	return &Endpoints{
		A: NewAEndpoint(s),
	}
}
// Use applies the given middleware to all the "Single" service endpoints.
func (e *Endpoints) Use(m func(goa.Endpoint) goa.Endpoint) {
	e.A = m(e.A)
}
// NewAEndpoint returns an endpoint function that calls the method "A" of
// service "Single".
func NewAEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*AType)
		return nil, s.A(ctx, p)
	}
}
`

		multipleMethods = `
// Endpoints wraps the "Multiple" service endpoints.
type Endpoints struct {
	B goa.Endpoint
	C goa.Endpoint
}
// NewEndpoints wraps the methods of the "Multiple" service with endpoints.
func NewEndpoints(s Service) *Endpoints {
	return &Endpoints{
		B: NewBEndpoint(s),
		C: NewCEndpoint(s),
	}
}
// Use applies the given middleware to all the "Multiple" service endpoints.
func (e *Endpoints) Use(m func(goa.Endpoint) goa.Endpoint) {
	e.B = m(e.B)
	e.C = m(e.C)
}
// NewBEndpoint returns an endpoint function that calls the method "B" of
// service "Multiple".
func NewBEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*BType)
		return nil, s.B(ctx, p)
	}
}
// NewCEndpoint returns an endpoint function that calls the method "C" of
// service "Multiple".
func NewCEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*CType)
		return nil, s.C(ctx, p)
	}
}
`

		nopayloadMethods = `
// Endpoints wraps the "NoPayload" service endpoints.
type Endpoints struct {
	NoPayload goa.Endpoint
}
// NewEndpoints wraps the methods of the "NoPayload" service with endpoints.
func NewEndpoints(s Service) *Endpoints {
	return &Endpoints{
		NoPayload: NewNoPayloadEndpoint(s),
	}
}
// Use applies the given middleware to all the "NoPayload" service endpoints.
func (e *Endpoints) Use(m func(goa.Endpoint) goa.Endpoint) {
	e.NoPayload = m(e.NoPayload)
}
// NewNoPayloadEndpoint returns an endpoint function that calls the method
// "NoPayload" of service "NoPayload".
func NewNoPayloadEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, s.NoPayload(ctx)
	}
}
`
	)
	cases := map[string]struct {
		Service  *design.ServiceExpr
		Expected string
	}{
		"single":    {Service: &singleEndpoint, Expected: singleMethod},
		"multiple":  {Service: &multipleEndpoints, Expected: multipleMethods},
		"nopayload": {Service: &nopayloadEndpoint, Expected: nopayloadMethods},
	}
	for k, tc := range cases {
		buf := new(bytes.Buffer)
		Services = make(ServicesData)
		design.Root.Services = []*design.ServiceExpr{tc.Service}
		design.Root.API = &design.APIExpr{Name: "test"}
		File("", tc.Service) // to initialize ServiceScope
		ef := EndpointFile("", tc.Service)
		for _, s := range ef.SectionTemplates {
			if err := s.Write(buf); err != nil {
				t.Fatal(err)
			}
		}
		actual := buf.String()
		if !strings.Contains(actual, tc.Expected) {
			d := codegen.Diff(t, actual, tc.Expected)
			t.Errorf("%s: got\n%v\n=============\nexpected to contain\n%v\ndiff\n%v", k, actual, tc.Expected, d)
		}
	}
}

var (
	a = design.MethodExpr{
		Name: "A",
		Payload: &design.AttributeExpr{
			Type: &design.UserTypeExpr{
				AttributeExpr: &design.AttributeExpr{Type: &design.Object{{Name: "a", Attribute: &design.AttributeExpr{Type: design.String}}}},
				TypeName:      "AType",
			}},
		Result: &design.AttributeExpr{Type: design.Empty},
	}

	b = design.MethodExpr{
		Name: "B",
		Payload: &design.AttributeExpr{
			Type: &design.UserTypeExpr{
				AttributeExpr: &design.AttributeExpr{Type: &design.Object{{Name: "b", Attribute: &design.AttributeExpr{Type: design.String}}}},
				TypeName:      "BType",
			}},
		Result: &design.AttributeExpr{Type: design.Empty},
	}

	c = design.MethodExpr{
		Name: "C",
		Payload: &design.AttributeExpr{
			Type: &design.UserTypeExpr{
				AttributeExpr: &design.AttributeExpr{Type: &design.Object{{Name: "c", Attribute: &design.AttributeExpr{Type: design.String}}}},
				TypeName:      "CType",
			}},
		Result: &design.AttributeExpr{Type: design.Empty},
	}

	nopayload = design.MethodExpr{
		Name:    "NoPayload",
		Payload: &design.AttributeExpr{Type: design.Empty},
		Result:  &design.AttributeExpr{Type: design.Empty},
	}

	singleEndpoint = design.ServiceExpr{
		Name: "Single",
		Methods: []*design.MethodExpr{
			&a,
		},
	}

	multipleEndpoints = design.ServiceExpr{
		Name: "Multiple",
		Methods: []*design.MethodExpr{
			&b,
			&c,
		},
	}

	nopayloadEndpoint = design.ServiceExpr{
		Name: "NoPayload",
		Methods: []*design.MethodExpr{
			&nopayload,
		},
	}

	genPkg = "goa.design/goa/example"
)

func init() {
	a.Service = &singleEndpoint
	b.Service = &multipleEndpoints
	c.Service = &multipleEndpoints
	nopayload.Service = &nopayloadEndpoint
}
