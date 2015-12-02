package design

import (
	. "github.com/raphael/goa/design"
	. "github.com/raphael/goa/design/dsl"
)

var _ = Resource("account", func() {

	DefaultMedia(Account)
	BasePath("/accounts")

	Action("show", func() {
		Routing(
			GET("/:accountID"),
		)
		Description("Retrieve account with given id")
		Params(func() {
			Param("accountID", Integer, "Account ID")
		})
		Response(OK)
		Response(NotFound)
	})

	Action("create", func() {
		Routing(
			POST(""),
		)
		Description("Create new account")
		Payload(func() {
			Member("name")
			Required("name")
		})
		Response(Created, "/accounts/[0-9]+")
	})

	Action("update", func() {
		Routing(
			PUT("/:accountID"),
		)
		Description("Change account name")
		Params(func() {
			Param("accountID", Integer, "Account ID")
		})
		Payload(func() {
			Member("name")
			Required("name")
		})
		Response(NoContent)
		Response(NotFound)
	})

	Action("delete", func() {
		Routing(
			DELETE("/:accountID"),
		)
		Params(func() {
			Param("accountID", Integer, "Account ID")
		})
		Response(NoContent)
		Response(NotFound)
	})
})

var _ = Resource("bottle", func() {

	DefaultMedia(Bottle)
	BasePath("bottles")
	Parent("account")

	Action("list", func() {
		Routing(
			GET(""),
		)
		Description("List all bottles in account optionally filtering by year")
		Params(func() {
			Param("years", ArrayOf(Integer), "Filter by years")
		})
		Response(OK, func() {
			Media(CollectionOf(Bottle, func() {
				View("default")
				View("tiny")
			}))
		})
		Response(NotFound)
	})

	Action("show", func() {
		Routing(
			GET("/:bottleID"),
		)
		Description("Retrieve bottle with given id")
		Params(func() {
			Param("bottleID", Integer)
		})
		Response(OK)
		Response(NotFound)
	})

	Action("create", func() {
		Routing(
			POST(""),
		)
		Description("Record new bottle")
		Payload(BottlePayload, func() {
			Required("name", "vineyard", "varietal", "vintage", "color")
		})
		Response(Created, "^/accounts/[0-9]+/bottles/[0-9]+$")
	})

	Action("update", func() {
		Routing(
			PATCH("/:bottleID"),
		)
		Params(func() {
			Param("bottleID", Integer)
		})
		Payload(BottlePayload)
		Response(NoContent)
		Response(NotFound)
	})

	Action("rate", func() {
		Routing(
			PUT("/:bottleID/actions/rate"),
		)
		Params(func() {
			Param("bottleID", Integer)
		})
		Payload(func() {
			Member("rating")
			Required("rating")
		})
		Response(NoContent)
		Response(NotFound)
	})

	Action("delete", func() {
		Routing(
			DELETE("/:bottleID"),
		)
		Params(func() {
			Param("bottleID", Integer)
		})
		Response(NoContent)
		Response(NotFound)
	})
})
