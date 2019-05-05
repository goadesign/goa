package dsl

import (
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// ConvertTo specifies an external type that instances of the generated struct
// are converted into. The generated struct is equipped with a method that makes
// it possible to instantiate the external type. The default algorithm used to
// match the external type fields to the design attributes is as follows:
//
//    1. Look for an attribute with the same name as the field
//    2. Look for an attribute with the same name as the field but with the
//       first letter being lowercase
//    3. Look for an attribute with a name corresponding to the snake_case
//       version of the field name
//
// This algorithm does not apply if the attribute is equipped with the
// "struct.field.external" meta. In this case the matching is done by
// looking up the field with a name corresponding to the value of the meta.
// If the value of the meta is "-" the attribute isn't matched and no
// conversion code is generated for it. In all other cases it is an error if no
// match is found or if the matching field type does not correspond to the
// attribute type.
//
// The following limitations apply on the external Go struct field types
// recursively:
//
//    * struct fields must use pointers
//    * pointers on slices or on maps are not supported
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
//        // The "rating" attribute is matched to the external
//        // typ "Rating" field.
//        Attribute("rating", Int)
//        Attribute("name", String, func() {
//            // The "name" attribute is matched to the external
//            // type "MyName" field.
//            Meta("struct.field.external", "MyName")
//        })
//        Attribute("vineyard", String, func() {
//            // The "vineyard" attribute is not converted.
//            Meta("struct.field.external", "-")
//        })
//    })
//
// External (i.e. non design) package:
//
//    package model
//
//    type Bottle struct {
//        Rating int
//        // Mapped field
//        MyName string
//        // Additional fields are OK
//        Description string
//    }
//
func ConvertTo(obj interface{}) {
	var ut expr.UserType
	switch actual := eval.Current().(type) {
	case *expr.AttributeExpr:
		for _, t := range expr.Root.Types {
			if t.Attribute() == actual {
				ut = t
			}
		}
	case *expr.ResultTypeExpr:
		ut = actual
	default:
		eval.IncompatibleDSL()
		return
	}
	expr.Root.Conversions =
		append(expr.Root.Conversions, &expr.TypeMap{User: ut, External: obj})
}

// CreateFrom specifies an external type that instances of the generated struct
// can be initialized from. The generated struct is equipped with a method that
// initializes its fields from an instance of the external type. The default
// algorithm used to match the external type fields to the design attributes is
// as follows:
//
//    1. Look for an attribute with the same name as the field
//    2. Look for an attribute with the same name as the field but with the
//       first letter being lowercase
//    3. Look for an attribute with a name corresponding to the snake_case
//       version of the field name
//
// This algorithm does not apply if the attribute is equipped with the
// "struct.field.external" meta. In this case the matching is done by
// looking up the field with a name corresponding to the value of the meta.
// If the value of the meta is "-" the attribute isn't matched and no
// conversion code is generated for it. In all other cases it is an error if no
// match is found or if the matching field type does not correspond to the
// attribute type.
//
// The following limitations apply on the external Go struct field types
// recursively:
//
//    * struct fields must use pointers
//    * pointers on slices or on maps are not supported
//
// CreateFrom must appear in Type or ResultType.
//
// CreateFrom accepts one arguments: an instance of the external type.
//
// Example:
//
// Service design:
//
//    var Bottle = Type("bottle", func() {
//        Description("A bottle")
//        CreateFrom(models.Bottle{})
//        Attribute("rating", Int)
//        Attribute("name", String, func() {
//            // The "name" attribute is matched to the external
//            // type "MyName" field.
//            Meta("struct.field.external", "MyName")
//        })
//        Attribute("vineyard", String, func() {
//            // The "vineyard" attribute is not initialized by the
//            // generated constructor method.
//            Meta("struct.field.external", "-")
//        })
//    })
//
// External (i.e. non design) package:
//
//    package model
//
//    type Bottle struct {
//        Rating int
//        // Mapped field
//        MyName string
//        // Additional fields are OK
//        Description string
//    }
//
func CreateFrom(obj interface{}) {
	var ut expr.UserType
	switch actual := eval.Current().(type) {
	case *expr.AttributeExpr:
		for _, t := range expr.Root.Types {
			if t.Attribute() == actual {
				ut = t
			}
		}
	case *expr.ResultTypeExpr:
		ut = actual
	default:
		eval.IncompatibleDSL()
		return
	}
	expr.Root.Creations =
		append(expr.Root.Creations, &expr.TypeMap{User: ut, External: obj})
}
