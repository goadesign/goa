package multiauth

import (
	"context"

	jwt "github.com/dgrijalva/jwt-go"
	securedservice "goa.design/goa/examples/security/gen/secured_service"
	"goa.design/goa/security"
)

var (
	// ErrUnauthorized is the error returned by Login when the request credentials
	// are invalid.
	ErrUnauthorized error = securedservice.Unauthorized("invalid username and password combination")

	// ErrInvalidToken is the error returned when the JWT token is invalid.
	ErrInvalidToken error = securedservice.Unauthorized("invalid token")

	// ErrInvalidTokenScopes is the error returned when the scopes provided in
	// the JWT token claims are invalid.
	ErrInvalidTokenScopes error = securedservice.InvalidScopes("invalid scopes in token")

	// Key is the key used in JWT authentication
	Key = []byte("secret")
)

// SecuredServiceBasicAuth implements the authorization logic for service
// "secured_service" for the "basic" security scheme.
func SecuredServiceBasicAuth(ctx context.Context, user, pass string, s *security.BasicScheme) (context.Context, error) {
	if user != "goa" {
		return ctx, ErrUnauthorized
	}
	if pass != "rocks" {
		return ctx, ErrUnauthorized
	}
	return ctx, nil
}

// SecuredServiceJWTAuth implements the authorization logic for service
// "secured_service" for the "jwt" security scheme.
func SecuredServiceJWTAuth(ctx context.Context, token string, s *security.JWTScheme) (context.Context, error) {
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
		return ctx, securedservice.InvalidScopes(err.Error())
	}
	return ctx, nil
}

// SecuredServiceAPIKeyAuth implements the authorization logic for service
// "secured_service" for the "api_key" security scheme.
func SecuredServiceAPIKeyAuth(ctx context.Context, key string, s *security.APIKeyScheme) (context.Context, error) {
	if key != "my_awesome_api_key" {
		return ctx, ErrUnauthorized
	}
	return ctx, nil
}

// SecuredServiceOAuth2Auth implements the authorization logic for service
// "secured_service" for the "oauth2" security scheme.
func SecuredServiceOAuth2Auth(ctx context.Context, token string, s *security.OAuth2Scheme) (context.Context, error) {
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
		return ctx, securedservice.InvalidScopes(err.Error())
	}
	return ctx, nil
}
