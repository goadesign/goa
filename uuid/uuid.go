// +build !js

//This is just a declaration of the uuid.UUID which doesn't work with gopherjs
//See uuid_js.go for the JS implementation

package uuid

import (
	"github.com/satori/go.uuid"
)

// FromString Wrapper around the real FromString
func FromString(input string) (UUID, error) {
	u, err := uuid.FromString(input)
	return UUID(u), err
}

// NewV4 Wrapper over the real NewV4 method
func NewV4() UUID {
	return UUID(uuid.NewV4())
}
