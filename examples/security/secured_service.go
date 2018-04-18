package multiauth

import (
	"context"
	"log"

	securedservice "goa.design/goa/examples/security/gen/secured_service"
)

// secured_service service example implementation.
// The example methods log the requests and return zero values.
type securedServiceSvc struct {
	logger *log.Logger
}

// NewSecuredService returns the secured_service service implementation.
func NewSecuredService(logger *log.Logger) securedservice.Service {
	return &securedServiceSvc{logger}
}

// Creates a valid JWT
func (s *securedServiceSvc) Signin(ctx context.Context, p *securedservice.SigninPayload) error {
	s.logger.Print("securedService.signin")
	return nil
}

// This action is secured with the jwt scheme
func (s *securedServiceSvc) Secure(ctx context.Context, p *securedservice.SecurePayload) (string, error) {
	var res string
	s.logger.Print("securedService.secure")
	return res, nil
}

// This action is secured with the jwt scheme and also requires an API key
// query string.
func (s *securedServiceSvc) DoublySecure(ctx context.Context, p *securedservice.DoublySecurePayload) (string, error) {
	var res string
	s.logger.Print("securedService.doubly_secure")
	return res, nil
}

// This action is secured with the jwt scheme and also requires an API key
// header.
func (s *securedServiceSvc) AlsoDoublySecure(ctx context.Context, p *securedservice.AlsoDoublySecurePayload) (string, error) {
	var res string
	s.logger.Print("securedService.also_doubly_secure")
	return res, nil
}
