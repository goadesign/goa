package design

import . "goa.design/goa/dsl"

var _ = Service("storage", func() {
	Description("The storage service makes it possible to view, add or remove wine bottles.")

	HTTP(func() {
		Path("/storage")
	})

	Method("list", func() {
		Description("List all stored bottles")
		Result(CollectionOf(StoredBottle), func() {
			View("tiny")
		})
		HTTP(func() {
			GET("/")
			Response(StatusOK)
		})
	})

	Method("show", func() {
		Description("Show bottle by ID")
		Payload(func() {
			Attribute("id", String, "ID of bottle to show")
			Attribute("view", String, "View to render", func() {
				Enum("default", "tiny")
			})
			Required("id")
		})
		Result(StoredBottle)
		Error("not_found", NotFound, "Bottle not found")
		HTTP(func() {
			GET("/{id}")
			Param("view")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})

	Method("add", func() {
		Description("Add new bottle and return its ID.")
		Payload(Bottle)
		Result(String)
		HTTP(func() {
			POST("/")
			Response(StatusCreated)
		})
	})

	Method("remove", func() {
		Description("Remove bottle from storage")
		Payload(func() {
			Attribute("id", String, "ID of bottle to remove")
			Required("id")
		})
		Error("not_found", NotFound, "Bottle not found")
		HTTP(func() {
			DELETE("/{id}")
			Response(StatusNoContent)
		})
	})

	Method("rate", func() {
		Description("Rate bottles by IDs")
		Payload(MapOf(UInt32, ArrayOf(String)))
		HTTP(func() {
			POST("/rate")
			MapParams()
			Response(StatusOK)
		})
	})

	Method("multi_add", func() {
		Description("Add n number of bottles and return their IDs. This is a multipart request and each part has field name 'bottle' and contains the encoded bottle info to be added.")
		Payload(ArrayOf(Bottle))
		Result(ArrayOf(String))
		HTTP(func() {
			POST("/multi_add")
			MultipartRequest()
		})
	})

	Method("multi_update", func() {
		Description("Update bottles with the given IDs. This is a multipart request and each part has field name 'bottle' and contains the encoded bottle info to be updated. The IDs in the query parameter is mapped to each part in the request.")
		Payload(func() {
			Attribute("ids", ArrayOf(String), "IDs of the bottles to be updated")
			Attribute("bottles", ArrayOf(Bottle), "Array of bottle info that matches the ids attribute")
			Required("ids", "bottles")
		})
		HTTP(func() {
			PUT("/multi_update")
			Param("ids")
			MultipartRequest()
			Response(StatusNoContent)
		})
	})
})
