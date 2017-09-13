package apidsl

import (
	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/dslengine"
)

// Type is a top level DSL.
//
// Type implements the type definition dsl. A type definition describes a data structure consisting
// of attributes. Each attribute has a type which can also refer to a type definition (or use a
// primitive type or nested attibutes). The dsl syntax for define a type definition is the
// Attribute dsl, see Attribute.
//
// On top of specifying any attribute type, type definitions can also be used to describe the data
// structure of a request payload. They can also be used by media type definitions as reference, see
// Reference. Here is an example:
//
//	var UpdatePayload = Type("UpdatePayload", func() {
//		Description("UpdatePayload describes the update action request bodies")
//		Attribute("origin", Origin, "Details on wine origin")  // See Origin definition below
//	})
//
//	Type("CreatePayload", func() {
//              Reference(UpdatePayload)
//		Description("CreatePayload describes the create action request bodies")
//		Attribute("name", String, "name of bottle")
//		Attribute("origin") // Inherits description, type from UpdatePayload
//		Required("name", "origin")
//	})
//
//	var Origin = Type("origin", func() {
//		Description("Origin of bottle")
//		Attribute("Country")
//	})
//
// This function returns the newly defined type so the value can be used throughout the dsl.
func Type(name string, dsl func()) *design.UserTypeDefinition {
	if design.Design.Types == nil {
		design.Design.Types = make(map[string]*design.UserTypeDefinition)
	} else if _, ok := design.Design.Types[name]; ok {
		dslengine.ReportError("type %#v defined twice", name)
		return nil
	}

	if !dslengine.IsTopLevelDefinition() {
		dslengine.IncompatibleDSL()
		return nil
	}

	t := &design.UserTypeDefinition{
		TypeName:            name,
		AttributeDefinition: &design.AttributeDefinition{DSLFunc: dsl},
	}
	if dsl == nil {
		t.Type = design.String
	} else {
		t.Type = make(design.Object)
	}
	design.Design.Types[name] = t
	return t
}

// ArrayOf creates an array type from its element type. The result can be used
// anywhere a type can. Examples:
//
//	var Bottle = Type("bottle", func() {
//		Attribute("name")
//	})
//
//	Action("update", func() {
//		Params(func() {
//			Param("ids", ArrayOf(Integer))
//		})
//		Payload(ArrayOf(Bottle))
//	})
//
// ArrayOf accepts an optional DSL as second argument which allows providing
// validations for the elements of the array:
//
//	Action("update", func() {
//		Params(func() {
//			Param("ids", ArrayOf(Integer, func() {
//				Minimum(1)
//			}))
//		})
//		Payload(ArrayOf(Bottle))
//	})
//
// If you are looking to return a collection of elements in a Response clause,
// refer to CollectionOf. ArrayOf creates a type, where CollectionOf creates a
// media type.
func ArrayOf(v interface{}, dsl ...func()) *design.Array {
	t := resolveType(v)
	// never return nil to avoid panics, errors are reported after DSL execution
	res := &design.Array{ElemType: &design.AttributeDefinition{Type: design.String}}
	if t == nil {
		dslengine.ReportError("invalid ArrayOf argument: not a type and not a known user type name")
		return res
	}
	if len(dsl) > 1 {
		dslengine.ReportError("ArrayOf: too many arguments")
		return res
	}
	at := design.AttributeDefinition{Type: t}
	if len(dsl) == 1 {
		dslengine.Execute(dsl[0], &at)
	}
	return &design.Array{ElemType: &at}
}

// HashOf creates a hash map from its key and element types. The result can be
// used anywhere a type can. Examples:
//
//	var Bottle = Type("bottle", func() {
//		Attribute("name")
//	})
//
//	var RatedBottles = HashOf(String, Bottle)
//
//	Action("updateRatings", func() {
//		Payload(func() {
//			Member("ratings", HashOf(String, Integer))
//			Member("bottles", RatedBottles)
//			// Member("bottles", "RatedBottles") // can use name of user type
//	})
//
// HashOf accepts optional DSLs as third and fourth argument which allows
// providing validations for the keys and values of the hash respectively:
//
//	var RatedBottles = HashOf(String, Bottle, func() {
//          Pattern("[a-zA-Z]+") // Validate bottle names
//      })
//
//      func ValidateKey() {
//          Pattern("^foo")
//      }
//
//      func TypeValue() {
//          Metadata("struct:field:type", "json.RawMessage", "encoding/json")
//      }
//
//	var Mappings = HashOf(String, String, ValidateKey, TypeValue)
//
func HashOf(k, v interface{}, dsls ...func()) *design.Hash {
	tk := resolveType(k)
	tv := resolveType(v)
	if tk == nil || tv == nil {
		// never return nil to avoid panics, errors are reported after DSL execution
		dslengine.ReportError("HashOf: invalid type name")
		return &design.Hash{
			KeyType:  &design.AttributeDefinition{Type: design.String},
			ElemType: &design.AttributeDefinition{Type: design.String},
		}
	}
	kat := design.AttributeDefinition{Type: tk}
	vat := design.AttributeDefinition{Type: tv}
	if len(dsls) > 2 {
		// never return nil to avoid panics, errors are reported after DSL execution
		dslengine.ReportError("HashOf: too many arguments")
		return &design.Hash{KeyType: &kat, ElemType: &vat}
	}
	if len(dsls) >= 1 {
		dslengine.Execute(dsls[0], &kat)
		if len(dsls) == 2 {
			dslengine.Execute(dsls[1], &vat)
		}
	}
	return &design.Hash{KeyType: &kat, ElemType: &vat}
}

func resolveType(v interface{}) design.DataType {
	if t, ok := v.(design.DataType); ok {
		return t
	}
	if name, ok := v.(string); ok {
		if ut, ok := design.Design.Types[name]; ok {
			return ut
		}
		if mt, ok := design.Design.MediaTypes[name]; ok {
			return mt
		}
	}
	return nil
}
