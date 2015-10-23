package design

import (
	. "github.com/raphael/goa/design"
	. "github.com/raphael/goa/design/dsl"
)

// Account is the account resource media type.
var Account = MediaType("application/vnd.goa.example.account", func() {
	Description("A tenant account")
	Attributes(func() {
		Attribute("id", Integer, "ID of account")
		Attribute("href", String, "API href of account")
		Attribute("name", String, "Name of account")
		Attribute("created_at", String, "Date of creation", func() {
			Format("date-time")
		})
		Attribute("created_by", String, "Email of account ownder", func() {
			Format("email")
		})

		Required("name")
	})

	View("default", func() {
		Attribute("id")
		Attribute("href")
		Attribute("name")
	})

	View("full", func() {
		Attribute("id")
		Attribute("href")
		Attribute("name")
		Attribute("created_at")
		Attribute("created_by")
	})

	View("link", func() {
		Attribute("href")
		Attribute("name")
	})
})

// Bottle is the bottle resource media type.
var Bottle = MediaType("application/vnd.goa.example.bottle", func() {
	Description("A bottle of wine")
	Reference(BottlePayload)
	Attributes(func() {
		Attribute("id", Integer, "ID of bottle")
		Attribute("href", String, "API href of bottle")
		Attribute("rating", Integer, "Rating of bottle between 1 and 5", func() {
			Minimum(1)
			Maximum(5)
		})
		Attribute("account", Account, "Account that owns bottle")
		Attribute("created_at", String, "Date of creation", func() {
			Format("date-time")
		})
		Attribute("updated_at", String, "Date of last update", func() {
			Format("date-time")
		})
		// Attributes below inherit from the base type
		Attribute("name")
		Attribute("vineyard")
		Attribute("varietal")
		Attribute("vintage")
		Attribute("color")
		Attribute("sweetness")
		Attribute("country")
		Attribute("region")
		Attribute("review")
		Attribute("characteristics")

		Links(func() {
			Link("account")
		})

		View("default", func() {
			Attribute("id")
			Attribute("href")
			Attribute("name")
			Attribute("vineyard")
			Attribute("varietal")
			Attribute("vintage")
			Attribute("links")
		})

		View("tiny", func() {
			Attribute("id")
			Attribute("href")
			Attribute("name")
			Attribute("links")
		})

		View("full", func() {
			Attribute("id")
			Attribute("href")
			Attribute("name")
			Attribute("vineyard")
			Attribute("varietal")
			Attribute("vintage")
			Attribute("color")
			Attribute("sweetness")
			Attribute("country")
			Attribute("region")
			Attribute("review")
			Attribute("characteristics")
			Attribute("created_at")
			Attribute("updated_at")
			Attribute("account", func() {
				View("full")
			})
		})

		Required("account", "name", "vineyard")
	})
})
