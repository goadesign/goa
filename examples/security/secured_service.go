package multiauth

import (
	"context"

	securedservice "goa.design/goa/examples/security/gen/secured_service"
	goalog "goa.design/goa/logging"
)

// secured_service service example implementation.
// The example methods log the requests and return zero values.
type securedServiceSvc struct {
	logger goalog.Logger
}

// Required for compatibility with Service interface
func (s *securedServiceSvc) GetLogger() goalog.Logger {
	return s.logger
}

// NewSecuredService returns the secured_service service implementation.
func NewSecuredService(logger goalog.Logger) securedservice.Service {
	return &securedServiceSvc{logger: logger}
}

// Creates a valid JWT
func (s *securedServiceSvc) Signin(ctx context.Context, p *securedservice.SigninPayload) (err error) {
	s.logger.Debug("securedService.signin")
	return
}

// This action is secured with the jwt scheme
func (s *securedServiceSvc) Secure(ctx context.Context, p *securedservice.SecurePayload) (res string, err error) {
	s.logger.Debug("securedService.secure")
	return
}

// This action is secured with the jwt scheme and also requires an API key
// query string.
func (s *securedServiceSvc) DoublySecure(ctx context.Context, p *securedservice.DoublySecurePayload) (res string, err error) {
	s.logger.Debug("securedService.doubly_secure")
	return
}

// This action is secured with the jwt scheme and also requires an API key
// header.
func (s *securedServiceSvc) AlsoDoublySecure(ctx context.Context, p *securedservice.AlsoDoublySecurePayload) (res string, err error) {
	s.logger.Debug("securedService.also_doubly_secure")
	return
}
