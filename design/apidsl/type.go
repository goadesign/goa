package apidsl

import (
	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/dslengine"
)

// Type implements the type definition dsl. A type definition describes a data structure consisting
// of attributes. Each attribute has a type which can also refer to a type definition (or use a
// primitive type or nested attibutes). The dsl syntax for define a type definition is the
// Attribute dsl, see Attribute.
//
// On top of specifying any attribute type, type definitions can also be used to describe the data
// structure of a request payload. They can also be used by media type definitions as reference, see
// Reference. Here is an example:
//
//	Type("createPayload", func() {
//		Description("Type of create and upload action payloads")
//		Attribute("name", String, "name of bottle")
//		Attribute("origin", Origin, "Details on wine origin")  // See Origin definition below
//		Required("name")
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

// ArrayOf creates an array type from its element type. The result can be used anywhere a type can.
// Examples:
//
//	var Bottle = Type("bottle", func() {
//		Attribute("name")
//	})
//
//	var Bottles = ArrayOf(Bottle)
//
//	Action("update", func() {
//		Params(func() {
//			Param("ids", ArrayOf(Integer))
//		})
//		Payload(ArrayOf(Bottle))  // Equivalent to Payload(Bottles)
//	})
//
// ArrayOf accepts an optional DSL as second argument which allows providing validations for the
// elements of the array:
//
//      var Names = ArrayOf(String, func() {
//          Pattern("[a-zA-Z]+")
//      })
//
// If you are looking to return a collection of elements in a Response clause, refer to
// CollectionOf.  ArrayOf creates a type, where CollectionOf creates a media type.
func ArrayOf(v interface{}, dsl ...func()) *design.Array {
	var t design.DataType
	var ok bool
	t, ok = v.(design.DataType)
	if !ok {
		if name, ok := v.(string); ok {
			if ut, ok := design.Design.Types[name]; ok {
				t = ut
			} else if mt, ok := design.Design.MediaTypes[name]; ok {
				t = mt
			}
		}
	}
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

// HashOf creates a hash map from its key and element types. The result can be used anywhere a type
// can. Examples:
//
//	var Bottle = Type("bottle", func() {
//		Attribute("name")
//	})
//
//	var RatedBottles = HashOf(String, Bottle)
//
//	Action("updateRatings", func() {
//		Payload(func() {
//			Member("ratings", HashOf(String, Integer))  // Artificial examples...
//			Member("bottles", RatedBottles)
//	})
//
// HashOf accepts optional DSLs as third and fourth argument which allows providing validations for
// the keys and values of the hash respectively:
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
func HashOf(k, v design.DataType, dsls ...func()) *design.Hash {
	kat := design.AttributeDefinition{Type: k}
	vat := design.AttributeDefinition{Type: v}
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
