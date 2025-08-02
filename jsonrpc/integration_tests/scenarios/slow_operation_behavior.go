package scenarios

import (
	"fmt"
)

// SlowOperationBehavior implements the slow_operation method pattern
type SlowOperationBehavior struct {
	typeRegistry *TypeHandlerRegistry
}

// GetName returns the behavior name
func (b *SlowOperationBehavior) GetName() string {
	return "slow_operation"
}

// GenerateImplementation creates the slow_operation method implementation
func (b *SlowOperationBehavior) GenerateImplementation(ctx ImplementationContext) (string, error) {
	if b.typeRegistry == nil {
		b.typeRegistry = NewTypeHandlerRegistry()
	}

	// Get the appropriate type handler
	payloadHandler := b.typeRegistry.Get(ctx.PayloadType)
	payloadParam := payloadHandler.GetParameterDeclaration(ctx.ServiceName, ctx.MethodCapitalized)
	delayLogic := payloadHandler.GetLogicTemplate("slow_operation")

	if ctx.ResultType == DataTypeNone {
		// Notification method - only return error
		return fmt.Sprintf(`// %s implements %s.
func (s *%s) %s(%s) (err error) {
	log.Printf(ctx, "%s.%s")
	
	// Simulate slow notification operation with delay
	%s
	return nil
}`,
			ctx.MethodCapitalized, ctx.MethodName, ctx.ServiceStruct, ctx.MethodCapitalized,
			payloadParam, ctx.ServiceName, ctx.MethodName, delayLogic,
		), nil
	} else {
		// Regular method - return result and error
		return fmt.Sprintf(`// %s implements %s.
func (s *%s) %s(%s) (res string, err error) {
	log.Printf(ctx, "%s.%s")
	
	// Simulate slow operation with delay
	%s
	return "operation completed", nil
}`,
			ctx.MethodCapitalized, ctx.MethodName, ctx.ServiceStruct, ctx.MethodCapitalized,
			payloadParam, ctx.ServiceName, ctx.MethodName, delayLogic,
		), nil
	}
}
