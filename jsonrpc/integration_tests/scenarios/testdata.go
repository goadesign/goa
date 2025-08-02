package scenarios

import "fmt"

// SSETestData provides a single source of truth for SSE test data.
// Both the server implementation and test validation use this.
type SSETestData struct {
	ResultType DataType
}

// GenerateData creates the test data for a given index.
// This is used by tests to know what to expect.
func (s SSETestData) GenerateData(index int) any {
	switch s.ResultType {
	case DataTypePrimitive:
		// Now wrapped in object for JSON-RPC streaming compliance
		// Use uppercase field name to match Go struct JSON marshaling
		return map[string]any{
			"Value": fmt.Sprintf("event %d", index),
		}

	case DataTypeArray:
		// Now wrapped in object for JSON-RPC streaming compliance
		// Use uppercase field name to match Go struct JSON marshaling
		return map[string]any{
			"Items": []any{
				fmt.Sprintf("event-%d-a", index),
				fmt.Sprintf("event-%d-b", index),
				index,
			},
		}

	case DataTypeObject:
		// Match the actual generated object type with field1, field2, field3
		return map[string]any{
			"field1": fmt.Sprintf("evt-%03d", index),
			"field2": index * 10,
			"field3": index%2 == 0,
		}

	case DataTypeUserType:
		// Match the actual generated UserType with id, name, email, age
		return map[string]any{
			"id":    fmt.Sprintf("evt-user-%d", index),
			"name":  fmt.Sprintf("Event User %d", index),
			"email": fmt.Sprintf("event%d@example.com", index),
			"age":   25 + index,
		}

	case DataTypeComplex:
		// Return the complex structure with metadata as a map
		return map[string]any{
			"sequence": index,
			"data": map[string]any{
				"event": fmt.Sprintf("complex-event-%d", index),
				"nested": map[string]any{
					"level": index,
					"info":  fmt.Sprintf("Level %d info", index),
				},
			},
			"metadata": map[string]any{
				"index": index,
				"type":  "sse",
			},
		}

	default:
		return fmt.Sprintf("sse-data-%d", index)
	}
}

// GenerateImplementationCode generates the Go code for the server to send this data.
// This ensures the server sends exactly what GenerateData returns.
func (s SSETestData) GenerateImplementationCode(serviceName string) string {
	switch s.ResultType {
	case DataTypePrimitive:
		// Now wrapped in object for JSON-RPC streaming compliance
		return fmt.Sprintf(`&%s.SubscribeResult{
			Value: fmt.Sprintf("event %%d", i),
		}`, serviceName)

	case DataTypeArray:
		// Now wrapped in object for JSON-RPC streaming compliance
		return fmt.Sprintf(`&%s.SubscribeResult{
			Items: []string{
				fmt.Sprintf("event-%%d-a", i),
				fmt.Sprintf("event-%%d-b", i),
				fmt.Sprintf("%%d", i),
			},
		}`, serviceName)

	case DataTypeObject:
		return fmt.Sprintf(`func() *%s.SubscribeResult {
			field2 := i * 10
			field3 := i%%2 == 0
			return &%s.SubscribeResult{
				Field1: fmt.Sprintf("evt-%%03d", i),
				Field2: &field2,
				Field3: &field3,
			}
		}()`, serviceName, serviceName)

	case DataTypeUserType:
		return fmt.Sprintf(`func() *%s.UserType {
			email := fmt.Sprintf("event%%d@example.com", i)
			age := 25 + i
			return &%s.UserType{
				ID:    fmt.Sprintf("evt-user-%%d", i),
				Name:  fmt.Sprintf("Event User %%d", i),
				Email: &email,
				Age:   &age,
			}
		}()`, serviceName, serviceName)

	case DataTypeComplex:
		// For complex type, generate the correct structure with metadata as a map
		return fmt.Sprintf(`&%s.SubscribeResult{
			Sequence: i,
			Data: map[string]any{
				"event": fmt.Sprintf("complex-event-%%d", i),
				"nested": map[string]any{
					"level": i,
					"info":  fmt.Sprintf("Level %%d info", i),
				},
			},
			Metadata: map[string]any{
				"index": i,
				"type":  "sse",
			},
		}`, serviceName)

	default:
		// Default fallback - wrap in object for JSON-RPC streaming compliance
		return fmt.Sprintf(`&%s.SubscribeResult{
			Value: fmt.Sprintf("event %%d", i),
		}`, serviceName)
	}
}
