package http

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"strings"

	goa "goa.design/goa"
)

const (
	// AcceptTypeKey is the context key used to store the value of the HTTP
	// request Accept-Type header. The value may be used by encoders and
	// decoders to implement a content type negotiation algorithm.
	AcceptTypeKey contextKey = iota + 1
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

// ResponseEncoder returns a HTTP response encoder and the corresponding mime
// type suitable for the given request. The encoder supports the following mime
// types:
//
//     * application/json using package encoding/json
//     * application/xml using package encoding/xml
//     * application/gob using package encoding/gob
//
// ResponseEncoder defaults to the JSON encoder if the request "Accept" header
// does not match any of the supported mime types or is missing altogether.
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
		}
		return nil, ""
	}
	var accept string
	if a := ctx.Value(AcceptTypeKey); a != nil {
		accept = a.(string)
	}
	var (
		enc Encoder
		mt  string
	)
	if enc, mt = negotiate(accept); enc == nil {
		// attempt to normalize
		if mt, _, err := mime.ParseMediaType(accept); err == nil {
			enc, mt = negotiate(mt)
		}
	}
	if enc == nil {
		enc, mt = negotiate("")
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
	default:
		return json.NewDecoder(resp.Body)
	}
}

// ErrorEncoder returns an encoder that checks whether the error is a goa Error
// and if so sets the response status code using the error status and encodes
// the corresponding ErrorResponse struct to the response body. If the error is
// not a goa.Error then it sets the response status code to InternalServerError
// (500) and writes the error message to the response body.
func ErrorEncoder(encoder func(context.Context, http.ResponseWriter) Encoder) func(context.Context, http.ResponseWriter, error) {
	return func(ctx context.Context, w http.ResponseWriter, v error) {
		switch t := v.(type) {

		case goa.Error:
			enc := encoder(ctx, w)
			w.WriteHeader(Status(t.Status()))
			enc.Encode(NewErrorResponse(t))

		default:
			b := make([]byte, 6)
			io.ReadFull(rand.Reader, b)
			id := base64.RawURLEncoding.EncodeToString(b)
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(id + ": " + t.Error()))
		}
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
