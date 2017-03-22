package files

import (
	"bytes"
	"strings"
	"testing"

	"goa.design/goa.v2/design"
)

func TestEndpoint(t *testing.T) {
	const (
		singleMethod = `type (
	// Single lists the Single service endpoints.
	Single struct {
		A goa.Endpoint
	}
)

// NewSingle wraps the methods of a Single service with endpoints.
func NewSingle(s services.Single) *Single {
	ep := &Single{}

	ep.A = func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*services.AType)
		return s.A(ctx, p)
	}

	return ep
}`

		multipleMethods = `type (
	// Multiple lists the Multiple service endpoints.
	Multiple struct {
		A goa.Endpoint
		B goa.Endpoint
	}
)

// NewMultiple wraps the methods of a Multiple service with endpoints.
func NewMultiple(s services.Multiple) *Multiple {
	ep := &Multiple{}

	ep.A = func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*services.AType)
		return s.A(ctx, p)
	}

	ep.B = func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*services.BType)
		return s.B(ctx, p)
	}

	return ep
}`

		nopayloadMethods = `type (
	// NoPayload lists the NoPayload service endpoints.
	NoPayload struct {
		NoPayload goa.Endpoint
	}
)

// NewNoPayload wraps the methods of a NoPayload service with endpoints.
func NewNoPayload(s services.NoPayload) *NoPayload {
	ep := &NoPayload{}

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
			Payload: &design.UserTypeExpr{
				TypeName: "AType",
			},
		}

		b = design.EndpointExpr{
			Name: "B",
			Payload: &design.UserTypeExpr{
				TypeName: "BType",
			},
		}

		nopayload = design.EndpointExpr{
			Name:    "NoPayload",
			Payload: design.Empty,
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
				&a,
				&b,
			},
		}

		nopayloadEndpoint = design.ServiceExpr{
			Name: "NoPayload",
			Endpoints: []*design.EndpointExpr{
				&nopayload,
			},
		}
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
		file := Endpoint(tc.Service)
		for _, s := range file.Sections(genPkg) {
			if err := s.Write(buf); err != nil {
				t.Fatal(err)
			}
		}
		actual := buf.String()
		if !strings.Contains(actual, tc.Expected) {
			t.Errorf("%s: got\n%v\n=============\nexpected to contain\n%v", k, actual, tc.Expected)
		}
	}
}
