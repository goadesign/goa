package files

import (
	"bytes"
	"strings"
	"testing"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/design"
)

func TestEndpoint(t *testing.T) {
	const (
		singleMethod = `type (
	// Endpoints wraps the Single service endpoints.
	Endpoints struct {
		A goa.Endpoint
	}
)

// NewEndpoints wraps the methods of a Single service with endpoints.
func NewEndpoints(s Service) *Endpoints {
	ep := new(Endpoints)

	ep.A = func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*AType)
		return nil, s.A(ctx, p)
	}

	return ep
}`

		multipleMethods = `type (
	// Endpoints wraps the Multiple service endpoints.
	Endpoints struct {
		B goa.Endpoint
		C goa.Endpoint
	}
)

// NewEndpoints wraps the methods of a Multiple service with endpoints.
func NewEndpoints(s Service) *Endpoints {
	ep := new(Endpoints)

	ep.B = func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*BType)
		return nil, s.B(ctx, p)
	}

	ep.C = func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*CType)
		return nil, s.C(ctx, p)
	}

	return ep
}`

		nopayloadMethods = `type (
	// Endpoints wraps the NoPayload service endpoints.
	Endpoints struct {
		NoPayload goa.Endpoint
	}
)

// NewEndpoints wraps the methods of a NoPayload service with endpoints.
func NewEndpoints(s Service) *Endpoints {
	ep := new(Endpoints)

	ep.NoPayload = func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, s.NoPayload(ctx, nil)
	}

	return ep
}`

		genPkg = "goa.design/goa.v2/example"
	)
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
			d := codegen.Diff(t, actual, tc.Expected)
			t.Errorf("%s: got\n%v\n=============\nexpected to contain\n%v\ndiff\n%v", k, actual, tc.Expected, d)
		}
	}
}
