package design

import . "goa.design/goa/dsl"

var _ = Service("sommelier", func() {
	Description("The sommelier service retrieves bottles given a set of criteria.")
	HTTP(func() {
		Path("/sommelier")
	})
	Method("pick", func() {
		Payload(Criteria)
		Result(CollectionOf(StoredBottle), func() {
			View("default")
		})
		Error("no_criteria", String, "Missing criteria")
		Error("no_match", String, "No bottle matched given criteria")
		HTTP(func() {
			POST("/")
			Response(StatusOK)
			Response("no_criteria", StatusBadRequest)
			Response("no_match", StatusNotFound)
		})
	})
})
