package goa

import (
	"bufio"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"io"
	"mime"
	"net/http"
	"sync"

	"github.com/golang/gddo/httputil"
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

	// The ResettableDecoder is used to determine whether or not a Decoder can be reset and
	// thus safely reused in a sync.Pool
	ResettableDecoder interface {
		Decoder
		Reset(r io.Reader)
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
		Reset(w io.Writer)
	}

	// encoderPool smartly determines whether to instantiate a new Encoder or reuse
	// one from a sync.Pool
	encoderPool struct {
		factory EncoderFactory
		pool    *sync.Pool
	}

	// jsonFactory uses encoding/json to act as an DecoderFactory and EncoderFactory
	jsonFactory struct{}

	// xmlFactory uses encoding/xml to act as an DecoderFactory and EncoderFactory
	xmlFactory struct{}

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

	// XMLContentTypes is a slice of default Content-Type headers that will use stdlib
	// encoding/xml to unmarshal unless overwritten using SetDecoder
	XMLContentTypes = []string{"application/xml"}

	// GobContentTypes is a slice of default Content-Type headers that will use stdlib
	// encoding/gob to unmarshal unless overwritten using SetDecoder
	GobContentTypes = []string{"application/gob"}
)

// DecodeRequest uses registered Decoders to unmarshal the request body based on
// the request `Content-Type` header
func (app *Application) DecodeRequest(ctx *Context, v interface{}) error {
	body := ctx.Request().Body
	contentType := ctx.Request().Header.Get("Content-Type")
	defer body.Close()

	var p *decoderPool
	if contentType == "" {
		// Default to JSON
		contentType = "application/json"
	} else {
		if mediaType, _, err := mime.ParseMediaType(contentType); err == nil {
			contentType = mediaType
		}
	}
	p = app.decoderPools[contentType]

	// Do not attempt to decode request bodies for which no decoder has been setup.
	// These may be handled differently by the service.
	if p != nil {
		// the decoderPool will handle whether or not a pool is actually in use
		decoder := p.Get(body)
		if err := decoder.Decode(v); err != nil {
			ctx.Error(err.Error(), "ContentType", contentType)
			return err
		}
		p.Put(decoder)
	}

	return nil
}

func detectContentType(ctx *Context, body io.Reader) string {
	bodyBuf := bufio.NewReader(body)
	// http.DetectContentType uses a max of 512 bytes
	peekSize := 512
	if bodyBuf.Buffered() < peekSize {
		peekSize = bodyBuf.Buffered()
	}
	b, err := bodyBuf.Peek(peekSize)
	if err != nil {
		return "*/*"
	}

	return http.DetectContentType(b)
}

// SetDecoder sets a specific decoder to be used for the specified content types. If
// a decoder is already registered, it will be overwritten.
func (app *Application) SetDecoder(f DecoderFactory, makeDefault bool, contentTypes ...string) {
	p := newDecodePool(f)

	for _, contentType := range contentTypes {
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil {
			mediaType = contentType
		}
		app.decoderPools[mediaType] = p
	}

	if makeDefault {
		app.decoderPools["*/*"] = p
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

// EncodeResponse uses registered Encoders to marshal the response body based on the request
// `Accept` header and writes it to the http.ResponseWriter
func (app *Application) EncodeResponse(ctx *Context, v interface{}) error {
	contentType := httputil.NegotiateContentType(ctx.Request(), app.encodableContentTypes, "*/*")
	p := app.encoderPools[contentType]

	// the encoderPool will handle whether or not a pool is actually in use
	encoder := p.Get(ctx)
	if err := encoder.Encode(v); err != nil {
		// TODO: log out error details
		return err
	}
	p.Put(encoder)

	return nil
}

// SetEncoder sets a specific encoder to be used for the specified content types. If
// a encoder is already registered, it will be overwritten.
func (app *Application) SetEncoder(f EncoderFactory, makeDefault bool, contentTypes ...string) {
	p := newEncodePool(f)
	for _, contentType := range contentTypes {
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil {
			mediaType = contentType
		}
		app.encoderPools[mediaType] = p
	}

	if makeDefault {
		app.encoderPools["*/*"] = p
	}

	// Rebuild a unique index of registered content encoders to be used in EncodeResponse
	app.encodableContentTypes = make([]string, 0, len(app.encoderPools))
	for contentType := range app.encoderPools {
		app.encodableContentTypes = append(app.encodableContentTypes, contentType)
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

// encoding/json default encoder/decoder

// JSONDecoderFactory returns a struct that can generate new json.Decoders
func JSONDecoderFactory() DecoderFactory {
	return &jsonFactory{}
}

// NewDecoder returns a new json.Decoder
func (f *jsonFactory) NewDecoder(r io.Reader) Decoder {
	return json.NewDecoder(r)
}

// JSONEncoderFactory returns a struct that can generate new json.Encoders
func JSONEncoderFactory() EncoderFactory {
	return &jsonFactory{}
}

// NewEncoder returns a new json.Encoder
func (f *jsonFactory) NewEncoder(w io.Writer) Encoder {
	return json.NewEncoder(w)
}

// encoding/xml default encoder/decoder

// XMLDecoderFactory returns a struct that can generate new xml.Decoders
func XMLDecoderFactory() DecoderFactory {
	return &xmlFactory{}
}

// NewDecoder returns a new xml.Decoder
func (f *xmlFactory) NewDecoder(r io.Reader) Decoder {
	return xml.NewDecoder(r)
}

// XMLEncoderFactory returns a struct that can generate new xml.Encoders
func XMLEncoderFactory() EncoderFactory {
	return &xmlFactory{}
}

// NewEncoder returns a new xml.Encoder
func (f *xmlFactory) NewEncoder(w io.Writer) Encoder {
	return xml.NewEncoder(w)
}

// encoding/gob default encoder/decoder

// GobDecoderFactory returns a struct that can generate new gob.Decoders
func GobDecoderFactory() DecoderFactory {
	return &gobFactory{}
}

// NewDecoder returns a new gob.Decoder
func (f *gobFactory) NewDecoder(r io.Reader) Decoder {
	return gob.NewDecoder(r)
}

// GobEncoderFactory returns a struct that can generate new gob.Encoders
func GobEncoderFactory() EncoderFactory {
	return &gobFactory{}
}

// NewEncoder returns a new gob.Encoder
func (f *gobFactory) NewEncoder(w io.Writer) Encoder {
	return gob.NewEncoder(w)
}
