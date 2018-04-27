package service

import (
	"bytes"
	"fmt"
	"go/format"
	"strings"
	"testing"

	"goa.design/goa/codegen"
	"goa.design/goa/design"
)

func TestService(t *testing.T) {
	const (
		singleMethodCode = `// Service is the Single service interface.
type Service interface {
	// A implements A.
	A(context.Context, *APayload) (*AResult, error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "Single"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"A"}

// APayload is the payload type of the Single service A method.
type APayload struct {
	IntField      int
	StringField   string
	BooleanField  bool
	BytesField    []byte
	OptionalField *string
}

// AResult is the result type of the Single service A method.
type AResult struct {
	IntField      int
	StringField   string
	BooleanField  bool
	BytesField    []byte
	OptionalField *string
}
`

		multipleMethodsCode = `// Service is the Multiple service interface.
type Service interface {
	// A implements A.
	A(context.Context, *APayload) (*AResult, error)
	// B implements B.
	B(context.Context, *BPayload) (*BResult, error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "Multiple"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [2]string{"A", "B"}

// APayload is the payload type of the Multiple service A method.
type APayload struct {
	IntField      int
	StringField   string
	BooleanField  bool
	BytesField    []byte
	OptionalField *string
}

// AResult is the result type of the Multiple service A method.
type AResult struct {
	IntField      int
	StringField   string
	BooleanField  bool
	BytesField    []byte
	OptionalField *string
}

// BPayload is the payload type of the Multiple service B method.
type BPayload struct {
	ArrayField  []bool
	MapField    map[int]string
	ObjectField *struct {
		IntField    *int
		StringField *string
	}
	UserTypeField *Parent
}

// BResult is the result type of the Multiple service B method.
type BResult struct {
	ArrayField  []bool
	MapField    map[int]string
	ObjectField *struct {
		IntField    *int
		StringField *string
	}
	UserTypeField *Parent
}

type Parent struct {
	C *Child
}

type Child struct {
	P *Parent
}
`

		emptyMethodsCode = `// Service is the Empty service interface.
type Service interface {
	// Empty implements Empty.
	Empty(context.Context) error
}
`

		emptyResultMethodsCode = `// Service is the EmptyResult service interface.
type Service interface {
	// EmptyResult implements EmptyResult.
	EmptyResult(context.Context, *APayload) error
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "EmptyResult"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"EmptyResult"}

// APayload is the payload type of the EmptyResult service EmptyResult method.
type APayload struct {
	IntField      int
	StringField   string
	BooleanField  bool
	BytesField    []byte
	OptionalField *string
}
`

		emptyPayloadMethodsCode = `// Service is the EmptyPayload service interface.
type Service interface {
	// EmptyPayload implements EmptyPayload.
	EmptyPayload(context.Context) (*AResult, error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "EmptyPayload"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"EmptyPayload"}

// AResult is the result type of the EmptyPayload service EmptyPayload method.
type AResult struct {
	IntField      int
	StringField   string
	BooleanField  bool
	BytesField    []byte
	OptionalField *string
}
`

		serviceLevelErrorCode = `
// MakeError builds a goa.ServiceError from an error.
func MakeError(err error) *goa.ServiceError {
	return &goa.ServiceError{
		Name:    "error",
		ID:      goa.NewErrorID(),
		Message: err.Error(),
	}
}
`

		genPkg = "goa.design/goa/example"
	)
	var (
		child = &design.UserTypeExpr{
			TypeName: "Child",
		}

		parent = &design.UserTypeExpr{
			TypeName: "Parent",
			AttributeExpr: &design.AttributeExpr{
				Type: &design.Object{
					{"c", &design.AttributeExpr{
						Type: child,
					}},
				},
			},
		}
	)

	child.AttributeExpr = &design.AttributeExpr{
		Type: &design.Object{
			{"p", &design.AttributeExpr{Type: parent}},
		},
	}

	var (
		apayload = design.AttributeExpr{
			Type: &design.UserTypeExpr{
				TypeName: "APayload",
				AttributeExpr: &design.AttributeExpr{
					Type: &design.Object{
						{"IntField", &design.AttributeExpr{Type: design.Int}},
						{"StringField", &design.AttributeExpr{Type: design.String}},
						{"BooleanField", &design.AttributeExpr{Type: design.Boolean}},
						{"BytesField", &design.AttributeExpr{Type: design.Bytes}},
						{"OptionalField", &design.AttributeExpr{Type: design.String}},
					},
					Validation: &design.ValidationExpr{
						Required: []string{"IntField", "StringField", "BooleanField", "BytesField"},
					},
				},
			}}

		bpayload = design.AttributeExpr{
			Type: &design.UserTypeExpr{
				TypeName: "BPayload",
				AttributeExpr: &design.AttributeExpr{Type: &design.Object{
					{"ArrayField", &design.AttributeExpr{Type: &design.Array{&design.AttributeExpr{Type: design.Boolean}}}},
					{"MapField", &design.AttributeExpr{Type: &design.Map{KeyType: &design.AttributeExpr{Type: design.Int}, ElemType: &design.AttributeExpr{Type: design.String}}}},
					{"ObjectField", &design.AttributeExpr{Type: &design.Object{{"IntField", &design.AttributeExpr{Type: design.Int}}, {"StringField", &design.AttributeExpr{Type: design.String}}}}},
					{"UserTypeField", &design.AttributeExpr{Type: parent}},
				}},
			}}

		aresult = design.AttributeExpr{
			Type: &design.UserTypeExpr{
				TypeName: "AResult",
				AttributeExpr: &design.AttributeExpr{Type: &design.Object{
					{"IntField", &design.AttributeExpr{Type: design.Int}},
					{"StringField", &design.AttributeExpr{Type: design.String}},
					{"BooleanField", &design.AttributeExpr{Type: design.Boolean}},
					{"BytesField", &design.AttributeExpr{Type: design.Bytes}},
					{"OptionalField", &design.AttributeExpr{Type: design.String}},
				},
					Validation: &design.ValidationExpr{
						Required: []string{"IntField", "StringField", "BooleanField", "BytesField"},
					},
				},
			}}

		bresult = design.AttributeExpr{
			Type: &design.UserTypeExpr{
				TypeName: "BResult",
				AttributeExpr: &design.AttributeExpr{Type: &design.Object{
					{"ArrayField", &design.AttributeExpr{Type: &design.Array{&design.AttributeExpr{Type: design.Boolean}}}},
					{"MapField", &design.AttributeExpr{Type: &design.Map{KeyType: &design.AttributeExpr{Type: design.Int}, ElemType: &design.AttributeExpr{Type: design.String}}}},
					{"ObjectField", &design.AttributeExpr{Type: &design.Object{{"IntField", &design.AttributeExpr{Type: design.Int}}, {"StringField", &design.AttributeExpr{Type: design.String}}}}},
					{"UserTypeField", &design.AttributeExpr{Type: parent}},
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

		errorExp = design.ErrorExpr{
			Name: "error",
			AttributeExpr: &design.AttributeExpr{
				Type: design.ErrorResult,
			},
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

		serviceError = design.ServiceExpr{
			Name: "ServiceError",
			Methods: []*design.MethodExpr{
				&a1,
			},
			Errors: []*design.ErrorExpr{
				&errorExp,
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
		"service level error":                {Service: &serviceError, Expected: serviceLevelErrorCode},
	}
	for k, tc := range cases {
		t.Run(k, func(t *testing.T) {
			buf := new(bytes.Buffer)
			Services = make(ServicesData)
			design.Root = new(design.RootExpr)
			design.Root.API = &design.APIExpr{Name: "test"}
			design.Root.Services = []*design.ServiceExpr{tc.Service}
			file := File("", tc.Service)
			for _, s := range file.SectionTemplates {
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
				t.Errorf("%s:\ngot:\n%s\ndiff:\n%s", k, actual, codegen.Diff(t, actual, tc.Expected))
			}
		})
	}
}
