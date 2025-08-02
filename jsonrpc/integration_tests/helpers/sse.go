package helpers

import (
	"fmt"
)

// SSETestImplementation generates a test implementation for an SSE streaming method
// that sends the provided data items with appropriate delays.
func SSETestImplementation(serviceName, methodName, serviceStruct, methodCapitalized string, dataItems []string) string {
	impl := fmt.Sprintf(`// %s implements %s.
func (s *%s) %s(ctx context.Context, stream %s.%sServerStream) (err error) {
	log.Printf("%s.%s")
	// Send test events
	for _, data := range []string{%s} {
		if err := stream.Send(%s); err != nil {
			return err
		}
		// Small delay between events
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}
	return nil
}`,
		methodCapitalized, methodName,
		serviceStruct, methodCapitalized, 
		serviceName, methodCapitalized,
		serviceName, methodName,
		formatDataItems(dataItems),
		"data")
	
	return impl
}

// SSEPrimitiveImplementation generates an implementation for primitive string streaming
func SSEPrimitiveImplementation(serviceName, methodName string, count int) string {
	serviceStruct := serviceName + "srvc"
	methodCapitalized := capitalize(methodName)
	
	return fmt.Sprintf(`// %s implements %s.
func (s *%s) %s(ctx context.Context, stream %s.%sServerStream) (err error) {
	log.Printf( "%s.%s")
	// Send %d test events
	for i := 1; i <= %d; i++ {
		event := fmt.Sprintf("event %%d", i)
		if err := stream.Send(event); err != nil {
			return err
		}
		// Small delay between events
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}
	return nil
}`,
		methodCapitalized, methodName,
		serviceStruct, methodCapitalized,
		serviceName, methodCapitalized,
		serviceName, methodName,
		count, count)
}

// SSEArrayImplementation generates an implementation for array streaming
func SSEArrayImplementation(serviceName, methodName string, count int) string {
	serviceStruct := serviceName + "srvc"
	methodCapitalized := capitalize(methodName)
	
	return fmt.Sprintf(`// %s implements %s.
func (s *%s) %s(ctx context.Context, stream %s.%sServerStream) (err error) {
	log.Printf( "%s.%s")
	// Send %d test events
	for i := 1; i <= %d; i++ {
		event := []string{
			fmt.Sprintf("event-%%d-a", i),
			fmt.Sprintf("event-%%d-b", i),
			fmt.Sprintf("%%d", i),
		}
		if err := stream.Send(event); err != nil {
			return err
		}
		// Small delay between events
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}
	return nil
}`,
		methodCapitalized, methodName,
		serviceStruct, methodCapitalized,
		serviceName, methodCapitalized,
		serviceName, methodName,
		count, count)
}

// SSEObjectImplementation generates an implementation for object streaming
func SSEObjectImplementation(serviceName, methodName, resultTypeName string, count int) string {
	serviceStruct := serviceName + "srvc"
	methodCapitalized := capitalize(methodName)
	
	return fmt.Sprintf(`// %s implements %s.
func (s *%s) %s(ctx context.Context, stream %s.%sServerStream) (err error) {
	log.Printf( "%s.%s")
	// Send %d test events
	for i := 1; i <= %d; i++ {
		event := &%s.%s{
			EventID:   fmt.Sprintf("evt-%%03d", i),
			Type:      "update",
			Data:      fmt.Sprintf("Event data %%d", i),
			Timestamp: fmt.Sprintf("2024-01-01T12:00:%%02dZ", i),
		}
		if err := stream.Send(event); err != nil {
			return err
		}
		// Small delay between events
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}
	return nil
}`,
		methodCapitalized, methodName,
		serviceStruct, methodCapitalized,
		serviceName, methodCapitalized,
		serviceName, methodName,
		count, count,
		serviceName, resultTypeName)
}

// SSEUserTypeImplementation generates an implementation for user type streaming
func SSEUserTypeImplementation(serviceName, methodName string, count int) string {
	serviceStruct := serviceName + "srvc"
	methodCapitalized := capitalize(methodName)
	
	return fmt.Sprintf(`// %s implements %s.
func (s *%s) %s(ctx context.Context, stream %s.%sServerStream) (err error) {
	log.Printf( "%s.%s")
	// Send %d test events
	for i := 1; i <= %d; i++ {
		event := &%s.UserType{
			ID:    fmt.Sprintf("evt-user-%%d", i),
			Name:  fmt.Sprintf("Event User %%d", i),
			Email: fmt.Sprintf("event%%d@example.com", i),
		}
		if err := stream.Send(event); err != nil {
			return err
		}
		// Small delay between events
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}
	return nil
}`,
		methodCapitalized, methodName,
		serviceStruct, methodCapitalized,
		serviceName, methodCapitalized,
		serviceName, methodName,
		count, count,
		serviceName)
}

// Helper functions

func formatDataItems(items []string) string {
	result := ""
	for i, item := range items {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf(`"%s"`, item)
	}
	return result
}

func capitalize(s string) string {
	if s == "" {
		return ""
	}
	return string(s[0]-32) + s[1:]
}