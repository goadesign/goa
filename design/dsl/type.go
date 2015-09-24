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
	at := &AttributeDefinition{}
	executeDSL(dsl, at)
	return &UserTypeDefinition{
		AttributeDefinition: at,
		TypeName:            name,
	}
}
