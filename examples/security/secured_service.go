package multiauth

import (
	"context"
	"fmt"
	"log"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
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

// Creates a valid JWT after authenticating using basic_auth scheme.
func (s *securedServiceSvc) Signin(ctx context.Context, p *securedservice.SigninPayload) (res *securedservice.Creds, err error) {
	// create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"nbf":    time.Date(2015, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
		"iat":    time.Now().Unix(),
		"scopes": []string{"api:read", "api:write"},
	})

	s.logger.Printf("user '%s' logged in", p.Username)

	// note that if "SignedString" returns an error then it is returned as
	// an internal error to the client
	t, err := token.SignedString(Key)
	if err != nil {
		return nil, err
	}
	return &securedservice.Creds{
		JWT:        t,
		OauthToken: t,
		APIKey:     "my_awesome_api_key",
	}, nil
}

// This action is secured with the jwt scheme
func (s *securedServiceSvc) Secure(ctx context.Context, p *securedservice.SecurePayload) (res string, err error) {
	res = fmt.Sprintf("User authorized using JWT token %q", p.Token)
	s.logger.Printf(res)
	if p.Fail != nil && *p.Fail {
		s.logger.Printf("Uh oh! `fail` passed in parameter. Auth failed!")
		return "", securedservice.Unauthorized("forced authentication failure")
	}
	return
}

// This action is secured with the jwt scheme and also requires an API key
// query string.
func (s *securedServiceSvc) DoublySecure(ctx context.Context, p *securedservice.DoublySecurePayload) (res string, err error) {
	res = fmt.Sprintf("User authorized using JWT token %q and API Key %q", p.Token, p.Key)
	s.logger.Printf(res)
	return
}

// This action is secured with the jwt scheme and also requires an API key
// header.
func (s *securedServiceSvc) AlsoDoublySecure(ctx context.Context, p *securedservice.AlsoDoublySecurePayload) (res string, err error) {
	if p.Username != nil && p.Password != nil && p.OauthToken != nil {
		res = fmt.Sprintf("User authorized using username %q/password %q and OAuth2 token %q", *p.Username, *p.Password, *p.OauthToken)
		s.logger.Printf(res)
		return
	}
	res = fmt.Sprintf("User authorized using JWT token %q and API Key %q", *p.Token, *p.Key)
	s.logger.Print(res)
	return
}
