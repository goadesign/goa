package capnproto2

import (
	"errors"
	"io"

	"zombiezen.com/go/capnproto2"

	"github.com/goadesign/goa"
)

// Enforce that codec.Decoder satisfies goa.Decoder at compile time
var (
	_ goa.Decoder = (*ProtoDecoder)(nil)
	_ goa.Encoder = (*ProtoEncoder)(nil)
)

type (
	// ProtoDecoder stores state between Reset and Decode
	ProtoDecoder struct {
		dec *capnp.Decoder
	}

	// ProtoEncoder stores state between Reset and Encode
	ProtoEncoder struct {
		enc *capnp.Encoder
	}
)

// NewDecoder returns a new capnp.Decoder that satisfies goa.Decoder
func NewDecoder(r io.Reader) goa.Decoder {
	return &ProtoDecoder{
		dec: capnp.NewDecoder(r),
	}
}

// NewPackedDecoder returns a new packed capnp.Decoder that satisfies goa.Decoder
func NewPackedDecoder(r io.Reader) goa.Decoder {
	return &ProtoDecoder{
		dec: capnp.NewPackedDecoder(r),
	}
}

// Decode unmarshals an io.Reader into *capnp.Message v
func (dec *ProtoDecoder) Decode(v interface{}) error {
	msg, ok := v.(*capnp.Message)
	if !ok {
		return errors.New("Cannot decode into struct that doesn't implement *capnp.Message")
	}

	newMsg, err := dec.dec.Decode()
	if err != nil {
		return err
	}

	if newMsg == nil {
		msg = nil
		return nil
	}

	*msg = *newMsg
	return nil
}

// NewEncoder returns a new capnp.Encoder that satisfies goa.Encoder
func NewEncoder(w io.Writer) goa.Encoder {
	return &ProtoEncoder{
		enc: capnp.NewEncoder(w),
	}
}

// NewPackedEncoder returns a new packed capnp.Encoder that satisfies goa.Encoder
func NewPackedEncoder(w io.Writer) goa.Encoder {
	return &ProtoEncoder{
		enc: capnp.NewPackedEncoder(w),
	}
}

// Encode marshals a *capnp.Message and writes it to an io.Writer
func (enc *ProtoEncoder) Encode(v interface{}) error {
	msg, ok := v.(*capnp.Message)
	if !ok {
		return errors.New("Cannot encode struct that doesn't implement *capnp.Message")
	}

	var err error

	if err = enc.enc.Encode(msg); err != nil {
		return err
	}

	return nil
}
