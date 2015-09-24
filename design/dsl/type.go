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
	if Design.Types == nil {
		Design.Types = make(map[string]*UserTypeDefinition)
	}
	var t *UserTypeDefinition
	at := &AttributeDefinition{}
	if executeDSL(dsl, at) {
		t = &UserTypeDefinition{
			TypeName:            name,
			AttributeDefinition: at,
		}
		Design.Types[name] = t
	}
	return t
}
