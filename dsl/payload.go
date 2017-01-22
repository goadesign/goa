package dsl

import (
	"fmt"
	"unicode"

	"goa.design/goa.v2/design"
	"goa.design/goa.v2/eval"
)

// Payload defines the data type which lists the payload attributes.
//
// Payload may appear in a Endpoint expression.
//
// Payload takes one or two arguments. The first argument is either a type or a
// DSL function. If the first argument is a type then an optional DSL may be
// passed as second argument that further specializes the type by providing
// additional validations (e.g. list of required attributes)
//
// Examples:
//
// Endpoint("add", func() {
//     // Define payload type inline
//     Payload(func() {
//         Attribute("left", Int32, "Left operand")
//         Attribute("right", Int32, "Left operand")
//         Required("left", "right")
//     })
// })
//
// Endpoint("add", func() {
//     // Define payload type by reference to user type
//     Payload(Operands)
// })
//
// Endpoint("divide", func() {
//     // Specify required attributes on user type
//     Payload(Operands, func() {
//         Required("left", "right")
//     })
// })
//
func Payload(val interface{}, dsls ...func()) {
	if len(dsls) > 1 {
		eval.ReportError("too many arguments")
	}
	e, ok := eval.Current().(*design.EndpointExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	e.Payload = endpointTypeDSL("Payload", val, dsls...)
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
	if len(dsls) > 0 && dsls[0] == nil {
		dsls = dsls[1:]
	}
	switch actual := p.(type) {
	case func():
		dsl = actual
		att = &design.AttributeExpr{Type: design.Object{}}
	case design.UserType:
		if len(dsls) == 0 {
			// Do not duplicate type if it is not customized
			return actual
		}
		ut = design.Dup(actual).(design.UserType)
		att = ut.Attribute()
	case design.DataType:
		att = &design.AttributeExpr{Type: actual}
	default:
		eval.ReportError("invalid %s argument, must be a type or a function", suffix)
		return nil
	}
	if len(dsls) == 1 {
		if dsl != nil {
			eval.ReportError("invalid arguments in %s call, must be (type), (func) or (type, func)", suffix)
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
