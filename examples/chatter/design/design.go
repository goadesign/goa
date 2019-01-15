package design

import (
	. "goa.design/goa/dsl"
)

var _ = API("chatter", func() {
	Title("Chatter service describing the streaming features of goa v2.")
})

var _ = Service("chatter", func() {
	Description("The chatter service implements a simple client and server chat.")

	Method("login", func() {
		Description("Creates a valid JWT token for auth to chat.")

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

		GRPC(func() {
			Response(CodeOK)
			Response("unauthorized", CodeUnauthenticated)
		})
	})

	Method("echoer", func() { // bidirectional streaming example
		Description("Echoes the message sent by the client.")

		Security(JWTAuth, func() {
			Scope("stream:write")
		})

		Payload(func() {
			Token("token", String, func() {
				Description("JWT used for authentication")
			})
			Required("token")
		})

		StreamingPayload(String)

		StreamingResult(String)

		Error("unauthorized", String)
		Error("invalid-scopes", String)

		HTTP(func() {
			GET("/echoer")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("invalid-scopes", StatusForbidden)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("unauthorized", CodeUnauthenticated)
			Response("invalid-scopes", CodeUnauthenticated)
		})
	})

	Method("listener", func() { // streaming payload example (server doesn't respond)
		Description("Listens to the messages sent by the client.")

		Security(JWTAuth, func() {
			Scope("stream:write")
		})

		Payload(func() {
			Token("token", String, func() {
				Description("JWT used for authentication")
			})
			Required("token")
		})

		StreamingPayload(String)

		Error("unauthorized", String)
		Error("invalid-scopes", String)

		HTTP(func() {
			GET("/listener")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("invalid-scopes", StatusForbidden)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("unauthorized", CodeUnauthenticated)
			Response("invalid-scopes", CodeUnauthenticated)
		})
	})

	Method("summary", func() { // streaming payload example (server responds)
		Description("Summarizes the chat messages sent by the client.")

		Security(JWTAuth, func() {
			Scope("stream:write")
		})

		Payload(func() {
			Token("token", String, func() {
				Description("JWT used for authentication")
			})
			Required("token")
		})

		StreamingPayload(String)

		Result(CollectionOf(ChatSummary), func() {
			View("default")
		})

		Error("unauthorized", String)
		Error("invalid-scopes", String)

		HTTP(func() {
			GET("/summary")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("invalid-scopes", StatusForbidden)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("unauthorized", CodeUnauthenticated)
			Response("invalid-scopes", CodeUnauthenticated)
		})
	})

	Method("history", func() { // streaming result example
		Description("Returns the chat messages sent to the server.")

		Security(JWTAuth, func() {
			Scope("stream:read")
		})

		Payload(func() {
			Token("token", String, func() {
				Description("JWT used for authentication")
			})
			Attribute("view", String, "View to use to render the result")
			Required("token")
		})

		StreamingResult(ChatSummary)

		Error("unauthorized", String)
		Error("invalid-scopes", String)

		HTTP(func() {
			GET("/history")
			Param("view")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("invalid-scopes", StatusForbidden)
		})

		GRPC(func() {
			Metadata(func() {
				Attribute("view")
			})
			Response(CodeOK)
			Response("unauthorized", CodeUnauthenticated)
			Response("invalid-scopes", CodeUnauthenticated)
		})
	})
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

var ChatSummary = ResultType("application/vnd.goa.summary", func() {
	TypeName("ChatSummary")
	Attributes(func() {
		Field(1, "message", String, "Message sent to the server")
		Field(2, "length", Int, "Length of the message sent")
		Field(3, "sent_at", String, "Time at which the message was sent", func() {
			Format(FormatDateTime)
		})
		Required("message")
	})
	View("tiny", func() {
		Attribute("message")
	})
})
