package scenarios

import (
	"fmt"
)

// EchoBehavior implements the echo method pattern
type EchoBehavior struct {
	typeRegistry *TypeHandlerRegistry
}

// GetName returns the behavior name
func (b *EchoBehavior) GetName() string {
	return "echo"
}

// GenerateImplementation creates the echo method implementation
func (b *EchoBehavior) GenerateImplementation(ctx ImplementationContext) (string, error) {
	if b.typeRegistry == nil {
		b.typeRegistry = NewTypeHandlerRegistry()
	}

	// Get the appropriate type handler
	payloadHandler := b.typeRegistry.Get(ctx.PayloadType)
	payloadParam := payloadHandler.GetParameterDeclaration(ctx.ServiceName, ctx.MethodCapitalized)
	echoLogic := payloadHandler.GetLogicTemplate("echo")

	if ctx.ResultType == DataTypeNone {
		// Notification method - only return error
		return fmt.Sprintf(`// %s implements %s.
func (s *%s) %s(%s) (err error) {
	log.Printf(ctx, "%s.%s")
	
	// Echo notification - no result returned
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
	
	// Echo back the message from the payload
	%s
}`,
			ctx.MethodCapitalized, ctx.MethodName, ctx.ServiceStruct, ctx.MethodCapitalized,
			payloadParam, ctx.ServiceName, ctx.MethodName, echoLogic,
		), nil
	}
}
