package files

import (
	"bytes"
	"go/format"
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
		BooleanField bool
		BytesField   []byte
		IntField     int
		StringField  string
	}

	AResult struct {
		BooleanField bool
		BytesField   []byte
		IntField     int
		StringField  string
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
		BooleanField bool
		BytesField   []byte
		IntField     int
		StringField  string
	}

	BPayload struct {
		ArrayField    []bool
		MapField      map[int]string
		ObjectField   map[string]interface{}
		UserTypeField Parent
	}

	AResult struct {
		BooleanField bool
		BytesField   []byte
		IntField     int
		StringField  string
	}

	BResult struct {
		ArrayField    []bool
		MapField      map[int]string
		ObjectField   map[string]interface{}
		UserTypeField Parent
	}

	Child struct {
		p Parent
	}

	Parent struct {
		c Child
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
		child = &design.UserTypeExpr{
			TypeName: "Child",
		}

		parent = &design.UserTypeExpr{
			TypeName: "Parent",
			AttributeExpr: &design.AttributeExpr{
				Type: design.Object{
					"c": &design.AttributeExpr{
						Type: child,
					},
				},
			},
		}
	)

	child.AttributeExpr = &design.AttributeExpr{
		Type: design.Object{
			"p": &design.AttributeExpr{Type: parent},
		},
	}

	var (
		a = design.EndpointExpr{
			Name: "A",
			Payload: &design.UserTypeExpr{
				TypeName: "APayload",
				AttributeExpr: &design.AttributeExpr{Type: design.Object{
					"IntField":     &design.AttributeExpr{Type: design.Int},
					"StringField":  &design.AttributeExpr{Type: design.String},
					"BooleanField": &design.AttributeExpr{Type: design.Boolean},
					"BytesField":   &design.AttributeExpr{Type: design.Bytes},
				}},
			},
			Result: &design.UserTypeExpr{
				TypeName: "AResult",
				AttributeExpr: &design.AttributeExpr{Type: design.Object{
					"IntField":     &design.AttributeExpr{Type: design.Int},
					"StringField":  &design.AttributeExpr{Type: design.String},
					"BooleanField": &design.AttributeExpr{Type: design.Boolean},
					"BytesField":   &design.AttributeExpr{Type: design.Bytes},
				}},
			},
		}

		b = design.EndpointExpr{
			Name: "B",
			Payload: &design.UserTypeExpr{
				TypeName: "BPayload",
				AttributeExpr: &design.AttributeExpr{Type: design.Object{
					"ArrayField":    &design.AttributeExpr{Type: &design.Array{&design.AttributeExpr{Type: design.Boolean}}},
					"MapField":      &design.AttributeExpr{Type: &design.Map{KeyType: &design.AttributeExpr{Type: design.Int}, ElemType: &design.AttributeExpr{Type: design.String}}},
					"ObjectField":   &design.AttributeExpr{Type: design.Object{"IntField": &design.AttributeExpr{Type: design.Int}, "StringField": &design.AttributeExpr{Type: design.String}}},
					"UserTypeField": &design.AttributeExpr{Type: parent},
				}},
			},
			Result: &design.UserTypeExpr{
				TypeName: "BResult",
				AttributeExpr: &design.AttributeExpr{Type: design.Object{
					"ArrayField":    &design.AttributeExpr{Type: &design.Array{&design.AttributeExpr{Type: design.Boolean}}},
					"MapField":      &design.AttributeExpr{Type: &design.Map{KeyType: &design.AttributeExpr{Type: design.Int}, ElemType: &design.AttributeExpr{Type: design.String}}},
					"ObjectField":   &design.AttributeExpr{Type: design.Object{"IntField": &design.AttributeExpr{Type: design.Int}, "StringField": &design.AttributeExpr{Type: design.String}}},
					"UserTypeField": &design.AttributeExpr{Type: parent},
				}},
			},
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
		file := Service(tc.Service)
		for _, s := range file.Sections(genPkg) {
			if err := s.Write(buf); err != nil {
				t.Fatal(err)
			}
		}
		bs, err := format.Source(buf.Bytes())
		if err != nil {
			t.Fatal(err)
		}
		actual := string(bs)
		if !strings.Contains(actual, tc.Expected) {
			t.Errorf("%s: got\n%v\n=============\nexpected to contain\n%v", k, actual, tc.Expected)
		}
	}
}
