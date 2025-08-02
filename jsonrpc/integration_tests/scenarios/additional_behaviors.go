package scenarios

import (
	"fmt"
)

// ValidateComplexBehavior implements the validate_complex method pattern
type ValidateComplexBehavior struct{}

func (b *ValidateComplexBehavior) GetName() string {
	return "validate_complex"
}

func (b *ValidateComplexBehavior) GenerateImplementation(ctx ImplementationContext) (string, error) {
	return fmt.Sprintf(`// %s implements %s.
func (s *%s) %s(ctx context.Context, p *%s.%sPayload) (res bool, err error) {
	log.Printf(ctx, "%s.%s")
	
	// Complex validation - return true if validation rules are satisfied
	return true, nil
}`,
		ctx.MethodCapitalized, ctx.MethodName, ctx.ServiceStruct, ctx.MethodCapitalized,
		ctx.ServiceName, ctx.MethodCapitalized, ctx.ServiceName, ctx.MethodName,
	), nil
}

// ProcessBehavior implements the process method pattern
type ProcessBehavior struct{}

func (b *ProcessBehavior) GetName() string {
	return "process"
}

func (b *ProcessBehavior) GenerateImplementation(ctx ImplementationContext) (string, error) {
	return fmt.Sprintf(`// %s implements %s.
func (s *%s) %s(ctx context.Context, p *%s.%sPayload) (res *%s.%sResult, err error) {
	log.Printf(ctx, "%s.%s")
	
	// Simple processing - return success result
	return &%s.%sResult{
		Data:   "processed: " + p.Data,
		Status: "success",
	}, nil
}`,
		ctx.MethodCapitalized, ctx.MethodName, ctx.ServiceStruct, ctx.MethodCapitalized,
		ctx.ServiceName, ctx.MethodCapitalized, ctx.ServiceName, ctx.MethodCapitalized, ctx.ServiceName, ctx.MethodName,
		ctx.ServiceName, ctx.MethodCapitalized,
	), nil
}

// StatusBehavior implements the status method pattern
type StatusBehavior struct{}

func (b *StatusBehavior) GetName() string {
	return "status"
}

func (b *StatusBehavior) GenerateImplementation(ctx ImplementationContext) (string, error) {
	return fmt.Sprintf(`// %s implements %s.
func (s *%s) %s(ctx context.Context) (res *%s.StatusResult, err error) {
	log.Printf(ctx, "%s.%s")
	
	// Return status information
	return &%s.StatusResult{
		Status:  "running",
		Uptime:  "1h30m",
		Version: "1.0.0",
	}, nil
}`,
		ctx.MethodCapitalized, ctx.MethodName, ctx.ServiceStruct, ctx.MethodCapitalized,
		ctx.ServiceName, ctx.ServiceName, ctx.MethodName, ctx.ServiceName,
	), nil
}

// ErrorTestBehavior implements the error_test method pattern
type ErrorTestBehavior struct{}

func (b *ErrorTestBehavior) GetName() string {
	return "error_test"
}

func (b *ErrorTestBehavior) GenerateImplementation(ctx ImplementationContext) (string, error) {
	return fmt.Sprintf(`// %s implements %s.
func (s *%s) %s(ctx context.Context, p *%s.%sPayload) (res string, err error) {
	log.Printf(ctx, "%s.%s")
	
	// Test error scenarios
	switch p.ErrorType {
	case "invalid_params":
		return "", %s.MakeInvalidParams(fmt.Errorf("invalid parameters"))
	case "not_found":
		return "", %s.MakeNotFound(fmt.Errorf("resource not found"))
	case "internal_error":
		return "", %s.MakeInternalError(fmt.Errorf("internal server error"))
	case "timeout":
		return "", %s.MakeTimeout(fmt.Errorf("request timeout"))
	case "conflict":
		return "", %s.MakeConflict(fmt.Errorf("conflict"))
	default:
		return "success", nil
	}
}`,
		ctx.MethodCapitalized, ctx.MethodName, ctx.ServiceStruct, ctx.MethodCapitalized,
		ctx.ServiceName, ctx.MethodCapitalized, ctx.ServiceName, ctx.MethodName,
		ctx.ServiceName, ctx.ServiceName, ctx.ServiceName, ctx.ServiceName, ctx.ServiceName,
	), nil
}
