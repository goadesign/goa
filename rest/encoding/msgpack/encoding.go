package msgpack

import (
	"io"

	"github.com/goadesign/goa/rest"
	"github.com/ugorji/go/codec"
)

// Enforce that codec.Decoder satisfies goa.ResettableDecoder at compile time
var (
	_ rest.ResettableDecoder = (*codec.Decoder)(nil)
	_ rest.ResettableEncoder = (*codec.Encoder)(nil)

	Handle codec.MsgpackHandle
)

// NewDecoder returns a msgpack decoder.
func NewDecoder(r io.Reader) rest.Decoder {
	return codec.NewDecoder(r, &Handle)
}

// NewEncoder returns a msgpack encoder.
func NewEncoder(w io.Writer) rest.Encoder {
	return codec.NewEncoder(w, &Handle)
}
