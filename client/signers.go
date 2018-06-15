package client

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"golang.org/x/net/websocket"
)

const (
	// AuthHeader is the standard Authorization header name.
	AuthHeader = "Authorization"
)

type (
	// Signer is the common interface implemented by all http signers.
	Signer interface {
		// Sign adds required headers, cookies etc.
		Sign(*http.Request) error
	}

	// WsSigner is the common interface for websocket signers
	WsSigner interface {
		// WsSign adds required headers, cookies, etc. to websocket config
		WsSign(*websocket.Config) error
	}

	// BasicSigner implements basic auth.
	BasicSigner struct {
		// Username is the basic auth user.
		Username string
		// Password is err guess what? the basic auth password.
		Password string
	}

	// WsBasicSigner implements basic auth for a websocket.
	WsBasicSigner struct {
		// Username is the basic auth user.
		Username string
		// Password is err guess what? the basic auth password.
		Password string
	}

	// APIKeySigner implements API Key auth.
	APIKeySigner struct {
		// SignQuery indicates whether to set the API key in the URL query with key KeyName
		// or whether to use a header with name KeyName.
		SignQuery bool
		// KeyName is the name of the HTTP header or query string that contains the API key.
		KeyName string
		// KeyValue stores the actual key.
		KeyValue string
		// Format is the format used to render the key, e.g. "Bearer %s"
		Format string
	}

	// JWTSigner implements JSON Web Token auth.
	JWTSigner struct {
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

	// Token is the interface to an OAuth2 token implementation.
	// It can be implemented with https://godoc.org/golang.org/x/oauth2#Token.
	Token interface {
		// SetAuthHeader sets the Authorization header to r.
		SetAuthHeader(r *http.Request)
		// SetWsAuthHeader sets the Authorization header to cfg.
		SetWsAuthHeader(cfg *websocket.Config)
		// Valid reports whether Token can be used to properly sign requests.
		Valid() bool
	}

	// A TokenSource is anything that can return a token.
	TokenSource interface {
		// Token returns a token or an error.
		// Token must be safe for concurrent use by multiple goroutines.
		// The returned Token must not be modified.
		Token() (Token, error)
	}

	// StaticTokenSource implements a token source that always returns the same token.
	StaticTokenSource struct {
		StaticToken *StaticToken
	}

	// StaticToken implements a token that sets the auth header with a given static value.
	StaticToken struct {
		// Value used to set the auth header.
		Value string
		// OAuth type, defaults to "Bearer".
		Type string
	}
)

// Sign adds the basic auth header to the http request.
func (s *BasicSigner) Sign(req *http.Request) error {
	if s.Username != "" && s.Password != "" {
		req.SetBasicAuth(s.Username, s.Password)
	}
	return nil
}

// WsSign adds the basic auth header to the websocket config
func (s *WsBasicSigner) WsSign(cfg *websocket.Config) error {
	if s.Username != "" && s.Password != "" {
		cfg.Header.Set(AuthHeader, "Basic "+base64.StdEncoding.EncodeToString([]byte(s.Username+":"+s.Password)))
	}
	return nil
}

// Sign adds the API key header to the http request.
func (s *APIKeySigner) Sign(req *http.Request) error {
	if s.KeyName == "" {
		s.KeyName = AuthHeader
	}
	if s.Format == "" {
		s.Format = "Bearer %s"
	}
	name := s.KeyName
	format := s.Format
	val := fmt.Sprintf(format, s.KeyValue)
	if s.SignQuery && val != "" {
		query := req.URL.Query()
		query.Set(name, val)
		req.URL.RawQuery = query.Encode()
	} else {
		req.Header.Set(name, val)
	}
	return nil
}

// WsSign adds the API key header to the websocket config
func (s *APIKeySigner) WsSign(cfg *websocket.Config) error {
	if s.KeyName == "" {
		s.KeyName = AuthHeader
	}
	if s.Format == "" {
		s.Format = "Bearer %s"
	}
	name := s.KeyName
	format := s.Format
	val := fmt.Sprintf(format, s.KeyValue)
	cfg.Header.Set(name, val)
	return nil
}

// Sign adds the JWT auth header to the http request.
func (s *JWTSigner) Sign(req *http.Request) error {
	return signFromSource(s.TokenSource, req)
}

// WsSign adds the JWT auth header to the websocket config.
func (s *JWTSigner) WsSign(cfg *websocket.Config) error {
	return signWsFromSource(s.TokenSource, cfg)
}

// Sign refreshes the access token if needed and adds the OAuth header to the http request.
func (s *OAuth2Signer) Sign(req *http.Request) error {
	return signFromSource(s.TokenSource, req)
}

// WsSign refreshes the access token if needed and adds the OAuth header to the websocket config.
func (s *OAuth2Signer) WsSign(cfg *websocket.Config) error {
	return signWsFromSource(s.TokenSource, cfg)
}

// signFromSource generates a token using the given source and uses it to sign the http request.
func signFromSource(source TokenSource, req *http.Request) error {
	token, err := source.Token()
	if err != nil {
		return err
	}
	if !token.Valid() {
		return fmt.Errorf("token expired or invalid")
	}
	token.SetAuthHeader(req)
	return nil
}

// signWsFromSource generates a token using the given source and uses it to create the websocket config.
func signWsFromSource(source TokenSource, cfg *websocket.Config) error {
	token, err := source.Token()
	if err != nil {
		return err
	}
	if !token.Valid() {
		return fmt.Errorf("token expired or invalid")
	}
	token.SetWsAuthHeader(cfg)
	return nil
}

// Token returns the static token.
func (s *StaticTokenSource) Token() (Token, error) {
	return s.StaticToken, nil
}

// SetAuthHeader sets the Authorization header to http request.
func (t *StaticToken) SetAuthHeader(r *http.Request) {
	typ := t.Type
	if typ == "" {
		typ = "Bearer"
	}
	r.Header.Set(AuthHeader, typ+" "+t.Value)
}

// SetWsAuthHeader sets the Authorization header to websocket config.
func (t *StaticToken) SetWsAuthHeader(c *websocket.Config) {
	typ := t.Type
	if typ == "" {
		typ = "Bearer"
	}
	c.Header.Set(AuthHeader, typ+" "+t.Value)
}

// Valid reports whether Token can be used to properly sign requests.
func (t *StaticToken) Valid() bool { return true }
