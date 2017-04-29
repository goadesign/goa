package files

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"

	"goa.design/goa.v2/codegen"
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
func NewSingle(s service.Single) *Single {
	ep := new(Single)

	ep.A = func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*service.AType)
		return s.A(ctx, p)
	}

	return ep
}`

		multipleMethods = `type (
	// Multiple lists the Multiple service endpoints.
	Multiple struct {
		B goa.Endpoint
		C goa.Endpoint
	}
)

// NewMultiple wraps the methods of a Multiple service with endpoints.
func NewMultiple(s service.Multiple) *Multiple {
	ep := new(Multiple)

	ep.B = func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*service.BType)
		return s.B(ctx, p)
	}

	ep.C = func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*service.CType)
		return s.C(ctx, p)
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
func NewNoPayload(s service.NoPayload) *NoPayload {
	ep := new(NoPayload)

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
		}

		b = design.EndpointExpr{
			Name: "B",
			Payload: &design.AttributeExpr{
				Type: &design.UserTypeExpr{
					AttributeExpr: &design.AttributeExpr{Type: design.String},
					TypeName:      "BType",
				}},
		}

		c = design.EndpointExpr{
			Name: "C",
			Payload: &design.AttributeExpr{
				Type: &design.UserTypeExpr{
					AttributeExpr: &design.AttributeExpr{Type: design.String},
					TypeName:      "CType",
				}},
		}

		nopayload = design.EndpointExpr{
			Name:    "NoPayload",
			Payload: &design.AttributeExpr{Type: design.Empty},
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
		ServiceScope = codegen.NewNameScope()
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
			dmp := diffmatchpatch.New()
			diffs := dmp.DiffMain(actual, tc.Expected, false)
			diff := dmp.DiffPrettyText(diffs)
			t.Errorf("%s: got\n%v\n=============\nexpected to contain\n%v\ndiff\n%v", k, actual, tc.Expected, diff)
		}
	}
}
