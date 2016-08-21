package json

import (
	"io"

	"github.com/goadesign/goa"
	"github.com/ugorji/go/codec"
)

// Enforce that codec.Decoder satisfies goa.ResettableDecoder at compile time
var (
	// Handle used by encoder and decoder.
	Handle codec.JsonHandle

	_ goa.ResettableDecoder = (*codec.Decoder)(nil)
	_ goa.ResettableEncoder = (*codec.Encoder)(nil)
)

// NewDecoder returns a JSON decoder.
func NewDecoder(r io.Reader) goa.Decoder {
	return codec.NewDecoder(r, &Handle)
}

// NewEncoder returns a JSON encoder.
func NewEncoder(w io.Writer) goa.Encoder {
	return codec.NewEncoder(w, &Handle)
}
