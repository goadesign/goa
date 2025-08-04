package framework

import (
	"fmt"
	"strings"
	"time"
)

// Sequence action types for streaming scenarios
const (
	SequenceActionSend    = "send"
	SequenceActionReceive = "receive"
	SequenceActionClose   = "close"
)

// Scenario represents a test scenario loaded from YAML
type Scenario struct {
	Name        string     `yaml:"name"`
	Method      string     `yaml:"method"`
	Transport   string     `yaml:"transport"`
	Request     Request    `yaml:"request"`
	RawRequest  string     `yaml:"raw_request,omitempty"`  // For testing invalid JSON
	Batch       []Request  `yaml:"batch,omitempty"`        // For batch requests
	Expect      Expect     `yaml:"expect"`
	ExpectBatch []Expect   `yaml:"expect_batch,omitempty"` // For batch responses
	Sequence    []Action   `yaml:"sequence,omitempty"`     // For streaming scenarios
}

// Request represents the JSON-RPC request to send
type Request struct {
	// Method overrides the scenario method if specified
	Method string `yaml:"method,omitempty"`
	Params any    `yaml:"params"`
	ID     any    `yaml:"id,omitempty"` // Omit for notifications
}

// Expect represents what we expect in response
type Expect struct {
	ID      any          `yaml:"id,omitempty"`
	Result  any          `yaml:"result,omitempty"`
	Error   *ExpectError `yaml:"error,omitempty"`
	// NoResponse indicates we expect no response (for notifications)
	NoResponse bool `yaml:"no_response,omitempty"`
}

// ExpectError represents an expected JSON-RPC error
type ExpectError struct {
	Code    int    `yaml:"code"`
	Message string `yaml:"message"`
	Data    any    `yaml:"data,omitempty"`
}

// Aliases for backward compatibility
type ErrorExpect = ExpectError

// Action represents a step in a streaming sequence
type Action struct {
	Type   string        `yaml:"type"` // Use Action* constants
	Data   any           `yaml:"data,omitempty"`
	Expect any           `yaml:"expect,omitempty"`
	Delay  time.Duration `yaml:"delay,omitempty"`
}

// Config holds test configuration
type Config struct {
	Scenarios []Scenario `yaml:"scenarios"`
	Settings  Settings   `yaml:"settings,omitempty"`
}

// Settings holds global test settings
type Settings struct {
	Timeout time.Duration `yaml:"timeout,omitempty"`
	BaseURL string        `yaml:"base_url,omitempty"`
}

// MethodInfo extracts information from a method name
type MethodInfo struct {
	Action    string // echo, transform, generate, etc.
	Type      string // string, array, object, etc.
	Modifier  string // notify, error, validate, final
	Transport string // sse, ws (extracted from method suffix)
}

// ParseMethod parses a method name into its components.
// Format: action_type[_modifier][_transport]
// Examples: echo_string, stream_object_final_sse, broadcast_string_ws
// Returns error if the method name is invalid.
func ParseMethod(method string) (MethodInfo, error) {
	parts := strings.Split(method, "_")
	if len(parts) < 2 {
		return MethodInfo{}, fmt.Errorf("invalid method name %q: must have format action_type[_modifier][_transport]", method)
	}
	
	info := MethodInfo{
		Action: parts[0],
		Type:   parts[1],
	}
	
	// Check if last part is a transport
	if len(parts) > 2 {
		lastPart := parts[len(parts)-1]
		if lastPart == "sse" || lastPart == "ws" {
			info.Transport = lastPart
			parts = parts[:len(parts)-1] // Remove transport from parts
		}
	}
	
	// Validate action
	validActions := map[string]bool{
		ActionEcho: true, ActionTransform: true, ActionGenerate: true,
		ActionStream: true, ActionCollect: true, ActionBroadcast: true,
	}
	if !validActions[info.Action] {
		return MethodInfo{}, fmt.Errorf("invalid action %q in method %q: must be one of: %s", 
			info.Action, method, strings.Join(getMapKeys(validActions), ", "))
	}
	
	// Validate type
	validTypes := map[string]bool{
		TypeString: true, TypeArray: true, TypeObject: true,
		TypeMap: true, TypeUser: true, TypeInt: true, TypeBool: true,
	}
	if !validTypes[info.Type] {
		return MethodInfo{}, fmt.Errorf("invalid type %q in method %q: must be one of: %s", 
			info.Type, method, strings.Join(getMapKeys(validTypes), ", "))
	}
	
	// Check for modifier (3rd part after action and type)
	if len(parts) >= 3 {
		info.Modifier = parts[2]
		// Validate modifier
		validModifiers := map[string]bool{
			ModifierNotify: true, ModifierError: true, ModifierValidate: true, ModifierFinal: true,
		}
		if !validModifiers[info.Modifier] {
			return MethodInfo{}, fmt.Errorf("invalid modifier %q in method %q: must be one of: %s", 
				info.Modifier, method, strings.Join(getMapKeys(validModifiers), ", "))
		}
	}
	
	return info, nil
}

// Name returns the full method name reconstructed from its components
func (info MethodInfo) Name() string {
	parts := []string{info.Action, info.Type}
	if info.Modifier != "" {
		parts = append(parts, info.Modifier)
	}
	if info.Transport != "" {
		parts = append(parts, info.Transport)
	}
	return strings.Join(parts, "_")
}

// IsNotification returns true if this scenario expects no response
func (s Scenario) IsNotification() bool {
	info, err := ParseMethod(s.Method)
	if err != nil {
		return false
	}
	return info.Modifier == ModifierNotify || s.Expect.NoResponse
}

// IsSSE returns true if this method uses SSE transport
func (info MethodInfo) IsSSE() bool {
	return info.Transport == "sse"
}

// IsWebSocket returns true if this method uses WebSocket transport
func (info MethodInfo) IsWebSocket() bool {
	return info.Transport == "ws"
}

// IsStreaming returns true if this method involves streaming
func (info MethodInfo) IsStreaming() bool {
	return info.IsSSE() || info.IsWebSocket() || info.Action == ActionStream || info.Action == ActionCollect || info.Action == ActionBroadcast
}

// HasStreamingResult returns true if this method streams results
func (info MethodInfo) HasStreamingResult() bool {
	if info.IsSSE() {
		return true // SSE always streams results
	}
	if info.IsWebSocket() {
		// WebSocket methods can stream results based on action
		return info.Action == ActionStream || info.Action == ActionBroadcast || 
			   info.Action == ActionEcho || info.Action == ActionTransform || info.Action == ActionGenerate ||
			   info.Action == ActionCollect
	}
	return false
}

// HasStreamingPayload returns true if this method streams payload
func (info MethodInfo) HasStreamingPayload() bool {
	if info.IsSSE() {
		return false // SSE doesn't support streaming payload
	}
	if info.IsWebSocket() {
		// Server-initiated broadcasts don't have streaming payload
		if info.Action == ActionBroadcast {
			return false
		}
		// Client notifications and bidirectional methods have streaming payload
		return true
	}
	return false
}

// GetMethod returns the effective method name for the request
func (r Request) GetMethod(fallback string) string {
	if r.Method != "" {
		return r.Method
	}
	return fallback
}

// getMapKeys returns sorted keys from a map[string]bool
func getMapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}