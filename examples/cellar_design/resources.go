package main

import . "github.com/raphael/goa/design"

var _ = Resource("bottle", func() {

	MediaType(BottleMediaType)
	Trait("Authenticated")

	Action("list", func() {
		Routing(Get(""))
		Description("List all bottles in account optionally filtering by year")
		Params(
			Param("year", Integer, "Filter by year"),
		)
		Response(Ok, MediaCollection(BottleMediaType))
	})

	Action("get", func() {
		Routing(Get("/:id"))
		Description("Retrieve bottle with given id")
		Params(
			Param("id", Required()),
		)
		Response(Ok, BottleMediaType)
		Response(NotFound)
	})

	Action("create", func() {
		Routing(Post(""))
		Description("Record new bottle")
		Payload(Object(
			Member("name", Required()),
			Member("vintage", Required()),
			Member("vineyard", Required()),
			Member("varietal"),
			Member("color"),
			Member("sweet"),
			Member("country"),
			Member("region"),
			Member("review"),
			Member("characteristics"),
		))
		Response(Created)
	})

	Action("update", func() {
		Route(Patch("/:id"))
		Params(
			Param("id", Required()),
		)
		Payload(
			Member("name"),
			Member("vineyard"),
			Member("varietal"),
			Member("vintage"),
			Member("color"),
			Member("sweet"),
			Member("country"),
			Member("region"),
			Member("review"),
			Member("characteristics"),
		)
	})

	Action("delete", func() {
		Route(Delete("/:id"))
		Params(
			Param("id", Required()),
		)
		Headers(
			Header("X-Force", Enum("true", "false")),
		)
		Response(NoContent)
		Response(NotFound)
	})
})
