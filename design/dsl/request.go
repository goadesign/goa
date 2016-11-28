package dsl

import (
	"fmt"
	"unicode"

	"goa.design/goa.v2/design"
	"goa.design/goa.v2/eval"
)

// Request defines the data type which lists the request parameters in its
// attributes. Transport specific DSL may provide a mapping between the
// attributes and incoming request state (e.g. which attributes are initialized
// from HTTP headers, query string values or body fields in the case of HTTP)
//
// Request may appear in a Endpoint expression.
//
// Request takes one or two arguments. The first argument is either a reference
// to a type, the name of a type or a DSL function.
// If the first argument is a type or the name of a type then an optional DSL
// may be passed as second argument that further specializes the type by
// providing additional validations (e.g. list of required attributes)
//
// Examples:
//
// Endpoint("add", func() {
//     // Define request type inline
//     Request(func() {
//         Attribute("left", Int32, "Left operand")
//         Attribute("right", Int32, "Left operand")
//         Required("left", "right")
//     })
// })
//
// Endpoint("add", func() {
//     // Define request type by reference to user type
//     Request(Operands)
// })
//
// Endpoint("divide", func() {
//     // Specify required attributes on user type
//     Request(Operands, func() {
//         Required("left", "right")
//     })
// })
//
func Request(val interface{}, dsls ...func()) {
	if len(dsls) > 1 {
		eval.ReportError("too many arguments")
	}
	e, ok := eval.Current().(*design.EndpointExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	e.Request = endpointTypeDSL("Request", val, dsls...)
}

func endpointTypeDSL(suffix string, p interface{}, dsls ...func()) design.UserType {
	var (
		e  = eval.Current().(*design.EndpointExpr)
		sn = camelize(e.Service.Name)
		en = camelize(e.Name)

		att *design.AttributeExpr
		ut  design.UserType
		dsl func()
	)
	switch actual := p.(type) {
	case func():
		dsl = actual
		att = &design.AttributeExpr{Type: design.Object{}}
	case *design.AttributeExpr:
		att = design.DupAtt(actual)
	case design.UserType:
		if design.IsObject(actual) {
			eval.ReportError("%s type must be an object, %s is a %s", suffix, actual.Name(), actual.Attribute().Type.Name())
			return nil
		}
		if len(dsls) == 0 {
			return actual
		}
		ut = design.Dup(actual).(design.UserType)
		ute, ok := ut.(*design.UserTypeExpr)
		if !ok {
			ute = ut.(*design.MediaTypeExpr).UserTypeExpr
		}
		ute.TypeName = fmt.Sprintf("%s%s%s", en, sn, suffix)
		att = ut.Attribute()
	case string:
		t := design.Root.UserType(actual)
		if t == nil {
			eval.ReportError("unknown request type %s", actual)
			return nil
		}
		if design.IsObject(t) {
			eval.ReportError("%s type must be an object, %s is a %s", suffix, actual, t.Attribute().Type.Name())
			return nil
		}
		ut = design.Dup(t).(design.UserType)
		att = ut.Attribute()
	default:
		eval.ReportError("invalid Request argument, must be a object type, a media type or a DSL building a type")
		return nil
	}
	if len(dsls) == 1 {
		if dsl != nil {
			eval.ReportError("invalid arguments in Request call, must be (type), (dsl) or (type, dsl)")
		}
		dsl = dsls[0]
	}
	if dsl != nil {
		eval.Execute(dsl, att)
	}
	if ut == nil {
		ut = &design.UserTypeExpr{
			AttributeExpr: att,
			TypeName:      fmt.Sprintf("%s%s%s", en, sn, suffix),
		}
	}
	return ut
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
