package http

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"io"
	"mime"
	"net/http"

	goa "goa.design/goa.v2"
	"goa.design/goa.v2/rest"
)

type (
	// Service Encoding/Decoding

	// DecodeRequestFunc decodes the HTTP request into a request payload.
	DecodeRequestFunc func(*http.Request) (interface{}, error)

	// EncodeResponseFunc encodes the request result into a HTTP response.
	// The encoding may perform content-type negotiation using the given
	// request.
	EncodeResponseFunc func(http.ResponseWriter, *http.Request, interface{}) error

	// EncodeErrorFunc encodes an error into a HTTP response.
	// The encoding may perform content-type negotiation using the given
	// request.
	EncodeErrorFunc func(http.ResponseWriter, *http.Request, error)

	// Client Encoding/Decoding

	// EncodeRequestFunc encodes a request payload into a HTTP request.
	EncodeRequestFunc func(interface{}) (*http.Request, error)

	// DecodeResponseFunc decodes a request result from a HTTP response.
	DecodeResponseFunc func(*http.Response) (interface{}, error)
)

// NewDecoder returns a HTTP request body decoder.
// The decoder handles the following content types:
//
// * application/json using package encoding/json
// * application/xml using package encoding/xml
// * application/gob using package encoding/gob
func NewDecoder(r *http.Request) rest.Decoder {
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

// NewEncoder returns a HTTP response encoder.
// The encoder handles the following content types:
//
// * application/json using package encoding/json
// * application/xml using package encoding/xml
// * application/gob using package encoding/gob
func NewEncoder(w http.ResponseWriter, r *http.Request) rest.Encoder {
	accept := r.Header.Get("Accept")
	if accept == "" {
		// Default to JSON
		accept = "application/json"
	} else {
		if mediaType, _, err := mime.ParseMediaType(accept); err == nil {
			accept = mediaType
		}
	}
	switch accept {
	case "application/json":
		return json.NewEncoder(w)
	case "application/gob":
		return gob.NewEncoder(w)
	case "application/xml":
		return xml.NewEncoder(w)
	default:
		return json.NewEncoder(w)
	}
}

// ResponseContentType returns the value of the Content-Type header for the
// given request.
func ResponseContentType(r *http.Request) string {
	accept := r.Header.Get("Accept")
	if accept == "" {
		// Default to JSON
		return "application/json"
	}
	if mediaType, _, err := mime.ParseMediaType(accept); err == nil {
		accept = mediaType
	}
	switch accept {
	case "application/json",
		"application/gob",
		"application/xml":
		return accept
	default:
		return "application/json"
	}
}

// EncodeError returns an encoder that checks whether the error is a goa
// Error and if so sets the response status code using the error status and
// encodes the corresponding ErrorResponse struct to the response body. If the
// error is not a goa.Error then it sets the response status code to 500, writes
// the error message to the response body and logs it.
func EncodeError(enc rest.ResponseEncoderFunc, logger goa.Logger) func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, r *http.Request, v error) {
		switch t := v.(type) {
		case goa.Error:
			w.Header().Set("Content-Type", ResponseContentType(r))
			w.WriteHeader(rest.HTTPStatus(t.Status()))
			err := enc(w, r).Encode(rest.NewErrorResponse(t))
			if err != nil {
				logger.Error(r.Context(), "encoding", err)
			}
		default:
			b := make([]byte, 6)
			io.ReadFull(rand.Reader, b)
			id := base64.RawURLEncoding.EncodeToString(b) + ": "
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(id + t.Error()))
			logger.Error(r.Context(), "id", id, "error", t.Error())
		}
	}
}
