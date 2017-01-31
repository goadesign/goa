package app

import (
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"mime"
	"net/http"

	"goa.design/goa.v2/rest"
)

// NewDecoder returns a request body decoder. The decoder handles the following
// content types:
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

// NewEncoder returns a response encoder. The encoder handles the following
// content types:
//
// * application/json using package encoding/json
// * application/xml using package encoding/xml
// * application/gob using package encoding/gob
func NewEncoder(w http.ResponseWriter, r *http.Request) (rest.Encoder, string) {
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
		return json.NewEncoder(w), accept
	case "application/gob":
		return gob.NewEncoder(w), accept
	case "application/xml":
		return xml.NewEncoder(w), accept
	default:
		return json.NewEncoder(w), "application/json"
	}
}
