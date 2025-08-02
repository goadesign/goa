package scenarios

import (
	"fmt"
)

// ValidateBehavior implements the validate method pattern
type ValidateBehavior struct {
	typeRegistry *TypeHandlerRegistry
}

// GetName returns the behavior name
func (b *ValidateBehavior) GetName() string {
	return "validate"
}

// GenerateImplementation creates the validate method implementation
func (b *ValidateBehavior) GenerateImplementation(ctx ImplementationContext) (string, error) {
	if b.typeRegistry == nil {
		b.typeRegistry = NewTypeHandlerRegistry()
	}

	// Get the appropriate type handler
	payloadHandler := b.typeRegistry.Get(ctx.PayloadType)
	payloadParam := payloadHandler.GetParameterDeclaration(ctx.ServiceName, ctx.MethodCapitalized)
	validationLogic := payloadHandler.GetLogicTemplate("validate")

	if ctx.ResultType == DataTypeNone {
		// Notification method - only return error
		return fmt.Sprintf(`// %s implements %s.
func (s *%s) %s(%s) (err error) {
	log.Printf(ctx, "%s.%s")
	
	// Validation notification - no result returned
	return nil
}`,
			ctx.MethodCapitalized, ctx.MethodName, ctx.ServiceStruct, ctx.MethodCapitalized,
			payloadParam, ctx.ServiceName, ctx.MethodName,
		), nil
	} else {
		// Regular method - return result and error
		return fmt.Sprintf(`// %s implements %s.
func (s *%s) %s(%s) (res bool, err error) {
	log.Printf(ctx, "%s.%s")
	
	// Simple validation - return true if required field is present
	%s
}`,
			ctx.MethodCapitalized, ctx.MethodName, ctx.ServiceStruct, ctx.MethodCapitalized,
			payloadParam, ctx.ServiceName, ctx.MethodName, validationLogic,
		), nil
	}
}
