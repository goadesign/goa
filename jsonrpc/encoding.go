package jsonrpc

import (
	"encoding/json"
	"io"
)

type (
	// Decoder provides the actual decoding algorithm used to load HTTP
	// request and response bodies.
	Decoder interface {
		// Decode decodes into v.
		Decode(v any) error
	}

	// Encoder provides the actual encoding algorithm used to write HTTP
	// request and response bodies.
	Encoder interface {
		// Encode encodes v.
		Encode(v any) error
	}
)

// StdDecoder uses the standard library JSON decoder.
func StdDecoder(r io.Reader) Decoder {
	return json.NewDecoder(r)
}

// StdEncoder uses the standard library JSON encoder.
func StdEncoder(w io.Writer) Encoder {
	return json.NewEncoder(w)
}
