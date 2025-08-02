package scenarios

import (
	"fmt"
)

// GenericBehavior implements the default method pattern for unknown methods
type GenericBehavior struct {
	typeRegistry *TypeHandlerRegistry
}

// GetName returns the behavior name
func (b *GenericBehavior) GetName() string {
	return "generic"
}

// GenerateImplementation creates a generic method implementation
func (b *GenericBehavior) GenerateImplementation(ctx ImplementationContext) (string, error) {
	if b.typeRegistry == nil {
		b.typeRegistry = NewTypeHandlerRegistry()
	}

	// Get the appropriate type handler
	payloadHandler := b.typeRegistry.Get(ctx.PayloadType)
	payloadParam := payloadHandler.GetParameterDeclaration(ctx.ServiceName, ctx.MethodCapitalized)

	if ctx.ResultType == DataTypeNone {
		// Notification method - only return error
		return fmt.Sprintf(`// %s implements %s.
func (s *%s) %s(%s) (err error) {
	log.Printf(ctx, "%s.%s")
	
	// Generic notification implementation
	return nil
}`,
			ctx.MethodCapitalized, ctx.MethodName, ctx.ServiceStruct, ctx.MethodCapitalized,
			payloadParam, ctx.ServiceName, ctx.MethodName,
		), nil
	} else {
		// Regular method - return result and error
		return fmt.Sprintf(`// %s implements %s.
func (s *%s) %s(%s) (res string, err error) {
	log.Printf(ctx, "%s.%s")
	
	// Generic implementation - return success message
	return "method executed successfully", nil
}`,
			ctx.MethodCapitalized, ctx.MethodName, ctx.ServiceStruct, ctx.MethodCapitalized,
			payloadParam, ctx.ServiceName, ctx.MethodName,
		), nil
	}
}

// CallBehavior implements the call method pattern - delegates to generateCallImplementation
type CallBehavior struct{}

// GetName returns the behavior name
func (b *CallBehavior) GetName() string {
	return "call"
}

// GenerateImplementation creates the call method implementation
// This delegates to the existing generateCallImplementation for now
func (b *CallBehavior) GenerateImplementation(ctx ImplementationContext) (string, error) {
	// For now, we keep the existing complex call logic
	// This could be refactored further in the future
	runner := &ScenarioRunner{} // Create temporary runner to access existing method
	return runner.generateCallImplementation(
		ctx.ServiceName,
		ctx.MethodName,
		ctx.ServiceStruct,
		ctx.MethodCapitalized,
		ctx.Scenario,
	), nil
}
