package files

import (
	"bytes"
	"strings"
	"testing"

	. "goa.design/goa.v2/codegen/testing"
	"goa.design/goa.v2/design"
)

func TestEndpoint(t *testing.T) {
	const (
		singleMethod = `type (
	// SingleEndpoint lists the Single service endpoints.
	SingleEndpoint struct {
		A goa.Endpoint
	}
)

// NewSingleEndpoint wraps the methods of a Single service with endpoints.
func NewSingleEndpoint(s Single) *SingleEndpoint {
	ep := new(SingleEndpoint)

	ep.A = func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*AType)
		return s.A(ctx, p)
	}

	return ep
}`

		multipleMethods = `type (
	// MultipleEndpoint lists the Multiple service endpoints.
	MultipleEndpoint struct {
		B goa.Endpoint
		C goa.Endpoint
	}
)

// NewMultipleEndpoint wraps the methods of a Multiple service with endpoints.
func NewMultipleEndpoint(s Multiple) *MultipleEndpoint {
	ep := new(MultipleEndpoint)

	ep.B = func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*BType)
		return s.B(ctx, p)
	}

	ep.C = func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*CType)
		return s.C(ctx, p)
	}

	return ep
}`

		nopayloadMethods = `type (
	// NoPayloadEndpoint lists the NoPayload service endpoints.
	NoPayloadEndpoint struct {
		NoPayload goa.Endpoint
	}
)

// NewNoPayloadEndpoint wraps the methods of a NoPayload service with endpoints.
func NewNoPayloadEndpoint(s NoPayload) *NoPayloadEndpoint {
	ep := new(NoPayloadEndpoint)

	ep.NoPayload = func(ctx context.Context, req interface{}) (interface{}, error) {
		return s.NoPayload(ctx, nil)
	}

	return ep
}`

		genPkg = "goa.design/goa.v2/example"
	)
	var (
		a = design.EndpointExpr{
			Name: "A",
			Payload: &design.AttributeExpr{
				Type: &design.UserTypeExpr{
					AttributeExpr: &design.AttributeExpr{Type: design.String},
					TypeName:      "AType",
				}},
			Result: &design.AttributeExpr{Type: design.Empty},
		}

		b = design.EndpointExpr{
			Name: "B",
			Payload: &design.AttributeExpr{
				Type: &design.UserTypeExpr{
					AttributeExpr: &design.AttributeExpr{Type: design.String},
					TypeName:      "BType",
				}},
			Result: &design.AttributeExpr{Type: design.Empty},
		}

		c = design.EndpointExpr{
			Name: "C",
			Payload: &design.AttributeExpr{
				Type: &design.UserTypeExpr{
					AttributeExpr: &design.AttributeExpr{Type: design.String},
					TypeName:      "CType",
				}},
			Result: &design.AttributeExpr{Type: design.Empty},
		}

		nopayload = design.EndpointExpr{
			Name:    "NoPayload",
			Payload: &design.AttributeExpr{Type: design.Empty},
			Result:  &design.AttributeExpr{Type: design.Empty},
		}

		singleEndpoint = design.ServiceExpr{
			Name: "Single",
			Endpoints: []*design.EndpointExpr{
				&a,
			},
		}

		multipleEndpoints = design.ServiceExpr{
			Name: "Multiple",
			Endpoints: []*design.EndpointExpr{
				&b,
				&c,
			},
		}

		nopayloadEndpoint = design.ServiceExpr{
			Name: "NoPayload",
			Endpoints: []*design.EndpointExpr{
				&nopayload,
			},
		}
	)
	a.Service = &singleEndpoint
	b.Service = &multipleEndpoints
	c.Service = &multipleEndpoints
	nopayload.Service = &nopayloadEndpoint

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
		s := Service(tc.Service) // to initialize ServiceScope
		s.Sections("")
		file := Endpoint(tc.Service)
		for _, s := range file.Sections(genPkg) {
			if err := s.Write(buf); err != nil {
				t.Fatal(err)
			}
		}
		actual := buf.String()
		if !strings.Contains(actual, tc.Expected) {
			d := Diff(t, actual, tc.Expected)
			t.Errorf("%s: got\n%v\n=============\nexpected to contain\n%v\ndiff\n%v", k, actual, tc.Expected, d)
		}
	}
}
