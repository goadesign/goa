package design

import (
	. "github.com/raphael/goa/design"
	. "github.com/raphael/goa/design/dsl"
)

// This is the cellar application API design used by goa to generate
// the application code, client, tests, documentation etc.
var _ = API("cellar", func() {
	Title("The virtual wine cellar")
	Description("A basic example of a CRUD API implemented with goa")
	Contact(func() {
		Name("goa team")
		Email("admin@goa.design")
		URL("http://goa.design")
	})
	License(func() {
		Name("MIT")
		URL("https://github.com/raphael/goa/blob/master/LICENSE")
	})
	Docs(func() {
		Description("goa guide")
		URL("http://goa.design/getting-started.html")
	})
	Host("cellar.goa.design")
	Scheme("http")
	BasePath("/cellar")

	ResponseTemplate(Created, func(pattern string) {
		Description("Resource created")
		Status(201)
		Headers(func() {
			Header("Location", String, "href to created resource", func() {
				Pattern(pattern)
			})
		})
	})
})
