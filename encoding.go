package goa

import (
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/context"
)

type (
	// DecoderFunc instantiates a decoder that decodes data read from the given io reader.
	DecoderFunc func(r io.Reader) Decoder

	// A Decoder unmarshals an io.Reader into an interface
	Decoder interface {
		Decode(v interface{}) error
	}

	// The ResettableDecoder is used to determine whether or not a Decoder can be reset and
	// thus safely reused in a sync.Pool
	ResettableDecoder interface {
		Decoder
		Reset(r io.Reader)
	}

	// decoderPool smartly determines whether to instantiate a new Decoder or reuse
	// one from a sync.Pool
	decoderPool struct {
		fn   DecoderFunc
		pool *sync.Pool
	}

	// EncoderFunc instantiates an encoder that encodes data into the given writer.
	EncoderFunc func(w io.Writer) Encoder

	// An Encoder marshals from an interface into an io.Writer
	Encoder interface {
		Encode(v interface{}) error
	}

	// The ResettableEncoder is used to determine whether or not a Encoder can be reset and
	// thus safely reused in a sync.Pool
	ResettableEncoder interface {
		Encoder
		Reset(w io.Writer)
	}

	// encoderPool smartly determines whether to instantiate a new Encoder or reuse
	// one from a sync.Pool
	encoderPool struct {
		fn   EncoderFunc
		pool *sync.Pool
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

// DecodeRequest retrives the request body and `Content-Type` header and uses Decode
// to unmarshal into the provided `interface{}`
func (service *Service) DecodeRequest(req *http.Request, v interface{}) error {
	body, contentType := req.Body, req.Header.Get("Content-Type")
	defer body.Close()

	if err := service.Decode(v, body, contentType); err != nil {
		return fmt.Errorf("failed to decode request body with content type %#v: %s", contentType, err)
	}

	return nil
}

// Decode uses registered Decoders to unmarshal a body based on the contentType
func (service *Service) Decode(v interface{}, body io.Reader, contentType string) error {
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
	p = service.decoderPools[contentType]
	if p == nil {
		p = service.decoderPools["*/*"]
	}
	if p == nil {
		return nil
	}

	// the decoderPool will handle whether or not a pool is actually in use
	decoder := p.Get(body)
	defer p.Put(decoder)
	if err := decoder.Decode(v); err != nil {
		return err
	}

	return nil
}

// Decoder sets a specific decoder to be used for the specified content types. If
// a decoder is already registered, it will be overwritten.
func (service *Service) Decoder(f DecoderFunc, contentTypes ...string) {
	p := newDecodePool(f)

	for _, contentType := range contentTypes {
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil {
			mediaType = contentType
		}
		service.decoderPools[mediaType] = p
	}
}

// newDecodePool checks to see if the DecoderFunc returns reusable decoders
// and if so, creates a pool
func newDecodePool(f DecoderFunc) *decoderPool {
	// get a new decoder and type assert to see if it can be reset
	decoder := f(nil)
	rd, ok := decoder.(ResettableDecoder)

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

// Get returns an already reset Decoder from the pool if possible
// or creates a new one if necessary
func (p *decoderPool) Get(r io.Reader) Decoder {
	if p.pool == nil {
		return p.fn(r)
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

// EncodeResponse uses registered Encoders to marshal the response body based on the request
// `Accept` header and writes it to the http.ResponseWriter
func (service *Service) EncodeResponse(ctx context.Context, v interface{}) error {
	now := time.Now()
	accept := Request(ctx).Header.Get("Accept")
	if accept == "" {
		accept = "*/*"
	}
	var contentType string
	for _, t := range service.encodableContentTypes {
		if accept == "*/*" || accept == t {
			contentType = accept
			break
		}
	}
	defer MeasureSince([]string{"goa", "encode", contentType}, now)
	p := service.encoderPools[contentType]
	if p == nil && contentType != "*/*" {
		p = service.encoderPools["*/*"]
	}
	if p == nil {
		return fmt.Errorf("No encoder registered for %s and no default encoder", contentType)
	}

	// the encoderPool will handle whether or not a pool is actually in use
	encoder := p.Get(Response(ctx))
	if err := encoder.Encode(v); err != nil {
		return err
	}
	p.Put(encoder)

	return nil
}

// Encoder sets a specific encoder to be used for the specified content types. If
// an encoder is already registered, it will be overwritten.
func (service *Service) Encoder(f EncoderFunc, contentTypes ...string) {
	p := newEncodePool(f)
	for _, contentType := range contentTypes {
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil {
			mediaType = contentType
		}
		service.encoderPools[mediaType] = p
	}

	// Rebuild a unique index of registered content encoders to be used in EncodeResponse
	service.encodableContentTypes = make([]string, 0, len(service.encoderPools))
	for contentType := range service.encoderPools {
		service.encodableContentTypes = append(service.encodableContentTypes, contentType)
	}

}

// newEncodePool checks to see if the EncoderFactory returns reusable encoders
// and if so, creates a pool
func newEncodePool(f EncoderFunc) *encoderPool {
	// get a new encoder and type assert to see if it can be reset
	encoder := f(nil)
	re, ok := encoder.(ResettableEncoder)

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

// Get returns an already reset Encoder from the pool if possible
// or creates a new one if necessary
func (p *encoderPool) Get(w io.Writer) Encoder {
	if p.pool == nil {
		return p.fn(w)
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
