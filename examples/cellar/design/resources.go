package main

import . "github.com/raphael/goa/design"

var _ = Resource("bottle", func() {

	MediaType(BottleMediaType)
	Trait("Authenticated")

	Action("list", func() {
		Route(Get(""))
		Description("List all bottles in account optionally filtering by year")
		Params(
			Param("years", Collection(Integer), "Filter by years"),
		)
		Response(Ok, MediaCollection(BottleMediaType))
	})

	Action("show", func() {
		Route(Get("/:id"))
		Description("Retrieve bottle with given id")
		Params(
			Param("id", Integer, "Bottle ID"),
		)
		Response(Ok, BottleMediaType)
		Response(NotFound)
	})

	Action("create", func() {
		Route(Post(""))
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
			Param("id", Integer, "Bottle ID"),
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
			Param("id", Integer, "Bottle ID"),
		)
		Headers(
			Header("X-Force", Enum("true", "false")),
		)
		Response(NoContent)
		Response(NotFound)
	})
})
