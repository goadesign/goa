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

		emptyMethods = `type (
	// Empty is the Empty service interface.
	Empty interface {
		// Empty implements the Empty endpoint.
		Empty(context.Context) error
	}
)
`

		emptyResultMethods = `type (
	// EmptyResult is the EmptyResult service interface.
	EmptyResult interface {
		// EmptyResult implements the EmptyResult endpoint.
		EmptyResult(context.Context, *APayload) error
	}

	APayload struct {
		BooleanField bool
		BytesField   []byte
		IntField     int
		StringField  string
	}
)
`

		emptyPayloadMethods = `type (
	// EmptyPayload is the EmptyPayload service interface.
	EmptyPayload interface {
		// EmptyPayload implements the EmptyPayload endpoint.
		EmptyPayload(context.Context) (AResult, error)
	}

	AResult struct {
		BooleanField bool
		BytesField   []byte
		IntField     int
		StringField  string
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
		apayload = design.UserTypeExpr{
			TypeName: "APayload",
			AttributeExpr: &design.AttributeExpr{Type: design.Object{
				"IntField":     &design.AttributeExpr{Type: design.Int},
				"StringField":  &design.AttributeExpr{Type: design.String},
				"BooleanField": &design.AttributeExpr{Type: design.Boolean},
				"BytesField":   &design.AttributeExpr{Type: design.Bytes},
			}},
		}

		bpayload = design.UserTypeExpr{
			TypeName: "BPayload",
			AttributeExpr: &design.AttributeExpr{Type: design.Object{
				"ArrayField":    &design.AttributeExpr{Type: &design.Array{&design.AttributeExpr{Type: design.Boolean}}},
				"MapField":      &design.AttributeExpr{Type: &design.Map{KeyType: &design.AttributeExpr{Type: design.Int}, ElemType: &design.AttributeExpr{Type: design.String}}},
				"ObjectField":   &design.AttributeExpr{Type: design.Object{"IntField": &design.AttributeExpr{Type: design.Int}, "StringField": &design.AttributeExpr{Type: design.String}}},
				"UserTypeField": &design.AttributeExpr{Type: parent},
			}},
		}

		aresult = design.UserTypeExpr{
			TypeName: "AResult",
			AttributeExpr: &design.AttributeExpr{Type: design.Object{
				"IntField":     &design.AttributeExpr{Type: design.Int},
				"StringField":  &design.AttributeExpr{Type: design.String},
				"BooleanField": &design.AttributeExpr{Type: design.Boolean},
				"BytesField":   &design.AttributeExpr{Type: design.Bytes},
			}},
		}

		bresult = design.UserTypeExpr{
			TypeName: "BResult",
			AttributeExpr: &design.AttributeExpr{Type: design.Object{
				"ArrayField":    &design.AttributeExpr{Type: &design.Array{&design.AttributeExpr{Type: design.Boolean}}},
				"MapField":      &design.AttributeExpr{Type: &design.Map{KeyType: &design.AttributeExpr{Type: design.Int}, ElemType: &design.AttributeExpr{Type: design.String}}},
				"ObjectField":   &design.AttributeExpr{Type: design.Object{"IntField": &design.AttributeExpr{Type: design.Int}, "StringField": &design.AttributeExpr{Type: design.String}}},
				"UserTypeField": &design.AttributeExpr{Type: parent},
			}},
		}

		a = design.EndpointExpr{
			Name:    "A",
			Payload: &apayload,
			Result:  &aresult,
		}

		b = design.EndpointExpr{
			Name:    "B",
			Payload: &bpayload,
			Result:  &bresult,
		}

		empty = design.EndpointExpr{
			Name:    "Empty",
			Payload: design.Empty,
			Result:  design.Empty,
		}

		emptyResult = design.EndpointExpr{
			Name:    "EmptyResult",
			Payload: &apayload,
			Result:  design.Empty,
		}

		emptyPayload = design.EndpointExpr{
			Name:    "EmptyPayload",
			Payload: design.Empty,
			Result:  &aresult,
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

		emptyEndpoint = design.ServiceExpr{
			Name: "Empty",
			Endpoints: []*design.EndpointExpr{
				&empty,
			},
		}

		emptyResultEndpoint = design.ServiceExpr{
			Name: "EmptyResult",
			Endpoints: []*design.EndpointExpr{
				&emptyResult,
			},
		}

		emptyPayloadEndpoint = design.ServiceExpr{
			Name: "EmptyPayload",
			Endpoints: []*design.EndpointExpr{
				&emptyPayload,
			},
		}
	)

	cases := map[string]struct {
		Service  *design.ServiceExpr
		Expected string
	}{
		"single":                             {Service: &singleEndpoint, Expected: singleMethod},
		"multiple":                           {Service: &multipleEndpoints, Expected: multipleMethods},
		"empty payload, empty result":        {Service: &emptyEndpoint, Expected: emptyMethods},
		"non empty payload but empty result": {Service: &emptyResultEndpoint, Expected: emptyResultMethods},
		"empty payload and non empty result": {Service: &emptyPayloadEndpoint, Expected: emptyPayloadMethods},
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
