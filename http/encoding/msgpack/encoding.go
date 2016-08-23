package msgpack

import (
	"io"

	"github.com/goadesign/goa"
	"github.com/ugorji/go/codec"
)

// Enforce that codec.Decoder satisfies goa.ResettableDecoder at compile time
var (
	_ goa.ResettableDecoder = (*codec.Decoder)(nil)
	_ goa.ResettableEncoder = (*codec.Encoder)(nil)

	Handle codec.MsgpackHandle
)

// NewDecoder returns a msgpack decoder.
func NewDecoder(r io.Reader) goa.Decoder {
	return codec.NewDecoder(r, &Handle)
}

// NewEncoder returns a msgpack encoder.
func NewEncoder(w io.Writer) goa.Encoder {
	return codec.NewEncoder(w, &Handle)
}
