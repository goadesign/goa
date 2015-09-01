package main

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

		Required("name")
	})
})

// BottleMediaType is the bottle resource media type.
var BottleMediaType = MediaType("application/vnd.goa.example.bottle", func() {
	Description("A bottle of wine")
	Attributes(func() {
		Attribute("id", Integer, "ID of bottle")
		Attribute("href", String, "API href of bottle")
		Attribute("name", String, "Name of wine")
		Attribute("vineyard", String, "Name of vineyard / winery")
		Attribute("varietal", String, "Wine varietal")
		Attribute("vintage", Integer, "Wine vintage")
		Attribute("color", String, "Type of wine", func() {
			Enum("red", "white", "rose", "yellow")
		})
		Attribute("sweet", Boolean, "Whether wine is sweet or dry")
		Attribute("country", String, "Country of origin")
		Attribute("region", String, "Region")
		Attribute("review", String, "Review")
		Attribute("characteristics", ArrayOf(String), "Wine characteristics")
		Attribute("created_at", String, "Date of creation", func() {
			Format("date-time")
		})
		Attribute("updated_at", String, "Date of last update", func() {
			Format("date-time")
		})

		Required("name", "vineyard")
	})
})
