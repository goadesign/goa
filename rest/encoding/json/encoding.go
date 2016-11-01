package json

import (
	"io"

	"github.com/goadesign/goa/rest"
	"github.com/ugorji/go/codec"
)

// Enforce that codec.Decoder satisfies goa.ResettableDecoder at compile time
var (
	// Handle used by encoder and decoder.
	Handle codec.JsonHandle

	_ rest.ResettableDecoder = (*codec.Decoder)(nil)
	_ rest.ResettableEncoder = (*codec.Encoder)(nil)
)

// NewDecoder returns a JSON decoder.
func NewDecoder(r io.Reader) rest.Decoder {
	return codec.NewDecoder(r, &Handle)
}

// NewEncoder returns a JSON encoder.
func NewEncoder(w io.Writer) rest.Encoder {
	return codec.NewEncoder(w, &Handle)
}
