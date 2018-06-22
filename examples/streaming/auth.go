package cars

import (
	"context"

	jwt "github.com/dgrijalva/jwt-go"
	carssvc "goa.design/goa/examples/streaming/gen/cars"
	"goa.design/goa/security"
)

var (
	// ErrUnauthorized is the error returned by Login when the request credentials
	// are invalid.
	ErrUnauthorized error = carssvc.Unauthorized("invalid username and password combination")

	// ErrInvalidToken is the error returned when the JWT token is invalid.
	ErrInvalidToken error = carssvc.Unauthorized("invalid token")

	// ErrInvalidScopes is the error returned when the scopes provided in
	// the JWT token claims are invalid.
	ErrInvalidScopes error = carssvc.InvalidScopes("invalid scopes, requires 'stream:read'")

	// Key is the key used in JWT authentication
	Key = []byte("secret")
)

// BasicAuthFunc implements the basic auth scheme.
func BasicAuthFunc(ctx context.Context, username, password string, s *security.BasicScheme) (context.Context, error) {
	if username != "goa" {
		return ctx, ErrUnauthorized
	}
	if password != "rocks" {
		return ctx, ErrUnauthorized
	}
	return ctx, nil
}

// CarsJWTAuth implements the authorization logic for service "cars" for the
// "jwt" security scheme.
func CarsJWTAuth(ctx context.Context, token string, s *security.JWTScheme) (context.Context, error) {
	claims := make(jwt.MapClaims)

	// authorize request
	// 1. parse JWT token, token key is hardcoded to "secret" in this example
	_, err := jwt.ParseWithClaims(token, claims, func(_ *jwt.Token) (interface{}, error) { return Key, nil })
	if err != nil {
		return ctx, ErrInvalidToken
	}

	// 2. validate provided "scopes" claim
	if claims["scopes"] == nil {
		return ctx, ErrInvalidScopes
	}
	scopes, ok := claims["scopes"].([]interface{})
	if !ok {
		return ctx, ErrInvalidScopes
	}
	hasScope := false
	for _, s := range scopes {
		if s == "stream:read" {
			hasScope = true
			break
		}
	}
	if !hasScope {
		return ctx, ErrInvalidScopes
	}

	return ctx, nil
}
