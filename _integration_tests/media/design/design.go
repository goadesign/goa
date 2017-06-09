package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var _ = API("media", func() {
	Title("An API exercising the DefaultMedia definition")
	Host("localhost:8080")
	Scheme("http")
})

var Greeting = Type("Greeting", func() {
	Attribute("id", Integer, func() {
		Description("A required int field in the parent type.")
	})
	Attribute("message", String, func() {
		Description("A required string field in the parent type.")
	})
	Attribute("parent_optional", Boolean, func() {
		Description("An optional boolean field in the parent type.")
	})
	Required("id", "message")
})

var GreetingMedia = MediaType("application/vnd.io.bluecanvas.helloworld.greeting.v1+json", func() {
	TypeName("GreetingMedia")
	Reference(Greeting)

	Attributes(func() {
		Attribute("id")
		Attribute("message")
		Attribute("parent_optional")
		Attribute("href", String, func() {
			Description("A required string field in the response media type.")
		})
		Required("id", "message", "href")
	})

	View("default", func() {
		Attribute("id")
		Attribute("message")
		Attribute("parent_optional")
		Attribute("href")
	})
})

var _ = Resource("Greeting", func() {
	DefaultMedia(GreetingMedia)

	Action("show", func() {
		Routing(
			GET("/"))
		Response(OK, GreetingMedia)
		Response(BadRequest)
	})

	Action("create", func() {
		Routing(
			POST("/"))
		Payload(func() {
			Member("message")
			Member("parent_optional")
			Required("message")
		})
		Response(Created)
		Response(BadRequest)
	})
})
