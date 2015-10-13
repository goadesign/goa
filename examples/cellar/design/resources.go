package design

import (
	. "github.com/raphael/goa/design"
	. "github.com/raphael/goa/design/dsl"
)

var _ = Resource("account", func() {

	MediaType(Account)
	BasePath("/accounts")
	CanonicalActionName("show")
	Trait("Authenticated")

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
		Response(Created)
	})

	Action("update", func() {
		Routing(
			PUT("/:accountID"),
		)
		Description("Change account name")
		Params(func() {
			Param("accountID", Integer, "Accoutn ID")
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

	MediaType(Bottle)
	BasePath("bottles")
	Parent("account")
	CanonicalActionName("show")
	Trait("Authenticated")

	Action("list", func() {
		Routing(
			GET(""),
		)
		Description("List all bottles in account optionally filtering by year")
		Params(func() {
			Param("years", ArrayOf(Integer), "Filter by years")
		})
		Response(OK, func() {
			MediaType(CollectionOf(Bottle, func() {
				View("default")
				View("tiny")
			}))
		})
	})

	Action("show", func() {
		Routing(
			GET("/:id"),
		)
		Description("Retrieve bottle with given id")
		Params(func() {
			Param("id")
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
		Response(Created)
	})

	Action("update", func() {
		Routing(
			PATCH("/:id"),
		)
		Params(func() {
			Param("id")
		})
		Payload(BottlePayload)
		Response(NoContent)
		Response(NotFound)
	})

	Action("rate", func() {
		Routing(
			PUT("/:id/actions/rate"),
		)
		Params(func() {
			Param("id")
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
			DELETE("/:id"),
		)
		Params(func() {
			Param("id")
		})
		Headers(func() {
			Header("X-Force", func() {
				Enum("true", "false")
			})
		})
		Response(NoContent)
		Response(NotFound)
	})
})
