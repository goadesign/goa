package harness

import (
	"context"
	"fmt"
	"time"
)

// EventsService provides a test implementation for SSE streaming
type EventsService struct{}

// Subscribe implements the SSE streaming method
func (s *EventsService) Subscribe(ctx context.Context, stream any) error {
	// Type assert to get the actual stream interface
	type serverStream interface {
		Send(string) error
	}

	sseStream, ok := stream.(serverStream)
	if !ok {
		return fmt.Errorf("invalid stream type")
	}

	// Send 5 events as expected by the tests
	for i := 1; i <= 5; i++ {
		event := fmt.Sprintf("event %d", i)
		if err := sseStream.Send(event); err != nil {
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
}
