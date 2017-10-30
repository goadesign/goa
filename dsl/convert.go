package dsl

import (
	"goa.design/goa/design"
	"goa.design/goa/eval"
)

// ConvertTo specifies an external type that instances of the generated struct
// corresponding to the outter type should map to. The generated struct is
// equipped with a method that makes it possible to instantiate the external
// type. The external type must be a struct with field types matching the
// attribute types defined in the DSL. Attribute expressions may take advantage
// of the "struct.field.external" metadata to specify the field of the external
// struct that matches if its name differs.
//
// ConvertTo must appear in Type or ResutType.
//
// ConvertTo accepts one arguments: an instance of the external type.
//
// Example:
//
// Service design:
//
//    var Bottle = Type("bottle", func() {
//        Description("A bottle")
//        ConvertTo(models.Bottle{})
//        Attribute("name", String, func() {
//	          // The generated type "Name" field is matched to the external
//            // type field "MyName".
//	          Metadata("struct.field.external", "MyName")
//        })
//    })
//
// External (i.e. non design) package:
//
//    package model
//
//    type Bottle struct {
//        // Mapped field
//        MyName string
//        // Other non-mapped fields are OK
//        Description string
//    }
//
func ConvertTo(obj interface{}) {
	at, ok := eval.Current().(*design.AttributeExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	for _, t := range design.Root.Types {
		if t.Attribute() == at {
			design.Root.Conversions =
				append(design.Root.Conversions, &design.TypeMap{User: t, External: obj})
		}
	}
}

// CreateFrom specifies an external type that instances of the generated struct
// corresponding to the outter type should map to. The generated struct is
// equipped with a method that makes it possible to instantiate it from an
// instance of the external type. The external type must be a struct with field
// types matching the attribute types defined in the DSL. Attribute expressions
// may take advantage of the "struct.field.external" metadata to specify the
// field of the external struct that matches if its name differs.
//
// CreateFrom must appear in Type or ResutType.
//
// CreateFrom accepts one arguments: an instance of the external type.
//
// Example:
//
// Service design:
//
//    var Bottle = Type("bottle", func() { Description("A bottle")
//        CreateFrom(models.Bottle{}) Attribute("name", String, func() {
//            // The generated type "Name" field is matched to the external
//            // type field "MyName".
//            Metadata("struct.field.external", "MyName")
//        })
//    })
//
// External (i.e. non design) package:
//
//    package model
//
//    type Bottle struct {
//        // Mapped field
//        MyName string
//        // Other non-mapped fields are OK
//        Description string
//    }
//
func CreateFrom(obj interface{}) {
	at, ok := eval.Current().(*design.AttributeExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	for _, t := range design.Root.Types {
		if t.Attribute() == at {
			design.Root.Creations =
				append(design.Root.Creations, &design.TypeMap{User: t, External: obj})
		}
	}
}
