package design

import (
	. "github.com/raphael/goa/design"
	. "github.com/raphael/goa/design/dsl"
)

// AccountMediaType is the account resource media type.
var AccountMediaType = MediaType("application/vnd.goa.example.account", func() {
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

// BottleMediaType is the bottle resource media type.
var BottleMediaType = MediaType("application/vnd.goa.example.bottle", func() {
	Description("A bottle of wine")
	Attributes(func() {
		Attribute("id", Integer, "ID of bottle")
		Attribute("href", String, "API href of bottle")
		Attribute("name", String, "Name of wine")
		Attribute("account", AccountMediaType, "Owner account")
		Attribute("vineyard", String, "Name of vineyard / winery")
		Attribute("varietal", String, "Wine varietal")
		Attribute("vintage", Integer, "Wine vintage", func() {
			Minimum(1900)
			Maximum(2020)
		})
		Attribute("color", String, "Type o9f wine", func() {
			Enum("red", "white", "rose", "yellow")
		})
		Attribute("sweet", Boolean, "Whether wine is sweet or dry")
		Attribute("country", String, "Country of origin")
		Attribute("region", String, "Region")
		Attribute("review", String, "Review", func() {
			MinLength(10)
		})
		Attribute("characteristics", ArrayOf(String), "Wine characteristics")
		Attribute("created_at", String, "Date of creation", func() {
			Format("date-time")
		})
		Attribute("updated_at", String, "Date of last update", func() {
			Format("date-time")
		})

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

		View("full", func() {
			Attribute("id")
			Attribute("href")
			Attribute("name")
			Attribute("vineyard")
			Attribute("varietal")
			Attribute("vintage")
			Attribute("color")
			Attribute("sweet")
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
