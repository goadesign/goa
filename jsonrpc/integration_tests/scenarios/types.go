package scenarios

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/jsonrpc/integration_tests/harness"
	"goa.design/goa/v3/jsonrpc/integration_tests/validators"
)

// Transport represents a JSON-RPC transport type
type Transport string

const (
	TransportHTTP      Transport = "http"
	TransportWebSocket Transport = "websocket"
	TransportSSE       Transport = "sse"
)

// DataType represents a data type used in payloads or results
type DataType string

const (
	DataTypeNone      DataType = "none"
	DataTypePrimitive DataType = "primitive"
	DataTypeArray     DataType = "array"
	DataTypeObject    DataType = "object"
	DataTypeMap       DataType = "map"
	DataTypeUserType  DataType = "user_type"
	DataTypeComplex   DataType = "complex"
)

// StreamingType represents the type of streaming
type StreamingType string

const (
	StreamingNone          StreamingType = "none"
	StreamingServer        StreamingType = "server"
	StreamingClient        StreamingType = "client"
	StreamingBidirectional StreamingType = "bidirectional"
)

// Feature represents a test feature
type Feature string

const (
	FeatureCore       Feature = "core"
	FeatureStreaming  Feature = "streaming"
	FeatureErrors     Feature = "errors"
	FeatureValidation Feature = "validation"
	FeatureViews      Feature = "views"
	FeatureBatch      Feature = "batch"
)

// Scenario represents a complete test scenario
type Scenario struct {
	// Name is the unique name for this scenario
	Name string

	// Description provides details about what this scenario tests
	Description string

	// Transport is the transport type being tested
	Transport Transport

	// PayloadType is the type of data in the request payload
	PayloadType DataType

	// ResultType is the type of data in the response
	ResultType DataType

	// Streaming specifies the streaming configuration
	Streaming StreamingType

	// Features lists the features being tested
	Features []Feature

	// DSLFile is the path to the DSL file (relative to testdata/dsls)
	DSLFile string

	// DSLCode is the actual DSL code (alternative to DSLFile)
	DSLCode string

	// Requests defines the test requests to execute
	Requests []TestRequest

	// Validators are the validation functions to run
	Validators []validators.Validator

	// Skip provides a reason to skip this test
	Skip string
}

// TestRequest represents a single test request
type TestRequest struct {
	// Method is the JSON-RPC method name
	Method string

	// Params are the request parameters
	Params any

	// ExpectedResult is the expected result (for non-streaming)
	ExpectedResult any

	// ExpectedError is the expected error (if any)
	ExpectedError *ExpectedError

	// StreamingMessages for streaming scenarios
	StreamingMessages []StreamMessage
}

// ExpectedError represents an expected JSON-RPC error
type ExpectedError struct {
	Code    int
	Message string
	Data    any
}

// StreamMessage represents a message in a streaming scenario
type StreamMessage struct {
	Direction MessageDirection
	Data      any
	Delay     int // milliseconds
}

// MessageDirection indicates the direction of a streaming message
type MessageDirection string

const (
	DirectionSend    MessageDirection = "send"
	DirectionReceive MessageDirection = "receive"
)

// ScenarioRunner coordinates the execution of test scenarios using a test
// harness. It handles the complete lifecycle of a scenario: generating code,
// starting servers, executing requests, and validating responses.
//
// The runner abstracts transport-specific logic, delegating to appropriate
// handlers for HTTP, WebSocket, and SSE scenarios.
type ScenarioRunner struct {
	harness *harness.TestHarness
}

// NewScenarioRunner creates a new scenario runner
func NewScenarioRunner(h *harness.TestHarness) *ScenarioRunner {
	return &ScenarioRunner{
		harness: h,
	}
}

// Run executes a complete test scenario from start to finish. It:
//  1. Generates code from the scenario's DSL
//  2. Compiles and starts a server
//  3. Creates a client and executes test requests
//  4. Validates responses using the scenario's validators
//  5. Cleans up all resources via the harness
//
// The method returns an error if any step fails. Cleanup is automatic and
// guaranteed by the test harness.
func (r *ScenarioRunner) Run(scenario Scenario) error {
	// Skip if needed
	if scenario.Skip != "" {
		return nil
	}

	// Create overall timeout context for the entire scenario
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Generate code from DSL
	var genDir string
	var err error

	if scenario.DSLFile != "" {
		// Use DSL file
		genDir, err = r.harness.GenerateCodeFromFile(ctx, scenario.Name, scenario.DSLFile)
	} else if scenario.DSLCode != "" {
		// Use DSL code directly
		genDir, err = r.harness.GenerateCode(ctx, scenario.Name, scenario.DSLCode)
	} else {
		return fmt.Errorf("scenario %s has neither DSLFile nor DSLCode", scenario.Name)
	}

	if err != nil {
		return err
	}

	// Start server
	port, err := r.harness.AllocatePort()
	if err != nil {
		return err
	}

	// Find the server directory - goa example generates cmd/<api-name>/
	// For our test scenarios, the API is always named "test"
	serverDir := filepath.Join(genDir, "cmd", "test")

	// Verify the directory exists
	if _, err := os.Stat(serverDir); os.IsNotExist(err) {
		// Fallback: look for any non-cli directory
		cmdDir := filepath.Join(genDir, "cmd")
		entries, err := os.ReadDir(cmdDir)
		if err != nil {
			return fmt.Errorf("failed to read cmd directory: %w", err)
		}

		for _, entry := range entries {
			if entry.IsDir() && !strings.HasSuffix(entry.Name(), "-cli") {
				serverDir = filepath.Join(cmdDir, entry.Name())
				break
			}
		}

		if _, err := os.Stat(serverDir); os.IsNotExist(err) {
			return fmt.Errorf("no server directory found in %s", cmdDir)
		}
	}

	serverConfig := harness.ServerConfig{
		SourceDir:      serverDir,
		Port:           port,
		StartupTimeout: 2 * time.Second,
		ReadyString:    "listening", // Match any server listening message
	}

	// Add service implementations based on scenario type
	if scenario.Transport == TransportSSE {
		serverConfig.ServiceImplementations = r.createSSEImplementations(scenario)
	} else if hasFeature(scenario.Features, FeatureErrors) {
		serverConfig.ServiceImplementations = r.createErrorImplementations(scenario)
	} else if scenario.Transport == TransportWebSocket && scenario.Streaming != StreamingNone {
		serverConfig.ServiceImplementations = r.createWebSocketImplementations(scenario)
	} else if hasFeature(scenario.Features, FeatureViews) {
		serverConfig.ServiceImplementations = r.createViewsImplementations(scenario)
	} else if hasFeature(scenario.Features, FeatureBatch) {
		// Batch tests need specialized implementations with correct method names and fields
		serverConfig.ServiceImplementations = r.createBatchImplementations(scenario)
	} else if strings.Contains(scenario.Name, "validation") {
		// Validation tests need specialized implementations with correct field names
		serverConfig.ServiceImplementations = r.createValidationImplementations(scenario)
	} else if strings.Contains(scenario.Name, "unicode") || strings.Contains(scenario.Name, "large_payload") || strings.Contains(scenario.Name, "deeply_nested") {
		// Complex payload tests need specialized implementations with correct field structures
		serverConfig.ServiceImplementations = r.createComplexPayloadImplementations(scenario)
	} else {
		// Default: create basic service implementations for core scenarios
		serverConfig.ServiceImplementations = r.createBasicImplementations(scenario)
	}

	server, err := r.harness.StartServer(ctx, scenario.Name, serverConfig)
	if err != nil {
		return fmt.Errorf("failed to start server %s: %w", scenario.Name, err)
	}

	// Create client
	clientConfig := harness.ClientConfig{
		SourceDir: genDir + "/client",
		ServerURL: server.URL(),
		Transport: string(scenario.Transport),
	}

	client, err := r.harness.StartClient(scenario.Name, clientConfig)
	if err != nil {
		return err
	}

	// Execute test requests based on transport
	switch scenario.Transport {
	case TransportHTTP:
		return r.runHTTPScenario(client, scenario)
	case TransportWebSocket:
		return r.runWebSocketScenario(client, scenario)
	case TransportSSE:
		return r.runSSEScenario(client, scenario)
	default:
		return fmt.Errorf("unknown transport: %s", scenario.Transport)
	}
}

// runHTTPScenario executes HTTP test requests
func (r *ScenarioRunner) runHTTPScenario(client *harness.ClientProcess, scenario Scenario) error {
	// Use a context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if this is a batch request scenario
	if hasFeature(scenario.Features, FeatureBatch) {
		return r.runBatchScenario(ctx, client, scenario)
	}

	for _, req := range scenario.Requests {
		// Check if this is a notification (no expected result or error)
		if req.ExpectedResult == nil && req.ExpectedError == nil {
			// This is a notification - no response expected
			err := client.SendNotification(ctx, req.Method, req.Params)
			if err != nil {
				return fmt.Errorf("notification failed: %w", err)
			}
			continue // Skip response validation for notifications
		}

		// Regular JSON-RPC call with expected response
		resp, err := client.CallHTTP(ctx, req.Method, req.Params)
		if err != nil {
			return fmt.Errorf("HTTP request failed: %w", err)
		}

		// Validate response against expected result/error
		if req.ExpectedError != nil {
			// Expecting an error response
			jsonResp, err := validators.AsJSONRPCResponse(resp)
			if err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			if jsonResp.Error == nil {
				return fmt.Errorf("expected error response but got result")
			}
			if jsonResp.Error.Code != req.ExpectedError.Code {
				return fmt.Errorf("expected error code %d but got %d", req.ExpectedError.Code, jsonResp.Error.Code)
			}
		} else if req.ExpectedResult != nil {
			// Expecting a result response
			jsonResp, err := validators.AsJSONRPCResponse(resp)
			if err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			if jsonResp.Error != nil {
				return fmt.Errorf("expected result but got error: %s", jsonResp.Error.Message)
			}
		}

		// Run additional validators
		for _, validator := range scenario.Validators {
			if err := validator.Validate(resp); err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}
		}
	}

	return nil
}

// runBatchScenario executes batch JSON-RPC requests
func (r *ScenarioRunner) runBatchScenario(ctx context.Context, client *harness.ClientProcess, scenario Scenario) error {
	// For batch requests, we need to extract the individual requests from the test data
	if len(scenario.Requests) != 1 {
		return fmt.Errorf("batch scenario should have exactly one test request containing the batch")
	}

	req := scenario.Requests[0]

	// The params should contain the array of requests
	batchRequests, ok := req.Params.([]any)
	if !ok {
		return fmt.Errorf("batch request params should be an array of requests")
	}

	// Convert to harness.Request objects
	var requests []harness.Request
	for i, batchReq := range batchRequests {
		reqMap, ok := batchReq.(map[string]any)
		if !ok {
			return fmt.Errorf("batch request %d is not a valid JSON-RPC request object", i)
		}

		request := harness.Request{
			JSONRPC: "2.0",
			Method:  reqMap["method"].(string),
			Params:  reqMap["params"],
			ID:      reqMap["id"],
		}
		requests = append(requests, request)
	}

	// Make batch request
	responses, err := client.CallHTTPBatch(ctx, requests)
	if err != nil {
		return fmt.Errorf("batch request failed: %w", err)
	}

	// Validate using batch validators
	for _, validator := range scenario.Validators {
		if err := validator.Validate(responses); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	}

	return nil
}

// runWebSocketScenario executes WebSocket streaming test
func (r *ScenarioRunner) runWebSocketScenario(client *harness.ClientProcess, scenario Scenario) error {
	// Connect WebSocket
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := client.ConnectWebSocket(ctx)
	if err != nil {
		return fmt.Errorf("WebSocket connection failed: %w", err)
	}
	defer client.Stop()

	// Collect all received messages for validation
	var responses []any

	// Execute streaming messages
	for _, req := range scenario.Requests {
		// Send initial request if method specified
		if req.Method != "" {
			request := harness.Request{
				JSONRPC: "2.0",
				Method:  req.Method,
				Params:  req.Params,
				ID:      1,
			}
			if err := client.SendWebSocketMessage(ctx, request); err != nil {
				return fmt.Errorf("failed to send request: %w", err)
			}
		}

		// Process streaming messages
		for i, msg := range req.StreamingMessages {
			switch msg.Direction {
			case DirectionSend:
				if msg.Delay > 0 {
					time.Sleep(time.Duration(msg.Delay) * time.Millisecond)
				}

				// Wrap streaming data in proper JSON-RPC request format
				jsonrpcRequest := harness.Request{
					JSONRPC: "2.0",
					Method:  req.Method,
					Params:  msg.Data,
					ID:      i + 2, // Start from 2 since first request used ID 1
				}

				if err := client.SendWebSocketMessage(ctx, jsonrpcRequest); err != nil {
					return fmt.Errorf("failed to send message: %w", err)
				}

			case DirectionReceive:
				data, err := client.ReceiveWebSocketMessage(ctx)
				if err != nil {
					return fmt.Errorf("failed to receive message: %w", err)
				}
				// Basic validation that we received data
				if data == nil {
					return fmt.Errorf("received empty message")
				}
				// Collect response for validation
				responses = append(responses, data)

				// Validate individual message immediately for streaming validators
				for _, validator := range scenario.Validators {
					if err := validator.Validate(data); err != nil {
						return fmt.Errorf("message validation failed: %w", err)
					}
				}
			}
		}
	}

	// For server streaming, we need to receive messages even if not specified in StreamingMessages
	// Keep receiving until we get all expected messages or timeout
	if scenario.Streaming == StreamingServer {
		for {
			select {
			case <-ctx.Done():
				// Timeout - break out of receive loop
				goto validateResponses
			default:
				data, err := client.ReceiveWebSocketMessage(ctx)
				if err != nil {
					// No more messages or connection closed - break out
					goto validateResponses
				}
				if data != nil {
					responses = append(responses, data)

					// Validate individual message immediately for streaming validators
					for _, validator := range scenario.Validators {
						if err := validator.Validate(data); err != nil {
							return fmt.Errorf("server streaming message validation failed: %w", err)
						}
					}
				}
			}
		}
	}

validateResponses:
	// Final validation for validators that need complete response set
	// (Individual messages were already validated above)
	for _, validator := range scenario.Validators {
		// Check if this validator has a Complete method (like StreamingValidator)
		if completer, ok := validator.(interface{ Complete() error }); ok {
			if err := completer.Complete(); err != nil {
				return fmt.Errorf("final validation failed: %w", err)
			}
		}
	}

	return nil
}

// runSSEScenario executes SSE streaming test
func (r *ScenarioRunner) runSSEScenario(client *harness.ClientProcess, scenario Scenario) error {
	for _, req := range scenario.Requests {
		// Make SSE request
		// For JSON-RPC SSE, the path is always /jsonrpc/sse based on our DSL convention
		sse, err := client.ConnectSSE(context.Background(), "/jsonrpc/sse", req.Params)
		if err != nil {
			return fmt.Errorf("SSE connection failed: %w", err)
		}
		defer sse.Close()

		// Read expected number of events
		for i, expectedMsg := range req.StreamingMessages {
			// Only process DirectionReceive messages for SSE
			if expectedMsg.Direction != DirectionReceive {
				continue
			}

			event, err := sse.ReadEvent()
			if err != nil {
				return fmt.Errorf("failed to read SSE event %d: %w", i+1, err)
			}

			// Basic validation
			if event.Data == "" {
				return fmt.Errorf("received empty SSE event")
			}

			// Parse the SSE event data (should be JSON result data)
			var eventData any
			if err := json.Unmarshal([]byte(event.Data), &eventData); err != nil {
				return fmt.Errorf("failed to parse SSE event JSON: %w", err)
			}

			// Keep the original for validators
			originalEventData := eventData

			// For JSON-RPC notifications, extract the params for comparison
			if eventMap, ok := eventData.(map[string]any); ok {
				if eventMap["jsonrpc"] == "2.0" && eventMap["method"] != nil {
					// This is a JSON-RPC notification, extract params
					if params, ok := eventMap["params"]; ok {
						eventData = params
					}
				}
			}

			// Validate the event content matches expected
			if expectedMsg.Data != nil {
				// Convert both to strings for comparison
				expectedStr := fmt.Sprintf("%v", expectedMsg.Data)
				actualStr := fmt.Sprintf("%v", eventData)
				if expectedStr != actualStr {
					return fmt.Errorf("SSE event %d content mismatch: expected %v, got %v", i+1, expectedMsg.Data, eventData)
				}
			}

			// Run validators on the original event data (full JSON-RPC notification)
			for _, validator := range scenario.Validators {
				if err := validator.Validate(originalEventData); err != nil {
					return fmt.Errorf("SSE validation failed: %w", err)
				}
			}
		}
	}

	return nil
}

// createSSEImplementations creates test implementations for SSE streaming methods.
// This is necessary because the generated example server implementations for streaming
// endpoints are empty stubs that just log and return immediately. For SSE tests to
// work, we need actual implementations that send events through the stream.
//
// The generated code looks like:
//
//	func (s *eventssrvc) Subscribe(ctx context.Context, stream events.SubscribeServerStream) (err error) {
//	    log.Printf(ctx, "events.subscribe")
//	    return  // <-- This doesn't send any events!
//	}
//
// We inject test implementations that actually send the expected test data.
func (r *ScenarioRunner) createSSEImplementations(scenario Scenario) []harness.ServiceImplementation {
	// Extract service name from DSL code
	serviceName := extractServiceNameFromDSL(scenario.DSLCode, "events")
	methodName := "subscribe"
	serviceStruct := serviceName + "srvc"
	methodCapitalized := "Subscribe"

	// Check if the scenario has a payload
	hasPayload := scenario.PayloadType != DataTypeNone

	// Use the same data generator that defines expected content
	// This ensures the server sends exactly what the tests expect
	implementation := r.generateSSEImplementationWithPayload(
		serviceName, methodName, serviceStruct, methodCapitalized,
		scenario.ResultType, hasPayload,
	)

	return []harness.ServiceImplementation{
		{
			ServiceName:    serviceName,
			MethodName:     methodName,
			Implementation: implementation,
		},
	}
}

// createValidationImplementations creates service implementations for validation tests
func (r *ScenarioRunner) createValidationImplementations(scenario Scenario) []harness.ServiceImplementation {
	var implementations []harness.ServiceImplementation

	// Determine the service and method based on scenario name and DSL
	if strings.Contains(scenario.Name, "required_field") {
		// users service with create_user method
		implementation := `// CreateUser implements create_user.
func (s *userssrvc) CreateUser(ctx context.Context, p *users.CreateUserPayload) (res *users.CreateUserResult, err error) {
	log.Printf(ctx, "users.create_user")
	
	// For validation testing, return a simple result
	return &users.CreateUserResult{
		ID:      "generated-id-" + p.Name,
		Created: true,
	}, nil
}`

		implementations = append(implementations, harness.ServiceImplementation{
			ServiceName:    "users",
			MethodName:     "create_user",
			Implementation: implementation,
		})
	} else if strings.Contains(scenario.Name, "validation") {
		// validation service - get method name from the first request
		methodName := "validate" // default
		if len(scenario.Requests) > 0 {
			// Method name is now always simple (not service.method format)
			methodName = scenario.Requests[0].Method
		}

		methodCapitalized := codegen.Goify(methodName, true)

		// Determine the result field name based on the scenario type
		var resultField string
		if strings.Contains(scenario.Name, "http_validation") {
			resultField = "Validated" // HTTP scenarios use "validated" -> "Validated"
		} else {
			resultField = "Valid" // Standalone tests use "valid" -> "Valid"
		}

		var implementation string
		if strings.Contains(scenario.Name, "format") || strings.Contains(methodName, "formats") {
			// Format validation has Email, URL, Date fields
			implementation = fmt.Sprintf(`// %s implements %s.
func (s *validationsrvc) %s(ctx context.Context, p *validation.%sPayload) (res *validation.%sResult, err error) {
	log.Printf(ctx, "validation.%s")
	
	// Check email format - simple validation without strings package
	hasAt := false
	for _, char := range p.Email {
		if char == '@' {
			hasAt = true
			break
		}
	}
	if p.Email != "" && !hasAt {
		// Return a goa validation error which will be mapped to -32602 Invalid params
		return nil, goa.InvalidFieldTypeError("email", p.Email, "valid email address")
	}
	
	// For validation testing, return result field: false when validation passes
	return &validation.%sResult{
		%s: false,
	}, nil
}`, methodCapitalized, methodName, methodCapitalized, methodCapitalized, methodCapitalized, methodName, methodCapitalized, resultField)
		} else if strings.Contains(methodName, "ranges") {
			// Range validation implementation
			implementation = fmt.Sprintf(`// %s implements %s.
func (s *validationsrvc) %s(ctx context.Context, p *validation.%sPayload) (res *validation.%sResult, err error) {
	log.Printf(ctx, "validation.%s")
	
	// For validation testing, return result field: false when validation passes
	return &validation.%sResult{
		%s: false,
	}, nil
}`, methodCapitalized, methodName, methodCapitalized, methodCapitalized, methodCapitalized, methodName, methodCapitalized, resultField)
		} else if strings.Contains(methodName, "strings") {
			// String validation implementation
			implementation = fmt.Sprintf(`// %s implements %s.
func (s *validationsrvc) %s(ctx context.Context, p *validation.%sPayload) (res *validation.%sResult, err error) {
	log.Printf(ctx, "validation.%s")
	
	// For validation testing, return result field: false when validation passes
	return &validation.%sResult{
		%s: false,
	}, nil
}`, methodCapitalized, methodName, methodCapitalized, methodCapitalized, methodCapitalized, methodName, methodCapitalized, resultField)
		} else if strings.Contains(methodName, "enums") {
			// Enum validation implementation
			implementation = fmt.Sprintf(`// %s implements %s.
func (s *validationsrvc) %s(ctx context.Context, p *validation.%sPayload) (res *validation.%sResult, err error) {
	log.Printf(ctx, "validation.%s")
	
	// For validation testing, return result field: false when validation passes
	return &validation.%sResult{
		%s: false,
	}, nil
}`, methodCapitalized, methodName, methodCapitalized, methodCapitalized, methodCapitalized, methodName, methodCapitalized, resultField)
		} else {
			// Required validation has RequiredField, OptionalField fields
			implementation = fmt.Sprintf(`// %s implements %s.
func (s *validationsrvc) %s(ctx context.Context, p *validation.%sPayload) (res *validation.%sResult, err error) {
	log.Printf(ctx, "validation.%s")
	
	// Check if required field is missing or empty - this should trigger a validation error
	if p.RequiredField == nil || (p.RequiredField != nil && *p.RequiredField == "") {
		// Return a goa validation error which will be mapped to -32602 Invalid params
		return nil, goa.MissingFieldError("required_field", "payload")
	}
	
	// For validation testing, return result field: false when validation passes
	return &validation.%sResult{
		%s: false,
	}, nil
}`, methodCapitalized, methodName, methodCapitalized, methodCapitalized, methodCapitalized, methodName, methodCapitalized, resultField)
		}

		implementations = append(implementations, harness.ServiceImplementation{
			ServiceName:    "validation",
			MethodName:     methodName,
			Implementation: implementation,
		})
	}

	return implementations
}

// createBatchImplementations creates test implementations for batch request methods
func (r *ScenarioRunner) createBatchImplementations(scenario Scenario) []harness.ServiceImplementation {
	var implementations []harness.ServiceImplementation

	// Batch scenarios use a "batch" service with "add" and "multiply" methods
	addImplementation := `// Add implements add.
func (s *batchsrvc) Add(ctx context.Context, p *batch.AddPayload) (res int, err error) {
	log.Printf(ctx, "batch.add")
	
	// Add the two numbers from the payload
	return p.A + p.B, nil
}`

	multiplyImplementation := `// Multiply implements multiply.
func (s *batchsrvc) Multiply(ctx context.Context, p *batch.MultiplyPayload) (res int, err error) {
	log.Printf(ctx, "batch.multiply")
	
	// Multiply the two numbers from the payload
	return p.A * p.B, nil
}`

	implementations = append(implementations,
		harness.ServiceImplementation{
			ServiceName:    "batch",
			MethodName:     "add",
			Implementation: addImplementation,
		},
		harness.ServiceImplementation{
			ServiceName:    "batch",
			MethodName:     "multiply",
			Implementation: multiplyImplementation,
		},
	)

	return implementations
}

// createComplexPayloadImplementations creates test implementations for complex payload scenarios
func (r *ScenarioRunner) createComplexPayloadImplementations(scenario Scenario) []harness.ServiceImplementation {
	var implementations []harness.ServiceImplementation

	if strings.Contains(scenario.Name, "unicode") {
		// Unicode scenarios use a "unicode" service with "echo" method
		implementation := `// Echo implements echo.
func (s *unicodesrvc) Echo(ctx context.Context, p *unicode.EchoPayload) (res *unicode.EchoResult, err error) {
	log.Printf(ctx, "unicode.echo")
	
	// Echo the text with unicode handling
	text := p.Text
	if p.Emoji != nil {
		text += " " + *p.Emoji
	}
	
	return &unicode.EchoResult{
		Echoed: text,
		Length: len(text),
	}, nil
}`

		implementations = append(implementations, harness.ServiceImplementation{
			ServiceName:    "unicode",
			MethodName:     "echo",
			Implementation: implementation,
		})
	} else if strings.Contains(scenario.Name, "large_payload") {
		// Large payload scenarios use a "large" service with "process" method
		implementation := `// Process implements process.
func (s *largesrvc) Process(ctx context.Context, p *large.ProcessPayload) (res *large.ProcessResult, err error) {
	log.Printf(ctx, "large.process")
	
	// Process the large payload
	totalSize := int64(0)
	for _, item := range p.Data {
		totalSize += int64(len(item))
	}
	
	return &large.ProcessResult{
		Count: len(p.Data),
		Size:  totalSize,
	}, nil
}`

		implementations = append(implementations, harness.ServiceImplementation{
			ServiceName:    "large",
			MethodName:     "process",
			Implementation: implementation,
		})
	} else if strings.Contains(scenario.Name, "deeply_nested") {
		// Deeply nested scenarios use a "complex" service with "process" method
		implementation := `// Process implements process.
func (s *complex_srvc) Process(ctx context.Context, p *complex_.Level1) (res *complex_.Level1, err error) {
	log.Printf(ctx, "complex.process")
	
	// Process the deeply nested structure - echo it back with some modifications
	result := &complex_.Level1{
		Nested: p.Nested,
		Map:    make(map[string]*complex_.Level2),
	}
	
	// Copy the map
	if p.Map != nil {
		for k, v := range p.Map {
			result.Map[k] = v
		}
	}
	
	return result, nil
}`

		implementations = append(implementations, harness.ServiceImplementation{
			ServiceName:    "complex_",
			MethodName:     "process",
			Implementation: implementation,
		})
	}

	return implementations
}

// hasFeature checks if a scenario has a specific feature
func hasFeature(features []Feature, feature Feature) bool {
	for _, f := range features {
		if f == feature {
			return true
		}
	}
	return false
}

// createWebSocketImplementations creates test implementations for WebSocket streaming methods
func (r *ScenarioRunner) createWebSocketImplementations(scenario Scenario) []harness.ServiceImplementation {
	// Determine service and method names from the first request
	if len(scenario.Requests) == 0 {
		return nil
	}

	// Method name is now always simple (not service.method format)
	methodName := scenario.Requests[0].Method

	// Extract service name from DSL code
	serviceName := extractServiceNameFromDSL(scenario.DSLCode, "streaming") // default to "streaming"

	// Convert to proper casing
	serviceStruct := serviceName + "srvc"
	methodCapitalized := toCamelCase(methodName)

	var implementation string
	switch scenario.Streaming {
	case StreamingServer:
		implementation = r.generateWebSocketServerStreamingImplementation(
			serviceName, methodName, serviceStruct, methodCapitalized,
			scenario.ResultType,
		)
	case StreamingClient:
		implementation = r.generateWebSocketClientStreamingImplementation(
			serviceName, methodName, serviceStruct, methodCapitalized,
			scenario.PayloadType, scenario.ResultType,
		)
	case StreamingBidirectional:
		implementation = r.generateWebSocketBidirectionalImplementation(
			serviceName, methodName, serviceStruct, methodCapitalized,
			scenario.PayloadType, scenario.ResultType,
		)
	default:
		return nil
	}

	// Different injection strategies for different streaming types
	var implementations []harness.ServiceImplementation

	switch scenario.Streaming {
	case StreamingClient:
		// For client streaming, override both the service method and HandleStream
		// HandleStream needs proper error handling for stream establishment messages
		implementations = []harness.ServiceImplementation{
			{
				ServiceName:    serviceName,
				MethodName:     methodName,
				Implementation: implementation,
			},
			{
				ServiceName:    serviceName,
				MethodName:     "HandleStream",
				Implementation: r.generateClientStreamingHandleStreamImplementation(serviceName, methodName, serviceStruct, methodCapitalized),
			},
		}
	case StreamingServer:
		// For server streaming, override both the service method and HandleStream
		implementations = []harness.ServiceImplementation{
			{
				ServiceName:    serviceName,
				MethodName:     methodName,
				Implementation: implementation,
			},
			{
				ServiceName:    serviceName,
				MethodName:     "HandleStream",
				Implementation: r.generateServerStreamingHandleStreamImplementation(serviceName, methodName, serviceStruct),
			},
		}
	case StreamingBidirectional:
		// For bidirectional streaming, override both the service method and HandleStream
		// HandleStream needs to dispatch JSON-RPC calls to the BidirectionalStream method
		implementations = []harness.ServiceImplementation{
			{
				ServiceName:    serviceName,
				MethodName:     methodName,
				Implementation: implementation,
			},
			{
				ServiceName:    serviceName,
				MethodName:     "HandleStream",
				Implementation: r.generateBidirectionalHandleStreamImplementation(serviceName, methodName, serviceStruct, methodCapitalized, scenario.PayloadType, scenario.ResultType),
			},
		}
	default:
		implementations = []harness.ServiceImplementation{
			{
				ServiceName:    serviceName,
				MethodName:     methodName,
				Implementation: implementation,
			},
		}
	}

	return implementations
}

// generateServerStreamingHandleStreamImplementation generates HandleStream implementation for server streaming
func (r *ScenarioRunner) generateServerStreamingHandleStreamImplementation(serviceName, methodName, serviceStruct string) string {
	return fmt.Sprintf(`// HandleStream handles the JSON-RPC WebSocket connection for server streaming.
func (s *%s) HandleStream(ctx context.Context, stream %s.Stream) error {
	fmt.Println("%s.HandleStream")
	
	// Process incoming requests - the server_stream method will be called
	// when a request is received, and it will handle the streaming
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := stream.Recv(ctx); err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}
		}
	}
}`,
		serviceStruct, serviceName, serviceName,
	)
}

// createErrorImplementations creates test implementations for error handling methods.
// This allows tests to trigger specific errors based on request parameters.
func (r *ScenarioRunner) createErrorImplementations(scenario Scenario) []harness.ServiceImplementation {
	// Extract service name from DSL code, handle the underscore suffix for Go package conflicts
	baseName := extractServiceNameFromDSL(scenario.DSLCode, "errors")
	serviceName := baseName + "_" // Goa appends underscore to service names that conflict with Go packages
	serviceStruct := "errors_srvc"

	// Determine the method from the scenario requests
	methodName := "test_error" // default
	methodCapitalized := "TestError"

	// Check if this is the custom errors scenario with "process" method
	if len(scenario.Requests) > 0 && scenario.Requests[0].Method == "process" {
		methodName = "process"
		methodCapitalized = "Process"
	} else if len(scenario.Requests) > 0 && scenario.Requests[0].Method == "error_stream" {
		methodName = "error_stream"
		methodCapitalized = "ErrorStream"
	}

	// Check if this scenario has custom errors
	hasCustomErrors := false
	for _, req := range scenario.Requests {
		if req.ExpectedError != nil && (req.ExpectedError.Code == -32001 || req.ExpectedError.Code == -32002) {
			hasCustomErrors = true
			break
		}
	}

	implementation := r.generateErrorImplementation(
		serviceName, methodName, serviceStruct, methodCapitalized, hasCustomErrors,
	)

	implementations := []harness.ServiceImplementation{
		{
			ServiceName:    serviceName,
			MethodName:     methodName,
			Implementation: implementation,
		},
	}

	// If this is a streaming error method, also inject HandleStream
	if methodName == "error_stream" {
		handleStreamImpl := r.generateErrorHandleStreamImplementation(serviceName)
		implementations = append(implementations, harness.ServiceImplementation{
			ServiceName:    serviceName,
			MethodName:     "HandleStream",
			Implementation: handleStreamImpl,
		})
	}

	return implementations
}

// generateErrorImplementation generates the error handling implementation
func (r *ScenarioRunner) generateErrorImplementation(serviceName, methodName, serviceStruct, methodCapitalized string, hasCustomErrors bool) string {

	// Special handling for streaming error methods
	if methodName == "error_stream" {
		return fmt.Sprintf(`// %s implements %s.
func (s *%s) %s(ctx context.Context, p *errors_.%sPayload, stream errors_.%sServerStream) error {
	log.Printf(ctx, "errors_.%s")
	
	// Bidirectional streaming method with stream parameter
	// This method gets called for each incoming payload
	
	// Check if this should trigger an error
	if p.Data == "trigger_error" {
		// Return a simple error - the framework will handle JSON-RPC error mapping
		return fmt.Errorf("internal error")
	}
	
	// Send response back through the stream
	return stream.Send(&errors_.%sResult{
		ID:   p.ID,
		Data: "processed: " + p.Data,
	})
}`,
			methodCapitalized, methodName, serviceStruct, methodCapitalized,
			methodCapitalized, methodCapitalized, methodName, methodCapitalized,
		)
	}

	if hasCustomErrors {
		// For custom errors with the process method
		if methodName == "process" {
			return fmt.Sprintf(`// %s implements %s.
func (s *%s) %s(ctx context.Context, p *errors_.%sPayload) (res *errors_.ProcessResult, err error) {
	log.Printf(ctx, "errors_.%s")
	
	// Trigger different errors based on the "action" parameter  
	switch p.Action {
	case "unauthorized":
		// Return the Unauthorized error
		return nil, &errors_.Unauthorized{Reason: "unauthorized"}
	case "not_found":
		// Return the NotFound error  
		return nil, &errors_.NotFound{Resource: "resource", ID: "123"}
	case "conflict":
		// Return the Conflict error
		return nil, &errors_.Conflict{Message: "conflict"}
	case "success":
		return &errors_.ProcessResult{Status: "success"}, nil
	default:
		// Default to success
		return &errors_.ProcessResult{Status: "ok"}, nil
	}
}`,
				methodCapitalized, methodName, serviceStruct, methodCapitalized,
				methodCapitalized, methodName,
			)
		}

		// For test_error method with Trigger field
		return fmt.Sprintf(`// %s implements %s.
func (s *%s) %s(ctx context.Context, p *errors_.%sPayload) (res string, err error) {
	log.Printf(ctx, "errors_.%s")
	
	// Trigger different errors based on the "trigger" parameter  
	switch p.Trigger {
	case "validation":
		// Return a validation error
		return "", &errors_.ValidationError{Field: "trigger", Message: "validation error"}
	case "notfound":
		// Return a not found error
		return "", &errors_.NotFoundError{Resource: "test", ID: "123"}
	case "success":
		return "success", nil
	default:
		// Default to success
		return "test result", nil
	}
}`,
			methodCapitalized, methodName, serviceStruct, methodCapitalized,
			methodCapitalized, methodName,
		)
	}

	// Standard errors implementation (no custom error types)
	return fmt.Sprintf(`// %s implements %s.
func (s *%s) %s(ctx context.Context, p *errors_.%sPayload) (res string, err error) {
	log.Printf(ctx, "errors_.%s")
	
	// Trigger different errors based on the "trigger" parameter
	// These will be mapped to JSON-RPC error codes by the framework
	switch p.Trigger {
	case "parse":
		// Parse error would typically be handled by the JSON-RPC layer
		// We can't really trigger it from here, so return a generic error
		return "", fmt.Errorf("parse error")
	case "invalid":
		// Invalid request - will be mapped to -32600
		return "", fmt.Errorf("invalid request")
	case "internal":
		// Internal error - any generic error gets mapped to -32603
		return "", fmt.Errorf("internal error")
	case "success":
		return "success", nil
	default:
		// Default to success
		return "test result", nil
	}
}`,
		methodCapitalized, methodName, serviceStruct, methodCapitalized,
		methodCapitalized, methodName,
	)
}

// generateSSEImplementation generates the streaming implementation using the same
// data that createSSEData uses for validation. This ensures consistency.
func (r *ScenarioRunner) generateSSEImplementation(serviceName, methodName, serviceStruct, methodCapitalized string, resultType DataType) string {
	// Use SSETestData to get the implementation code
	testData := SSETestData{ResultType: resultType}

	// For JSON-RPC SSE in the new architecture:
	// 1. The service method returns a result (no stream parameter)
	// 2. The endpoint receives a stream and calls the service method
	// 3. The endpoint then uses the stream to send the result
	// 4. For testing, we need to intercept at the endpoint level
	//
	// Since the test harness generates service implementations, we need to
	// work with what the endpoint expects. However, the actual streaming
	// happens in the endpoint, not the service.
	//
	// We'll generate a service that works with the generated code and
	// provide a custom endpoint handler that does the streaming.

	return fmt.Sprintf(`// %s implements %s.
func (s *%s) %s(ctx context.Context) (res *%s.%sResult, err error) {
	log.Printf(ctx, "%s.%s")
	// For JSON-RPC SSE, we just return a result
	// The endpoint will handle streaming
	res = %s
	return
}

// NewSubscribeEndpoint returns a custom endpoint that streams test data.
// This overrides the default endpoint to provide test-specific streaming behavior.
func NewSubscribeEndpoint(s %s.Service) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		// Extract the stream from the endpoint input
		input := req.(*%s.SubscribeEndpointInput)
		stream := input.Stream
		
		// Send 5 test events
		for i := 1; i <= 5; i++ {
			event := %s
			if err := stream.Send(ctx, event); err != nil {
				return nil, err
			}
			// Small delay between events
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(10 * time.Millisecond):
			}
		}
		
		return nil, nil
	}
}`,
		methodCapitalized, methodName, serviceStruct, methodCapitalized,
		serviceName, methodCapitalized, serviceName, methodName,
		testData.GenerateImplementationCode(serviceName),
		serviceName, serviceName, testData.GenerateImplementationCode(serviceName),
	)
}

func (r *ScenarioRunner) generateSSEImplementationWithPayload(serviceName, methodName, serviceStruct, methodCapitalized string, resultType DataType, hasPayload bool) string {
	// Use SSETestData to get the implementation code
	testData := SSETestData{ResultType: resultType}

	// Generate the appropriate method signature based on whether there's a payload
	var methodSignature string
	if hasPayload {
		methodSignature = fmt.Sprintf("func (s *%s) %s(ctx context.Context, p *%s.%sPayload, stream %s.%sServerStream) (err error)",
			serviceStruct, methodCapitalized, serviceName, methodCapitalized, serviceName, methodCapitalized)
	} else {
		methodSignature = fmt.Sprintf("func (s *%s) %s(ctx context.Context, stream %s.%sServerStream) (err error)",
			serviceStruct, methodCapitalized, serviceName, methodCapitalized)
	}

	// Generate the implementation that sends data matching createSSEData
	return fmt.Sprintf(`// %s implements %s.
%s {
	log.Printf(ctx, "%s.%s")
	// Send 5 test events using the same data generator as the test expectations
	for i := 1; i <= 5; i++ {
		event := %s
		if err := stream.Send(ctx, event); err != nil {
			return err
		}
		// Small delay between events
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}
	return nil
}`,
		methodCapitalized, methodName, methodSignature, serviceName, methodName,
		testData.GenerateImplementationCode(serviceName),
	)
}

// toCamelCase converts snake_case to CamelCase
func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

// generateWebSocketServerStreamingImplementation generates server streaming service method implementation
func (r *ScenarioRunner) generateWebSocketServerStreamingImplementation(serviceName, methodName, serviceStruct, methodCapitalized string, resultType DataType) string {
	// Generate result creation based on result type
	var resultCreation string
	switch resultType {
	case DataTypePrimitive:
		resultCreation = fmt.Sprintf(`&%s.%sResult{
			ID:   fmt.Sprintf("req-%%d", i+1),
			Data: fmt.Sprintf("message %%d", i+1),
		}`, serviceName, methodCapitalized)
	case DataTypeArray:
		resultCreation = fmt.Sprintf(`&%s.%sResult{
			ID:    fmt.Sprintf("req-%%d", i+1),
			Items: []string{fmt.Sprintf("item%%d-1", i+1), fmt.Sprintf("item%%d-2", i+1)},
		}`, serviceName, methodCapitalized)
	case DataTypeObject:
		resultCreation = fmt.Sprintf(`&%s.%sResult{
			ID:     fmt.Sprintf("req-%%d", i+1),
			Field1: fmt.Sprintf("Message %%d", i+1),
			Field2: func() *int { v := i+1; return &v }(),
			Field3: func() *bool { v := (i+1)%%2 == 0; return &v }(),
		}`, serviceName, methodCapitalized)
	case DataTypeUserType:
		resultCreation = fmt.Sprintf(`&%s.%sResult{
			ID:      fmt.Sprintf("req-%%d", i+1),
			UserID:  fmt.Sprintf("user%%d", i+1),
			Name:    fmt.Sprintf("Stream User %%d", i+1),
			Email:   func() *string { s := fmt.Sprintf("stream%%d@example.com", i+1); return &s }(),
		}`, serviceName, methodCapitalized)
	case DataTypeComplex:
		resultCreation = fmt.Sprintf(`&%s.%sResult{
			ID:       fmt.Sprintf("req-%%d", i+1),
			Sequence: i + 1,
			Data: map[string]any{
				"value": fmt.Sprintf("complex-%%d", i+1),
			},
			Metadata: map[string]any{
				"index": i + 1,
				"type":  "stream",
			},
		}`, serviceName, methodCapitalized)
	default:
		resultCreation = fmt.Sprintf(`&%s.%sResult{
			ID:   fmt.Sprintf("req-%%d", i+1),
			Data: fmt.Sprintf("data-%%d", i+1),
		}`, serviceName, methodCapitalized)
	}

	// For JSON-RPC WebSocket server streaming with non-streaming payload
	// Method receives payload and stream for sending multiple results
	return fmt.Sprintf(`// %s implements %s.
func (s *%s) %s(ctx context.Context, p *%s.%sPayload, stream %s.%sServerStream) (err error) {
	log.Printf(ctx, "%s.%s with count: %%d", p.Count)
	
	// Send multiple results based on the count requested
	for i := 0; i < p.Count; i++ {
		result := %s
		if err := stream.Send(result); err != nil {
			return err
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}`,
		methodCapitalized, methodName, serviceStruct, methodCapitalized,
		serviceName, methodCapitalized, serviceName, methodCapitalized,
		serviceName, methodName,
		resultCreation,
	)
}

// generateHandleStreamImplementation generates HandleStream implementation for server streaming
func (r *ScenarioRunner) generateHandleStreamImplementation(serviceName, methodName, serviceStruct, methodCapitalized string, resultType DataType) string {
	// For server streaming with non-streaming payload, HandleStream just processes incoming requests
	// The actual streaming happens in the service method when it receives the payload
	return fmt.Sprintf(`// HandleStream handles the JSON-RPC WebSocket connection.
func (s *%s) HandleStream(ctx context.Context, stream %s.Stream) error {
	log.Printf(ctx, "%s.HandleStream")
	
	// Process incoming requests - the server_stream method will be called
	// when a request is received, and it will handle the streaming
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := stream.Recv(ctx); err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}
		}
	}
}`, serviceStruct, serviceName, serviceName)
}

// generateBidirectionalHandleStreamImplementation generates a HandleStream implementation
// for bidirectional streaming that processes incoming JSON-RPC requests by calling the
// service's BidirectionalStream method and sending responses back.
func (r *ScenarioRunner) generateBidirectionalHandleStreamImplementation(serviceName, methodName, serviceStruct, methodCapitalized string, payloadType, resultType DataType) string {
	return fmt.Sprintf(`// HandleStream handles the JSON-RPC WebSocket connection for bidirectional streaming.
func (s *%s) HandleStream(ctx context.Context, stream %s.Stream) error {
	log.Printf(ctx, "%s.HandleStream starting bidirectional processing")
	
	// Process incoming requests via Recv which dispatches to the appropriate method
	// For bidirectional streaming, each incoming message should trigger the BidirectionalStream method
	// The first message without params establishes the stream, subsequent messages process data
	for {
		select {
		case <-ctx.Done():
			log.Printf(ctx, "%s.HandleStream context cancelled")
			return ctx.Err()
		default:
			// Call Recv to process incoming JSON-RPC requests
			// This will automatically dispatch to the BidirectionalStream method
			// The Recv method handles messages with and without params appropriately
			if err := stream.Recv(ctx); err != nil {
				log.Printf(ctx, "%s.HandleStream recv error: %%v", err)
				// For bidirectional streaming, ignore missing payload errors from stream establishment
				if err.Error() == "handler error for %s: missing required payload" {
					log.Printf(ctx, "%s.HandleStream ignoring stream establishment message")
					continue
				}
				return err
			}
		}
	}
}`,
		serviceStruct, serviceName, serviceName,
		serviceName,
		serviceName,
		methodName,
		serviceName,
	)
}

// generateWebSocketClientStreamingImplementation generates client streaming implementation
func (r *ScenarioRunner) generateWebSocketClientStreamingImplementation(serviceName, methodName, serviceStruct, methodCapitalized string, payloadType, resultType DataType) string {
	// For JSON-RPC WebSocket client streaming, the service method takes only payload parameter
	// No stream parameter - the method processes individual payloads and returns final result
	// The method is called by stream.Recv() in HandleStream for each incoming payload
	// All streaming methods use structured payloads (never raw primitives)
	var payloadParam string
	if payloadType == DataTypeNone {
		payloadParam = ""
	} else {
		payloadParam = fmt.Sprintf("p *%s.%sPayload", serviceName, methodCapitalized)
	}

	// Client streaming returns a final result
	var resultReturn string
	if resultType == DataTypeNone {
		resultReturn = "nil, nil"
	} else {
		// Generate result based on type
		// For client streaming, we accumulate payloads and return a final result
		// The result structure depends on the result type, not the payload type
		switch resultType {
		case DataTypePrimitive:
			// For primitive results, we need to access the appropriate field from payload
			var dataAccess string
			switch payloadType {
			case DataTypePrimitive:
				dataAccess = "p.Data"
			case DataTypeArray:
				dataAccess = `"accumulated"`
			case DataTypeObject:
				dataAccess = `"field1: " + p.Field1`
			case DataTypeUserType:
				dataAccess = `"user: " + p.Name`
			case DataTypeComplex:
				dataAccess = `"complex data"`
			default:
				dataAccess = `"final"`
			}
			resultReturn = fmt.Sprintf(`&%s.%sResult{ID: p.ID, Data: "final result: " + %s}, nil`, serviceName, methodCapitalized, dataAccess)
		case DataTypeArray:
			resultReturn = fmt.Sprintf(`&%s.%sResult{ID: p.ID, Items: []string{"final1", "final2"}}, nil`, serviceName, methodCapitalized)
		case DataTypeObject:
			resultReturn = fmt.Sprintf(`&%s.%sResult{ID: p.ID, Field1: "final"}, nil`, serviceName, methodCapitalized)
		case DataTypeUserType:
			resultReturn = fmt.Sprintf(`&%s.%sResult{ID: p.ID, UserID: "u123", Name: "Final User"}, nil`, serviceName, methodCapitalized)
		case DataTypeComplex:
			resultReturn = fmt.Sprintf(`&%s.%sResult{ID: p.ID, Sequence: 999}, nil`, serviceName, methodCapitalized)
		default:
			resultReturn = fmt.Sprintf(`&%s.%sResult{ID: p.ID, Data: "final"}, nil`, serviceName, methodCapitalized)
		}
	}

	return fmt.Sprintf(`// %s implements %s (client streaming).
func (s *%s) %s(ctx context.Context, %s) (*%s.%sResult, error) {
	log.Printf(ctx, "%s.%s")
	// In a real implementation, you would accumulate payloads
	// For testing, we just return a final result
	return %s
}`,
		methodCapitalized, methodName, serviceStruct, methodCapitalized,
		payloadParam, serviceName, methodCapitalized, serviceName, methodName,
		resultReturn,
	)
}

// generateWebSocketBidirectionalImplementation generates bidirectional streaming implementation
func (r *ScenarioRunner) generateWebSocketBidirectionalImplementation(serviceName, methodName, serviceStruct, methodCapitalized string, payloadType, resultType DataType) string {
	// For JSON-RPC WebSocket bidirectional streaming, the service method takes payload and stream parameters
	// The method is called by stream.Recv() in HandleStream for each incoming payload
	// The stream is used to send results back to the client
	// All streaming methods use structured payloads (never raw primitives)
	var payloadParam string
	if payloadType == DataTypeNone {
		payloadParam = ""
	} else {
		payloadParam = fmt.Sprintf("p *%s.%sPayload, ", serviceName, methodCapitalized)
	}

	streamParam := fmt.Sprintf("stream %s.%sServerStream", serviceName, methodCapitalized)

	// For bidirectional streaming, we don't return a result directly - we send via stream
	// Generate response based on result type
	var sendCode string
	switch resultType {
	case DataTypePrimitive:
		sendCode = fmt.Sprintf(`// Echo back the payload data
	return stream.Send(&%s.%sResult{
		ID:   p.ID,
		Data: "echo: " + p.Data,
	})`, serviceName, methodCapitalized)
	case DataTypeArray:
		sendCode = fmt.Sprintf(`// Send back array result
	return stream.Send(&%s.%sResult{
		ID:    p.ID,
		Items: append([]string{"echo"}, p.Items...),
	})`, serviceName, methodCapitalized)
	case DataTypeObject:
		sendCode = fmt.Sprintf(`// Send back object result
	return stream.Send(&%s.%sResult{
		ID:     p.ID,
		Field1: "echo: " + p.Field1,
		Field2: p.Field2,
		Field3: p.Field3,
	})`, serviceName, methodCapitalized)
	case DataTypeUserType:
		sendCode = fmt.Sprintf(`// Send back user type result
	return stream.Send(&%s.%sResult{
		ID:     p.ID,
		UserID: p.UserID,
		Name:   "echo: " + p.Name,
	})`, serviceName, methodCapitalized)
	case DataTypeComplex:
		sendCode = fmt.Sprintf(`// Send back complex result
	return stream.Send(&%s.%sResult{
		ID:       p.ID,
		Sequence: p.Sequence + 1000,
		Data:     p.Data,
	})`, serviceName, methodCapitalized)
	default:
		sendCode = fmt.Sprintf(`// Send back default result
	return stream.Send(&%s.%sResult{
		ID:   p.ID,
		Data: "echo",
	})`, serviceName, methodCapitalized)
	}

	return fmt.Sprintf(`// %s implements %s (bidirectional streaming).
func (s *%s) %s(ctx context.Context, %s%s) (err error) {
	log.Printf(ctx, "%s.%s")
	// For bidirectional streaming, echo back the payload
	%s
}`,
		methodCapitalized, methodName, serviceStruct, methodCapitalized,
		payloadParam, streamParam, serviceName, methodName,
		sendCode,
	)
}

// generateBidirectionalResultType generates the result type signature for bidirectional streaming
func (r *ScenarioRunner) generateBidirectionalResultType(serviceName, methodCapitalized string, resultType DataType) string {
	switch resultType {
	case DataTypePrimitive:
		return "string"
	case DataTypeArray:
		return "[]string"
	case DataTypeObject, DataTypeUserType:
		return fmt.Sprintf("*%s.BidirectionalStreamResult", serviceName)
	default:
		return "string"
	}
}

// generateBidirectionalPayloadResultResponse generates response code for payload/result pattern
func (r *ScenarioRunner) generateBidirectionalPayloadResultResponse(serviceName, methodCapitalized string, payloadType, resultType DataType) string {
	// Generate result struct initialization based on result type
	// Field names must match the DSL attributes defined in createWebSocketStreamingType
	switch resultType {
	case DataTypePrimitive:
		return fmt.Sprintf(`res = &%s.%sResult{
		ID:   p.ID,
		Data: "echo: " + p.Data,
	}`, serviceName, methodCapitalized)
	case DataTypeArray:
		return fmt.Sprintf(`res = &%s.%sResult{
		ID:    p.ID,
		Items: append([]string{"echo:"}, p.Items...),
	}`, serviceName, methodCapitalized)
	case DataTypeObject:
		return fmt.Sprintf(`res = &%s.%sResult{
		ID:     p.ID,
		Field1: "echo: " + p.Field1,
		Field2: p.Field2,
		Field3: p.Field3,
	}`, serviceName, methodCapitalized)
	case DataTypeUserType:
		return fmt.Sprintf(`res = &%s.%sResult{
		ID:     p.ID,
		UserID: p.UserID,
		Name:   "echo: " + p.Name,
		Email:  p.Email,
	}`, serviceName, methodCapitalized)
	case DataTypeComplex:
		return fmt.Sprintf(`res = &%s.%sResult{
		ID:       p.ID,
		Sequence: p.Sequence + 1000, // Modified sequence to show processing
		Data:     p.Data,
		Metadata: p.Metadata,
	}`, serviceName, methodCapitalized)
	default:
		return fmt.Sprintf(`res = &%s.%sResult{
		ID:   p.ID,
		Data: "echo: " + p.Data,
	}`, serviceName, methodCapitalized)
	}
}

// generateBidirectionalResponse generates appropriate response code for bidirectional streaming
func (r *ScenarioRunner) generateBidirectionalResponse(serviceName, methodCapitalized string, resultType DataType) string {
	switch resultType {
	case DataTypePrimitive:
		return `if err := stream.Send("echo response"); err != nil {
			return err
		}`
	case DataTypeArray:
		return `if err := stream.Send([]string{"echo", "response"}); err != nil {
			return err
		}`
	case DataTypeObject:
		return fmt.Sprintf(`// Create a response object - actual fields depend on generated types
		var result %s.%sResult
		if err := stream.Send(&result); err != nil {
			return err
		}`, serviceName, methodCapitalized)
	case DataTypeUserType:
		return fmt.Sprintf(`// Create a user type response - actual structure depends on generated types
		result := &%s.%sResult{}
		if err := stream.Send(result); err != nil {
			return err
		}`, serviceName, methodCapitalized)
	case DataTypeComplex:
		return fmt.Sprintf(`// Create a complex response - actual structure depends on generated types
		var result %s.%sResult
		if err := stream.Send(&result); err != nil {
			return err
		}`, serviceName, methodCapitalized)
	default:
		return `// Send empty response for unknown type
		if err := stream.Send(nil); err != nil {
			return err
		}`
	}
}

// generateErrorHandleStreamImplementation generates HandleStream implementation for error handling tests
func (r *ScenarioRunner) generateErrorHandleStreamImplementation(serviceName string) string {
	return fmt.Sprintf(`// HandleStream handles the JSON-RPC WebSocket connection for error handling tests.
func (s *errors_srvc) HandleStream(ctx context.Context, stream %s.Stream) error {
	log.Printf(ctx, "%s.HandleStream")
	
	// Simple HandleStream that processes incoming requests
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Recv automatically dispatches JSON-RPC requests to service methods
			err := stream.Recv(ctx)
			if err != nil {
				return err
			}
		}
	}
}`,
		serviceName, serviceName,
	)
}

// generateClientStreamingHandleStreamImplementation generates HandleStream implementation for client streaming
// with proper error handling for stream establishment messages
func (r *ScenarioRunner) generateClientStreamingHandleStreamImplementation(serviceName, methodName, serviceStruct, methodCapitalized string) string {
	return fmt.Sprintf(`// HandleStream handles the JSON-RPC WebSocket connection for client streaming.
func (s *%s) HandleStream(ctx context.Context, stream %s.Stream) error {
	log.Printf(ctx, "%s.HandleStream starting client streaming processing")
	
	// Process incoming requests via Recv which dispatches to the appropriate method
	// For client streaming, multiple incoming messages get processed by the %s method
	// The first message without params establishes the stream, subsequent messages contain data
	for {
		select {
		case <-ctx.Done():
			log.Printf(ctx, "%s.HandleStream context cancelled")
			return ctx.Err()
		default:
			// Call Recv to process incoming JSON-RPC requests
			// This will automatically dispatch to the %s method
			// The Recv method handles messages with and without params appropriately
			if err := stream.Recv(ctx); err != nil {
				log.Printf(ctx, "%s.HandleStream recv error: %%v", err)
				// For client streaming, ignore missing payload errors from stream establishment
				if err.Error() == "handler error for %s: missing required payload" {
					log.Printf(ctx, "%s.HandleStream ignoring stream establishment message")
					continue
				}
				return err
			}
		}
	}
}`,
		serviceStruct, serviceName, serviceName, methodCapitalized, serviceName, methodCapitalized, serviceName, methodName, serviceName,
	)
}

// createViewsImplementations creates test implementations for methods that return views.
// The generated example server implementations don't return actual data, so we need
// to inject implementations that return proper view data for testing.
func (r *ScenarioRunner) createViewsImplementations(scenario Scenario) []harness.ServiceImplementation {
	// Extract service name from DSL code
	serviceName := extractServiceNameFromDSL(scenario.DSLCode, "users")
	methodName := "get"
	serviceStruct := serviceName + "srvc"
	methodCapitalized := "Get"

	implementation := r.generateViewsImplementation(
		serviceName, methodName, serviceStruct, methodCapitalized,
	)

	return []harness.ServiceImplementation{
		{
			ServiceName:    serviceName,
			MethodName:     methodName,
			Implementation: implementation,
		},
	}
}

// generateViewsImplementation generates the views implementation that returns
// data matching what the tests expect
func (r *ScenarioRunner) generateViewsImplementation(serviceName, methodName, serviceStruct, methodCapitalized string) string {
	// The service method returns the service type, not the view type
	return fmt.Sprintf(`// %s implements %s.
func (s *%s) %s(ctx context.Context, p *%s.%sPayload) (res *%s.User, view string, err error) {
	log.Printf(ctx, "%s.%s")
	
	// Helper function to get string pointer
	stringPtr := func(s string) *string {
		return &s
	}
	
	// Create a user result with all fields populated
	// The generated User type has ID, Name, Email, Profile fields (all capitalized)
	// Profile is an anonymous struct, not a named type
	user := &%s.User{
		ID:    p.ID,
		Name:  "Test User",
		Email: stringPtr("test@example.com"),
		Profile: &struct {
			Bio    *string
			Avatar *string
		}{
			Bio:    stringPtr("Test bio"),
			Avatar: stringPtr("test-avatar.png"),
		},
	}
	
	// Return the requested view (default if not specified)
	requestedView := "default"
	if p.View != nil {
		requestedView = *p.View
	}
	
	return user, requestedView, nil
}`,
		methodCapitalized, methodName, serviceStruct, methodCapitalized,
		serviceName, methodCapitalized, serviceName,
		serviceName, methodName,
		serviceName,
	)
}

// createBasicImplementations creates implementations for basic/core service scenarios
// These scenarios typically involve simple request/response patterns without streaming,
// custom errors, or views - just basic JSON-RPC method calls.
func (r *ScenarioRunner) createBasicImplementations(scenario Scenario) []harness.ServiceImplementation {
	var implementations []harness.ServiceImplementation

	// Only create implementations for methods that actually exist in DSL
	// For method_not_found tests, we shouldn't create implementations for non-existent methods

	// Create implementations based on the scenario's specific needs
	serviceMethodMap := make(map[string][]string)

	// For method_not_found tests, we need implementations for methods that DO exist in DSL
	// For regular tests, we need implementations for the requested methods
	// Extract the actual service and method names from DSL code instead of guessing

	serviceName := extractServiceNameFromDSL(scenario.DSLCode, "test")

	// For each requested method, determine what implementations we need
	for _, req := range scenario.Requests {
		methodName := req.Method

		// Skip nonexistent methods - they're meant to fail
		if strings.Contains(methodName, "nonexistent") {
			continue
		}

		// Skip batch method - it's handled separately
		if methodName == "batch" {
			continue
		}

		// Use the service from DSL and the requested method
		serviceMethodMap[serviceName] = []string{methodName}
	}

	// Special case: if no valid methods were found (e.g., method_not_found test),
	// we need to add the methods that DO exist in the DSL so the server can start
	if len(serviceMethodMap) == 0 {
		// Extract method names from DSL - look for Method("name", func() patterns
		if strings.Contains(scenario.DSLCode, `Method("echo"`) {
			serviceMethodMap[serviceName] = []string{"echo"}
		} else if strings.Contains(scenario.DSLCode, `Method("call"`) {
			serviceMethodMap[serviceName] = []string{"call"}
		} else if strings.Contains(scenario.DSLCode, `Method("validate"`) {
			serviceMethodMap[serviceName] = []string{"validate"}
		}
	}

	// Generate implementations for each service
	for serviceName, methods := range serviceMethodMap {
		serviceStruct := serviceName + "srvc"

		// Remove duplicates from methods
		uniqueMethods := make(map[string]bool)
		for _, method := range methods {
			uniqueMethods[method] = true
		}

		// Generate implementation for each unique method
		for methodName := range uniqueMethods {
			methodCapitalized := toCamelCase(methodName)

			implementation := r.generateBasicImplementation(
				serviceName, methodName, serviceStruct, methodCapitalized, scenario,
			)

			implementations = append(implementations, harness.ServiceImplementation{
				ServiceName:    serviceName,
				MethodName:     methodName,
				Implementation: implementation,
			})

		}
	}

	return implementations
}

// generateBasicImplementation generates a basic service implementation for non-streaming methods
func (r *ScenarioRunner) generateBasicImplementation(serviceName, methodName, serviceStruct, methodCapitalized string, scenario Scenario) string {
	// Use strategy pattern to generate implementations
	registry := NewMethodBehaviorRegistry()
	behavior, _ := registry.Get(methodName)

	ctx := ImplementationContext{
		ServiceName:       serviceName,
		MethodName:        methodName,
		MethodCapitalized: methodCapitalized,
		ServiceStruct:     serviceStruct,
		PayloadType:       scenario.PayloadType,
		ResultType:        scenario.ResultType,
		Scenario:          scenario,
	}

	implementation, err := behavior.GenerateImplementation(ctx)
	if err != nil {
		// Fallback to generic behavior on error
		generic := &GenericBehavior{}
		implementation, _ = generic.GenerateImplementation(ctx)
	}

	return implementation
}

// generateCallImplementation generates implementation for the "call" method based on scenario data types
func (r *ScenarioRunner) generateCallImplementation(serviceName, methodName, serviceStruct, methodCapitalized string, scenario Scenario) string {
	// Extract payload and result types from scenario
	payloadType := scenario.PayloadType
	resultType := scenario.ResultType

	// Generate the appropriate method signature based on data types
	var payloadParam string
	var resultReturn string
	var implementation string

	// Handle payload parameter based on type
	if payloadType == DataTypeNone {
		payloadParam = "ctx context.Context"
	} else if payloadType == DataTypePrimitive {
		// Primitive payloads don't generate payload structs
		payloadParam = "ctx context.Context, p string"
	} else if payloadType == DataTypeMap {
		// Map payloads use map[string]interface{} directly
		payloadParam = "ctx context.Context, p map[string]interface{}"
	} else if payloadType == DataTypeUserType {
		// User type payloads use the user type directly, not a generated payload struct
		payloadParam = fmt.Sprintf("ctx context.Context, p *%s.UserType", serviceName)
	} else {
		payloadParam = fmt.Sprintf("ctx context.Context, p *%s.%sPayload", serviceName, methodCapitalized)
	}

	// Handle result return type
	switch resultType {
	case DataTypeNone:
		// Notification method - only return error
		resultReturn = "(err error)"
		implementation = `return nil`
	case DataTypePrimitive:
		resultReturn = "(res string, err error)"
		if payloadType == DataTypePrimitive {
			// Primitive to primitive - echo the payload
			implementation = `return "echo: " + p, nil`
		} else {
			// Other to primitive - return test result
			implementation = `return "test result", nil`
		}
	case DataTypeArray:
		resultReturn = "(res []string, err error)"
		implementation = `return []string{"item1", "item2"}, nil`
	case DataTypeObject:
		resultReturn = fmt.Sprintf("(res *%s.%sResult, err error)", serviceName, methodCapitalized)
		if payloadType == DataTypeObject {
			// Object to object - copy fields
			implementation = fmt.Sprintf(`return &%s.%sResult{
		Field1: p.Field1,
		Field2: p.Field2,
		Field3: p.Field3,
	}, nil`, serviceName, methodCapitalized)
		} else if payloadType == DataTypeMap {
			// Map to object - use map data to populate fields
			implementation = fmt.Sprintf(`return &%s.%sResult{
		Field1: fmt.Sprintf("map-data: %%v", p),
		Field2: func() *int { i := len(p); return &i }(), 
		Field3: func() *bool { b := len(p) > 0; return &b }(),
	}, nil`, serviceName, methodCapitalized)
		} else {
			// Other to object - create default (Field1 is string, Field2/Field3 are pointers)
			implementation = fmt.Sprintf(`return &%s.%sResult{
		Field1: "default",
		Field2: func() *int { i := 42; return &i }(), 
		Field3: func() *bool { b := true; return &b }(),
	}, nil`, serviceName, methodCapitalized)
		}
	case DataTypeMap:
		resultReturn = "(res map[string]interface{}, err error)"
		if payloadType == DataTypeMap {
			// Map to map - return the map data directly
			implementation = `return p, nil`
		} else {
			// Other to map - create default map
			implementation = `return map[string]interface{}{"key": "value"}, nil`
		}
	case DataTypeUserType:
		resultReturn = fmt.Sprintf("(res *%s.UserType, err error)", serviceName)
		// Use helper function to get pointer to string and int
		emailPtr := `func() *string { s := "test@example.com"; return &s }()`
		agePtr := `func() *int { i := 25; return &i }()`
		// The generated Go struct has capitalized field names: ID, Name, Email, Age
		implementation = fmt.Sprintf(`return &%s.UserType{
		ID:    "test-id",
		Name:  "test name",
		Email: %s,
		Age:   %s,
	}, nil`, serviceName, emailPtr, agePtr)
	default:
		resultReturn = "(res string, err error)"
		implementation = `return "unknown type", nil`
	}

	return fmt.Sprintf(`// %s implements %s.
func (s *%s) %s(%s) %s {
	log.Printf(ctx, "%s.%s")
	
	%s
}`,
		methodCapitalized, methodName, serviceStruct, methodCapitalized,
		payloadParam, resultReturn, serviceName, methodName, implementation)
}

// extractServiceNameFromDSL extracts the service name from DSL code using regex
// Returns the first service name found, or the defaultName if none found
func extractServiceNameFromDSL(dslCode, defaultName string) string {
	// Use regex to find Service("name", func() pattern
	re := regexp.MustCompile(`Service\("([^"]+)"`)
	matches := re.FindStringSubmatch(dslCode)
	if len(matches) >= 2 {
		return matches[1]
	}
	return defaultName
}
