package chatter

import (
	"context"
	"io"
	"log"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	chattersvc "goa.design/goa/examples/streaming/gen/chatter"
)

// chatter service example implementation.
// The example methods log the requests and return zero values.
type chatterSvc struct {
	logger         *log.Logger
	storedMessages []*chattersvc.ChatSummary
}

// NewChatter returns the chatter service implementation.
func NewChatter(logger *log.Logger) chattersvc.Service {
	return &chatterSvc{
		logger:         logger,
		storedMessages: make([]*chattersvc.ChatSummary, 0, 10),
	}
}

// Creates a valid JWT token for auth to chat.
func (s *chatterSvc) Login(ctx context.Context, p *chattersvc.LoginPayload) (res string, err error) {
	// create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"nbf":    time.Date(2015, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
		"iat":    time.Now().Unix(),
		"scopes": []string{"stream:read", "stream:write"},
	})

	s.logger.Printf("user '%s' logged in", p.User)

	// note that if "SignedString" returns an error then it is returned as
	// an internal error to the client
	return token.SignedString(Key)
}

// Echoes the message sent by the client.
// NOTE: An example for bidirectional streaming.
func (s *chatterSvc) Echoer(ctx context.Context, p *chattersvc.EchoerPayload, stream chattersvc.EchoerServerStream) (err error) {
	s.logger.Printf("authentication successful")

	// Receive from the stream in a separate go routine so we can listen for and handle GracefulStops.
	strCh := make(chan string)
	errCh := make(chan error)
	go func() {
		for {
			str, err := stream.Recv()
			if err != nil {
				if err != io.EOF {
					errCh <- err
				}
				close(strCh)
				close(errCh)
				return
			}
			strCh <- str
		}
	}()

	// Listen for context cancellation and stream input simultaneously.
	for done := false; !done; {
		select {
		case str := <-strCh:
			s.storeMessage(str)
			if err = stream.Send(str); err != nil {
				return err
			}
		case err := <-errCh:
			if err != nil {
				return err
			}
			done = true
		case <-ctx.Done():
			done = true
		}
	}
	return stream.Close()
}

// Listens to the messages sent by the client.
// NOTE: An example for payload streaming where server does not respond with a
// result type.
func (s *chatterSvc) Listener(ctx context.Context, p *chattersvc.ListenerPayload, stream chattersvc.ListenerServerStream) (err error) {
	s.logger.Printf("authentication successful")

	// Receive from the stream in a separate go routine so we can listen for and handle GracefulStops.
	strCh := make(chan string)
	errCh := make(chan error)
	go func() {
		for {
			str, err := stream.Recv()
			if err != nil {
				if err != io.EOF {
					errCh <- err
				}
				close(strCh)
				close(errCh)
				return
			}
			strCh <- str
		}
	}()

	// Listen for context cancellation and stream input simultaneously.
	for done := false; !done; {
		select {
		case str := <-strCh:
			s.storeMessage(str)
		case err := <-errCh:
			if err != nil {
				return err
			}
			done = true
		case <-ctx.Done():
			done = true
		}
	}
	return stream.Close()
}

// Summarizes the messages sent by the client.
// NOTE: An example for payload streaming where server responds with a result
// type.
func (s *chatterSvc) Summary(ctx context.Context, p *chattersvc.SummaryPayload, stream chattersvc.SummaryServerStream) (err error) {
	var summary chattersvc.ChatSummaryCollection
	s.logger.Printf("authentication successful")

	// Receive from the stream in a separate go routine so we can listen for and handle GracefulStops.
	strCh := make(chan string)
	errCh := make(chan error)
	go func() {
		for {
			str, err := stream.Recv()
			if err != nil {
				if err != io.EOF {
					errCh <- err
				}
				close(strCh)
				close(errCh)
				return
			}
			strCh <- str
		}
	}()

	// Listen for context cancellation and stream input simultaneously.
	for done := false; !done; {
		select {
		case str := <-strCh:
			s.storeMessage(str)
			lastMsg := s.storedMessages[len(s.storedMessages)-1]
			summary = append(summary, lastMsg)
		case err := <-errCh:
			if err != nil {
				return err
			}
			done = true
		case <-ctx.Done():
			done = true
		}
	}
	return stream.SendAndClose(summary)
}

// Subscribe to events sent when new chat messages are added or deleted.
// NOTE: An example for result streaming.
func (s *chatterSvc) Subscribe(ctx context.Context, p *chattersvc.SubscribePayload, stream chattersvc.SubscribeServerStream) (err error) {
	s.logger.Printf("authentication successful")
	old := s.storedMessages
	done := false
	for {
		select {
		case <-ctx.Done():
			done = true
			break
		default:
			if ev := diff(s.storedMessages, old); ev != nil {
				old = s.storedMessages
				if err := stream.Send(ev); err != nil {
					return err
				}
			}
		}
		if done {
			break
		}
	}
	s.logger.Printf("subscription ended")
	return stream.Close()
}

// Returns the chat messages sent to the server.
// NOTE: An example for result streaming with views.
func (s *chatterSvc) History(ctx context.Context, p *chattersvc.HistoryPayload, stream chattersvc.HistoryServerStream) (err error) {
	s.logger.Printf("authentication successful")
	stream.SetView("default")
	if p.View != nil {
		stream.SetView(*p.View)
	}
	for _, summ := range s.storedMessages {
		if err := stream.Send(summ); err != nil {
			return err
		}
	}
	return stream.Close()
}

// storeMessage stores the incoming message into an in-memory variable.
func (s *chatterSvc) storeMessage(message string) {
	mlen := len(message)
	sentAt := time.Now().Format(time.RFC3339)
	s.logger.Printf("Storing message: %q (length: %d, sent_at: %s)", message, mlen, sentAt)
	s.storedMessages = append(s.storedMessages, &chattersvc.ChatSummary{
		Message: message,
		Length:  &mlen,
		SentAt:  sentAt,
	})
}

// A very basic function to figure out if chat messages were added.
func diff(new, old []*chattersvc.ChatSummary) *chattersvc.Event {
	m := make(map[string]struct{})
	for _, c := range old {
		m[c.Message] = struct{}{}
	}
	for _, c := range new {
		if _, ok := m[c.Message]; !ok {
			return &chattersvc.Event{
				Message: c.Message,
				Action:  "added",
				AddedAt: c.SentAt,
			}
		}
	}
	return nil
}
