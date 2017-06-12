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
		singleMethodCode = `type (
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

		multipleMethodsCode = `type (
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

		emptyMethodsCode = `type (
	// Empty is the Empty service interface.
	Empty interface {
		// Empty implements Empty.
		Empty(context.Context) error
	}
)
`

		emptyResultMethodsCode = `type (
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

		emptyPayloadMethodsCode = `type (
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

		a1 = design.MethodExpr{
			Name:    "A",
			Payload: &apayload,
			Result:  &aresult,
		}

		a2 = design.MethodExpr{
			Name:    "A",
			Payload: &apayload,
			Result:  &aresult,
		}

		b = design.MethodExpr{
			Name:    "B",
			Payload: &bpayload,
			Result:  &bresult,
		}

		empty = design.MethodExpr{
			Name:    "Empty",
			Payload: &design.AttributeExpr{Type: design.Empty},
			Result:  &design.AttributeExpr{Type: design.Empty},
		}

		emptyResult = design.MethodExpr{
			Name:    "EmptyResult",
			Payload: &apayload,
			Result:  &design.AttributeExpr{Type: design.Empty},
		}

		emptyPayload = design.MethodExpr{
			Name:    "EmptyPayload",
			Payload: &design.AttributeExpr{Type: design.Empty},
			Result:  &aresult,
		}

		singleMethod = design.ServiceExpr{
			Name: "Single",
			Methods: []*design.MethodExpr{
				&a1,
			},
		}

		multipleMethods = design.ServiceExpr{
			Name: "Multiple",
			Methods: []*design.MethodExpr{
				&a2,
				&b,
			},
		}

		emptyMethod = design.ServiceExpr{
			Name: "Empty",
			Methods: []*design.MethodExpr{
				&empty,
			},
		}

		emptyResultMethod = design.ServiceExpr{
			Name: "EmptyResult",
			Methods: []*design.MethodExpr{
				&emptyResult,
			},
		}

		emptyPayloadMethod = design.ServiceExpr{
			Name: "EmptyPayload",
			Methods: []*design.MethodExpr{
				&emptyPayload,
			},
		}
	)
	singleMethod.Methods[0].Service = &singleMethod
	multipleMethods.Methods[0].Service = &multipleMethods
	multipleMethods.Methods[1].Service = &multipleMethods
	emptyMethod.Methods[0].Service = &emptyMethod
	emptyResultMethod.Methods[0].Service = &emptyResultMethod
	emptyPayloadMethod.Methods[0].Service = &emptyPayloadMethod

	cases := map[string]struct {
		Service  *design.ServiceExpr
		Expected string
	}{
		"single":                             {Service: &singleMethod, Expected: singleMethodCode},
		"multiple":                           {Service: &multipleMethods, Expected: multipleMethodsCode},
		"empty payload, empty result":        {Service: &emptyMethod, Expected: emptyMethodsCode},
		"non empty payload but empty result": {Service: &emptyResultMethod, Expected: emptyResultMethodsCode},
		"empty payload and non empty result": {Service: &emptyPayloadMethod, Expected: emptyPayloadMethodsCode},
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
