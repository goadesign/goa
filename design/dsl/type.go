package dsl

import . "github.com/raphael/goa/design"

// Type defines a user type DSL.
//
// Type("CreatePayload", func() {
// 	Description("Type of create and upload action payloads")
//	Attribute("name", String, "name of bottle")
//	Attribute("origin", Origin, "Details on wine origin")
// 	Required("name")
// })
//
// This function returns the newly defined user type.
func Type(name string, dsl func()) *UserTypeDefinition {
	if Design == nil {
		InitDesign()
	}
	if Design.Types == nil {
		Design.Types = make(map[string]*UserTypeDefinition)
	} else if _, ok := Design.Types[name]; ok {
		ReportError("type %#v defined twice", name)
		return nil
	}
	var t *UserTypeDefinition
	if topLevelDefinition(true) {
		t = &UserTypeDefinition{
			TypeName:            name,
			AttributeDefinition: &AttributeDefinition{},
			DSL:                 dsl,
		}
		if dsl == nil {
			t.Type = String
		}
		Design.Types[name] = t
	}
	return t
}

// ArrayOf creates an array from its element type.
func ArrayOf(t DataType) *Array {
	at := AttributeDefinition{Type: t}
	return &Array{ElemType: &at}
}

// HashOf creates a hash map from its key and element types.
func HashOf(k, v DataType) *Hash {
	kat := AttributeDefinition{Type: k}
	vat := AttributeDefinition{Type: v}
	return &Hash{KeyType: &kat, ElemType: &vat}
}
