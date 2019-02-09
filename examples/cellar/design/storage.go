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
		GRPC(func() {
			Response(CodeOK)
		})
	})

	Method("show", func() {
		Description("Show bottle by ID")
		Payload(func() {
			Field(1, "id", String, "ID of bottle to show")
			Field(2, "view", String, "View to render", func() {
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
		GRPC(func() {
			Metadata(func() {
				Attribute("view")
			})
			Response(CodeOK)
			Response("not_found", CodeNotFound)
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
		GRPC(func() {
			Response(CodeOK)
		})
	})

	Method("remove", func() {
		Description("Remove bottle from storage")
		Payload(func() {
			Field(1, "id", String, "ID of bottle to remove")
			Required("id")
		})
		Error("not_found", NotFound, "Bottle not found")
		HTTP(func() {
			DELETE("/{id}")
			Response(StatusNoContent)
		})
		GRPC(func() {
			Response(CodeOK)
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
		GRPC(func() {
			Response(CodeOK)
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
		GRPC(func() {
			Response(CodeOK)
		})
	})

	Method("multi_update", func() {
		Description("Update bottles with the given IDs. This is a multipart request and each part has field name 'bottle' and contains the encoded bottle info to be updated. The IDs in the query parameter is mapped to each part in the request.")
		Payload(func() {
			Field(1, "ids", ArrayOf(String), "IDs of the bottles to be updated")
			Field(2, "bottles", ArrayOf(Bottle), "Array of bottle info that matches the ids attribute")
			Required("ids", "bottles")
		})
		HTTP(func() {
			PUT("/multi_update")
			Param("ids")
			MultipartRequest()
			Response(StatusNoContent)
		})
		GRPC(func() {
			Response(CodeOK)
		})
	})
})
