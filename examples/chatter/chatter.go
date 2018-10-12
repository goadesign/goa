package chatter

import (
	"context"
	"io"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	chattersvc "goa.design/goa/examples/chatter/gen/chatter"
	goalog "goa.design/goa/logging"
)

// chatter service example implementation.
// The example methods log the requests and return zero values.
type chatterSvc struct {
	logger         goalog.Logger
	storedMessages []*chattersvc.ChatSummary
}

// Required for compatibility with Service interface
func (s *chatterSvc) GetLogger() goalog.Logger {
	return s.logger
}

// NewChatter returns the chatter service implementation.
func NewChatter(logger goalog.Logger) chattersvc.Service {
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

	s.logger.Infof("user '%s' logged in", p.User)

	// note that if "SignedString" returns an error then it is returned as
	// an internal error to the client
	return token.SignedString(Key)
}

// Echoes the message sent by the client.
func (s *chatterSvc) Echoer(ctx context.Context, p *chattersvc.EchoerPayload, stream chattersvc.EchoerServerStream) (err error) {
	s.logger.Infof("authentication successful")
	for {
		str, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		s.storeMessage(str)
		if err = stream.Send(str); err != nil {
			return err
		}
	}
	return stream.Close()
}

// Listens to the messages sent by the client.
func (s *chatterSvc) Listener(ctx context.Context, p *chattersvc.ListenerPayload, stream chattersvc.ListenerServerStream) (err error) {
	s.logger.Infof("authentication successful")
	for {
		str, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		s.storeMessage(str)
	}
	return nil
}

// Summarizes the chat messages sent by the client.
func (s *chatterSvc) Summary(ctx context.Context, p *chattersvc.SummaryPayload, stream chattersvc.SummaryServerStream) (err error) {
	var summary chattersvc.ChatSummaryCollection
	s.logger.Infof("authentication successful")
	for {
		str, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		s.storeMessage(str)
		lastMsg := s.storedMessages[len(s.storedMessages)-1]
		summary = append(summary, lastMsg)
	}
	return stream.SendAndClose(summary)
}

// Returns the chat messages sent to the server.
func (s *chatterSvc) History(ctx context.Context, p *chattersvc.HistoryPayload, stream chattersvc.HistoryServerStream) (err error) {
	s.logger.Infof("authentication successful")
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
	s.logger.Infof("Storing message: %q (length: %d, sent_at: %s)", message, mlen, sentAt)
	s.storedMessages = append(s.storedMessages, &chattersvc.ChatSummary{
		Message: message,
		Length:  &mlen,
		SentAt:  &sentAt,
	})
}
