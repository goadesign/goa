package dsl

import (
	"fmt"
	"unicode"

	"github.com/goadesign/goa/eval"
	"github.com/goadesign/goa/rpc/design"
)

// Endpoint defines a single service endpoint.
//
// Endpoint may appear in a Service expression.
// Endpoint takes two arguments: the name of the endpoint and the defining DSL.
//
// Example:
//
//    Endpoint("add", func() {
//        Description("The add endpoint returns the sum of A and B")
//        Docs(func() {
//            Description("Add docs")
//            URL("http//adder.goa.design/docs/actions/add")
//        })
//        Request(Operands)
//        Response(Sum)
//        Error(ErrInvalidOperands)
//    })
//
func Endpoint(name string, dsl func()) {
	s, ok := eval.Current().(*design.ServiceExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	ep := &design.EndpointExpr{Name: name, DSLFunc: dsl}
	s.Endpoints = append(s.Endpoints, ep)
}

// Request defines the data type which lists the request parameters in its
// attributes. Transport specific DSL may provide a mapping between the
// attributes and incoming request state (e.g. which attributes are initialized
// from HTTP headers, query string values or body fields in the case of HTTP)
//
// Request may appear in a Endpoint expression.
// Request takes one
func Request(p interface{}, dsls ...func()) {
	if len(dsls) > 1 {
		eval.ReportError("too many arguments given to Payload")
		return
	}
	a, ok := eval.Current().(*design.ActionExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	var att *apidesign.AttributeExpr
	var dsl func()
	switch actual := p.(type) {
	case func():
		dsl = actual
		att = newAttribute(a.Parent.MediaType)
		att.Type = apidesign.Object{}
	case *apidesign.AttributeExpr:
		att = apidesign.DupAtt(actual)
	case *apidesign.UserTypeExpr:
		if len(dsls) == 0 {
			a.Payload = actual
			a.PayloadOptional = isOptional
			return
		}
		att = apidesign.DupAtt(actual.Attribute())
	case *design.MediaTypeExpr:
		att = apidesign.DupAtt(actual.AttributeExpr)
	case string:
		ut := apidesign.Root.UserType(actual)
		if ut == nil {
			eval.ReportError("unknown payload type %s", actual)
			return
		}
		att = apidesign.DupAtt(ut.Attribute())
	case *apidesign.Array:
		att = &apidesign.AttributeExpr{Type: actual}
	case *apidesign.Map:
		att = &apidesign.AttributeExpr{Type: actual}
	case apidesign.Primitive:
		att = &apidesign.AttributeExpr{Type: actual}
	default:
		eval.ReportError("invalid Payload argument, must be a type, a media type or a DSL building a type")
		return
	}
	if len(dsls) == 1 {
		if dsl != nil {
			eval.ReportError("invalid arguments in Payload call, must be (type), (dsl) or (type, dsl)")
		}
		dsl = dsls[0]
	}
	if dsl != nil {
		eval.Execute(dsl, att)
	}
	rn := camelize(a.Parent.Name)
	an := camelize(a.Name)
	a.Payload = &apidesign.UserTypeExpr{
		AttributeExpr: att,
		TypeName:      fmt.Sprintf("%s%sPayload", an, rn),
	}
	a.PayloadOptional = isOptional
}

// newAttribute creates a new attribute definition using the media type with the given identifier
// as base type.
func newAttribute(baseMT string) *apidesign.AttributeExpr {
	var base apidesign.DataType
	if mt := design.Root.MediaType(baseMT); mt != nil {
		base = mt.Type
	}
	return &apidesign.AttributeExpr{Reference: base}
}

func camelize(str string) string {
	runes := []rune(str)
	w, i := 0, 0
	for i+1 <= len(runes) {
		eow := false
		if i+1 == len(runes) {
			eow = true
		} else if !validIdentifier(runes[i]) {
			runes = append(runes[:i], runes[i+1:]...)
		} else if spacer(runes[i+1]) {
			eow = true
			n := 1
			for i+n+1 < len(runes) && spacer(runes[i+n+1]) {
				n++
			}
			copy(runes[i+1:], runes[i+n+1:])
			runes = runes[:len(runes)-n]
		} else if unicode.IsLower(runes[i]) && !unicode.IsLower(runes[i+1]) {
			eow = true
		}
		i++
		if !eow {
			continue
		}
		runes[w] = unicode.ToUpper(runes[w])
		w = i
	}
	return string(runes)
}

// validIdentifier returns true if the rune is a letter or number
func validIdentifier(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

func spacer(c rune) bool {
	switch c {
	case '_', ' ', ':', '-':
		return true
	}
	return false
}
