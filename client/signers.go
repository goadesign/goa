package client

import (
	"fmt"
	"net/http"

	"golang.org/x/net/context"

	"github.com/spf13/cobra"
)

type (
	// Signer is the common interface implemented by all signers.
	Signer interface {
		// Sign adds required headers, cookies etc.
		Sign(context.Context, *http.Request) error
		// RegisterFlags registers the command line flags that defines the values used to
		// initialize the signer.
		RegisterFlags(cmd *cobra.Command)
	}

	// BasicSigner implements basic auth.
	BasicSigner struct {
		// Username is the basic auth user.
		Username string
		// Password is err guess what? the basic auth password.
		Password string
	}

	// APIKeySigner implements API Key auth.
	APIKeySigner struct {
		// SignHeader indicates whether to set the API key in the header with name KeyName
		// or whether to use a query string with name KeyName.
		SignHeader bool
		// KeyName is the name of the HTTP header or query string that contains the API key.
		KeyName string
		// KeyValue stores the actual key.
		KeyValue string
		// Format is the format used to render the key, defaults to "Bearer %s"
		Format string
	}

	// JWTSigner implements JSON Web Token auth.
	JWTSigner struct {
		// Header is the name of the HTTP header which contains the JWT.
		// The default is "Authentication"
		Header string
		// Format represents the format used to render the JWT.
		// The default is "Bearer %s"
		Format string
		// TokenSource is a JWT token source.
		// See https://godoc.org/golang.org/x/oauth2/jwt#Config.TokenSource for an example
		// of an implementation.
		TokenSource TokenSource
	}

	// OAuth2Signer adds a authorization header to the request using the given OAuth2 token
	// source to produce the header value.
	OAuth2Signer struct {
		// TokenSource is an OAuth2 access token source.
		// See package golang/oauth2 and its subpackage for implementations of token
		// sources.
		TokenSource TokenSource
	}

	// A TokenSource is anything that can return a token.
	TokenSource interface {
		// Token returns a token or an error.
		// Token must be safe for concurrent use by multiple goroutines.
		// The returned Token must not be modified.
		Token() (*Token, error)
	}
)

// Sign adds the basic auth header to the request.
func (s *BasicSigner) Sign(ctx context.Context, req *http.Request) error {
	if s.Username != "" && s.Password != "" {
		req.SetBasicAuth(s.Username, s.Password)
	}
	return nil
}

// Sign adds the API key header to the request.
func (s *APIKeySigner) Sign(ctx context.Context, req *http.Request) error {
	if s.KeyName == "" {
		s.KeyName = "Authorization"
	}
	if s.Format == "" {
		s.Format = "Bearer %s"
	}
	name := s.KeyName
	format := s.Format
	val := fmt.Sprintf(format, s.KeyValue)
	if s.SignHeader {
		req.Header.Set(name, val)
	} else {
		req.URL.Query().Set(name, val)
	}
	return nil
}

// Sign adds the JWT auth header.
func (s *JWTSigner) Sign(ctx context.Context, req *http.Request) error {
	header := s.Header
	if header == "" {
		header = "Authorization"
	}
	format := s.Format
	if format == "" {
		format = "Bearer %s"
	}
	token, err := s.TokenSource()
	if err != nil {
		return err
	}
	req.Header.Set(header, token)
	return nil
}

// Sign refreshes the access token if needed and adds the OAuth header.
func (s *OAuth2Signer) Sign(ctx context.Context, req *http.Request) error {
	token, err := s.TokenSource.Token()
	if err != nil {
		return err
	}
	token.SetAuthHeader(req)
	return nil
}
