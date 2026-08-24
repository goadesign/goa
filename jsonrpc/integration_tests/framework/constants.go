package framework

// Transport constants define available transport protocols
const (
	TransportHTTP = "http"
	TransportSSE  = "sse"
)

// Action constants define server behavior patterns
const (
	ActionEcho      = "echo"      // Returns input unchanged
	ActionTransform = "transform" // Modifies input predictably
	ActionGenerate  = "generate"  // Returns fixed values
	ActionStream    = "stream"    // Sends results with server-sent events
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
	ModifierIDMap    = "idmap"    // Map envelope ID to payload/result field
)
