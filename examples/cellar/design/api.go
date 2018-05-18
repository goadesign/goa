package design

import . "goa.design/goa/http/design"
import . "goa.design/goa/http/dsl"

var _ = API("cellar", func() {
	Title("Cellar Service")
	Description("HTTP service for managing your wine cellar")
	Server("http://localhost:8080")
})

var StoredBottle = ResultType("application/vnd.cellar.stored-bottle", func() {
	Description("A StoredBottle describes a bottle retrieved by the storage service.")
	Reference(Bottle)
	TypeName("StoredBottle")

	Attributes(func() {
		Attribute("id", String, "ID is the unique id of the bottle.", func() {
			Example("123abc")
		})
		Attribute("name")
		Attribute("winery")
		Attribute("vintage")
		Attribute("composition")
		Attribute("description")
		Attribute("rating")
	})

	View("default", func() {
		Attribute("id")
		Attribute("name")
		Attribute("winery", func() {
			View("tiny")
		})
		Attribute("vintage")
		Attribute("composition")
		Attribute("description")
		Attribute("rating")
	})

	View("tiny", func() {
		Attribute("id")
		Attribute("name")
		Attribute("winery", func() {
			View("tiny")
		})
	})

	Required("id", "name", "winery", "vintage")
})

var Bottle = Type("Bottle", func() {
	Description("Bottle describes a bottle of wine to be stored.")
	Attribute("name", String, "Name of bottle", func() {
		MaxLength(100)
		Example("Blue's Cuvee")
	})
	Attribute("winery", Winery, "Winery that produces wine")
	Attribute("vintage", UInt32, "Vintage of bottle", func() {
		Minimum(1900)
		Maximum(2020)
	})
	Attribute("composition", ArrayOf(Component), "Composition is the list of grape varietals and associated percentage.")
	Attribute("description", String, "Description of bottle", func() {
		MaxLength(2000)
		Example("Red wine blend with an emphasis on the Cabernet Franc grape and including other Bordeaux grape varietals and some Syrah")
	})
	Attribute("rating", UInt32, "Rating of bottle from 1 (worst) to 5 (best)", func() {
		Minimum(1)
		Maximum(5)
	})
	Required("name", "winery", "vintage")
})

var Winery = ResultType("Winery", func() {
	Attributes(func() {
		Attribute("name", String, "Name of winery", func() {
			Example("Longoria")
		})
		Attribute("region", String, "Region of winery", func() {
			Pattern(`(?i)[a-z '\.]+`)
			Example("Central Coast, California")
		})
		Attribute("country", String, "Country of winery", func() {
			Pattern(`(?i)[a-z '\.]+`)
			Example("USA")
		})
		Attribute("url", String, "Winery website URL", func() {
			Pattern(`(?i)^(https?|ftp)://[^\s/$.?#].[^\s]*$`)
			Example("http://www.longoriawine.com/")
		})
	})
	View("default", func() {
		Attribute("name")
		Attribute("region")
		Attribute("country")
		Attribute("url")
	})
	View("tiny", func() {
		Attribute("name")
	})
	Required("name", "region", "country")
})

var Component = Type("Component", func() {
	Attribute("varietal", String, "Grape varietal", func() {
		Pattern(`[A-Za-z' ]+`)
		MaxLength(100)
		Example("Syrah")
	})
	Attribute("percentage", UInt32, "Percentage of varietal in wine", func() {
		Minimum(1)
		Maximum(100)
	})
	Required("varietal")
})

var NotFound = Type("NotFound", func() {
	Description("NotFound is the type returned when attempting to show or delete a bottle that does not exist.")
	Attribute("message", String, "Message of error", func() {
		Metadata("struct:error:name")
		Example("bottle 1 not found")
	})
	Attribute("id", String, "ID of missing bottle")
	Required("message", "id")
})

var Criteria = Type("Criteria", func() {
	Description("Criteria described a set of criteria used to pick a bottle. All criteria are optional, at least one must be provided.")
	Attribute("name", String, "Name of bottle to pick", func() {
		Example("Blue's Cuvee")
	})
	Attribute("varietal", ArrayOf(String), "Varietals in preference order", func() {
		Example([]string{"pinot noir", "merlot", "cabernet franc"})
	})
	Attribute("winery", String, "Winery of bottle to pick", func() {
		Example("longoria")
	})
})
