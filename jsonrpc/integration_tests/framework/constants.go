package framework

// Transport constants define available transport protocols
const (
	TransportHTTP      = "http"
	TransportWebSocket = "websocket"
	TransportSSE       = "sse"
)

// Action constants define server behavior patterns
const (
	ActionEcho      = "echo"      // Returns input unchanged
	ActionTransform = "transform" // Modifies input predictably
	ActionGenerate  = "generate"  // Returns fixed values
	ActionStream    = "stream"    // Server-side streaming
	ActionCollect   = "collect"   // Client-side streaming
	ActionBroadcast = "broadcast" // Server-initiated messages
)

// Type constants define data structures
const (
	TypeString = "string"
	TypeArray  = "array"
	TypeObject = "object"
	TypeMap    = "map"
	TypeUser   = "user" // Goa user-defined type
	TypeInt    = "int"
	TypeBool   = "bool"
)

// Modifier constants alter behavior
const (
	ModifierNotify   = "notify"   // No response expected
	ModifierError    = "error"    // Always returns error
	ModifierValidate = "validate" // Includes validation
	ModifierFinal    = "final"    // SSE: final response
)

