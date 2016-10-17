package jwt

import (
	"errors"

	"github.com/goadesign/goa"
)

var (
	// ErrEmptyHeaderName is returned when the header value given to the standard key resolver
	// constructor is empty.
	ErrEmptyHeaderName = errors.New("header name must not be empty")

	// ErrInvalidKey is returned when a key is not of type string, []string, *rsa.PublicKey or
	// []*rsa.PublicKey.
	ErrInvalidKey = errors.New("invalid parameter, the only keys accepted " +
		"are *rsa.publicKey, []*rsa.PublicKey (for RSA-based algorithms) or a " +
		"signing secret string, []string (for HS algorithms)")

	// ErrKeyDoesNotExist is returned when a key cannot be found by the provided key name.
	ErrKeyDoesNotExist = errors.New("key does not exist")

	// ErrJWTError is the error returned by this middleware when any sort of validation or
	// assertion fails during processing.
	ErrJWTError = goa.NewErrorClass("jwt_security_error", 401)
)
