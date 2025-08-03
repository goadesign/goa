package scenarios

import (
	"fmt"
	"strings"
)

// GenerateDSLCode generates DSL code for a specific payload and result type combination
func GenerateDSLCode(payloadType, resultType DataType) string {
	var dsl strings.Builder

	// Add user type definitions if needed - use variable assignment like gRPC tests
	if payloadType == DataTypeUserType || resultType == DataTypeUserType {
		dsl.WriteString(`	var UserType = Type("UserType", func() {
		Attribute("id", String)
		Attribute("name", String)
		Attribute("email", String, func() {
			Format(FormatEmail)
		})
		Attribute("age", Int, func() {
			Minimum(0)
			Maximum(150)
		})
		Required("id", "name")
	})

`)
	}

	// Service definition
	dsl.WriteString(`	Service("test", func() {
		JSONRPC(func() {
			POST("/jsonrpc")
		})
		Method("call", func() {
`)

	// Payload definition
	if payloadType != DataTypeNone {
		// For array payloads, wrap in an object to avoid CLI generation issues
		if payloadType == DataTypeArray {
			dsl.WriteString(`			Payload(func() {
				Attribute("items", ArrayOf(String))
				Required("items")
			})
`)
		} else {
			dsl.WriteString(fmt.Sprintf(`			Payload(%s)
`, generateTypeExpression(payloadType)))
		}
	}

	// Result definition
	dsl.WriteString(fmt.Sprintf(`			Result(%s)
`, generateTypeExpression(resultType)))

	// JSON-RPC endpoint
	dsl.WriteString(`			JSONRPC(func() {
			})
		})
	})`)

	return dsl.String()
}

// generateTypeExpression generates the DSL type expression for a data type
func generateTypeExpression(dataType DataType) string {
	switch dataType {
	case DataTypePrimitive:
		return "String"

	case DataTypeArray:
		return "ArrayOf(String)"

	case DataTypeObject:
		return `func() {
				Attribute("field1", String)
				Attribute("field2", Int)
				Attribute("field3", Boolean)
				Required("field1")
			}`

	case DataTypeMap:
		return "MapOf(String, Any)"

	case DataTypeUserType:
		// When referencing a defined type, use the variable name
		return "UserType"

	case DataTypeComplex:
		// Return the complex structure with metadata as a map
		return `func() {
				Attribute("sequence", Int)
				Attribute("data", MapOf(String, Any))
				Attribute("metadata", MapOf(String, Any))
				Required("sequence")
			}`

	default:
		return "String"
	}
}

// generateStreamingResultExpression generates DSL type expressions for streaming results
// JSON-RPC streaming requires all results to be objects, so primitive types are wrapped
func generateStreamingResultExpression(dataType DataType) string {
	switch dataType {
	case DataTypePrimitive:
		// Wrap primitive in an object for JSON-RPC streaming compliance
		return `func() {
				Attribute("value", String, "The streamed value")
				Required("value")
			}`

	case DataTypeArray:
		// Wrap array in an object for JSON-RPC streaming compliance  
		return `func() {
				Attribute("items", ArrayOf(String), "The streamed array")
				Required("items")
			}`

	case DataTypeObject:
		// Already an object, use as-is
		return generateTypeExpression(dataType)

	case DataTypeMap:
		// Wrap map in an object for JSON-RPC streaming compliance
		return `func() {
				Attribute("data", MapOf(String, Any), "The streamed map")
				Required("data")
			}`

	case DataTypeUserType:
		// UserType should already be an object
		return "UserType"

	case DataTypeComplex:
		// Already an object, use as-is
		return generateTypeExpression(dataType)

	default:
		// Fallback: wrap in object
		return `func() {
				Attribute("value", String, "The streamed value")
				Required("value")
			}`
	}
}

// GenerateNotificationDSL generates DSL for notification scenarios
func GenerateNotificationDSL(payloadType DataType) string {
	var dsl strings.Builder

	dsl.WriteString(`	API("test", func() {
		Title("Notification Test API")
		Version("1.0")
	})

`)

	// Add user type definition if needed - use variable assignment
	if payloadType == DataTypeUserType {
		dsl.WriteString(`	var UserType = Type("UserType", func() {
		Attribute("id", String)
		Attribute("name", String)
		Attribute("email", String, func() {
			Format(FormatEmail)
		})
		Attribute("age", Int, func() {
			Minimum(0)
			Maximum(150)
		})
		Required("id", "name")
	})

`)
	}

	dsl.WriteString(`	Service("notifier", func() {
		JSONRPC(func() {
			POST("/jsonrpc")
		})
		Method("notify", func() {
`)

	// Payload definition
	dsl.WriteString(fmt.Sprintf(`			Payload(%s)
`, generateTypeExpression(payloadType)))

	// No Result for notifications
	dsl.WriteString(`			
			JSONRPC(func() {
			})
		})
	})`)

	return dsl.String()
}

// GenerateWebSocketDSL generates DSL for WebSocket streaming scenarios
func GenerateWebSocketDSL(payloadType, resultType DataType, streaming StreamingType) string {
	var dsl strings.Builder

	dsl.WriteString(`	API("test", func() {
		Title("WebSocket Test API")
		Version("1.0")
	})

`)

	// Add user type definitions if needed
	if payloadType == DataTypeUserType || resultType == DataTypeUserType {
		dsl.WriteString(`	Type("UserType", func() {
		Attribute("id", String)
		Attribute("name", String)
		Attribute("email", String, func() {
			Format(FormatEmail)
		})
		Attribute("age", Int, func() {
			Minimum(0)
			Maximum(150)
		})
		Required("id", "name")
	})

`)
	}

	// Determine method name based on streaming type
	var methodName string
	switch streaming {
	case StreamingServer:
		methodName = "server_stream"
	case StreamingClient:
		methodName = "client_stream"
	case StreamingBidirectional:
		methodName = "bidirectional_stream"
	}

	dsl.WriteString(fmt.Sprintf(`	Service("streaming", func() {
		JSONRPC(func() {
			GET("/jsonrpc/ws")
		})
		Method("%s", func() {
`, methodName))

	// Streaming configuration
	switch streaming {
	case StreamingServer:
		// Server streaming: non-streaming payload, streaming results
		dsl.WriteString(`			Payload(func() {
				Attribute("id", String, func() {
					Meta("jsonrpc:id")
				})
				Attribute("count", Int, "Number of messages to stream")
				Required("id", "count")
			})
`)
		dsl.WriteString(fmt.Sprintf("\t\t\tStreamingResult(%s)\n", generateJSONRPCStreamingTypeExpression(resultType)))

	case StreamingClient:
		dsl.WriteString(fmt.Sprintf("\t\t\tStreamingPayload(%s)\n", generateJSONRPCStreamingTypeExpression(payloadType)))
		if resultType != DataTypeNone {
			dsl.WriteString(fmt.Sprintf("\t\t\tResult(%s)\n", generateJSONRPCStreamingTypeExpression(resultType)))
		}

	case StreamingBidirectional:
		dsl.WriteString(fmt.Sprintf("\t\t\tStreamingPayload(%s)\n\t\t\tStreamingResult(%s)\n",
			generateJSONRPCStreamingTypeExpression(payloadType), generateJSONRPCStreamingTypeExpression(resultType)))
	}

	// JSON-RPC method endpoint
	dsl.WriteString("\t\t\t\n\t\t\tJSONRPC(func() {\n\t\t\t})\n\t\t})\n\t})")

	return dsl.String()
}

// generateJSONRPCStreamingTypeExpression generates proper JSON-RPC streaming type expressions
// that include ID attributes for request tracking. JSON-RPC streaming payloads and results
// must be objects with ID attributes for protocol compliance.
func generateJSONRPCStreamingTypeExpression(dataType DataType) string {
	switch dataType {
	case DataTypePrimitive:
		return `func() {
				ID("id", String, "Request ID")
				Attribute("data", String, "Data")
				Required("id", "data")
			}`

	case DataTypeArray:
		return `func() {
				ID("id", String, "Request ID")
				Attribute("items", ArrayOf(String), "Array items")
				Required("id", "items")
			}`

	case DataTypeObject:
		return `func() {
				ID("id", String, "Request ID")
				Attribute("field1", String, "Field 1")
				Attribute("field2", Int, "Field 2")
				Attribute("field3", Boolean, "Field 3")
				Required("id", "field1")
			}`

	case DataTypeMap:
		return `func() {
				ID("id", String, "Request ID")
				Attribute("data", MapOf(String, Any), "Map data")
				Required("id", "data")
			}`

	case DataTypeUserType:
		return `func() {
				ID("id", String, "Request ID")
				Attribute("user_id", String, "User ID")
				Attribute("name", String, "User name")
				Attribute("email", String, "User email")
				Required("id", "user_id", "name")
			}`

	case DataTypeComplex:
		return `func() {
				ID("id", String, "Request ID")
				Attribute("sequence", Int, "Sequence number")
				Attribute("data", MapOf(String, Any), "Complex data")
				Attribute("metadata", MapOf(String, Any), "Metadata")
				Required("id", "sequence")
			}`

	default:
		return `func() {
				ID("id", String, "Request ID")
				Attribute("data", String, "Default data")
				Required("id", "data")
			}`
	}
}

// GenerateSSEDSL generates DSL for SSE streaming scenarios
func GenerateSSEDSL(payloadType, resultType DataType) string {
	var dsl strings.Builder

	dsl.WriteString(`	API("test", func() {
		Title("SSE Test API")
		Version("1.0")
	})

`)

	// Add user type definitions if needed - use variable assignment like other generators
	if payloadType == DataTypeUserType || resultType == DataTypeUserType {
		dsl.WriteString(`	var UserType = Type("UserType", func() {
		Attribute("id", String)
		Attribute("name", String)
		Attribute("email", String, func() {
			Format(FormatEmail)
		})
		Attribute("age", Int, func() {
			Minimum(0)
			Maximum(150)
		})
		Required("id", "name")
	})

`)
	}

	dsl.WriteString(`	Service("events", func() {
		JSONRPC(func() {
			POST("/jsonrpc/sse")
			ServerSentEvents()
		})
		Method("subscribe", func() {
`)

	// For JSON-RPC SSE, we use POST and can send payload in the request body
	// However, for now, SSE will only support streaming results without payload
	// to keep the test scenarios simple

	// Streaming result - JSON-RPC streaming requires object results
	dsl.WriteString(fmt.Sprintf(`			StreamingResult(%s)
`, generateStreamingResultExpression(resultType)))

	// JSON-RPC endpoint
	dsl.WriteString(`			
			JSONRPC(func() {
			})
		})
	})`)

	return dsl.String()
}

// GenerateHTTPDSL is an alias for GenerateDSLCode for consistency
func GenerateHTTPDSL(payloadType, resultType DataType) string {
	return GenerateDSLCode(payloadType, resultType)
}

// GenerateErrorDSL generates DSL for error handling scenarios
func GenerateErrorDSL(customErrors bool) string {
	var dsl strings.Builder

	dsl.WriteString(`	API("test", func() {
		Title("Error Test API")
		Version("1.0")
	})

	Service("errors", func() {
		JSONRPC(func() {
			POST("/jsonrpc")
		})
`)

	if customErrors {
		dsl.WriteString(`		Error("ValidationError", func() {
			Attribute("field", String)
			Attribute("message", String)
			Required("field", "message")
		})
		
		Error("NotFoundError", func() {
			Attribute("resource", String)
			Attribute("id", String)
			Required("resource", "id")
		})

`)
	}

	dsl.WriteString(`		Method("test_error", func() {
			Payload(func() {
				Attribute("trigger", String)
				Required("trigger")
			})
			
			Result(String)
`)

	if customErrors {
		dsl.WriteString(`			
			Error("ValidationError")
			Error("NotFoundError")
`)
	}

	dsl.WriteString(`			
			JSONRPC(func() {
`)

	if customErrors {
		dsl.WriteString(`				Response("ValidationError", func() {
					Code(-32001)
				})
				Response("NotFoundError", func() {
					Code(-32002)
				})
`)
	}

	dsl.WriteString(`			})
		})
	})`)

	return dsl.String()
}

// GenerateValidationDSL generates DSL for validation scenarios
func GenerateValidationDSL(validationType string) string {
	var dsl strings.Builder

	dsl.WriteString(`	API("test", func() {
		Title("Validation Test API")
		Version("1.0")
	})

	Service("validation", func() {
		JSONRPC(func() {
			POST("/jsonrpc")
		})
		Method("validate", func() {
`)

	switch validationType {
	case "required":
		dsl.WriteString(`			Payload(func() {
				Attribute("required_field", String)
				Attribute("optional_field", String)
				// NOTE: Not using Required() here so validation happens in service, not transport
			})
			
			// Define a validation error that maps to -32602 Invalid params
			Error("invalid_params", ErrorResult, "Invalid parameters")
`)

	case "format":
		dsl.WriteString(`			Payload(func() {
				Attribute("email", String, func() {
					Format(FormatEmail)
				})
				Attribute("url", String, func() {
					Format(FormatURI)
				})
				Attribute("date", String, func() {
					Format(FormatDate)
				})
				Required("email")
			})
			
			// Define a validation error that maps to -32602 Invalid params
			Error("invalid_params", ErrorResult, "Invalid parameters")
`)
	}

	dsl.WriteString(`			
			Result(func() {
				Attribute("validated", Boolean)
				Required("validated")
			})
			
			JSONRPC(func() {
			})
		})
	})`)

	return dsl.String()
}

// GenerateBatchDSL generates DSL for batch request testing
func GenerateBatchDSL() string {
	return `	API("test", func() {
		Title("Batch Test API")
		Version("1.0")
	})

	Service("batch", func() {
		JSONRPC(func() {
			POST("/jsonrpc")
		})
		Method("add", func() {
			Payload(func() {
				Attribute("a", Int)
				Attribute("b", Int)
				Required("a", "b")
			})
			Result(Int)
			JSONRPC(func() {
			})
		})
		
		Method("multiply", func() {
			Payload(func() {
				Attribute("a", Int)
				Attribute("b", Int)
				Required("a", "b")
			})
			Result(Int)
			JSONRPC(func() {
			})
		})
	})`
}

// GenerateViewsDSL generates DSL for testing result views
func GenerateViewsDSL() string {
	return `	API("test", func() {
		Title("Views Test API")
		Version("1.0")
	})

	User := ResultType("User", func() {
		Attribute("id", String)
		Attribute("name", String)
		Attribute("email", String)
		Attribute("profile", func() {
			Attribute("bio", String)
			Attribute("avatar", String)
		})
		
		View("default", func() {
			Attribute("id")
			Attribute("name")
		})
		
		View("full", func() {
			Attribute("id")
			Attribute("name")
			Attribute("email")
			Attribute("profile")
		})
		
		Required("id", "name")
	})

	Service("users", func() {
		JSONRPC(func() {
			POST("/jsonrpc")
		})
		Method("get", func() {
			Payload(func() {
				Attribute("id", String)
				Attribute("view", String)
				Required("id")
			})
			Result(User)
			JSONRPC(func() {
			})
		})
	})`
}

// GenerateComplexDSL generates DSL for complex nested types
func GenerateComplexDSL() string {
	return `	API("test", func() {
		Title("Complex Types Test API")
		Version("1.0")
	})

	Level3 := Type("Level3", func() {
		Attribute("value", String)
		Required("value")
	})
	
	Level2 := Type("Level2", func() {
		Attribute("data", Level3)
		Attribute("items", ArrayOf(Level3))
		Required("data")
	})
	
	Level1 := Type("Level1", func() {
		Attribute("nested", Level2)
		Attribute("map", MapOf(String, Level2))
		Required("nested")
	})

	Service("complex", func() {
		JSONRPC(func() {
			POST("/jsonrpc")
		})
		Method("process", func() {
			Payload(Level1)
			Result(Level1)
			JSONRPC(func() {
			})
		})
	})`
}

// GenerateLargePayloadDSL generates DSL for large payload testing
func GenerateLargePayloadDSL() string {
	return `	API("test", func() {
		Title("Large Payload Test API")
		Version("1.0")
	})

	Service("large", func() {
		JSONRPC(func() {
			POST("/jsonrpc")
		})
		Method("process", func() {
			Payload(func() {
				Attribute("data", ArrayOf(String))
				Required("data")
			})
			Result(func() {
				Attribute("count", Int)
				Attribute("size", Int64)
				Required("count", "size")
			})
			JSONRPC(func() {
			})
		})
	})`
}

// GenerateUnicodeDSL generates DSL for unicode testing
func GenerateUnicodeDSL() string {
	return `	API("test", func() {
		Title("Unicode Test API")
		Version("1.0")
	})

	Service("unicode", func() {
		JSONRPC(func() {
			POST("/jsonrpc")
		})
		Method("echo", func() {
			Payload(func() {
				Attribute("text", String)
				Attribute("emoji", String)
				Attribute("languages", MapOf(String, String))
				Required("text")
			})
			Result(func() {
				Attribute("echoed", String)
				Attribute("length", Int)
				Required("echoed", "length")
			})
			JSONRPC(func() {
			})
		})
	})`
}
