package http

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"strings"
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
		Decode(v interface{}) error
	}

	// Encoder provides the actual encoding algorithm used to write HTTP
	// request and response bodies.
	Encoder interface {
		// Encode encodes v.
		Encode(v interface{}) error
	}

	// EncodingFunc allows a function with appropriate signature to act as a
	// Decoder/Encoder.
	EncodingFunc func(v interface{}) error

	// private type used to define context keys.
	contextKey int
)

// RequestDecoder returns a HTTP request body decoder suitable for the given
// request. The decoder handles the following mime types:
//
//     * application/json using package encoding/json
//     * application/xml using package encoding/xml
//     * application/gob using package encoding/gob
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
	default:
		return json.NewDecoder(r.Body)
	}
}

// ResponseEncoder returns a HTTP response encoder leveraging the mime type
// set in the context under the AcceptTypeKey or the ContentTypeKey if any.
// The encoder supports the following mime types:
//
//     * application/json using package encoding/json
//     * application/xml using package encoding/xml
//     * application/gob using package encoding/gob
//     * text/html and text/plain for strings
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
				case ct == "application/json" || strings.HasSuffix(ct, "+json"):
					enc = json.NewEncoder(w)
				case ct == "application/xml" || strings.HasSuffix(ct, "+xml"):
					enc = xml.NewEncoder(w)
				case ct == "application/gob" || strings.HasSuffix(ct, "+gob"):
					enc = gob.NewEncoder(w)
				case ct == "text/html" || ct == "text/plain" ||
					strings.HasSuffix(ct, "+html") || strings.HasSuffix(ct, "+txt"):
					enc = newTextEncoder(w, ct)
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
	var buf bytes.Buffer
	r.Body = ioutil.NopCloser(&buf)
	return json.NewEncoder(&buf)
}

// ResponseDecoder returns a HTTP response decoder.
// The decoder handles the following content types:
//
//   * application/json using package encoding/json (default)
//   * application/xml using package encoding/xml
//   * application/gob using package encoding/gob
//   * text/html and text/plain for strings
//
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
// methods. The encoder checks whether the error is a goa ServiceError struct
// and if so uses the error temporary and timeout fields to infer a proper HTTP
// status code and marshals the error struct to the body using the provided
// encoder. If the error is not a goa ServiceError struct then it is encoded
// as a permanent internal server error.
func ErrorEncoder(encoder func(context.Context, http.ResponseWriter) Encoder) func(context.Context, http.ResponseWriter, error) error {
	return func(ctx context.Context, w http.ResponseWriter, err error) error {
		enc := encoder(ctx, w)
		resp := NewErrorResponse(err)
		w.WriteHeader(resp.StatusCode())
		return enc.Encode(resp)
	}
}

// Decode implements the Decoder interface. It simply calls f(v).
func (f EncodingFunc) Decode(v interface{}) error { return f(v) }

// Encode implements the Encoder interface. It simply calls f(v).
func (f EncodingFunc) Encode(v interface{}) error { return f(v) }

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

func (e *textEncoder) Encode(v interface{}) error {
	var err error

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

	return err
}

func newTextDecoder(r io.Reader, ct string) Decoder {
	return &textDecoder{r, ct}
}

type textDecoder struct {
	r  io.Reader
	ct string
}

func (e *textDecoder) Decode(v interface{}) error {
	b, err := ioutil.ReadAll(e.r)
	if err != nil {
		return err
	}

	switch c := v.(type) {
	case *string:
		*c = string(b)
	case *[]byte:
		*c = []byte(string(b))
	default:
		err = fmt.Errorf("can't decode %s to %T", e.ct, c)
	}

	return err
}
