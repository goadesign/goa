package goa

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"io"
	"strings"
	"sync"
)

type (
	// A DecoderFactory generates custom decoders
	DecoderFactory interface {
		NewDecoder(r io.Reader) Decoder
	}

	// A Decoder unmarshals an io.Reader into an interface
	Decoder interface {
		Decode(v interface{}) error
	}

	// A DecodeFunc unmarshals body into v
	DecodeFunc func(ctx *Context, body io.ReadCloser, v interface{}) error

	// The ResettableDecoder is used to determine whether or not a Decoder can be reset and
	// thus safely reused in a sync.Pool
	ResettableDecoder interface {
		Decoder
		Reset(r io.Reader) error
	}

	// decoderPool smartly determines whether to instantiate a new Decoder or reuse
	// one from a sync.Pool
	decoderPool struct {
		factory DecoderFactory
		pool    *sync.Pool
	}

	// A EncoderFactory generates custom encoders
	EncoderFactory interface {
		NewEncoder(w io.Writer) Encoder
	}

	// An Encoder marshals from an interface into an io.Writer
	Encoder interface {
		Encode(v interface{}) error
	}

	// The ResettableEncoder is used to determine whether or not a Encoder can be reset and
	// thus safely reused in a sync.Pool
	ResettableEncoder interface {
		Encoder
		Reset(w io.Writer) error
	}

	// encoderPool smartly determines whether to instantiate a new Encoder or reuse
	// one from a sync.Pool
	encoderPool struct {
		factory EncoderFactory
		pool    *sync.Pool
	}

	// jsonFactory uses encoding/json to act as an DecoderFactory and EncoderFactory
	jsonFactory struct{}

	// gobFactory uses encoding/gob to act as an DecoderFactory and EncoderFactory
	gobFactory struct{}

	// // Unmarshaler is the interface implemented by objects that can unmarshal themselves.
	// // The input can be assumed to be a valid encoding that matches the Content-Type request header.
	// // Unmarshal must copy the data if it wishes to retain the data after returning.
	// Unmarshaler interface {
	// 	Unmarshal(v interface{}, shouldUnmarshal bool) error
	// }
)

var (
	// JSONContentTypes is a slice of default Content-Type headers that will use stdlib
	// encoding/json to unmarshal unless overwritten using SetDecoder
	JSONContentTypes = []string{"application/json", "application/text+json"}

	// GobContentTypes is a slice of default Content-Type headers that will use stdlib
	// encoding/gob to unmarshal unless overwritten using SetDecoder
	GobContentTypes = []string{"application/gob"}
)

// initEncoding initializes all the decoder/encoder pools with the Content-Types found
// in JSONContentTypes and GobContentTypes. JSON is set as the default decoder.
func (app *Application) initEncoding() {
	jf := &jsonFactory{}
	dp := newDecodePool(jf)
	ep := newEncodePool(jf)
	app.defaultDecoderPool = dp
	app.defaultEncoderPool = ep

	contentTypeCount := len(JSONContentTypes) + len(GobContentTypes)
	decoders := make(map[string]*decoderPool, contentTypeCount)
	encoders := make(map[string]*encoderPool, contentTypeCount)

	for _, contentType := range JSONContentTypes {
		decoders[contentType] = dp
		encoders[contentType] = ep
	}

	gf := &gobFactory{}
	dp = newDecodePool(gf)
	ep = newEncodePool(gf)
	for _, contentType := range GobContentTypes {
		decoders[contentType] = dp
		encoders[contentType] = ep
	}

	app.encoderPools = encoders
	app.decoderPools = decoders
}

// GetDecodeFunc returns a func that executes a registered Decoder based on
// the request "Content-Type" header
func (app *Application) GetDecodeFunc(contentType string) DecodeFunc {
	return func(ctx *Context, body io.ReadCloser, v interface{}) error {
		defer body.Close()

		p, ok := app.decoderPools[strings.ToLower(contentType)] // headers are supposed to be case insensitive
		if !ok {
			p = app.defaultDecoderPool
		}

		// the decoderPool will handle whether or not a pool is actually in use
		decoder := p.Get(body)
		if err := decoder.Decode(v); err != nil {
			// TODO: log out error details
			return err
		}
		p.Put(decoder)

		return nil
	}
}

// SetDecoder sets a specific decoder to be used for the specified content types. If
// a decoder is already registered, it will be overwritten.
func (app *Application) SetDecoder(f DecoderFactory, makeDefault bool, contentTypes ...string) {
	p := newDecodePool(f)
	for _, contentType := range contentTypes {
		app.decoderPools[strings.ToLower(contentType)] = p
	}

	if makeDefault {
		app.defaultDecoderPool = p
	}
}

// newDecodePool checks to see if the DecoderFactory returns reusable decoders
// and if so, creates a pool
func newDecodePool(f DecoderFactory) *decoderPool {
	// get a new decoder and type assert to see if it can be reset
	decoder := f.NewDecoder(nil)
	rd, ok := decoder.(ResettableDecoder)

	p := &decoderPool{
		factory: f,
	}

	// if the decoder can be reset, create a pool and put the typed decoder in
	if ok {
		p.pool = &sync.Pool{
			New: func() interface{} { return f.NewDecoder(nil) },
		}
		p.pool.Put(rd)
	}

	return p
}

// Get returns an already reset Decoder from the pool if possible
// or creates a new one if necessary
func (p *decoderPool) Get(r io.Reader) Decoder {
	if p.pool == nil {
		return p.factory.NewDecoder(r)
	}

	decoder := p.pool.Get().(ResettableDecoder)
	decoder.Reset(r)
	return decoder
}

// Put returns a Decoder into the pool if possible
func (p *decoderPool) Put(d Decoder) {
	if p.pool == nil {
		return
	}
	p.pool.Put(d)
}

// Encode uses registered Encoders to marshal the response body based on
// the request "Accpt" header
func (app *Application) Encode(ctx *Context, v interface{}, contentType string) ([]byte, error) {
	p, ok := app.encoderPools[strings.ToLower(contentType)] // headers are supposed to be case insensitive
	if !ok {
		p = app.defaultEncoderPool
	}

	// TODO: write directly to ctx.ResponseWriter
	buf := &bytes.Buffer{}

	// the encoderPool will handle whether or not a pool is actually in use
	encoder := p.Get(buf)
	if err := encoder.Encode(v); err != nil {
		// TODO: log out error details
		return nil, err
	}
	p.Put(encoder)

	return buf.Bytes(), nil
}

// SetEncoder sets a specific encoder to be used for the specified content types. If
// a encoder is already registered, it will be overwritten.
func (app *Application) SetEncoder(f EncoderFactory, makeDefault bool, contentTypes ...string) {
	p := newEncodePool(f)
	for _, contentType := range contentTypes {
		app.encoderPools[strings.ToLower(contentType)] = p
	}

	if makeDefault {
		app.defaultEncoderPool = p
	}
}

// newEncodePool checks to see if the EncoderFactory returns reusable encoders
// and if so, creates a pool
func newEncodePool(f EncoderFactory) *encoderPool {
	// get a new encoder and type assert to see if it can be reset
	encoder := f.NewEncoder(nil)
	re, ok := encoder.(ResettableEncoder)

	p := &encoderPool{
		factory: f,
	}

	// if the encoder can be reset, create a pool and put the typed encoder in
	if ok {
		p.pool = &sync.Pool{
			New: func() interface{} { return f.NewEncoder(nil) },
		}
		p.pool.Put(re)
	}

	return p
}

// Get returns an already reset Encoder from the pool if possible
// or creates a new one if necessary
func (p *encoderPool) Get(w io.Writer) Encoder {
	if p.pool == nil {
		return p.factory.NewEncoder(w)
	}

	encoder := p.pool.Get().(ResettableEncoder)
	encoder.Reset(w)
	return encoder
}

// Put returns a Decoder into the pool if possible
func (p *encoderPool) Put(e Encoder) {
	if p.pool == nil {
		return
	}
	p.pool.Put(e)
}

// NewDecoder returns a new json.Decoder every time because it can't be reset
func (f *jsonFactory) NewDecoder(r io.Reader) Decoder {
	return json.NewDecoder(r)
}

// NewEncoder returns a new json.Encoder every time because it can't be reset
func (f *jsonFactory) NewEncoder(w io.Writer) Encoder {
	return json.NewEncoder(w)
}

// NewDecoder returns a new gob.Decoder every time because it can't be reset
func (f *gobFactory) NewDecoder(r io.Reader) Decoder {
	return gob.NewDecoder(r)
}

// NewEncoder returns a new gob.Encoder every time because it can't be reset
func (f *gobFactory) NewEncoder(w io.Writer) Encoder {
	return gob.NewEncoder(w)
}
