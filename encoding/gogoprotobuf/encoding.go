package gogoprotobuf

import (
	"bytes"
	"errors"
	"io"

	"github.com/goadesign/goa"
	"github.com/gogo/protobuf/proto"
)

// Enforce that codec.Decoder satisfies goa.ResettableDecoder at compile time
var (
	_ goa.ResettableDecoder = (*ProtoDecoder)(nil)
	_ goa.ResettableEncoder = (*ProtoEncoder)(nil)
)

type (
	// ProtoDecoder stores state between Reset and Decode
	ProtoDecoder struct {
		pBuf *proto.Buffer
		bBuf *bytes.Buffer
		r    io.Reader
	}

	// ProtoEncoder stores state between Reset and Encode
	ProtoEncoder struct {
		pBuf *proto.Buffer
		w    io.Writer
	}
)

// NewDecoder returns a new proto.Decoder that satisfies goa.Decoder
func NewDecoder(r io.Reader) goa.Decoder {
	return &ProtoDecoder{
		pBuf: proto.NewBuffer(nil),
		bBuf: &bytes.Buffer{},
		r:    r,
	}
}

// Decode unmarshals an io.Reader into proto.Message v
func (dec *ProtoDecoder) Decode(v interface{}) error {
	msg, ok := v.(proto.Message)
	if !ok {
		return errors.New("Cannot decode into struct that doesn't implement proto.Message")
	}

	var err error

	// derekperkins TODO: pipe reader directly to proto.Buffer
	if _, err = dec.bBuf.ReadFrom(dec.r); err != nil {
		return err
	}
	dec.pBuf.SetBuf(dec.bBuf.Bytes())

	return dec.pBuf.Unmarshal(msg)
}

// Reset stores the new reader and resets its bytes.Buffer and proto.Buffer
func (dec *ProtoDecoder) Reset(r io.Reader) {
	dec.pBuf.Reset()
	dec.bBuf.Reset()
	dec.r = r
}

// NewEncoder returns a new proto.Encoder that satisfies goa.Encoder
func NewEncoder(w io.Writer) goa.Encoder {
	return &ProtoEncoder{
		pBuf: proto.NewBuffer(nil),
		w:    w,
	}
}

// Encode marshals a proto.Message and writes it to an io.Writer
func (enc *ProtoEncoder) Encode(v interface{}) error {
	msg, ok := v.(proto.Message)
	if !ok {
		return errors.New("Cannot encode struct that doesn't implement proto.Message")
	}

	var err error

	// derekperkins TODO: pipe marshal directly to writer
	if err = enc.pBuf.Marshal(msg); err != nil {
		return err
	}

	if _, err = enc.w.Write(enc.pBuf.Bytes()); err != nil {
		return err
	}
	return nil
}

// Reset stores the new writer and resets its proto.Buffer
func (enc *ProtoEncoder) Reset(w io.Writer) {
	enc.pBuf.Reset()
	enc.w = w
}
