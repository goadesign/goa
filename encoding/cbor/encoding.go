package cbor

import (
	"io"

	"github.com/goadesign/goa"
	"github.com/ugorji/go/codec"
)

var (
	// Handle used by encoder and decoder.
	Handle codec.CborHandle

	// Enforce that codec.Decoder satisfies goa.ResettableDecoder at compile time
	_ goa.ResettableDecoder = (*codec.Decoder)(nil)
	_ goa.ResettableEncoder = (*codec.Encoder)(nil)
)

// NewDecoder returns a cbor decoder.
func NewDecoder(r io.Reader) goa.Decoder {
	return codec.NewDecoder(r, &Handle)
}

// NewEncoder returns a cbor encoder.
func NewEncoder(w io.Writer) goa.Encoder {
	return codec.NewEncoder(w, &Handle)
}
