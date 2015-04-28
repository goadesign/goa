package design

import . "github.com/raphael/goa/design"

var _ = Resource("bottle", func() {
	MediaType(BottleMediaType)
	Trait("Authenticated")

	Action("list", func() {
		Routing(Get(""))
		Description("List all bottles in account optionally filtering by year")
		Params(
			Attribute("year", Integer, func() {
				Description("Filter by year")
			}), // Equivalent to A("year")
		)
		Response(Ok, MediaCollection(BottleMediaType))
	})

	Action("get", func() {
		Routing(Get("/:id"))
		Description("Retrieve bottle with given id")
		Params(
			Attribute("id", Required()),
		)
		Response(Ok, BottleMediaType)
		Response(NotFound)
	})

	Action("create", func() {
		Routing(Post(""))
		Description("Record new bottle")
		Payload(
			Attribute("name", func() { Required() }),
			Attribute("vintage", func() { Required() }),
			Attribute("vineyard", func() { Required() }),
			Attribute("varietal"),
			Attribute("color"),
			Attribute("sweet"),
			Attribute("country"),
			Attribute("region"),
			Attribute("review"),
			Attribute("characteristics"),
		)
		Response(Created)
	})

	Action("update", func() {
		Route(Patch("/:id"))
		Params(
			Attribute("id", func() { Required() }),
		)
		Payload(
			Attribute("name"),
			Attribute("vineyard"),
			Attribute("varietal"),
			Attribute("vintage"),
			Attribute("color"),
			Attribute("sweet"),
			Attribute("country"),
			Attribute("region"),
			Attribute("review"),
			Attribute("characteristics"),
		)
	})

	Action("delete", func() {
		Route(Delete("/:id"))
		Params(
			Attribute("id", func() { Required() }),
		)
		Headers(
			Header("X-Force", func() {
				//Required()
				Enum("true", "false")
			})
		)
		Response(NoContent)
		Response(NotFound)
	})
})
