package scenarios

import (
	"fmt"

	"goa.design/goa/v3/dsl"
)

// createValidationDSL creates a DSL for validation scenarios that test the
// framework's input validation capabilities. Different validation types
// exercise various validation rules:
//   - "required": Tests required field validation
//   - "format": Tests format validation (email, URL, date)
//
// The generated service includes validation constraints that should trigger
// JSON-RPC error responses when violated, ensuring proper error handling
// in the transport layer.
func createValidationDSL(validationType string) func() {
	return func() {
		dsl.API("test", func() {
			dsl.Title("Validation Test API")
		})

		dsl.Service("validation", func() {
			dsl.Method("validate", func() {
				switch validationType {
				case "required":
					dsl.Payload(func() {
						dsl.Attribute("required_field", dsl.String)
						dsl.Attribute("optional_field", dsl.String)
						dsl.Required("required_field")
					})

				case "format":
					dsl.Payload(func() {
						dsl.Attribute("email", dsl.String, func() {
							dsl.Format(dsl.FormatEmail)
						})
						dsl.Attribute("url", dsl.String, func() {
							dsl.Format(dsl.FormatURI)
						})
						dsl.Attribute("date", dsl.String, func() {
							dsl.Format(dsl.FormatDate)
						})
						dsl.Required("email")
					})
				}

				dsl.Result(func() {
					dsl.Attribute("validated", dsl.Boolean)
					dsl.Required("validated")
				})

				dsl.JSONRPC(func() {
					dsl.POST("/jsonrpc")
				})
			})
		})
	}
}

// createValidationRequests creates test requests for validation scenarios
// including both valid and invalid inputs. The requests are designed to
// trigger specific validation errors and verify the framework correctly
// returns JSON-RPC error responses with appropriate error codes.
//
// Each scenario includes requests that should succeed and requests that
// should fail with -32602 (Invalid params) errors, testing the complete
// validation pipeline from transport to service layer.
func createValidationRequests(validationType string) []TestRequest {
	switch validationType {
	case "required":
		return []TestRequest{
			{
				Method: "validate", // Use simple method name from DSL
				Params: map[string]any{
					"required_field": "value",
				},
				ExpectedResult: map[string]any{"validated": false},
			},
			{
				Method: "validate", // Use simple method name from DSL
				Params: map[string]any{
					"optional_field": "only optional",
				},
				ExpectedError: &ExpectedError{
					Code:    -32602,
					Message: "invalid params",
				},
			},
		}

	case "format":
		return []TestRequest{
			{
				Method: "validate", // Use simple method name from DSL
				Params: map[string]any{
					"email": "test@example.com",
					"url":   "https://example.com",
					"date":  "2024-01-01",
				},
				ExpectedResult: map[string]any{"validated": false},
			},
			{
				Method: "validate", // Use simple method name from DSL
				Params: map[string]any{
					"email": "invalid-email",
				},
				ExpectedError: &ExpectedError{
					Code:    -32602,
					Message: "invalid params",
				},
			},
		}

	default:
		return nil
	}
}

// createBatchDSL creates a DSL for batch request testing
func createBatchDSL() func() {
	return func() {
		dsl.API("test", func() {
			dsl.Title("Batch Test API")
		})

		dsl.Service("batch", func() {
			dsl.Method("add", func() {
				dsl.Payload(func() {
					dsl.Attribute("a", dsl.Int)
					dsl.Attribute("b", dsl.Int)
					dsl.Required("a", "b")
				})
				dsl.Result(dsl.Int)
				dsl.JSONRPC(func() {
					dsl.POST("/jsonrpc")
				})
			})

			dsl.Method("multiply", func() {
				dsl.Payload(func() {
					dsl.Attribute("a", dsl.Int)
					dsl.Attribute("b", dsl.Int)
					dsl.Required("a", "b")
				})
				dsl.Result(dsl.Int)
				dsl.JSONRPC(func() {
					dsl.POST("/jsonrpc")
				})
			})
		})
	}
}

// createBatchRequests creates batch test requests
func createBatchRequests() []TestRequest {
	// Batch requests are handled differently - this is a placeholder
	return []TestRequest{
		{
			Method: "batch",
			Params: []any{
				map[string]any{
					"jsonrpc": "2.0",
					"method":  "add", // Use simple method name from DSL
					"params":  map[string]any{"a": 5, "b": 3},
					"id":      1,
				},
				map[string]any{
					"jsonrpc": "2.0",
					"method":  "multiply", // Use simple method name from DSL
					"params":  map[string]any{"a": 4, "b": 2},
					"id":      2,
				},
			},
		},
	}
}

// createViewsDSL creates a DSL for testing result views
func createViewsDSL() func() {
	return func() {
		dsl.API("test", func() {
			dsl.Title("Views Test API")
		})

		var UserResult = dsl.ResultType("User", func() {
			dsl.Attribute("id", dsl.String)
			dsl.Attribute("name", dsl.String)
			dsl.Attribute("email", dsl.String)
			dsl.Attribute("profile", func() {
				dsl.Attribute("bio", dsl.String)
				dsl.Attribute("avatar", dsl.String)
			})

			dsl.View("default", func() {
				dsl.Attribute("id")
				dsl.Attribute("name")
			})

			dsl.View("full", func() {
				dsl.Attribute("id")
				dsl.Attribute("name")
				dsl.Attribute("email")
				dsl.Attribute("profile")
			})

			dsl.Required("id", "name")
		})

		dsl.Service("users", func() {
			dsl.Method("get", func() {
				dsl.Payload(func() {
					dsl.Attribute("id", dsl.String)
					dsl.Attribute("view", dsl.String)
					dsl.Required("id")
				})
				dsl.Result(UserResult)
				dsl.JSONRPC(func() {
					dsl.POST("/jsonrpc")
				})
			})
		})
	}
}

// createViewsRequests creates test requests for views
func createViewsRequests() []TestRequest {
	return []TestRequest{
		{
			Method: "get", // Use simple method name from DSL
			Params: map[string]any{
				"id": "user123",
			},
			ExpectedResult: map[string]any{
				"id":   "user123",
				"name": "Test User",
			},
		},
		{
			Method: "get", // Use simple method name from DSL
			Params: map[string]any{
				"id":   "user123",
				"view": "full",
			},
			ExpectedResult: map[string]any{
				"id":    "user123",
				"name":  "Test User",
				"email": "test@example.com",
			},
		},
	}
}

// createComplexDSL creates a DSL for complex nested types
func createComplexDSL() func() {
	return func() {
		dsl.API("test", func() {
			dsl.Title("Complex Types Test API")
		})

		dsl.Type("Level3", func() {
			dsl.Attribute("value", dsl.String)
			dsl.Required("value")
		})

		dsl.Type("Level2", func() {
			dsl.Attribute("data", "Level3")
			dsl.Attribute("items", dsl.ArrayOf("Level3"))
			dsl.Required("data")
		})

		dsl.Type("Level1", func() {
			dsl.Attribute("nested", "Level2")
			dsl.Attribute("map", dsl.MapOf(dsl.String, "Level2"))
			dsl.Required("nested")
		})

		dsl.Service("complex", func() {
			dsl.Method("process", func() {
				dsl.Payload("Level1")
				dsl.Result("Level1")
				dsl.JSONRPC(func() {
					dsl.POST("/jsonrpc")
				})
			})
		})
	}
}

// createComplexRequests creates test requests for complex types
func createComplexRequests() []TestRequest {
	return []TestRequest{
		{
			Method: "process", // Use simple method name from DSL
			Params: map[string]any{
				"nested": map[string]any{
					"data": map[string]any{
						"value": "deep",
					},
					"items": []map[string]any{
						{"value": "item1"},
						{"value": "item2"},
					},
				},
				"map": map[string]any{
					"key1": map[string]any{
						"data": map[string]any{
							"value": "mapped",
						},
					},
				},
			},
		},
	}
}

// createLargePayloadDSL creates a DSL for large payload testing
func createLargePayloadDSL() func() {
	return func() {
		dsl.API("test", func() {
			dsl.Title("Large Payload Test API")
		})

		dsl.Service("large", func() {
			dsl.Method("process", func() {
				dsl.Payload(func() {
					dsl.Attribute("data", dsl.ArrayOf(dsl.String))
					dsl.Required("data")
				})
				dsl.Result(func() {
					dsl.Attribute("count", dsl.Int)
					dsl.Attribute("size", dsl.Int64)
					dsl.Required("count", "size")
				})
				dsl.JSONRPC(func() {
					dsl.POST("/jsonrpc")
				})
			})
		})
	}
}

// createLargePayloadRequests creates test requests with large payloads
func createLargePayloadRequests() []TestRequest {
	// Create a large array
	largeData := make([]string, 10000)
	for i := range largeData {
		largeData[i] = fmt.Sprintf("item-%d-with-some-additional-data-to-make-it-larger", i)
	}

	return []TestRequest{
		{
			Method: "process", // Use simple method name from DSL
			Params: map[string]any{
				"data": largeData,
			},
			ExpectedResult: map[string]any{
				"count": float64(10000),
			},
		},
	}
}

// createUnicodeDSL creates a DSL for unicode testing
func createUnicodeDSL() func() {
	return func() {
		dsl.API("test", func() {
			dsl.Title("Unicode Test API")
		})

		dsl.Service("unicode", func() {
			dsl.Method("echo", func() {
				dsl.Payload(func() {
					dsl.Attribute("text", dsl.String)
					dsl.Attribute("emoji", dsl.String)
					dsl.Attribute("languages", dsl.MapOf(dsl.String, dsl.String))
					dsl.Required("text")
				})
				dsl.Result(func() {
					dsl.Attribute("echoed", dsl.String)
					dsl.Attribute("length", dsl.Int)
					dsl.Required("echoed", "length")
				})
				dsl.JSONRPC(func() {
					dsl.POST("/jsonrpc")
				})
			})
		})
	}
}

// createUnicodeRequests creates test requests with unicode data
func createUnicodeRequests() []TestRequest {
	return []TestRequest{
		{
			Method: "echo", // Use simple method name from DSL
			Params: map[string]any{
				"text":  "Hello 世界 🌍",
				"emoji": "🚀🌟💻🎉",
				"languages": map[string]string{
					"english":  "Hello",
					"chinese":  "你好",
					"japanese": "こんにちは",
					"arabic":   "مرحبا",
					"hebrew":   "שלום",
				},
			},
			ExpectedResult: map[string]any{
				"echoed": "Hello 世界 🌍",
			},
		},
	}
}
