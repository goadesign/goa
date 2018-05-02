// +build !js

//This is just a declaration of the uuid.UUID which doesn't work with gopherjs
//See uuid_js.go for the JS implementation

package uuid

import (
	"database/sql/driver"
	"fmt"

	"github.com/satori/go.uuid"
)

// FromString Wrapper around the real FromString
func FromString(input string) (UUID, error) {
	u, err := uuid.FromString(input)
	return UUID(u), err
}

// NewV4 Wrapper over the real NewV4 method
func NewV4() UUID {
	return UUID(uuid.Must(uuid.NewV4()))
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

// Value implements the driver.Valuer interface.
func (u UUID) Value() (driver.Value, error) {
	return u.String(), nil
}

// Scan implements the sql.Scanner interface.
// A 16-byte slice is handled by UnmarshalBinary, while
// a longer byte slice or a string is handled by UnmarshalText.
func (u *UUID) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		if len(src) == uuid.Size {
			return u.UnmarshalBinary(src)
		}
		return u.UnmarshalText(src)

	case string:
		return u.UnmarshalText([]byte(src))
	}

	return fmt.Errorf("uuid: cannot convert %T to UUID", src)
}
