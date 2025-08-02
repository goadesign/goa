package scenarios

// MethodBehavior defines how a specific method should be implemented.
// This strategy pattern replaces the large switch statement with composable behaviors.
type MethodBehavior interface {
	// GenerateImplementation creates the Go function implementation for this behavior
	GenerateImplementation(ctx ImplementationContext) (string, error)

	// GetName returns the name of this behavior (e.g., "echo", "validate")
	GetName() string
}

// ImplementationContext provides all the context needed to generate a method implementation
type ImplementationContext struct {
	ServiceName       string
	MethodName        string
	MethodCapitalized string
	ServiceStruct     string
	PayloadType       DataType
	ResultType        DataType
	Scenario          Scenario
}

// TypeHandler abstracts how different data types are handled in method signatures and logic
type TypeHandler interface {
	// GetParameterDeclaration returns the parameter part of method signature
	GetParameterDeclaration(serviceName, methodCapitalized string) string

	// GetLogicTemplate returns the business logic template for this type
	GetLogicTemplate(behaviorName string) string
}

// MethodBehaviorRegistry manages available method behaviors
type MethodBehaviorRegistry struct {
	behaviors map[string]MethodBehavior
}

// NewMethodBehaviorRegistry creates a new registry with standard behaviors
func NewMethodBehaviorRegistry() *MethodBehaviorRegistry {
	registry := &MethodBehaviorRegistry{
		behaviors: make(map[string]MethodBehavior),
	}

	// Register standard behaviors
	registry.Register(&EchoBehavior{})
	registry.Register(&ValidateBehavior{})
	registry.Register(&ValidateComplexBehavior{})
	registry.Register(&SlowOperationBehavior{})
	registry.Register(&ProcessBehavior{})
	registry.Register(&StatusBehavior{})
	registry.Register(&ErrorTestBehavior{})
	registry.Register(&CallBehavior{})
	registry.Register(&GenericBehavior{})

	return registry
}

// Register adds a behavior to the registry
func (r *MethodBehaviorRegistry) Register(behavior MethodBehavior) {
	r.behaviors[behavior.GetName()] = behavior
}

// Get retrieves a behavior by name
func (r *MethodBehaviorRegistry) Get(name string) (MethodBehavior, error) {
	behavior, exists := r.behaviors[name]
	if !exists {
		return &GenericBehavior{}, nil // Default to generic behavior
	}
	return behavior, nil
}

// GetDefaultBehavior returns the default behavior for unknown method names
func (r *MethodBehaviorRegistry) GetDefaultBehavior() MethodBehavior {
	return &GenericBehavior{}
}
