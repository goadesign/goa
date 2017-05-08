package rest

import (
	"bytes"
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

	goa "goa.design/goa.v2"
)

type (
	// Decoder is the low level decoder interface.
	Decoder interface {
		// Decode decodes into v.
		Decode(v interface{}) error
	}

	// Encoder is the low level encoder interface.
	Encoder interface {
		// Encode encodes v.
		Encode(v interface{}) error
	}
)

// DefaultRequestDecoder returns a HTTP request body decoder.
// The decoder handles the following content types:
//
// * application/json using package encoding/json
// * application/xml using package encoding/xml
// * application/gob using package encoding/gob
func DefaultRequestDecoder(r *http.Request) Decoder {
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		// Default to JSON
		contentType = "application/json"
	} else {
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

// DefaultResponseEncoder returns a HTTP response encoder and the corresponding
// mime type according to the request "Accept" header. The encoder supports the
// following mime types:
//
//   - application/json using package encoding/json
//   - application/xml using package encoding/xml
//   - application/gob using package encoding/gob
func DefaultResponseEncoder(w http.ResponseWriter, r *http.Request) (enc Encoder, mt string) {
	accept := r.Header.Get("Accept")
	builtin := func(a string) (Encoder, string) {
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
	if enc, mt = builtin(accept); enc == nil {
		// attempt to normalize
		if mt, _, err := mime.ParseMediaType(accept); err == nil {
			enc, mt = builtin(mt)
		}
	}
	if enc == nil {
		enc, mt = builtin("")
	}
	return
}

// DefaultRequestEncoder returns a HTTP request encoder.
// The encoder uses package encoding/json.
func DefaultRequestEncoder(r *http.Request) Encoder {
	var buf bytes.Buffer
	r.Body = ioutil.NopCloser(&buf)
	return json.NewEncoder(&buf)
}

// DefaultResponseDecoder returns a HTTP response decoder.
// The decoder handles the following content types:
//
// * application/json using package encoding/json (default)
// * application/xml using package encoding/xml
// * application/gob using package encoding/gob
func DefaultResponseDecoder(resp *http.Response) Decoder {
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

// EncodeError returns an encoder that checks whether the error is a goa
// Error and if so sets the response status code using the error status and
// encodes the corresponding ErrorResponse struct to the response body. If the
// error is not a goa.Error then it sets the response status code to 500, writes
// the error message to the response body and logs it.
func EncodeError(encoder func(http.ResponseWriter, *http.Request) (Encoder, string), logger goa.LogAdapter) func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, r *http.Request, v error) {
		switch t := v.(type) {

		case goa.Error:
			enc, ct := encoder(w, r)
			SetContentType(w, ct)
			w.WriteHeader(HTTPStatus(t.Status()))
			err := enc.Encode(NewErrorResponse(t))
			if err != nil {
				logger.Error(r.Context(), "encoding", err)
			}

		default:
			b := make([]byte, 6)
			io.ReadFull(rand.Reader, b)
			id := base64.RawURLEncoding.EncodeToString(b)
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(id + ": " + t.Error()))
			logger.Error(r.Context(), "id", id, "error", t.Error())
		}
	}
}

// SetContentType initializes the response Content-Type header given a mime type
// and optionally an already existing value.
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
