package jwt

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"context"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/goadesign/goa"
)

// New returns a middleware to be used with the JWTSecurity DSL definitions of goa.  It supports the
// scopes claim in the JWT and ensures goa-defined Security DSLs are properly validated.
//
// The steps taken by the middleware are:
//
//     1. Validate the "Bearer" token present in the "Authorization" header against the key(s)
//        given to New
//     2. If scopes are defined in the design for the action, validate them
//        against the scopes presented by the JWT in the claim "scope", or if
//        that's not defined, "scopes".
//
// The `exp` (expiration) and `nbf` (not before) date checks are validated by the JWT library.
//
// validationKeys can be one of these:
//
//     * []byte
//     * string
//     * an *rsa.PublicKey
//     * an *ecdsa.PublicKey
//     * a slice of any of the above
//
// Keys of type string or []byte are interpreted according to the signing method defined in the JWT
// token's `typ` header element: `HS`, `RS`, `ES`, etc.
//
// You can define an optional function to do additional validations on the token once the signature
// and the claims requirements are proven to be valid.  Example:
//
//	validationHandler, _ := goa.NewMiddleware(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
//		token := jwt.ContextJWT(ctx)
//		claims, ok := token.Claims.(jwtgo.MapClaims)
//		if !ok {
//			return jwt.ErrJWTError("unsupported claims shape")
//		}
//		if val, ok := claims["is_uncle"].(string); !ok || val != "ben" {
//			return jwt.ErrJWTError("you are not uncle ben's")
//		}
//		return nil
//	})
//
// Mount the middleware with the generated UseXX function where XX is the name of the scheme as
// defined in the design, e.g.:
//
//    jwtResolver, _ := jwt.NewSimpleResolver("secret")
//    app.UseJWT(jwt.New(jwtResolver, validationHandler, app.NewJWTSecurity()))
//
func New(resolver KeyResolver, validationFunc goa.Middleware, scheme *goa.JWTSecurity) goa.Middleware {
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
				return ErrJWTError(fmt.Sprintf("invalid or malformed %q header, expected 'Bearer JWT-token...'", val))
			}

			incomingToken := strings.Split(val, " ")[1]

			rsaKeys, ecdsaKeys, hmacKeys := partitionKeys(resolver.SelectKeys(req))

			var (
				token     *jwt.Token
				err       error
				validated = false
			)

			if len(rsaKeys) > 0 {
				token, err = validateRSAKeys(rsaKeys, "RS", incomingToken)
				if err == nil {
					validated = true
				}
			}

			if !validated && len(ecdsaKeys) > 0 {
				token, err = validateECDSAKeys(ecdsaKeys, "ES", incomingToken)
				if err == nil {
					validated = true
				}
			}

			if !validated && len(hmacKeys) > 0 {
				token, err = validateHMACKeys(hmacKeys, "HS", incomingToken)
				if err == nil {
					validated = true
				}
			}

			if !validated {
				return ErrJWTError("JWT validation failed")
			}

			scopesInClaim, scopesInClaimList, err := parseClaimScopes(token)
			if err != nil {
				goa.LogError(ctx, err.Error())
				return ErrJWTError(err)
			}

			requiredScopes := goa.ContextRequiredScopes(ctx)

			for _, scope := range requiredScopes {
				if !scopesInClaim[scope] {
					msg := "authorization failed: required 'scope' or 'scopes' not present in JWT claim"
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

// partitionKeys sorts keys by their type.
func partitionKeys(keys []Key) ([]*rsa.PublicKey, []*ecdsa.PublicKey, [][]byte) {
	var (
		rsaKeys   []*rsa.PublicKey
		ecdsaKeys []*ecdsa.PublicKey
		hmacKeys  [][]byte
	)

	for _, key := range keys {
		switch k := key.(type) {
		case *rsa.PublicKey:
			rsaKeys = append(rsaKeys, k)
		case *ecdsa.PublicKey:
			ecdsaKeys = append(ecdsaKeys, k)
		case []byte:
			hmacKeys = append(hmacKeys, k)
		case string:
			hmacKeys = append(hmacKeys, []byte(k))
		}
	}

	return rsaKeys, ecdsaKeys, hmacKeys
}

// validScopeClaimKeys are the claims under which scopes may be found in a token
var validScopeClaimKeys = []string{"scope", "scopes"}

// parseClaimScopes parses the "scope" or "scopes" parameter in the Claims. It
// supports two formats:
//
// * a list of strings
//
// * a single string with space-separated scopes (akin to OAuth2's "scope").
//
// An empty string is an explicit claim of no scopes.
func parseClaimScopes(token *jwt.Token) (map[string]bool, []string, error) {
	scopesInClaim := make(map[string]bool)
	var scopesInClaimList []string
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, nil, fmt.Errorf("unsupported claims shape")
	}
	for _, k := range validScopeClaimKeys {
		if rawscopes, ok := claims[k]; ok && rawscopes != nil {
			switch scopes := rawscopes.(type) {
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
				return nil, nil, fmt.Errorf("unsupported scope format in incoming JWT claim, was type %T", scopes)
			}
			break
		}
	}
	sort.Strings(scopesInClaimList)
	return scopesInClaim, scopesInClaimList, nil
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

func validateECDSAKeys(ecdsaKeys []*ecdsa.PublicKey, algo, incomingToken string) (token *jwt.Token, err error) {
	for _, pubkey := range ecdsaKeys {
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

func validateHMACKeys(hmacKeys [][]byte, algo, incomingToken string) (token *jwt.Token, err error) {
	for _, key := range hmacKeys {
		token, err = jwt.Parse(incomingToken, func(token *jwt.Token) (interface{}, error) {
			if !strings.HasPrefix(token.Method.Alg(), algo) {
				return nil, ErrJWTError(fmt.Sprintf("Unexpected signing method: %v", token.Header["alg"]))
			}
			return key, nil
		})
		if err == nil {
			return
		}
	}
	return
}
