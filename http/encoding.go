package goa

import (
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime"
	"sync"
	"time"
)

type (
	// DecoderFunc instantiates a decoder that decodes data read from the given io reader.
	DecoderFunc func(r io.Reader) Decoder

	// A Decoder unmarshals an io.Reader into an interface.
	Decoder interface {
		Decode(v interface{}) error
	}

	// ResettableDecoder is used to determine whether or not a Decoder can be reset and thus
	// safely reused in a sync.Pool.
	ResettableDecoder interface {
		Decoder
		Reset(r io.Reader)
	}

	// decoderPool smartly determines whether to instantiate a new Decoder or reuse one from a
	// sync.Pool.
	decoderPool struct {
		fn   DecoderFunc
		pool *sync.Pool
	}

	// EncoderFunc instantiates an encoder that encodes data into the given writer.
	EncoderFunc func(w io.Writer) Encoder

	// An Encoder marshals from an interface into an io.Writer.
	Encoder interface {
		Encode(v interface{}) error
	}

	// The ResettableEncoder is used to determine whether or not a Encoder can be reset and
	// thus safely reused in a sync.Pool.
	ResettableEncoder interface {
		Encoder
		Reset(w io.Writer)
	}

	// encoderPool smartly determines whether to instantiate a new Encoder or reuse one from a
	// sync.Pool.
	encoderPool struct {
		fn   EncoderFunc
		pool *sync.Pool
	}

	// HTTPDecoder is a Decoder that decodes HTTP request or response bodies given a set of
	// known Content-Type to decoder mapping.
	HTTPDecoder struct {
		pools map[string]*decoderPool // Registered decoders
	}

	// HTTPEncoder is a Encoder that encodes HTTP request or response bodies given a set of
	// known Content-Type to encoder mapping.
	HTTPEncoder struct {
		pools        map[string]*encoderPool // Registered encoders
		contentTypes []string                // List of content types for type negotiation
	}
)

// NewJSONEncoder is an adapter for the encoding package JSON encoder.
func NewJSONEncoder(w io.Writer) Encoder { return json.NewEncoder(w) }

// NewJSONDecoder is an adapter for the encoding package JSON decoder.
func NewJSONDecoder(r io.Reader) Decoder { return json.NewDecoder(r) }

// NewXMLEncoder is an adapter for the encoding package XML encoder.
func NewXMLEncoder(w io.Writer) Encoder { return xml.NewEncoder(w) }

// NewXMLDecoder is an adapter for the encoding package XML decoder.
func NewXMLDecoder(r io.Reader) Decoder { return xml.NewDecoder(r) }

// NewGobEncoder is an adapter for the encoding package gob encoder.
func NewGobEncoder(w io.Writer) Encoder { return gob.NewEncoder(w) }

// NewGobDecoder is an adapter for the encoding package gob decoder.
func NewGobDecoder(r io.Reader) Decoder { return gob.NewDecoder(r) }

// NewHTTPEncoder creates an encoder that maps HTTP content types to low level encoders.
func NewHTTPEncoder() *HTTPEncoder {
	return &HTTPEncoder{
		pools: make(map[string]*encoderPool),
	}
}

// NewHTTPDecoder creates a decoder that maps HTTP content types to low level decoders.
func NewHTTPDecoder() *HTTPDecoder {
	return &HTTPDecoder{
		pools: make(map[string]*decoderPool),
	}
}

// Decode uses registered Decoders to unmarshal a body based on the contentType.
func (decoder *HTTPDecoder) Decode(v interface{}, body io.Reader, contentType string) error {
	now := time.Now()
	defer MeasureSince([]string{"goa", "decode", contentType}, now)
	var p *decoderPool
	if contentType == "" {
		// Default to JSON
		contentType = "application/json"
	} else {
		if mediaType, _, err := mime.ParseMediaType(contentType); err == nil {
			contentType = mediaType
		}
	}
	p = decoder.pools[contentType]
	if p == nil {
		p = decoder.pools["*/*"]
	}
	if p == nil {
		return nil
	}

	// the decoderPool will handle whether or not a pool is actually in use
	d := p.Get(body)
	defer p.Put(d)
	if err := d.Decode(v); err != nil {
		return err
	}

	return nil
}

// Register sets a specific decoder to be used for the specified content types. If a decoder is
// already registered, it is overwritten.
func (decoder *HTTPDecoder) Register(f DecoderFunc, contentTypes ...string) {
	p := newDecodePool(f)

	for _, contentType := range contentTypes {
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil {
			mediaType = contentType
		}
		decoder.pools[mediaType] = p
	}
}

// newDecodePool checks to see if the DecoderFunc returns reusable decoders and if so, creates a
// pool.
func newDecodePool(f DecoderFunc) *decoderPool {
	// get a new decoder and type assert to see if it can be reset
	d := f(nil)
	rd, ok := d.(ResettableDecoder)

	p := &decoderPool{fn: f}

	// if the decoder can be reset, create a pool and put the typed decoder in
	if ok {
		p.pool = &sync.Pool{
			New: func() interface{} { return f(nil) },
		}
		p.pool.Put(rd)
	}

	return p
}

// Get returns an already reset Decoder from the pool or creates a new one if necessary.
func (p *decoderPool) Get(r io.Reader) Decoder {
	if p.pool == nil {
		return p.fn(r)
	}

	d := p.pool.Get().(ResettableDecoder)
	d.Reset(r)
	return d
}

// Put returns a Decoder into the pool if possible.
func (p *decoderPool) Put(d Decoder) {
	if p.pool == nil {
		return
	}
	p.pool.Put(d)
}

// Encode uses the registered encoders and given content type to marshal and write the given value
// using the given writer.
func (encoder *HTTPEncoder) Encode(v interface{}, resp io.Writer, accept string) error {
	now := time.Now()
	if accept == "" {
		accept = "*/*"
	}
	var contentType string
	for _, t := range encoder.contentTypes {
		if accept == "*/*" || accept == t {
			contentType = accept
			break
		}
	}
	defer MeasureSince([]string{"goa", "encode", contentType}, now)
	p := encoder.pools[contentType]
	if p == nil && contentType != "*/*" {
		p = encoder.pools["*/*"]
	}
	if p == nil {
		return fmt.Errorf("No encoder registered for %s and no default encoder", contentType)
	}

	// the encoderPool will handle whether or not a pool is actually in use
	e := p.Get(resp)
	if err := e.Encode(v); err != nil {
		return err
	}
	p.Put(e)

	return nil
}

// Register sets a specific encoder to be used for the specified content types. If an encoder is
// already registered, it is overwritten.
func (encoder *HTTPEncoder) Register(f EncoderFunc, contentTypes ...string) {
	p := newEncodePool(f)
	for _, contentType := range contentTypes {
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil {
			mediaType = contentType
		}
		encoder.pools[mediaType] = p
	}

	// Rebuild a unique index of registered content encoders to be used in EncodeResponse
	encoder.contentTypes = make([]string, 0, len(encoder.pools))
	for contentType := range encoder.pools {
		encoder.contentTypes = append(encoder.contentTypes, contentType)
	}
}

// newEncodePool checks to see if the EncoderFactory returns reusable encoders and if so, creates
// a pool.
func newEncodePool(f EncoderFunc) *encoderPool {
	// get a new encoder and type assert to see if it can be reset
	e := f(nil)
	re, ok := e.(ResettableEncoder)

	p := &encoderPool{fn: f}

	// if the encoder can be reset, create a pool and put the typed encoder in
	if ok {
		p.pool = &sync.Pool{
			New: func() interface{} { return f(nil) },
		}
		p.pool.Put(re)
	}

	return p
}

// Get returns an already reset Encoder from the pool or creates a new one if necessary.
func (p *encoderPool) Get(w io.Writer) Encoder {
	if p.pool == nil {
		return p.fn(w)
	}

	e := p.pool.Get().(ResettableEncoder)
	e.Reset(w)
	return e
}

// Put returns a Decoder into the pool if possible.
func (p *encoderPool) Put(e Encoder) {
	if p.pool == nil {
		return
	}
	p.pool.Put(e)
}
