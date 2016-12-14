package jwt

import (
	"crypto/rsa"
	"fmt"
	"net/http"
	"sort"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/goadesign/goa"
	"golang.org/x/net/context"
)

// New returns a middleware to be used with the JWTSecurity DSL definitions of goa.  It supports the
// scopes claim in the JWT and ensures goa-defined Security DSLs are properly validated.
//
// The steps taken by the middleware are:
//
//     1. Validate the "Bearer" token present in the "Authorization" header against the key(s)
//        given to New
//     2. If scopes are defined in the design for the action validate them against the "scopes" JWT
//        claim
//
// The `exp` (expiration) and `nbf` (not before) date checks are validated by the JWT library.
//
// validationKeys can be one of these:
//
//     * a single string
//     * a single []byte
//     * a list of string
//     * a list of []byte
//     * a single rsa.PublicKey
//     * a list of rsa.PublicKey
//
// The type of the keys determine the algorithms that will be used to do the check.  The goal of
// having lists of keys is to allow for key rotation, still check the previous keys until rotation
// has been completed.
//
// You can define an optional function to do additional validations on the token once the signature
// and the claims requirements are proven to be valid.  Example:
//
//    validationHandler, _ := goa.NewMiddleware(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
//        token := jwt.ContextJWT(ctx)
//        if val, ok := token.Claims["is_uncle"].(string); !ok || val != "ben" {
//            return jwt.ErrJWTError("you are not uncle ben's")
//        }
//    })
//
// Mount the middleware with the generated UseXX function where XX is the name of the scheme as
// defined in the design, e.g.:
//
//    app.UseJWT(jwt.New("secret", validationHandler, app.NewJWTSecurity()))
//
func New(validationKeys interface{}, validationFunc goa.Middleware, scheme *goa.JWTSecurity) goa.Middleware {
	var algo string
	var rsaKeys []*rsa.PublicKey
	var hmacKeys []string

	switch keys := validationKeys.(type) {
	case []*rsa.PublicKey:
		rsaKeys = keys
		algo = "RS"
	case *rsa.PublicKey:
		rsaKeys = []*rsa.PublicKey{keys}
		algo = "RS"
	case string:
		hmacKeys = []string{keys}
		algo = "HS"
	case []string:
		hmacKeys = keys
		algo = "HS"
	default:
		panic("invalid parameter to `jwt.New()`, only accepts *rsa.publicKey, []*rsa.PublicKey (for RSA-based algorithms) or a signing secret string (for HS algorithms)")
	}

	return func(nextHandler goa.Handler) goa.Handler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			// TODO: implement the QUERY string handler too
			if scheme.In != goa.LocHeader {
				return fmt.Errorf("whoops, security scheme with location (in) %q not supported", scheme.In)
			}
			val := req.Header.Get(scheme.Name)
			if val == "" {
				return ErrJWTError(fmt.Sprintf("missing header %q", scheme.Name))
			}

			if !strings.HasPrefix(strings.ToLower(val), "bearer ") {
				return ErrJWTError(fmt.Sprintf("invalid or malformed %q header, expected 'Authorization: Bearer JWT-token...'", val))
			}

			incomingToken := strings.Split(val, " ")[1]

			var token *jwt.Token
			var err error
			switch algo {
			case "RS":
				token, err = validateRSAKeys(rsaKeys, algo, incomingToken)
			case "HS":
				token, err = validateHMACKeys(hmacKeys, algo, incomingToken)
			default:
				panic("how did this happen ? unsupported algo in jwt middleware")
			}
			if err != nil {
				return ErrJWTError(fmt.Sprintf("JWT validation failed: %s", err))
			}

			scopesInClaim, scopesInClaimList, err := parseClaimScopes(token)
			if err != nil {
				goa.LogError(ctx, err.Error())
				return ErrJWTError(err)
			}

			requiredScopes := goa.ContextRequiredScopes(ctx)

			for _, scope := range requiredScopes {
				if !scopesInClaim[scope] {
					msg := "authorization failed: required 'scopes' not present in JWT claim"
					return ErrJWTError(msg, "required", requiredScopes, "scopes", scopesInClaimList)
				}
			}

			ctx = WithJWT(ctx, token)
			if validationFunc != nil {
				nextHandler = validationFunc(nextHandler)
			}
			return nextHandler(ctx, rw, req)
		}
	}
}

// parseClaimScopes parses the "scopes" parameter in the Claims. It supports two formats:
//
// * a list of string
//
// * a single string with space-separated scopes (akin to OAuth2's "scope").
func parseClaimScopes(token *jwt.Token) (map[string]bool, []string, error) {
	scopesInClaim := make(map[string]bool)
	var scopesInClaimList []string
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, nil, fmt.Errorf("unsupport claims shape")
	}
	if claims["scopes"] != nil {
		switch scopes := claims["scopes"].(type) {
		case string:
			for _, scope := range strings.Split(scopes, " ") {
				scopesInClaim[scope] = true
				scopesInClaimList = append(scopesInClaimList, scope)
			}
		case []interface{}:
			for _, scope := range scopes {
				if val, ok := scope.(string); ok {
					scopesInClaim[val] = true
					scopesInClaimList = append(scopesInClaimList, val)
				}
			}
		default:
			return nil, nil, fmt.Errorf("unsupported 'scopes' format in incoming JWT claim, was type %T", scopes)
		}
	}
	sort.Strings(scopesInClaimList)
	return scopesInClaim, scopesInClaimList, nil
}

// ErrJWTError is the error returned by this middleware when any sort of validation or assertion
// fails during processing.
var ErrJWTError = goa.NewErrorClass("jwt_security_error", 401)

type contextKey int

const (
	jwtKey contextKey = iota + 1
)

// WithJWT creates a child context containing the given JWT.
func WithJWT(ctx context.Context, t *jwt.Token) context.Context {
	return context.WithValue(ctx, jwtKey, t)
}

// ContextJWT retrieves the JWT token from a `context` that went through our security middleware.
func ContextJWT(ctx context.Context) *jwt.Token {
	token, ok := ctx.Value(jwtKey).(*jwt.Token)
	if !ok {
		return nil
	}
	return token
}

func validateRSAKeys(rsaKeys []*rsa.PublicKey, algo, incomingToken string) (token *jwt.Token, err error) {
	for _, pubkey := range rsaKeys {
		token, err = jwt.Parse(incomingToken, func(token *jwt.Token) (interface{}, error) {
			if !strings.HasPrefix(token.Method.Alg(), algo) {
				return nil, ErrJWTError(fmt.Sprintf("Unexpected signing method: %v", token.Header["alg"]))
			}
			return pubkey, nil
		})
		if err == nil {
			return
		}
	}
	return
}

func validateHMACKeys(hmacKeys []string, algo, incomingToken string) (token *jwt.Token, err error) {
	for _, key := range hmacKeys {
		token, err = jwt.Parse(incomingToken, func(token *jwt.Token) (interface{}, error) {
			if !strings.HasPrefix(token.Method.Alg(), algo) {
				return nil, ErrJWTError(fmt.Sprintf("Unexpected signing method: %v", token.Header["alg"]))
			}
			return []byte(key), nil
		})
		if err == nil {
			return
		}
	}
	return
}
