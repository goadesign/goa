package scenarios

import (
	"fmt"
)

// TypeHandlerRegistry manages type handlers for different data types
type TypeHandlerRegistry struct {
	handlers map[DataType]TypeHandler
}

// NewTypeHandlerRegistry creates a registry with all standard type handlers
func NewTypeHandlerRegistry() *TypeHandlerRegistry {
	registry := &TypeHandlerRegistry{
		handlers: make(map[DataType]TypeHandler),
	}

	// Register all standard type handlers
	registry.Register(DataTypeNone, &NoneTypeHandler{})
	registry.Register(DataTypePrimitive, &PrimitiveTypeHandler{})
	registry.Register(DataTypeArray, &ArrayTypeHandler{})
	registry.Register(DataTypeMap, &MapTypeHandler{})
	registry.Register(DataTypeUserType, &UserTypeHandler{})
	registry.Register(DataTypeObject, &ObjectTypeHandler{})

	return registry
}

// Register adds a type handler for the given data type
func (r *TypeHandlerRegistry) Register(dataType DataType, handler TypeHandler) {
	r.handlers[dataType] = handler
}

// Get retrieves the type handler for the given data type
func (r *TypeHandlerRegistry) Get(dataType DataType) TypeHandler {
	handler, exists := r.handlers[dataType]
	if !exists {
		return &ObjectTypeHandler{} // Default to object type
	}
	return handler
}

// NoneTypeHandler handles methods with no payload
type NoneTypeHandler struct{}

func (h *NoneTypeHandler) GetParameterDeclaration(serviceName, methodCapitalized string) string {
	return "ctx context.Context"
}

func (h *NoneTypeHandler) GetLogicTemplate(behaviorName string) string {
	switch behaviorName {
	case "echo":
		return `return "echo: <no payload>", nil`
	case "validate":
		return `return true, nil`
	case "slow_operation":
		return `// No delay parameter for no payload
	time.Sleep(100 * time.Millisecond)`
	default:
		return ""
	}
}

// PrimitiveTypeHandler handles string payloads
type PrimitiveTypeHandler struct{}

func (h *PrimitiveTypeHandler) GetParameterDeclaration(serviceName, methodCapitalized string) string {
	return "ctx context.Context, p string"
}

func (h *PrimitiveTypeHandler) GetLogicTemplate(behaviorName string) string {
	switch behaviorName {
	case "echo":
		return `return "echo: " + p, nil`
	case "validate":
		return `return p != "", nil`
	case "slow_operation":
		return `// Primitive payload - no DelayMs field
	time.Sleep(100 * time.Millisecond)`
	default:
		return ""
	}
}

// ArrayTypeHandler handles array payloads
type ArrayTypeHandler struct{}

func (h *ArrayTypeHandler) GetParameterDeclaration(serviceName, methodCapitalized string) string {
	return "ctx context.Context, p []string"
}

func (h *ArrayTypeHandler) GetLogicTemplate(behaviorName string) string {
	switch behaviorName {
	case "echo":
		return `return fmt.Sprintf("echo: %v", p), nil`
	case "validate":
		return `return len(p) > 0, nil`
	case "slow_operation":
		return `// Array payload - no DelayMs field
	time.Sleep(100 * time.Millisecond)`
	default:
		return ""
	}
}

// MapTypeHandler handles map payloads
type MapTypeHandler struct{}

func (h *MapTypeHandler) GetParameterDeclaration(serviceName, methodCapitalized string) string {
	return "ctx context.Context, p map[string]interface{}"
}

func (h *MapTypeHandler) GetLogicTemplate(behaviorName string) string {
	switch behaviorName {
	case "echo":
		return `return fmt.Sprintf("echo: %v", p), nil`
	case "validate":
		return `return len(p) > 0, nil`
	case "slow_operation":
		return `// Check for delay in map
	if delayVal, ok := p["delay_ms"]; ok {
		if delayMs, ok := delayVal.(float64); ok && delayMs > 0 {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}
	}`
	default:
		return ""
	}
}

// UserTypeHandler handles user-defined type payloads
type UserTypeHandler struct{}

func (h *UserTypeHandler) GetParameterDeclaration(serviceName, methodCapitalized string) string {
	return fmt.Sprintf("ctx context.Context, p *%s.UserType", serviceName)
}

func (h *UserTypeHandler) GetLogicTemplate(behaviorName string) string {
	switch behaviorName {
	case "echo":
		return `return fmt.Sprintf("echo: %v", p), nil`
	case "validate":
		return `return p != nil, nil`
	case "slow_operation":
		return `// UserType payload - no DelayMs field
	time.Sleep(100 * time.Millisecond)`
	default:
		return ""
	}
}

// ObjectTypeHandler handles object payloads (default)
type ObjectTypeHandler struct{}

func (h *ObjectTypeHandler) GetParameterDeclaration(serviceName, methodCapitalized string) string {
	return fmt.Sprintf("ctx context.Context, p *%s.%sPayload", serviceName, methodCapitalized)
}

func (h *ObjectTypeHandler) GetLogicTemplate(behaviorName string) string {
	switch behaviorName {
	case "echo":
		return `if p.Message != "" {
		return "echo: " + p.Message, nil
	}
	return "echo: <empty>", nil`
	case "validate":
		return `return p.Required != "", nil`
	case "slow_operation":
		return `if p.DelayMs > 0 {
		time.Sleep(time.Duration(p.DelayMs) * time.Millisecond)
	}`
	default:
		return ""
	}
}
