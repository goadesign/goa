package files

import (
	"bytes"
	"strings"
	"testing"

	"goa.design/goa.v2/design"
)

func TestService(t *testing.T) {
	const (
		singleMethod = `type (
	// Single is the Single service interface.
	Single interface {
		// A implements the A endpoint.
		A(context.Context, *APayload) (AResult, error)
	}

	APayload struct {
		IntField int
		StringField string
	}
)
`

		multipleMethods = `type (
	// Multiple is the Multiple service interface.
	Multiple interface {
		// A implements the A endpoint.
		A(context.Context, *APayload) (AResult, error)
		// B implements the B endpoint.
		B(context.Context, *BPayload) (BResult, error)
	}

	APayload struct {
		IntField int
		StringField string
	}

	BPayload struct {
		BooleanField bool
		BytesField []byte
	}
)
`

		nopayloadMethods = `type (
	// NoPayload is the NoPayload service interface.
	NoPayload interface {
		// NoPayload implements the NoPayload endpoint.
		NoPayload(context.Context) error
	}
)
`

		genPkg = "goa.design/goa.v2/example"
	)
	var (
		a = design.EndpointExpr{
			Name: "A",
			Payload: &design.UserTypeExpr{
				TypeName: "APayload",
				AttributeExpr: &design.AttributeExpr{Type: design.Object{
					"IntField":    &design.AttributeExpr{Type: design.Int},
					"StringField": &design.AttributeExpr{Type: design.String},
				}},
			},
			Result: design.NewUserTypeExpr("AResult", nil),
		}

		b = design.EndpointExpr{
			Name: "B",
			Payload: &design.UserTypeExpr{
				TypeName: "BPayload",
				AttributeExpr: &design.AttributeExpr{Type: design.Object{
					"BooleanField": &design.AttributeExpr{Type: design.Boolean},
					"BytesField":   &design.AttributeExpr{Type: design.Bytes},
				}},
			},
			Result: design.NewUserTypeExpr("BResult", nil),
		}

		nopayload = design.EndpointExpr{
			Name:    "NoPayload",
			Payload: design.Empty,
			Result:  design.Empty,
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
		_ = multipleEndpoints

		nopayloadEndpoint = design.ServiceExpr{
			Name: "NoPayload",
			Endpoints: []*design.EndpointExpr{
				&nopayload,
			},
		}
		_ = nopayloadEndpoint
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
		file := Service(tc.Service)
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
