package main

import (
	. "github.com/raphael/goa/design"
	. "github.com/raphael/goa/design/dsl"
)

var _ = Resource("account", func() {

	MediaType(AccountMediaType)
	BasePath("/accounts")
	CanonicalAction("show")
	Trait("Authenticated")

	Action("show", func() {
		Route(Get("/:id"))
		Description("Retrieve account with given id")
		Params(
			Param("id", Integer, "Account ID"),
		)
		Response(OK, AccountMediaType)
		Response(NotFound)
	})

	Action("create", func() {
		Route(Post(""))
		Description("Create new account")
		Payload(
			Member("name", Required()),
		)
		Response(Created)
	})

	Action("update", func() {
		Route(Put("/:id"))
		Description("Change account name")
		Params(
			Param("id", Integer, "Accoutn ID"),
		)
		Payload(
			Member("name"),
		)
		Response(NoContent)
		Response(NotFound)
	})

	Action("delete", func() {
		Route(Delete("/:id"))
		Params(
			Param("id", Integer, "Account ID"),
		)
		Response(NoContent)
		Response(NotFound)
	})
})

var _ = Resource("bottle", func() {

	MediaType(BottleMediaType)
	Prefix("bottles")
	Parent("accounts")
	CanonicalAction("show")
	Trait("Authenticated")

	Action("list", func() {
		Route(Get(""))
		Description("List all bottles in account optionally filtering by year")
		Params(
			Param("years", Collection(Integer), "Filter by years"),
		)
		Response(OK, MediaCollection(BottleMediaType))
	})

	Action("show", func() {
		Route(Get("/:id"))
		Description("Retrieve bottle with given id")
		Params(
			Param("id", Integer, "Bottle ID"),
		)
		Response(OK, BottleMediaType)
		Response(NotFound)
	})

	Action("create", func() {
		Route(Post(""))
		Description("Record new bottle")
		Payload(
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
		)
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
		Response(NoContent)
		Response(NotFound)
	})

	Action("rate", func() {
		Route(Put("/:id/actions/rate"))
		Params(
			Param("id", Integer, "Bottle ID"),
		)
		Payload(
			Member("rating"),
		)
		Response(NoContent)
		Response(NotFound)
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
