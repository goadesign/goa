package design

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
		Routing(
			GET("/:id"),
		)
		Description("Retrieve account with given id")
		Params(func() {
			Param("id", Integer, "Account ID")
		})
		Response(OK, func() {
			MediaType(AccountMediaType)
		})
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
			PUT("/:id"),
		)
		Description("Change account name")
		Params(func() {
			Param("id", Integer, "Accoutn ID")
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
			DELETE("/:id"),
		)
		Params(func() {
			Param("id", Integer, "Account ID")
		})
		Response(NoContent)
		Response(NotFound)
	})
})

var _ = Resource("bottle", func() {

	MediaType(BottleMediaType)
	BasePath("bottles")
	Parent("accounts")
	CanonicalAction("show")
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
			MediaType(CollectionOf(BottleMediaType))
		})
	})

	Action("show", func() {
		Routing(
			GET("/:id"),
		)
		Description("Retrieve bottle with given id")
		Params(func() {
			Param("id", Integer, "Bottle ID")
		})
		Response(OK, func() {
			MediaType(BottleMediaType)
		})
		Response(NotFound)
	})

	Action("create", func() {
		Routing(
			POST(""),
		)
		Description("Record new bottle")
		Payload(func() {
			Member("name")
			Member("vintage")
			Member("vineyard")
			Member("varietal")
			Member("color")
			Member("sweet")
			Member("country")
			Member("region")
			Member("review")
			Member("characteristics")

			Required("name", "vintage", "vineyard")
		})
		Response(Created)
	})

	Action("update", func() {
		Routing(
			PATCH("/:id"),
		)
		Params(func() {
			Param("id", Integer, "Bottle ID")
		})
		Payload(func() {
			Member("name")
			Member("vineyard")
			Member("varietal")
			Member("vintage")
			Member("color")
			Member("sweet")
			Member("country")
			Member("region")
			Member("review")
			Member("characteristics")
		})
		Response(NoContent)
		Response(NotFound)
	})

	Action("rate", func() {
		Routing(
			PUT("/:id/actions/rate"),
		)
		Params(func() {
			Param("id", Integer, "Bottle ID")
		})
		Payload(func() {
			Member("rating")
		})
		Response(NoContent)
		Response(NotFound)
	})

	Action("delete", func() {
		Routing(
			DELETE("/:id"),
		)
		Params(func() {
			Param("id", Integer, "Bottle ID")
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
