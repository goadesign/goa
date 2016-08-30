package dsl

import (
	"fmt"
	"unicode"

	"github.com/goadesign/goa/eval"
	"github.com/goadesign/goa/http/design"
)

// Payload implements the action payload DSL. An action payload describes the HTTP request body
// data structure. The function accepts either a type or a DSL that describes the payload members.
// The Member DSL accepts the same syntax as the Attribute DSL. This function can be called passing
// in a type, a DSL or both. Examples:
//
//	Payload(BottlePayload)		// Request payload is described by the BottlePayload type
//
//	Payload(func() {		// Request payload is an object and is described inline
//		Member("Name")
//	})
//
//	Payload(BottlePayload, func() {	// Request payload is described by merging the inline
//		Required("Name")	// definition into the BottlePayload type.
//	})
//
func Payload(p interface{}, dsls ...func()) {
	payload(false, p, dsls...)
}

// OptionalPayload implements the action optional payload DSL. The function works identically to the
// Payload DSL except it sets a bit in the action definition to denote that the payload is not
// required. Example:
//
//	OptionalPayload(BottlePayload)		// Request payload is described by the BottlePayload type and is optional
//
func OptionalPayload(p interface{}, dsls ...func()) {
	payload(true, p, dsls...)
}

func payload(isOptional bool, p interface{}, dsls ...func()) {
	if len(dsls) > 1 {
		eval.ReportError("too many arguments given to Payload")
		return
	}
	if a, ok := eval.Current().(*design.ActionDefinition); ok {
		var att *design.AttributeDefinition
		var dsl func()
		switch actual := p.(type) {
		case func():
			dsl = actual
			att = newAttribute(a.Parent.MediaType)
			att.Type = design.Object{}
		case *design.AttributeDefinition:
			att = design.DupAtt(actual)
		case *design.UserTypeDefinition:
			if len(dsls) == 0 {
				a.Payload = actual
				a.PayloadOptional = isOptional
				return
			}
			att = design.DupAtt(actual.Definition())
		case *design.MediaTypeDefinition:
			att = design.DupAtt(actual.AttributeDefinition)
		case string:
			ut, ok := design.Design.Types[actual]
			if !ok {
				eval.ReportError("unknown payload type %s", actual)
			}
			att = design.DupAtt(ut.AttributeDefinition)
		case *design.Array:
			att = &design.AttributeDefinition{Type: actual}
		case *design.Hash:
			att = &design.AttributeDefinition{Type: actual}
		case design.Primitive:
			att = &design.AttributeDefinition{Type: actual}
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
		a.Payload = &design.UserTypeDefinition{
			AttributeDefinition: att,
			TypeName:            fmt.Sprintf("%s%sPayload", an, rn),
		}
		a.PayloadOptional = isOptional
	} else {
		eval.IncompatibleDSL()
	}
}

// newAttribute creates a new attribute definition using the media type with the given identifier
// as base type.
func newAttribute(baseMT string) *design.AttributeDefinition {
	var base design.DataType
	if mt := design.Design.MediaTypeWithIdentifier(baseMT); mt != nil {
		base = mt.Type
	}
	return &design.AttributeDefinition{Reference: base}
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
