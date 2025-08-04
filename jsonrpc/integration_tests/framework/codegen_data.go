package framework

import (
	"goa.design/goa/v3/codegen"
)

// DesignData holds the semantic data for generating the design file
type DesignData struct {
	// APIName is the name of the API
	APIName string
	// APITitle is the title of the API
	APITitle string
	// APIDescription is the description of the API
	APIDescription string
	// Services contains the service definitions
	Services []*ServiceData
}

// ServiceData holds the semantic data for a service
type ServiceData struct {
	// Name is the service name (e.g., "test", "testsse")
	Name string
	// Title is the capitalized service name
	Title string
	// Description is the service description
	Description string
	// JSONRPCPath is the JSON-RPC endpoint path
	JSONRPCPath string
	// Methods contains the method definitions
	Methods []*MethodData
	// HasErrors indicates if any method returns errors
	HasErrors bool
}

// MethodData holds the semantic data for a method
type MethodData struct {
	// Name is the method name (e.g., "echo_string_http")
	Name string
	// GoName is the Go method name (e.g., "EchoStringHTTP")
	GoName string
	// Description is the method description
	Description string
	// Info contains the parsed method information
	Info MethodInfo
	
	// Type information
	Payload          *TypeSpec  // Initial payload (if any)
	StreamingPayload *TypeSpec  // Streaming payload (if any)
	Result           *TypeSpec  // Result - can be regular or streaming
	
	// Behavior flags
	IsNotification bool     // No response expected
	ReturnsError   bool     // Always returns error
	HasValidation  bool     // Payload has validation rules
	
	// Streaming information
	IsStreaming    bool
	StreamKind     string   // "payload", "result", "bidirectional"
	Transport      string   // "http", "sse", "ws"
	
	// For SSE with final response
	HasFinalResponse bool
}

// TypeSpec describes a type semantically
type TypeSpec struct {
	// Kind is the type category: "primitive", "array", "object", "map", "any"
	Kind string
	
	// For primitives
	Primitive string // "String", "Int", "Boolean"
	
	// For arrays
	ArrayElem *TypeSpec
	
	// For objects
	Fields []FieldSpec
	
	// For maps
	MapKey   *TypeSpec
	MapValue *TypeSpec
	
	// Validation rules
	Validations []ValidationSpec
	
	// Whether this type needs ID field (for bidirectional WebSocket)
	NeedsID bool
}

// FieldSpec describes a field in an object
type FieldSpec struct {
	Position    int
	Name        string
	GoName      string
	Type        *TypeSpec
	Description string
	Required    bool
}

// ValidationSpec describes a validation rule
type ValidationSpec struct {
	Type  string // "MinLength", "MaxLength", "Pattern", etc.
	Value interface{}
}

// ImplementationData holds the semantic data for generating service implementations
type ImplementationData struct {
	// PackageName is the Go package name
	PackageName string
	// Services contains the service implementations
	Services []*ServiceImplData
	// Imports contains the required imports
	Imports []*codegen.ImportSpec
}

// ServiceImplData holds the data for a service implementation
type ServiceImplData struct {
	*ServiceData
	// ServicePackage is the generated service package name
	ServicePackage string
	// Methods with implementation details
	Methods []*MethodImplData
}

// MethodImplData holds the data for a method implementation
type MethodImplData struct {
	*MethodData
	// ServicePackage is the service package name (for accessing service types)
	ServicePackage string
	// HasPayload indicates if the method accepts a payload
	HasPayload bool
	// HasResult indicates if the method returns a result
	HasResult bool
	// PayloadRef is the type reference for the payload parameter
	PayloadRef string
	// ResultRef is the type reference for the result
	ResultRef string
	// StreamInterface is the name of the stream interface (if streaming)
	StreamInterface string
}

// ActionBehavior describes how a method should behave based on its action
type ActionBehavior struct {
	// Action type (echo, transform, generate, collect, stream, broadcast)
	Action string
	// Type being operated on (string, array, object, map)
	Type string
	// Additional context (e.g., for streaming methods)
	Context map[string]interface{}
}

// Helper methods

// IsSSE returns true if this method uses SSE transport
func (m *MethodData) IsSSE() bool {
	return m.Transport == "sse"
}

// IsWebSocket returns true if this method uses WebSocket transport
func (m *MethodData) IsWebSocket() bool {
	return m.Transport == "ws"
}

// IsBidirectional returns true if this is a bidirectional streaming method
func (m *MethodData) IsBidirectional() bool {
	return m.StreamKind == "bidirectional"
}

// NeedsStreamingService returns true if this method requires a separate streaming service
func (m *MethodData) NeedsStreamingService() bool {
	return m.IsStreaming && (m.IsSSE() || m.IsWebSocket())
}