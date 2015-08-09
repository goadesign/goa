package main

import . "github.com/raphael/goa/design/dsl"

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
		Attribute("sweet", Bool, "Whether wine is sweet or dry")
		Attribute("country", String, "Country of origin")
		Attribute("region", String, "Region")
		Attribute("review", String, "Review")
		Attribute("characteristics", CollectionOf(String), "Wine characteristics")

		Required("name", "vineyard")
	})
})
