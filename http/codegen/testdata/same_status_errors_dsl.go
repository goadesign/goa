// This file defines an HTTP design whose endpoint maps several errors to the
// same status code so OpenAPI tests can check that every error stays visible.
package testdata

import . "goa.design/goa/v3/dsl"

// SameStatusErrorsDSL declares one endpoint with three errors sharing status
// 409 and two errors sharing status 404. The 409 errors mix two custom types
// using the default JSON content type with Goa's standard error result, which
// uses its own content type; only one of them sets the X-Retry-After header.
// The 404 errors both use the standard error result so their bodies share one
// schema.
var SameStatusErrorsDSL = func() {
	buildOpen := Type("BuildOpen", func() {
		Description("BuildOpen reports the build that blocks a new one.")
		ErrorName("name", String, "Error name")
		Attribute("message", String, "Refusal reason", func() {
			Example("another build is open")
		})
		Attribute("build_id", String, "Open build", func() {
			Example("b-17")
		})
		Required("name", "message", "build_id")
	})
	quotaExceeded := Type("QuotaExceeded", func() {
		Description("QuotaExceeded reports the build quota that was reached.")
		ErrorName("name", String, "Error name")
		Attribute("limit", Int, "Maximum number of builds", func() {
			Example(3)
		})
		Attribute("retry_after", Int, "Seconds until a build slot frees up", func() {
			Example(60)
		})
		Required("name", "limit")
	})

	Service("builds", func() {
		Method("start", func() {
			Result(String, func() {
				Example("b-18")
			})
			Error("build_open", buildOpen, "Another build is open")
			Error("conflict", ErrorResult, "Request conflicts with the current state")
			Error("quota_exceeded", quotaExceeded, "Build quota reached")
			Error("not_found", ErrorResult, "Project not found")
			Error("gone", ErrorResult, "Project deleted")
			HTTP(func() {
				POST("/builds")
				Response(StatusCreated)
				Response("build_open", StatusConflict)
				Response("conflict", StatusConflict)
				Response("quota_exceeded", StatusConflict, func() {
					Header("retry_after:X-Retry-After")
				})
				Response("not_found", StatusNotFound)
				Response("gone", StatusNotFound)
			})
		})
	})
}
