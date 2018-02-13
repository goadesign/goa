package testdata

import (
	. "goa.design/goa/http/dsl"
)

var ValidRouteDSL = func() {
	Service("ValidRoute", func() {
		HTTP(func() {
			Path("/{base_id}")
		})
		Method("Method", func() {
			HTTP(func() {
				POST("/{id}")
			})
		})
	})
}

var DuplicateWCRouteDSL = func() {
	Service("InvalidRoute", func() {
		HTTP(func() {
			Path("/{id}")
		})
		Method("Method", func() {
			HTTP(func() {
				POST("/{id}")
			})
		})
	})
}
