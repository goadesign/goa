package framework

import (
	"fmt"
	"strings"
	"time"
)

// Scenario represents a test scenario loaded from YAML
type Scenario struct {
	Name        string    `yaml:"name"`
	Method      string    `yaml:"method"`
	Transport   string    `yaml:"transport"`
	Request     Request   `yaml:"request"`
	RawRequest  string    `yaml:"raw_request,omitempty"` // For testing invalid JSON
	Batch       []Request `yaml:"batch,omitempty"`       // For batch requests
	Expect      Expect    `yaml:"expect"`
	ExpectBatch []Expect  `yaml:"expect_batch,omitempty"` // For batch responses
	Sequence    []Action  `yaml:"sequence,omitempty"`     // For streaming scenarios
}

// Request represents the JSON-RPC request to send
type Request struct {
	// JSONRPC allows overriding the JSON-RPC version field (defaults to "2.0", empty string means omit)
	JSONRPC string `yaml:"jsonrpc,omitempty"`
	// Method overrides the scenario method if specified
	Method string `yaml:"method,omitempty"`
	Params any    `yaml:"params"`
	ID     any    `yaml:"id,omitempty"` // Omit for notifications
}

// Expect represents what we expect in response
type Expect struct {
	ID     any          `yaml:"id,omitempty"`
	Result any          `yaml:"result,omitempty"`
	Error  *ExpectError `yaml:"error,omitempty"`
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
	Modifier  string // notify, error, validate, or idmap
	Transport string // sse when the method sends server-sent events
	GRPC      bool   // whether to emit GRPC endpoint (method name starts with grpc_)
}

// ParseMethod parses a method name into its components.
// Format: action_type[_modifier][_sse]
// Examples: echo_string, stream_object_sse
// Returns error if the method name is invalid.
func ParseMethod(method string) (MethodInfo, error) {
	// Check for grpc_ prefix convention
	grpc := false
	if strings.HasPrefix(method, "grpc_") {
		grpc = true
		method = strings.TrimPrefix(method, "grpc_")
	}

	parts := strings.Split(method, "_")
	if len(parts) < 2 {
		return MethodInfo{}, fmt.Errorf("invalid method name %q: must have format action_type[_modifier][_sse]", method)
	}

	info := MethodInfo{
		Action: parts[0],
		Type:   parts[1],
		GRPC:   grpc,
	}

	// Check if last part is a transport
	if len(parts) > 2 {
		lastPart := parts[len(parts)-1]
		if lastPart == "sse" {
			info.Transport = lastPart
			parts = parts[:len(parts)-1] // Remove transport from parts
		}
	}

	// Validate action
	validActions := map[string]bool{
		ActionEcho: true, ActionTransform: true, ActionGenerate: true,
		ActionStream: true,
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
			ModifierNotify: true, ModifierError: true, ModifierValidate: true, ModifierIDMap: true,
		}
		if !validModifiers[info.Modifier] {
			return MethodInfo{}, fmt.Errorf("invalid modifier %q in method %q: must be one of: %s",
				info.Modifier, method, strings.Join(getMapKeys(validModifiers), ", "))
		}
	}
	if info.Action == ActionStream && !info.IsSSE() {
		return MethodInfo{}, fmt.Errorf("streaming method %q must end with _sse", method)
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

// IsStreaming returns true if this method involves streaming
func (info MethodInfo) IsStreaming() bool {
	return info.IsSSE()
}

// GetMethod returns the effective JSON-RPC method name to use on the wire.
// It strips the grpc_ prefix convention and preserves transport suffixes.
func (r Request) GetMethod(fallback string) string {
	base := r.Method
	if base == "" {
		base = fallback
	}
	info, err := ParseMethod(base)
	if err != nil {
		return base
	}
	return info.Name()
}

// getMapKeys returns sorted keys from a map[string]bool
func getMapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
