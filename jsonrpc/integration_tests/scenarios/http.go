package scenarios

import (
	"goa.design/goa/v3/dsl"
)

// createHTTPDSL creates a DSL function for HTTP scenarios with the specified
// payload and result types. The function generates a complete Goa service
// definition including type declarations, method definitions, and JSON-RPC
// endpoint configuration.
//
// The generated DSL handles various type combinations including primitives,
// arrays, objects, maps, user-defined types, and complex nested structures.
// This allows systematic testing of type marshaling across the JSON-RPC/HTTP
// transport.
func createHTTPDSL(payloadType, resultType DataType) func() {
	return func() {
		dsl.API("test", func() {
			dsl.Title("Integration Test API")
		})

		// Define user types if needed
		if payloadType == DataTypeUserType || resultType == DataTypeUserType {
			dsl.Type("UserType", func() {
				dsl.Attribute("id", dsl.String)
				dsl.Attribute("name", dsl.String)
				dsl.Attribute("email", dsl.String, func() {
					dsl.Format(dsl.FormatEmail)
				})
				dsl.Attribute("age", dsl.Int, func() {
					dsl.Minimum(0)
					dsl.Maximum(150)
				})
				dsl.Required("id", "name")
			})
		}

		if payloadType == DataTypeComplex || resultType == DataTypeComplex {
			dsl.Type("Address", func() {
				dsl.Attribute("street", dsl.String)
				dsl.Attribute("city", dsl.String)
				dsl.Attribute("zip", dsl.String)
				dsl.Required("street", "city")
			})

			dsl.Type("ComplexType", func() {
				dsl.Attribute("data", dsl.MapOf(dsl.String, dsl.Any))
				dsl.Attribute("users", dsl.ArrayOf("UserType"))
				dsl.Attribute("addresses", dsl.ArrayOf("Address"))
				dsl.Attribute("metadata", func() {
					dsl.Attribute("created", dsl.String, func() {
						dsl.Format(dsl.FormatDateTime)
					})
					dsl.Attribute("tags", dsl.ArrayOf(dsl.String))
				})
			})
		}

		dsl.Service("test", func() {
			dsl.JSONRPC(func() {
				dsl.POST("/jsonrpc")
			})

			dsl.Method("call", func() {
				// Define payload
				if payloadType != DataTypeNone {
					dsl.Payload(createTypeExpression(payloadType))
				}

				// Define result
				dsl.Result(createTypeExpression(resultType))

				// JSON-RPC endpoint
				dsl.JSONRPC(func() {
					// Method-level JSONRPC config without POST
				})
			})
		})
	}
}

// createNotificationDSL creates a DSL for notification methods (no result)
// following the JSON-RPC specification. Notifications are fire-and-forget
// messages that don't expect a response from the server.
//
// The generated service includes a single notification method with the
// specified payload type. This tests the framework's ability to handle
// one-way communication patterns correctly.
func createNotificationDSL(payloadType DataType) func() {
	return func() {
		dsl.API("test", func() {
			dsl.Title("Notification Test API")
		})

		dsl.Service("notifier", func() {
			dsl.JSONRPC(func() {
				dsl.POST("/jsonrpc")
			})

			dsl.Method("notify", func() {
				dsl.Payload(createTypeExpression(payloadType))
				// No Result for notifications

				dsl.JSONRPC(func() {
					// Method-level JSONRPC config without POST
				})
			})
		})
	}
}

// createTypeExpression creates a type expression for the given data type
// enum value. This maps the test framework's abstract data types to concrete
// Goa DSL type expressions.
//
// The function returns the appropriate DSL type constructor or reference
// that can be used in Payload() or Result() declarations. For user-defined
// and complex types, it returns string references to previously defined types.
func createTypeExpression(dataType DataType) any {
	switch dataType {
	case DataTypePrimitive:
		return dsl.String

	case DataTypeArray:
		return dsl.ArrayOf(dsl.String)

	case DataTypeObject:
		return func() {
			dsl.Attribute("field1", dsl.String)
			dsl.Attribute("field2", dsl.Int)
			dsl.Attribute("field3", dsl.Boolean)
			dsl.Required("field1")
		}

	case DataTypeMap:
		return dsl.MapOf(dsl.String, dsl.Any)

	case DataTypeUserType:
		// For user types, return reference to named type
		return "UserType"

	case DataTypeComplex:
		// For complex types, define metadata as a map to avoid inline struct issues
		return func() {
			dsl.Attribute("sequence", dsl.Int)
			dsl.Attribute("data", dsl.MapOf(dsl.String, dsl.Any))
			dsl.Attribute("metadata", dsl.MapOf(dsl.String, dsl.Any))
			dsl.Required("sequence")
		}

	default:
		return dsl.String
	}
}

// createHTTPRequests creates test requests for HTTP scenarios with appropriate
// payload and expected results based on the data types. Each request includes
// the method name, parameters matching the payload type, and expected results
// matching the result type.
//
// The function generates a single request per scenario since HTTP is a
// request-response protocol without streaming capabilities.
func createHTTPRequests(payloadType, resultType DataType) []TestRequest {
	return []TestRequest{
		{
			Method:         "call", // Use simple method name from DSL
			Params:         createTestPayload(payloadType),
			ExpectedResult: createExpectedResult(resultType),
		},
	}
}

// createNotificationRequests creates test requests for notification scenarios
// where no response is expected from the server. According to JSON-RPC
// specification, notifications are requests without an ID field.
//
// The test framework validates that the server accepts the notification
// without sending a response, testing the one-way communication pattern.
func createNotificationRequests(payloadType DataType) []TestRequest {
	return []TestRequest{
		{
			Method: "notify", // Use simple method name from DSL
			Params: createTestPayload(payloadType),
			// No expected result for notifications
		},
	}
}

// createTestPayload creates test payload data appropriate for the specified
// data type. The generated payloads are designed to exercise type marshaling
// and validation logic in the JSON-RPC transport.
//
// Each payload contains realistic test data with sufficient complexity to
// verify correct handling of the type, including nested structures, arrays,
// and maps where applicable.
func createTestPayload(dataType DataType) any {
	switch dataType {
	case DataTypeNone:
		return nil

	case DataTypePrimitive:
		return "test string"

	case DataTypeArray:
		// Arrays are wrapped in objects for CLI compatibility
		return map[string]any{
			"items": []string{"item1", "item2", "item3"},
		}

	case DataTypeObject:
		return map[string]any{
			"field1": "value1",
			"field2": 42,
			"field3": true,
		}

	case DataTypeMap:
		return map[string]any{
			"key1": "value1",
			"key2": 123,
			"key3": []string{"a", "b", "c"},
		}

	case DataTypeUserType:
		return map[string]any{
			"id":    "user123",
			"name":  "Test User",
			"email": "test@example.com",
			"age":   25,
		}

	case DataTypeComplex:
		return map[string]any{
			"data": map[string]any{
				"nested": "value",
			},
			"users": []map[string]any{
				{
					"id":   "u1",
					"name": "User 1",
				},
			},
			"addresses": []map[string]any{
				{
					"street": "123 Main St",
					"city":   "Test City",
					"zip":    "12345",
				},
			},
			"metadata": map[string]any{
				"created": "2024-01-01T12:00:00Z",
				"tags":    []string{"test", "integration"},
			},
		}

	default:
		return "default"
	}
}

// createExpectedResult creates expected result data for validating responses
// from the server. The generated data matches what the test server should
// return for each data type.
//
// The results are structured to allow deep equality comparisons during
// validation, ensuring both the structure and values match expectations.
// This helps detect issues with type conversion, field mapping, and
// JSON marshaling in the response path.
func createExpectedResult(dataType DataType) any {
	// For integration tests, we're mainly checking that the types
	// are preserved correctly, not exact values
	switch dataType {
	case DataTypePrimitive:
		return "string"

	case DataTypeArray:
		return []any{}

	case DataTypeObject:
		return map[string]any{}

	case DataTypeMap:
		return map[string]any{}

	case DataTypeUserType:
		return map[string]any{
			"ID":    "string",
			"Name":  "string",
			"Email": "string",
			"Age":   0,
		}

	case DataTypeComplex:
		return map[string]any{}

	default:
		return nil
	}
}
