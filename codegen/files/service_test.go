package files

import (
	"bytes"
	"fmt"
	"go/format"
	"strings"
	"testing"

	"goa.design/goa.v2/design"

	. "goa.design/goa.v2/codegen/testing"
)

func TestService(t *testing.T) {
	const (
		singleMethod = `type (
	// Single is the Single service interface.
	Single interface {
		// A implements A.
		A(context.Context, *APayload) (*AResult, error)
	}

	// APayload is the payload type of the Single service A method.
	APayload struct {
		BooleanField  bool
		BytesField    []byte
		IntField      int
		OptionalField *string
		StringField   string
	}

	// AResult is the result type of the Single service A method.
	AResult struct {
		BooleanField  bool
		BytesField    []byte
		IntField      int
		OptionalField *string
		StringField   string
	}
)
`

		multipleMethods = `type (
	// Multiple is the Multiple service interface.
	Multiple interface {
		// A implements A.
		A(context.Context, *APayload) (*AResult, error)
		// B implements B.
		B(context.Context, *BPayload) (*BResult, error)
	}

	// APayload is the payload type of the Multiple service A method.
	APayload struct {
		BooleanField  bool
		BytesField    []byte
		IntField      int
		OptionalField *string
		StringField   string
	}

	// BPayload is the payload type of the Multiple service B method.
	BPayload struct {
		ArrayField  []bool
		MapField    map[int]string
		ObjectField *struct {
			IntField    *int
			StringField *string
		}
		UserTypeField *Parent
	}

	// AResult is the result type of the Multiple service A method.
	AResult struct {
		BooleanField  bool
		BytesField    []byte
		IntField      int
		OptionalField *string
		StringField   string
	}

	// BResult is the result type of the Multiple service B method.
	BResult struct {
		ArrayField  []bool
		MapField    map[int]string
		ObjectField *struct {
			IntField    *int
			StringField *string
		}
		UserTypeField *Parent
	}

	Parent struct {
		C *Child
	}

	Child struct {
		P *Parent
	}
)
`

		emptyMethods = `type (
	// Empty is the Empty service interface.
	Empty interface {
		// Empty implements Empty.
		Empty(context.Context) error
	}
)
`

		emptyResultMethods = `type (
	// EmptyResult is the EmptyResult service interface.
	EmptyResult interface {
		// EmptyResult implements EmptyResult.
		EmptyResult(context.Context, *APayload) error
	}

	// APayload is the payload type of the EmptyResult service EmptyResult method.
	APayload struct {
		BooleanField  bool
		BytesField    []byte
		IntField      int
		OptionalField *string
		StringField   string
	}
)
`

		emptyPayloadMethods = `type (
	// EmptyPayload is the EmptyPayload service interface.
	EmptyPayload interface {
		// EmptyPayload implements EmptyPayload.
		EmptyPayload(context.Context) (*AResult, error)
	}

	// AResult is the result type of the EmptyPayload service EmptyPayload method.
	AResult struct {
		BooleanField  bool
		BytesField    []byte
		IntField      int
		OptionalField *string
		StringField   string
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
		apayload = design.AttributeExpr{
			Type: &design.UserTypeExpr{
				TypeName: "APayload",
				AttributeExpr: &design.AttributeExpr{
					Type: design.Object{
						"IntField":      &design.AttributeExpr{Type: design.Int},
						"StringField":   &design.AttributeExpr{Type: design.String},
						"BooleanField":  &design.AttributeExpr{Type: design.Boolean},
						"BytesField":    &design.AttributeExpr{Type: design.Bytes},
						"OptionalField": &design.AttributeExpr{Type: design.String},
					},
					Validation: &design.ValidationExpr{
						Required: []string{"IntField", "StringField", "BooleanField", "BytesField"},
					},
				},
			}}

		bpayload = design.AttributeExpr{
			Type: &design.UserTypeExpr{
				TypeName: "BPayload",
				AttributeExpr: &design.AttributeExpr{Type: design.Object{
					"ArrayField":    &design.AttributeExpr{Type: &design.Array{&design.AttributeExpr{Type: design.Boolean}}},
					"MapField":      &design.AttributeExpr{Type: &design.Map{KeyType: &design.AttributeExpr{Type: design.Int}, ElemType: &design.AttributeExpr{Type: design.String}}},
					"ObjectField":   &design.AttributeExpr{Type: design.Object{"IntField": &design.AttributeExpr{Type: design.Int}, "StringField": &design.AttributeExpr{Type: design.String}}},
					"UserTypeField": &design.AttributeExpr{Type: parent},
				}},
			}}

		aresult = design.AttributeExpr{
			Type: &design.UserTypeExpr{
				TypeName: "AResult",
				AttributeExpr: &design.AttributeExpr{Type: design.Object{
					"IntField":      &design.AttributeExpr{Type: design.Int},
					"StringField":   &design.AttributeExpr{Type: design.String},
					"BooleanField":  &design.AttributeExpr{Type: design.Boolean},
					"BytesField":    &design.AttributeExpr{Type: design.Bytes},
					"OptionalField": &design.AttributeExpr{Type: design.String},
				},
					Validation: &design.ValidationExpr{
						Required: []string{"IntField", "StringField", "BooleanField", "BytesField"},
					},
				},
			}}

		bresult = design.AttributeExpr{
			Type: &design.UserTypeExpr{
				TypeName: "BResult",
				AttributeExpr: &design.AttributeExpr{Type: design.Object{
					"ArrayField":    &design.AttributeExpr{Type: &design.Array{&design.AttributeExpr{Type: design.Boolean}}},
					"MapField":      &design.AttributeExpr{Type: &design.Map{KeyType: &design.AttributeExpr{Type: design.Int}, ElemType: &design.AttributeExpr{Type: design.String}}},
					"ObjectField":   &design.AttributeExpr{Type: design.Object{"IntField": &design.AttributeExpr{Type: design.Int}, "StringField": &design.AttributeExpr{Type: design.String}}},
					"UserTypeField": &design.AttributeExpr{Type: parent},
				}},
			}}

		a1 = design.EndpointExpr{
			Name:    "A",
			Payload: &apayload,
			Result:  &aresult,
		}

		a2 = design.EndpointExpr{
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
			Payload: &design.AttributeExpr{Type: design.Empty},
			Result:  &design.AttributeExpr{Type: design.Empty},
		}

		emptyResult = design.EndpointExpr{
			Name:    "EmptyResult",
			Payload: &apayload,
			Result:  &design.AttributeExpr{Type: design.Empty},
		}

		emptyPayload = design.EndpointExpr{
			Name:    "EmptyPayload",
			Payload: &design.AttributeExpr{Type: design.Empty},
			Result:  &aresult,
		}

		singleEndpoint = design.ServiceExpr{
			Name: "Single",
			Endpoints: []*design.EndpointExpr{
				&a1,
			},
		}

		multipleEndpoints = design.ServiceExpr{
			Name: "Multiple",
			Endpoints: []*design.EndpointExpr{
				&a2,
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
	singleEndpoint.Endpoints[0].Service = &singleEndpoint
	multipleEndpoints.Endpoints[0].Service = &multipleEndpoints
	multipleEndpoints.Endpoints[1].Service = &multipleEndpoints
	emptyEndpoint.Endpoints[0].Service = &emptyEndpoint
	emptyResultEndpoint.Endpoints[0].Service = &emptyResultEndpoint
	emptyPayloadEndpoint.Endpoints[0].Service = &emptyPayloadEndpoint

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
		Services = make(ServicesData)
		design.Root = new(design.RootExpr)
		design.Root.Services = []*design.ServiceExpr{tc.Service}
		file := Service(tc.Service)
		for _, s := range file.Sections(genPkg) {
			if err := s.Write(buf); err != nil {
				t.Fatal(err)
			}
		}
		bs, err := format.Source(buf.Bytes())
		if err != nil {
			fmt.Println(buf.String())
			t.Fatal(err)
		}
		actual := string(bs)
		if !strings.Contains(actual, tc.Expected) {
			t.Errorf("%s:\ngot:\n%s\ndiff:\n%s", k, actual, Diff(t, actual, tc.Expected))
		}
	}
}
