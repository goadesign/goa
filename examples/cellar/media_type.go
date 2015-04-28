package design

import . "github.com/raphael/goa/design"

var BottleMediaType = MediaType("application/vnd.goa.example.bottle", func() {
	Description("A Droplet is a DigitalOcean virtual machine.")
	Attributes(
		Attribute("id", Integer, func() {
			Description("ID of bottle")
		}),
		Attribute("href", String, func() {
			Description("API href of bottle")
		}),
		Attribute("name", String, func() {
			Description("Name of wine")
		}),
		Attribute("vineyard", String, func() {
			Description("Name of vineyard / winery")
		}),
		Attribute("varietal", String, func() {
			Description("Wine varietal")
		}),
		Attribute("vintage", Integer, func() {
			Description("Wine vintage")
		}),
		Attribute("color", String, func() {
			Description("Type of wine")
			Enum("red", "white", "rose", "yellow")
		}),
		Attribute("sweet", Bool, func() {
			Description("Whether wine is sweet or dry")
		}),
		Attribute("country", String, func() {
			Description("Country of origin")
		}),
		Attribute("region", String, func() {
			Description("Region")
		}),
		Attribute("review", String, func() {
			Description("Review")
		}),
		Attribute("characteristics", CollectionOf(String), func() {
			Description("Wine characteristics")
		}),
	)
})
