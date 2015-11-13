package design

import (
	. "github.com/raphael/goa/design"
	. "github.com/raphael/goa/design/dsl"
)

// This is the cellar application API metadata used by goa to generate
// the application code, client, tests, documentation etc.
var _ = API("cellar", func() {

	Title("The virtual wine cellar")
	Description("A basic example of a CRUD API implemented with goa")
	BasePath("/cellar")

	ResponseTemplate("Created", func() {
		Description("Resource created")
		Status(201)
		Headers(func() {
			Header("Location", String, "href to created resource")
		})
	})

	Trait("Authenticated", func() {
		Headers(func() {
			Header("Auth-Token")
		})
	})
})
