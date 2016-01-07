package dsl

import "github.com/raphael/goa/design"

// Type implements the type definition DSL. A type definition describes a data structure consisting
// of attributes. Each attribute has a type which can also refer to a type definition (or use a
// primitive type or nested attibutes). The DSL syntax for define a type definition is the
// Attribute DSL, see Attribute.
//
// On top of specifying any attribute type, type definitions can also be used to describe the data
// structure of a request payload. They can also be used by media type definitions as reference, see
// Reference. Here is an example:
//
// Type("createPayload", func() {
// 	Description("Type of create and upload action payloads")
//	APIVersion("1.0")
//	Attribute("name", String, "name of bottle")
//	Attribute("origin", Origin, "Details on wine origin")  // See Origin definition below
// 	Required("name")
// })
//
// var Origin = Type("origin", func() {
//	Description("Origin of bottle")
//	Attribute("Country")
// })
//
// This function returns the newly defined type so the value can be used throughout the DSL.
func Type(name string, dsl func()) *design.UserTypeDefinition {
	if design.Design == nil {
		InitDesign()
	}
	if design.Design.Types == nil {
		design.Design.Types = make(map[string]*design.UserTypeDefinition)
	} else if _, ok := design.Design.Types[name]; ok {
		ReportError("type %#v defined twice", name)
		return nil
	}
	var t *design.UserTypeDefinition
	if topLevelDefinition(true) {
		t = &design.UserTypeDefinition{
			TypeName:            name,
			AttributeDefinition: &design.AttributeDefinition{},
			DSL:                 dsl,
		}
		if dsl == nil {
			t.Type = design.String
		}
		design.Design.Types[name] = t
	}
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
func ArrayOf(t design.DataType) *design.Array {
	at := design.AttributeDefinition{Type: t}
	if ds, ok := t.(design.DataStructure); ok {
		at.APIVersions = ds.Definition().APIVersions
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
//			Member("ratings", HashOf(String, Integer)) // Artificial examples...
//			Member("bottles", RatedBottles)
//	})
func HashOf(k, v design.DataType) *design.Hash {
	kat := design.AttributeDefinition{Type: k}
	vat := design.AttributeDefinition{Type: v}
	return &design.Hash{KeyType: &kat, ElemType: &vat}
}
