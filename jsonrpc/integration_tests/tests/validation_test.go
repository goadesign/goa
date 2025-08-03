package tests

import (
	"testing"

	"goa.design/goa/v3/jsonrpc/integration_tests/harness"
	"goa.design/goa/v3/jsonrpc/integration_tests/scenarios"
	"goa.design/goa/v3/jsonrpc/integration_tests/validators"
)

// TestRequiredFieldValidation tests required field validation
func TestRequiredFieldValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration tests skipped in short mode")
	}

	// Create test harness
	h := harness.New(t)

	// Create scenario with required fields
	scenario := scenarios.Scenario{
		Name:        "required_field_validation",
		Description: "Test required field validation",
		Transport:   scenarios.TransportHTTP,
		DSLCode:     createRequiredFieldsDSLCode(),
		Requests: []scenarios.TestRequest{
			// Valid request with all required fields
			{
				Method: "create_user",
				Params: map[string]any{
					"name":  "Test User",
					"email": "test@example.com",
					"age":   25,
				},
				ExpectedResult: map[string]any{
					"id":      "",
					"created": false,
				},
			},
			// Missing required field 'name'
			{
				Method: "create_user",
				Params: map[string]any{
					"email": "test@example.com",
					"age":   25,
				},
				ExpectedError: &scenarios.ExpectedError{
					Code:    -32602,
					Message: "invalid params",
				},
			},
			// Missing required field 'email'
			{
				Method: "create_user",
				Params: map[string]any{
					"name": "Test User",
					"age":  25,
				},
				ExpectedError: &scenarios.ExpectedError{
					Code:    -32602,
					Message: "invalid params",
				},
			},
			// All optional fields missing (valid)
			{
				Method: "create_user",
				Params: map[string]any{
					"name":  "Minimal User",
					"email": "minimal@example.com",
				},
				ExpectedResult: map[string]any{
					"id":      "",
					"created": false,
				},
			},
		},
		Validators: []validators.Validator{
			validators.ProtocolValidator(),
			validators.DataIntegrityValidator(),
		},
	}

	// Run scenario
	runner := scenarios.NewScenarioRunner(h)
	if err := runner.Run(scenario); err != nil {
		t.Fatalf("Required field validation scenario failed: %v", err)
	}
}

// TestFormatValidation tests format validation (email, URL, etc.)
func TestFormatValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration tests skipped in short mode")
	}

	// Create test harness
	h := harness.New(t)

	// Create scenario with format validation
	scenario := scenarios.Scenario{
		Name:        "format_validation",
		Description: "Test format validation",
		Transport:   scenarios.TransportHTTP,
		DSLCode:     createFormatValidationDSLCode(),
		Requests: []scenarios.TestRequest{
			// Valid formats
			{
				Method: "validate_formats",
				Params: map[string]any{
					"email":    "valid@example.com",
					"url":      "https://example.com",
					"date":     "2024-01-01",
					"datetime": "2024-01-01T12:00:00Z",
				},
				ExpectedResult: map[string]any{
					"valid": false,
				},
			},
			// Invalid email format
			{
				Method: "validate_formats",
				Params: map[string]any{
					"email":    "invalid-email",
					"url":      "https://example.com",
					"date":     "2024-01-01",
					"datetime": "2024-01-01T12:00:00Z",
				},
				ExpectedError: &scenarios.ExpectedError{
					Code:    -32602,
					Message: "invalid params",
				},
			},
			// Invalid URL format
			{
				Method: "validate_formats",
				Params: map[string]any{
					"email":    "valid@example.com",
					"url":      "not-a-url",
					"date":     "2024-01-01",
					"datetime": "2024-01-01T12:00:00Z",
				},
				ExpectedError: &scenarios.ExpectedError{
					Code:    -32602,
					Message: "invalid params",
				},
			},
			// Invalid date format
			{
				Method: "validate_formats",
				Params: map[string]any{
					"email":    "valid@example.com",
					"url":      "https://example.com",
					"date":     "01/01/2024", // Wrong format
					"datetime": "2024-01-01T12:00:00Z",
				},
				ExpectedError: &scenarios.ExpectedError{
					Code:    -32602,
					Message: "invalid params",
				},
			},
		},
		Validators: []validators.Validator{
			validators.ProtocolValidator(),
			validators.DataIntegrityValidator(),
		},
	}

	// Run scenario
	runner := scenarios.NewScenarioRunner(h)
	if err := runner.Run(scenario); err != nil {
		t.Fatalf("Format validation scenario failed: %v", err)
	}
}

// TestRangeValidation tests numeric range validation
func TestRangeValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration tests skipped in short mode")
	}

	// Create test harness
	h := harness.New(t)

	// Create scenario with range validation
	scenario := scenarios.Scenario{
		Name:        "range_validation",
		Description: "Test numeric range validation",
		Transport:   scenarios.TransportHTTP,
		DSLCode:     createRangeValidationDSLCode(),
		Requests: []scenarios.TestRequest{
			// Valid ranges
			{
				Method: "validate_ranges",
				Params: map[string]any{
					"age":        25,
					"score":      75.5,
					"count":      100,
					"percentage": 50.0,
				},
				ExpectedResult: map[string]any{
					"valid": false,
				},
			},
			// Age too low
			{
				Method: "validate_ranges",
				Params: map[string]any{
					"age":        17, // Min is 18
					"score":      75.5,
					"count":      100,
					"percentage": 50.0,
				},
				ExpectedError: &scenarios.ExpectedError{
					Code:    -32602,
					Message: "invalid params",
				},
			},
			// Age too high
			{
				Method: "validate_ranges",
				Params: map[string]any{
					"age":        151, // Max is 150
					"score":      75.5,
					"count":      100,
					"percentage": 50.0,
				},
				ExpectedError: &scenarios.ExpectedError{
					Code:    -32602,
					Message: "invalid params",
				},
			},
			// Percentage out of range
			{
				Method: "validate_ranges",
				Params: map[string]any{
					"age":        25,
					"score":      75.5,
					"count":      100,
					"percentage": 150.0, // Max is 100
				},
				ExpectedError: &scenarios.ExpectedError{
					Code:    -32602,
					Message: "invalid params",
				},
			},
		},
		Validators: []validators.Validator{
			validators.ProtocolValidator(),
			validators.DataIntegrityValidator(),
		},
	}

	// Run scenario
	runner := scenarios.NewScenarioRunner(h)
	if err := runner.Run(scenario); err != nil {
		t.Fatalf("Range validation scenario failed: %v", err)
	}
}

// TestStringValidation tests string validation (length, pattern)
func TestStringValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration tests skipped in short mode")
	}

	// Create test harness
	h := harness.New(t)

	// Create scenario with string validation
	scenario := scenarios.Scenario{
		Name:        "string_validation",
		Description: "Test string validation",
		Transport:   scenarios.TransportHTTP,
		DSLCode:     createStringValidationDSLCode(),
		Requests: []scenarios.TestRequest{
			// Valid strings
			{
				Method: "validate_strings",
				Params: map[string]any{
					"username": "john_doe",
					"password": "SecurePass123!",
					"code":     "ABC123",
					"bio":      "A short bio about me.",
				},
				ExpectedResult: map[string]any{
					"valid": false,
				},
			},
			// Username too short
			{
				Method: "validate_strings",
				Params: map[string]any{
					"username": "ab", // Min length 3
					"password": "SecurePass123!",
					"code":     "ABC123",
					"bio":      "A short bio about me.",
				},
				ExpectedError: &scenarios.ExpectedError{
					Code:    -32602,
					Message: "invalid params",
				},
			},
			// Password too short
			{
				Method: "validate_strings",
				Params: map[string]any{
					"username": "john_doe",
					"password": "Short1!", // Min length 8
					"code":     "ABC123",
					"bio":      "A short bio about me.",
				},
				ExpectedError: &scenarios.ExpectedError{
					Code:    -32602,
					Message: "invalid params",
				},
			},
			// Invalid code pattern
			{
				Method: "validate_strings",
				Params: map[string]any{
					"username": "john_doe",
					"password": "SecurePass123!",
					"code":     "abc123", // Must be uppercase
					"bio":      "A short bio about me.",
				},
				ExpectedError: &scenarios.ExpectedError{
					Code:    -32602,
					Message: "invalid params",
				},
			},
		},
		Validators: []validators.Validator{
			validators.ProtocolValidator(),
			validators.DataIntegrityValidator(),
		},
	}

	// Run scenario
	runner := scenarios.NewScenarioRunner(h)
	if err := runner.Run(scenario); err != nil {
		t.Fatalf("String validation scenario failed: %v", err)
	}
}

// TestEnumValidation tests enum validation
func TestEnumValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration tests skipped in short mode")
	}

	// Create test harness
	h := harness.New(t)

	// Create scenario with enum validation
	scenario := scenarios.Scenario{
		Name:        "enum_validation",
		Description: "Test enum validation",
		Transport:   scenarios.TransportHTTP,
		DSLCode:     createEnumValidationDSLCode(),
		Requests: []scenarios.TestRequest{
			// Valid enum values
			{
				Method: "validate_enums",
				Params: map[string]any{
					"status":   "active",
					"role":     "admin",
					"priority": "high",
				},
				ExpectedResult: map[string]any{
					"valid": false,
				},
			},
			// Invalid status
			{
				Method: "validate_enums",
				Params: map[string]any{
					"status":   "unknown", // Not in enum
					"role":     "admin",
					"priority": "high",
				},
				ExpectedError: &scenarios.ExpectedError{
					Code:    -32602,
					Message: "invalid params",
				},
			},
			// Invalid role
			{
				Method: "validate_enums",
				Params: map[string]any{
					"status":   "active",
					"role":     "superuser", // Not in enum
					"priority": "high",
				},
				ExpectedError: &scenarios.ExpectedError{
					Code:    -32602,
					Message: "invalid params",
				},
			},
		},
		Validators: []validators.Validator{
			validators.ProtocolValidator(),
			validators.DataIntegrityValidator(),
		},
	}

	// Run scenario
	runner := scenarios.NewScenarioRunner(h)
	if err := runner.Run(scenario); err != nil {
		t.Fatalf("Enum validation scenario failed: %v", err)
	}
}

// Helper DSL creation functions

func createRequiredFieldsDSLCode() string {
	return `	API("test", func() {
		Title("Required Fields Test API")
	})
	
	Service("users", func() {
		JSONRPC(func() {
			POST("/jsonrpc")
		})
		Method("create_user", func() {
			Payload(func() {
				Attribute("name", String)
				Attribute("email", String)
				Attribute("age", Int)
				Attribute("bio", String)      // Optional
				Attribute("website", String)  // Optional
				Required("name", "email") // age is optional despite being in params
			})
			Result(func() {
				Attribute("id", String)
				Attribute("created", Boolean)
				Required("id", "created")
			})
			JSONRPC(func() {
			})
		})
	})`
}

func createFormatValidationDSLCode() string {
	return `	API("test", func() {
		Title("Format Validation Test API")
	})
	
	Service("validation", func() {
		JSONRPC(func() {
			POST("/jsonrpc")
		})
		Method("validate_formats", func() {
			Payload(func() {
				Attribute("email", String, func() {
					Format(FormatEmail)
				})
				Attribute("url", String, func() {
					Format(FormatURI)
				})
				Attribute("date", String, func() {
					Format(FormatDate)
				})
				Attribute("datetime", String, func() {
					Format(FormatDateTime)
				})
				Required("email", "url", "date", "datetime")
			})
			Result(func() {
				Attribute("valid", Boolean)
				Required("valid")
			})
			JSONRPC(func() {
			})
		})
	})`
}

func createRangeValidationDSLCode() string {
	return `	API("test", func() {
		Title("Range Validation Test API")
	})
	
	Service("validation", func() {
		JSONRPC(func() {
			POST("/jsonrpc")
		})
		Method("validate_ranges", func() {
			Payload(func() {
				Attribute("age", Int, func() {
					Minimum(18)
					Maximum(150)
				})
				Attribute("score", Float64, func() {
					Minimum(0.0)
					Maximum(100.0)
				})
				Attribute("count", Int, func() {
					Minimum(1)
				})
				Attribute("percentage", Float64, func() {
					Minimum(0.0)
					Maximum(100.0)
				})
				Required("age", "score", "count", "percentage")
			})
			Result(func() {
				Attribute("valid", Boolean)
				Required("valid")
			})
			JSONRPC(func() {
			})
		})
	})`
}

func createStringValidationDSLCode() string {
	return `	API("test", func() {
		Title("String Validation Test API")
	})
	
	Service("validation", func() {
		JSONRPC(func() {
			POST("/jsonrpc")
		})
		Method("validate_strings", func() {
			Payload(func() {
				Attribute("username", String, func() {
					MinLength(3)
					MaxLength(20)
					Pattern("^[a-zA-Z0-9_]+$")
				})
				Attribute("password", String, func() {
					MinLength(8)
					MaxLength(128)
				})
				Attribute("code", String, func() {
					Pattern("^[A-Z]{3}[0-9]{3}$")
				})
				Attribute("bio", String, func() {
					MaxLength(500)
				})
				Required("username", "password", "code")
			})
			Result(func() {
				Attribute("valid", Boolean)
				Required("valid")
			})
			JSONRPC(func() {
			})
		})
	})`
}

func createEnumValidationDSLCode() string {
	return `	API("test", func() {
		Title("Enum Validation Test API")
	})
	
	Service("validation", func() {
		JSONRPC(func() {
			POST("/jsonrpc")
		})
		Method("validate_enums", func() {
			Payload(func() {
				Attribute("status", String, func() {
					Enum("active", "inactive", "pending", "suspended")
				})
				Attribute("role", String, func() {
					Enum("admin", "user", "moderator", "guest")
				})
				Attribute("priority", String, func() {
					Enum("low", "medium", "high", "critical")
				})
				Required("status", "role", "priority")
			})
			Result(func() {
				Attribute("valid", Boolean)
				Required("valid")
			})
			JSONRPC(func() {
			})
		})
	})`
}
