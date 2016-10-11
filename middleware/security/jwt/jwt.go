package jwt

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/goadesign/goa"
	"golang.org/x/net/context"
)

// ErrInvalidKey is returned when a key is not of type string, []string,
// *rsa.PublicKey or []*rsa.PublicKey.
var ErrInvalidKey = errors.New("invalid parameter, the only keys accepted " +
	"are *rsa.publicKey, []*rsa.PublicKey (for RSA-based algorithms) or a " +
	"signing secret string, []string (for HS algorithms)")

// ErrKeyDoesNotExist is returned when a key cannot be found by the provided
// key name.
var ErrKeyDoesNotExist = errors.New("key does not exist")

// KeyResolver is a struct that is passed into the New() function, which allows
// the user to add/remove keys from the jwt goa.middleware. The use of a
// resolver provides for better scalability/performance as the number of
// valid keys grows. If an incoming http.Request contains the header field
// "jwtkeyname", then the handler will only attempt to validate the incoming
// JWT against the keys stored in the resolver under that name. Otherwise,
// the handler will attempt to validate the incoming JWT against all keys
// stored in the resolver.
type KeyResolver struct {
	*sync.RWMutex
	jwtKeyNameField string
	keyMap          map[string][]interface{}
}

// AddKeys can be used to add keys to the resolver which will be referenced
// by the provided name. Acceptable types for keys include string, []string,
// *rsa.PublicKey or []*rsa.PublicKey. Multiple keys are allowed for a single
// key name to allow for key rotation.
func (kr *KeyResolver) AddKeys(name string, keys interface{}) error {
	kr.Lock()
	defer kr.Unlock()
	switch keys := keys.(type) {
	case *rsa.PublicKey:
		kr.keyMap[name] = append(kr.keyMap[name], keys)
	case []*rsa.PublicKey:
		for _, key := range keys {
			kr.keyMap[name] = append(kr.keyMap[name], key)
		}
	case string:
		kr.keyMap[name] = append(kr.keyMap[name], keys)
	case []string:
		for _, key := range keys {
			kr.keyMap[name] = append(kr.keyMap[name], key)
		}
	default:
		return ErrInvalidKey
	}
	return nil
}

// RemoveAllKeys removes all keys from the resolver.
func (kr *KeyResolver) RemoveAllKeys() {
	kr.Lock()
	defer kr.Unlock()
	kr.keyMap = make(map[string][]interface{})
	return
}

// RemoveKeys removes all keys from the resolver stored under the provided name.
func (kr *KeyResolver) RemoveKeys(name string) {
	kr.Lock()
	defer kr.Unlock()
	delete(kr.keyMap, name)
	return
}

// RemoveKey removes only the provided key stored under the provided name from
// the resolver.
func (kr *KeyResolver) RemoveKey(name string, key interface{}) {
	kr.Lock()
	defer kr.Unlock()
	if keys, ok := kr.keyMap[name]; ok {
		for i, keyItem := range keys {
			if keyItem == key {
				kr.keyMap[name] = append(keys[:i], keys[i+1:]...)
			}
		}
	}
	return
}

// GetAllKeys returns a list of all the keys stored in the resolver.
func (kr *KeyResolver) GetAllKeys() []interface{} {
	kr.RLock()
	defer kr.RUnlock()
	var keys []interface{}
	for name := range kr.keyMap {
		for _, key := range kr.keyMap[name] {
			keys = append(keys, key)
		}
	}
	return keys
}

// GetKeys returns a list of all the keys stored in the resolver under the
// provided name.
func (kr *KeyResolver) GetKeys(name string) ([]interface{}, error) {
	kr.RLock()
	defer kr.RUnlock()
	if keys, ok := kr.keyMap[name]; ok {
		return keys, nil
	}
	return nil, ErrKeyDoesNotExist
}

// NewResolver returns a KeyResolver populated with the provided map of key
// names to key lists. NewResolver will also set the HTTP header param name
// (jwtKeyNameField) to use for reading the JWT key name from HTTP requests.
func NewResolver(validationKeys map[string][]interface{},
	jwtKeyNameField string) (*KeyResolver, error) {
	keyMap := make(map[string][]interface{})
	for name := range validationKeys {
		for _, keys := range validationKeys[name] {
			switch keys := keys.(type) {
			case *rsa.PublicKey:
				keyMap[name] = append(keyMap[name], keys)
			case []*rsa.PublicKey:
				for _, key := range keys {
					keyMap[name] = append(keyMap[name], key)
				}
			case string:
				keyMap[name] = append(keyMap[name], keys)
			case []string:
				for _, key := range keys {
					keyMap[name] = append(keyMap[name], key)
				}
			default:
				return nil, ErrInvalidKey
			}
		}
	}
	return &KeyResolver{RWMutex: &sync.RWMutex{}, keyMap: keyMap,
		jwtKeyNameField: jwtKeyNameField}, nil
}

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
// 	  jwtResolver, _ := jwt.NewResolver("secret")
//    app.UseJWT(jwt.New(jwtResolver, validationHandler, app.NewJWTSecurity()))
//
func New(resolver *KeyResolver, validationFunc goa.Middleware, scheme *goa.JWTSecurity) goa.Middleware {
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

			var jwtKeyName string
			if resolver.jwtKeyNameField != "" {
				jwtKeyName = req.Header.Get(resolver.jwtKeyNameField)
			}

			// Make a copy of the current keys in the KeyResolver's key map
			resolver.RLock()
			keyMap := make(map[string][]interface{})
			for name, keys := range resolver.keyMap {
				keyMap[name] = keys
			}
			resolver.RUnlock()

			rsaKeys, hmacKeys := getKeys(jwtKeyName, keyMap)

			var token *jwt.Token
			var err error
			validated := false

			if len(rsaKeys) > 0 {
				token, err = validateRSAKeys(rsaKeys, "RS", incomingToken)
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
				return ErrJWTError(fmt.Sprint("JWT validation failed"))
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

func getKeys(jwtKeyName string, keyMap map[string][]interface{}) (
	rsaKeys []*rsa.PublicKey, hmacKeys []string) {
	// if jwtKeyName is a non-empty string, we will include only keys
	// under that name for validation, otherwise we will try all keys.
	if jwtKeyName != "" {
		for _, key := range keyMap[jwtKeyName] {
			switch key.(type) {
			case *rsa.PublicKey:
				rsaKeys = append(rsaKeys, key.(*rsa.PublicKey))
			case string:
				hmacKeys = append(hmacKeys, key.(string))
			}
		}
	} else {
		for _, keyList := range keyMap {
			for _, key := range keyList {
				switch key.(type) {
				case *rsa.PublicKey:
					rsaKeys = append(rsaKeys, key.(*rsa.PublicKey))
				case string:
					hmacKeys = append(hmacKeys, key.(string))
				}
			}
		}
	}
	return
}
