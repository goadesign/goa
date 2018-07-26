package design

import . "goa.design/goa/http/design"
import . "goa.design/goa/http/dsl"

var _ = API("cars", func() {
	Title("Cars Service")
	Description("HTTP service to lookup car models by body style.")
	Server("http://localhost:8080")
	Server("ws://localhost:8080")
})

// BasicAuth defines a security scheme that uses basic authentication.
var BasicAuth = BasicAuthSecurity("basic", func() {
	Description("Secures the login endpoint.")
})

// JWTAuth defines a security scheme that uses JWT tokens.
var JWTAuth = JWTSecurity("jwt", func() {
	Description(`Secures endpoint by requiring a valid JWT token. Supports scopes "stream:read" and "stream:write".`)
	Scope("stream:read", "Read-only access")
	Scope("stream:write", "Read and write access")
})

// StoredCar is the car result type.
var StoredCar = ResultType("application/vnd.goa.car", func() {
	TypeName("StoredCar")
	Reference(Car)
	Attributes(func() {
		Attribute("make", String, "The make of the car")
		Attribute("model", String, "The car model")
		Attribute("body_style", String, "The car body style")
	})
})

// Car is the car type.
var Car = Type("car", func() {
	Attribute("make", String, "The make of the car")
	Attribute("model", String, "The car model")
	Attribute("body_style", String, "The car body style")
	Required("make", "model", "body_style")
})

var _ = Service("cars", func() {
	HTTP(func() {
		Path("/cars")
	})

	Description("The cars service lists car models by body style.")

	Method("login", func() {
		Description("Creates a valid JWT")

		Security(BasicAuth)

		Payload(func() {
			Description("Credentials used to authenticate to retrieve JWT token")
			Username("user", String, func() {
				Example("username")
			})
			Password("password", String, func() {
				Example("password")
			})
			Required("user", "password")
		})

		Result(String, func() {
			Description("New JWT token")
		})

		Error("unauthorized", String, "Credentials are invalid")

		HTTP(func() {
			POST("/login")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
		})
	})

	Method("list", func() {
		Description("Lists car models by body style.")

		Security(JWTAuth, func() {
			Scope("stream:read")
		})

		Payload(func() {
			Attribute("style", String, "The car body style.", func() {
				Enum("sedan", "hatchback")
			})
			Token("token", String, func() {
				Description("JWT used for authentication")
			})
			Required("style", "token")
		})

		StreamingResult(StoredCar)

		Error("unauthorized", String)
		Error("invalid-scopes", String)

		HTTP(func() {
			GET("")
			Param("style")
			Header("token:Authorization")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("invalid-scopes", StatusForbidden)
		})
	})

	Method("add", func() {
		Description("Add car models.")

		Security(JWTAuth, func() {
			Scope("stream:write")
		})

		Payload(func() {
			Token("token", String, func() {
				Description("JWT used for authentication")
			})
		})

		StreamingPayload(func() {
			Attribute("car", Car, "Car to add.")
		})

		Result(CollectionOf(StoredCar))

		Error("unauthorized", String)
		Error("invalid-scopes", String)

		HTTP(func() {
			GET("/add")
			Header("token:Authorization")
			Response(StatusCreated)
			Response("unauthorized", StatusUnauthorized)
			Response("invalid-scopes", StatusForbidden)
		})
	})

	Method("update", func() {
		Description("Update car models.")

		Security(JWTAuth, func() {
			Scope("stream:write")
		})

		Payload(func() {
			Token("token", String, func() {
				Description("JWT used for authentication")
			})
		})

		StreamingPayload(ArrayOf(Car))

		StreamingResult(CollectionOf(StoredCar))

		Error("unauthorized", String)
		Error("invalid-scopes", String)

		HTTP(func() {
			GET("/update")
			Header("token:Authorization")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("invalid-scopes", StatusForbidden)
		})
	})
})
