package testdata

import (
	. "goa.design/goa/v3/dsl"
)

// MixedJSONRPCTransportsAPI defines an API with mixed JSON-RPC transports.
var MixedJSONRPCTransportsAPI = func() {
	API("MixedTransports", func() {
		Title("Mixed JSON-RPC Transports API")
		Description("API demonstrating mixed HTTP and SSE JSON-RPC transports")
	})
	
	Service("MixedService", func() {
		Description("Service with both HTTP and SSE JSON-RPC methods")
		
		// Regular HTTP method
		Method("GetUser", func() {
			Payload(func() {
				ID("id", String, "User ID")
				Required("id")
			})
			Result(func() {
				ID("id", String, "User ID")
				Field(1, "name", String)
				Field(2, "email", String)
				Required("id")
			})
			HTTP(func() {
				POST("/users/{id}")
			})
			JSONRPC(func() {
			})
		})
		
		// SSE streaming method  
		Method("WatchUsers", func() {
			Payload(func() {
				ID("request_id", String, "Request ID")
				Field(1, "filter", String, "Filter expression")
				Required("request_id")
			})
			StreamingResult(func() {
				Field(1, "user_id", String)
				Field(2, "event", String)
				Field(3, "timestamp", String)
			})
			HTTP(func() {
				POST("/users/watch")
				ServerSentEvents() // Enable SSE for this method
			})
			JSONRPC(func() {
			})
		})
		
		// Another regular HTTP method
		Method("CreateUser", func() {
			Payload(func() {
				Field(1, "name", String)
				Field(2, "email", String)
				Required("name", "email")
			})
			Result(func() {
				Field(1, "id", String, "Created user ID")
			})
			HTTP(func() {
				POST("/users")
			})
			JSONRPC(func() {
				// Notification - no ID needed
			})
		})
		
		// Configure JSON-RPC endpoint
		JSONRPC(func() {
			Path("/api/rpc")
		})
	})
}

// ValidWebSocketOnlyAPI shows WebSocket cannot mix with other transports.
var ValidWebSocketOnlyAPI = func() {
	API("WebSocketOnly", func() {
		Title("WebSocket Only API")
	})
	
	Service("WebSocketService", func() {
		Description("Service with only WebSocket JSON-RPC methods")
		
		Method("Connect", func() {
			Payload(func() {
				ID("token", String, "Request token used as ID")
				Required("token")
			})
			StreamingPayload(func() {
				Field(1, "message", String)
			})
			StreamingResult(func() {
				Field(1, "response", String)
			})
			HTTP(func() {
				GET("/ws")
			})
			JSONRPC(func() {
			})
		})
		
		JSONRPC(func() {
			Path("/ws")
		})
	})
}

// InvalidMixedWebSocketAPI shows invalid mixing of WebSocket with other transports.
var InvalidMixedWebSocketAPI = func() {
	API("InvalidMixed", func() {
		Title("Invalid Mixed API")
	})
	
	Service("InvalidService", func() {
		Description("Service incorrectly mixing WebSocket with HTTP")
		
		// WebSocket method
		Method("Stream", func() {
			StreamingPayload(String)
			StreamingResult(String)
			HTTP(func() {
				GET("/stream")
			})
			JSONRPC(func() {
				// Streaming methods typically don't use ID
			})
		})
		
		// Regular HTTP method - THIS SHOULD CAUSE VALIDATION ERROR
		Method("Get", func() {
			Payload(String)
			Result(String)
			HTTP(func() {
				POST("/get")
			})
			JSONRPC(func() {
				// This method mixes with WebSocket - should error
			})
		})
		
		JSONRPC(func() {
			Path("/invalid")
		})
	})
}