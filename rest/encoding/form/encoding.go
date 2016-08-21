/*
Package form provides a "application/x-www-form-encoding" encoder and decoder.  It uses
github.com/ajg/form for the actual implementation which can be used directly as well.  The goal of
this package is to raise awareness of the package above and its direct compatibility with goa.
*/
package form

import (
	"io"

	"github.com/ajg/form"
	"github.com/goadesign/goa"
)

// NewEncoder returns a form encoder that writes to w.
func NewEncoder(w io.Writer) goa.Encoder {
	return form.NewEncoder(w)
}

// NewDecoder returns a form decoder that reads from r.
func NewDecoder(r io.Reader) goa.Decoder {
	return form.NewDecoder(r)
}
