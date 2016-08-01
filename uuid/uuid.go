// +build !js

//This is just a declaration of the uuid.UUID which doesn't work with gopherjs
//See uuid_js.go for the JS implementation

package uuid

import "github.com/satori/go.uuid"

// FromString Wrapper around the real FromString
func FromString(input string) (UUID, error) {
	u, err := uuid.FromString(input)
	return UUID(u), err
}

// NewV4 Wrapper over the real NewV4 method
func NewV4() UUID {
	return UUID(uuid.NewV4())
}

// String Wrapper over the real String method
func (u UUID) String() string {
	return uuid.UUID(u).String()
}

// MarshalText Wrapper over the real MarshalText method
func (u UUID) MarshalText() (text []byte, err error) {
	return uuid.UUID(u).MarshalText()
}

// MarshalBinary Wrapper over the real MarshalBinary method
func (u UUID) MarshalBinary() ([]byte, error) {
	return uuid.UUID(u).MarshalBinary()
}

// UnmarshalBinary Wrapper over the real UnmarshalBinary method
func (u *UUID) UnmarshalBinary(data []byte) error {
	t := uuid.UUID{}
	err := t.UnmarshalBinary(data)
	for i, b := range t.Bytes() {
		u[i] = b
	}
	return err
}

// UnmarshalText Wrapper over the real UnmarshalText method
func (u *UUID) UnmarshalText(text []byte) error {
	t := uuid.UUID{}
	err := t.UnmarshalText(text)
	for i, b := range t.Bytes() {
		u[i] = b
	}
	return err
}
