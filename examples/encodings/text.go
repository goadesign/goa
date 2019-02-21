package encodings

import (
	"context"
	"fmt"
	"log"

	text "goa.design/goa/examples/encodings/gen/text"
)

// text service example implementation.
// The example methods log the requests and return zero values.
type textsrvc struct {
	logger *log.Logger
}

// NewText returns the text service implementation.
func NewText(logger *log.Logger) text.Service {
	return &textsrvc{logger}
}

// Concatstrings implements Concatstrings.
func (s *textsrvc) Concatstrings(ctx context.Context, p *text.ConcatstringsPayload) (res string, err error) {
	return fmt.Sprintf("%v%v", p.A, p.B), nil
}

// Concatbytes implements Concatbytes.
func (s *textsrvc) Concatbytes(ctx context.Context, p *text.ConcatbytesPayload) (res []byte, err error) {
	return []byte(fmt.Sprintf("%v%v", p.A, p.B)), nil
}

// Concatstringfield implements Concatstringfield.
func (s *textsrvc) Concatstringfield(ctx context.Context, p *text.ConcatstringfieldPayload) (res *text.MyConcatenation, err error) {
	str := fmt.Sprintf("%v%v", p.A, p.B)
	return &text.MyConcatenation{Stringfield: &str}, nil
}

// Concatbytesfield implements Concatbytesfield.
func (s *textsrvc) Concatbytesfield(ctx context.Context, p *text.ConcatbytesfieldPayload) (res *text.MyConcatenation, err error) {
	b := []byte(fmt.Sprintf("%v%v", p.A, p.B))
	return &text.MyConcatenation{Bytesfield: b}, nil
}
