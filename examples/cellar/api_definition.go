package main

import . "github.com/raphael/goa/design"

var _ = API("cellar", func() {
	Title("The virtual wine cellar")
	Description("A basic example of a CRUD API implemented with goa")
	BasePath("/:accountID")
	BaseParams(
		Attribute("accountID", Integer, func() {
			Description("API request account. All actions operate on resources belonging to the account.")
		}),
	)
	ResponseTemplate("NotFound", func() {
		Description("Resource not found")
		Status(404)
		MediaType("application/json")
	})
	ResponseTemplate("Ok", func(mt string) {
		Description("Resource listing")
		Status(200)
		MediaType(mt)
	})
})

// Multiple calls to ApiDefinition for the same API are possible
var _ = API("cellar", func() {
	Trait("Authenticated", func() {
		Headers(
			Key("Auth-Token", String, func() {
				Required()
			}),
		)
	})
})
