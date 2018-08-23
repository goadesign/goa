package chatter

import (
	"context"

	jwt "github.com/dgrijalva/jwt-go"
	chattersvc "goa.design/goa/examples/chatter/gen/chatter"
	"goa.design/goa/security"
)

var (
	// ErrUnauthorized is the error returned by Login when the request credentials
	// are invalid.
	ErrUnauthorized error = chattersvc.Unauthorized("invalid username and password combination")

	// ErrInvalidToken is the error returned when the JWT token is invalid.
	ErrInvalidToken error = chattersvc.Unauthorized("invalid token")

	// ErrInvalidScopes is the error returned when the scopes provided in
	// the JWT token claims are invalid.
	ErrInvalidTokenScopes error = chattersvc.InvalidScopes("invalid scopes in token")

	// Key is the key used in JWT authentication
	Key = []byte("secret")
)

// ChatterBasicAuth implements the authorization logic for service "chatter"
// for the "basic" security scheme.
func ChatterBasicAuth(ctx context.Context, user, pass string, s *security.BasicScheme) (context.Context, error) {
	if user != "goa" {
		return ctx, ErrUnauthorized
	}
	if pass != "rocks" {
		return ctx, ErrUnauthorized
	}
	return ctx, nil
}

// ChatterJWTAuth implements the authorization logic for service "chatter" for
// the "jwt" security scheme.
func ChatterJWTAuth(ctx context.Context, token string, s *security.JWTScheme) (context.Context, error) {
	claims := make(jwt.MapClaims)

	// authorize request
	// 1. parse JWT token, token key is hardcoded to "secret" in this example
	_, err := jwt.ParseWithClaims(token, claims, func(_ *jwt.Token) (interface{}, error) { return Key, nil })
	if err != nil {
		return ctx, ErrInvalidToken
	}

	// 2. validate provided "scopes" claim
	if claims["scopes"] == nil {
		return ctx, ErrInvalidTokenScopes
	}
	scopes, ok := claims["scopes"].([]interface{})
	if !ok {
		return ctx, ErrInvalidTokenScopes
	}
	scopesInToken := make([]string, len(scopes))
	for _, scp := range scopes {
		scopesInToken = append(scopesInToken, scp.(string))
	}
	if err := s.Validate(scopesInToken); err != nil {
		return ctx, chattersvc.InvalidScopes(err.Error())
	}
	return ctx, nil
}
