package http

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"

	goa "goa.design/goa/v3/pkg"
)

const (
	// AcceptTypeKey is the context key used to store the value of the HTTP
	// request Accept-Type header. The value may be used by encoders and
	// decoders to implement a content type negotiation algorithm.
	AcceptTypeKey contextKey = iota + 1

	// ContentTypeKey is the context key used to store the value of the HTTP
	// response Content-Type header when explicitly set in the DSL. The value
	// may be used by encoders to set the header appropriately.
	ContentTypeKey
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

	// EncodingFunc allows a function with appropriate signature to act as a
	// Decoder/Encoder.
	EncodingFunc func(v any) error

	// private type used to define context keys.
	contextKey int
)

// RequestDecoder returns a HTTP request body decoder suitable for the given
// request. The decoder handles the following mime types:
//
//   - application/json using package encoding/json
//   - application/xml using package encoding/xml
//   - application/gob using package encoding/gob
//   - text/html and text/plain for strings
//
// RequestDecoder defaults to the JSON decoder if the request "Content-Type"
// header does not match any of the supported mime type or is missing
// altogether.
func RequestDecoder(r *http.Request) Decoder {
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		// default to JSON
		contentType = "application/json"
	} else {
		// sanitize
		if mediaType, _, err := mime.ParseMediaType(contentType); err == nil {
			contentType = mediaType
		}
	}
	switch contentType {
	case "application/json":
		return json.NewDecoder(r.Body)
	case "application/gob":
		return gob.NewDecoder(r.Body)
	case "application/xml":
		return xml.NewDecoder(r.Body)
	case "text/html", "text/plain":
		return newTextDecoder(r.Body, contentType)
	default:
		return newUnsupportedDecoder(contentType)
	}
}

// ResponseEncoder returns a HTTP response encoder leveraging the mime type
// set in the context under the AcceptTypeKey or the ContentTypeKey if any.
// The encoder supports the following mime types:
//
//   - application/json using package encoding/json
//   - application/xml using package encoding/xml
//   - application/gob using package encoding/gob
//   - text/html and text/plain for strings
//
// ResponseEncoder defaults to the JSON encoder if the context AcceptTypeKey or
// ContentTypeKey value does not match any of the supported mime types or is
// missing altogether.
func ResponseEncoder(ctx context.Context, w http.ResponseWriter) Encoder {
	negotiate := func(a string) (Encoder, string) {
		switch a {
		case "", "application/json":
			// default to JSON
			return json.NewEncoder(w), "application/json"
		case "application/xml":
			return xml.NewEncoder(w), "application/xml"
		case "application/gob":
			return gob.NewEncoder(w), "application/gob"
		case "text/html", "text/plain":
			return newTextEncoder(w, a), a
		}
		return nil, ""
	}
	var accept string
	{
		if a := ctx.Value(AcceptTypeKey); a != nil {
			accept = a.(string)
		}
	}
	var ct string
	{
		if a := ctx.Value(ContentTypeKey); a != nil {
			ct = a.(string)
		}
	}
	var (
		enc Encoder
		mt  string
		err error
	)
	{
		if ct != "" {
			// If content type explicitly set in the DSL, infer the response encoder
			// from the content type context key.
			if mt, _, err = mime.ParseMediaType(ct); err == nil {
				switch {
				case mt == "application/json" || strings.HasSuffix(mt, "+json"):
					enc = json.NewEncoder(w)
				case mt == "application/xml" || strings.HasSuffix(mt, "+xml"):
					enc = xml.NewEncoder(w)
				case mt == "application/gob" || strings.HasSuffix(mt, "+gob"):
					enc = gob.NewEncoder(w)
				case mt == "text/html" || mt == "text/plain" ||
					strings.HasSuffix(mt, "+html") || strings.HasSuffix(mt, "+txt"):
					enc = newTextEncoder(w, mt)
				default:
					enc = json.NewEncoder(w)
				}
			}
			SetContentType(w, mt)
			return enc
		}
		// If Accept header exists in the request, infer the response encoder
		// from the header value.
		if enc, mt = negotiate(accept); enc == nil {
			// attempt to normalize
			if mt, _, err = mime.ParseMediaType(accept); err == nil {
				enc, mt = negotiate(mt)
			}
		}
		if enc == nil {
			enc, mt = negotiate("")
		}
	}
	SetContentType(w, mt)
	return enc
}

// RequestEncoder returns a HTTP request encoder.
// The encoder uses package encoding/json.
func RequestEncoder(r *http.Request) Encoder {
	const k = "Content-Type"
	if h := r.Header.Get(k); h == "" {
		r.Header.Set(k, "application/json")
	}
	enc := new(jsonEncoder)
	r.Body = enc
	// GetBody enables request retry on HTTP/2 connections when the server
	// sends GOAWAY during graceful shutdown. Without GetBody, the HTTP transport
	// cannot retry because the request body has already been consumed.
	r.GetBody = enc.GetBody
	return enc
}

// jsonEncoder implements io.ReadCloser and provides GetBody functionality
// to support HTTP/2 request retries during server graceful shutdown (GOAWAY).
type jsonEncoder struct {
	b []byte
	r bytes.Reader
}

var errEncodeNotCalled = errors.New("RequestEncoder: Encode must be called prior to reading")

func (je *jsonEncoder) Read(b []byte) (n int, err error) {
	if len(je.b) == 0 {
		return 0, errEncodeNotCalled
	}
	return je.r.Read(b)
}

func (*jsonEncoder) Close() (err error) { return nil }

func (je *jsonEncoder) Encode(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	je.b = b
	je.r = *bytes.NewReader(b)
	return nil
}

// GetBody returns a new reader of the encoded bytes, enabling request retries.
// This is required for HTTP/2 connections to handle server GOAWAY during graceful shutdown.
func (je *jsonEncoder) GetBody() (io.ReadCloser, error) {
	if len(je.b) == 0 {
		return nil, errEncodeNotCalled
	}
	return io.NopCloser(bytes.NewReader(je.b)), nil
}

// ResponseDecoder returns a HTTP response decoder.
// The decoder handles the following content types:
//
//   - application/json using package encoding/json (default)
//   - application/xml using package encoding/xml
//   - application/gob using package encoding/gob
//   - text/html and text/plain for strings
func ResponseDecoder(resp *http.Response) Decoder {
	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		return json.NewDecoder(resp.Body)
	}
	if mediaType, _, err := mime.ParseMediaType(ct); err == nil {
		ct = mediaType
	}
	switch {
	case ct == "application/json" || strings.HasSuffix(ct, "+json"):
		return json.NewDecoder(resp.Body)
	case ct == "application/xml" || strings.HasSuffix(ct, "+xml"):
		return xml.NewDecoder(resp.Body)
	case ct == "application/gob" || strings.HasSuffix(ct, "+gob"):
		return gob.NewDecoder(resp.Body)
	case ct == "text/html" || ct == "text/plain" ||
		strings.HasSuffix(ct, "+html") || strings.HasSuffix(ct, "+txt"):
		return newTextDecoder(resp.Body, ct)
	default:
		return json.NewDecoder(resp.Body)
	}
}

// ErrorEncoder returns an encoder that encodes errors returned by service
// methods. The default encoder checks whether the error is a goa ServiceError
// struct and if so uses the error temporary and timeout fields to infer a
// proper HTTP status code and marshals the error struct to the body using the
// provided encoder. If the error is not a goa ServiceError struct then it is
// encoded as a permanent internal server error. This behavior as well as the
// shape of the response can be overridden by providing a non-nil formatter.
func ErrorEncoder(encoder func(context.Context, http.ResponseWriter) Encoder, formatter func(ctx context.Context, err error) Statuser) func(context.Context, http.ResponseWriter, error) error {
	return func(ctx context.Context, w http.ResponseWriter, err error) error {
		enc := encoder(ctx, w)
		if formatter == nil {
			formatter = NewErrorResponse
		}
		resp := formatter(ctx, err)
		w.WriteHeader(resp.StatusCode())
		return enc.Encode(resp)
	}
}

// Decode implements the Decoder interface. It simply calls f(v).
func (f EncodingFunc) Decode(v any) error { return f(v) }

// Encode implements the Encoder interface. It simply calls f(v).
func (f EncodingFunc) Encode(v any) error { return f(v) }

// SetContentType initializes the response Content-Type header given a MIME
// type. If the Content-Type header is already set and the MIME type is
// "application/json" or "application/xml" then SetContentType appends a suffix
// to the header ("+json" or "+xml" respectively).
func SetContentType(w http.ResponseWriter, ct string) {
	h := w.Header().Get("Content-Type")
	if h == "" {
		w.Header().Set("Content-Type", ct)
		return
	}
	// RFC6839 only defines suffixes for a few mime types, we only concern
	// ourselves with JSON and XML.
	if ct != "application/json" && ct != "application/xml" {
		w.Header().Set("Content-Type", ct)
		return
	}
	if strings.Contains(h, "+") {
		return
	}
	suffix := "+json"
	if ct == "application/xml" {
		suffix = "+xml"
	}
	w.Header().Set("Content-Type", h+suffix)
}

func newTextEncoder(w io.Writer, ct string) Encoder {
	return &textEncoder{w, ct}
}

type textEncoder struct {
	w  io.Writer
	ct string
}

func (e *textEncoder) Encode(v any) (err error) {
	switch c := v.(type) {
	case string:
		_, err = e.w.Write([]byte(c))
	case *string: // v may be a string pointer when the Response Body is set to the field of a custom response type.
		_, err = e.w.Write([]byte(*c))
	case []byte:
		_, err = e.w.Write(c)
	default:
		err = fmt.Errorf("can't encode %T as %s", c, e.ct)
	}
	return
}

func newTextDecoder(r io.Reader, ct string) Decoder {
	return &textDecoder{r, ct}
}

type textDecoder struct {
	r  io.Reader
	ct string
}

func (e *textDecoder) Decode(v any) error {
	b, err := io.ReadAll(e.r)
	if err != nil {
		return err
	}
	switch c := v.(type) {
	case *string:
		*c = string(b)
	case *[]byte:
		*c = b
	default:
		return fmt.Errorf("can't decode %s to %T", e.ct, c)
	}
	return nil
}

// newUnsupportedDecoder returns a decoder that returns an error indicating that
// the content type is not supported.
func newUnsupportedDecoder(ct string) Decoder {
	return &unsupportedDecoder{ct}
}

type unsupportedDecoder struct {
	ct string
}

func (e *unsupportedDecoder) Decode(_ any) error {
	return goa.UnsupportedMediaTypeError(e.ct)
}
