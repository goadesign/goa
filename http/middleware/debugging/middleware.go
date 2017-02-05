package debug

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dimfeld/httptreemux"
	goa "goa.design/goa.v2"
	"goa.design/goa.v2/rest/middleware/tracing"
)

// responseDupper tees the response to a buffer and a response writer.
type responseDupper struct {
	http.ResponseWriter
	Buffer *bytes.Buffer
	Status int
}

// New returns a debug middleware which prints all the details about incoming
// requests and outgoing responses.
func New(logger goa.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := tracing.ContextTraceID(r.Context())
			if requestID == "" {
				requestID = shortID()
			}

			if len(r.Header) > 0 {
				entries := make([]interface{}, 4+2*len(r.Header))
				entries[0] = "id"
				entries[1] = requestID
				entries[2] = "headers"
				entries[3] = len(r.Header)
				i := 0
				for k, v := range r.Header {
					entries[i+4] = k
					entries[i+5] = interface{}(strings.Join(v, ", "))
					i = i + 2
				}
				logger.Info(r.Context(), entries...)
			}
			params := httptreemux.ContextParams(r.Context())
			if len(params) > 0 {
				entries := make([]interface{}, 4+2*len(params))
				entries[0] = "id"
				entries[1] = requestID
				entries[2] = "params"
				entries[3] = len(params)
				i := 0
				for k, v := range params {
					entries[i] = k
					entries[i+1] = v
					i = i + 2
				}
				logger.Info(r.Context(), entries...)
			}
			buf, err := ioutil.ReadAll(r.Body)
			if err != nil {
				buf = []byte("failed to read body: " + err.Error())
			}
			r.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
			if len(buf) == 0 {
				buf = []byte("<empty>")
			}
			logger.Info(r.Context(), "id", requestID, "payload", string(buf))

			dupper := &responseDupper{ResponseWriter: w, Buffer: &bytes.Buffer{}}
			h.ServeHTTP(dupper, r)

			if len(dupper.Header()) > 0 {
				entries := make([]interface{}, 4+2*len(dupper.Header()))
				entries[0] = "id"
				entries[1] = requestID
				entries[2] = "response headers"
				entries[3] = len(dupper.Header())
				i := 0
				for k, v := range dupper.Header() {
					entries[i+4] = k
					entries[i+5] = interface{}(strings.Join(v, ", "))
					i = i + 2
				}
				logger.Info(r.Context(), entries...)
			}
			if dupper.Buffer.Len() > 0 {
				logger.Info(r.Context(), "id", requestID, "response body", dupper.Buffer.String())
			}
		})
	}
}

// Write writes the data to the buffer and connection as part of an HTTP reply.
func (r *responseDupper) Write(b []byte) (int, error) {
	r.Buffer.Write(b)
	return r.ResponseWriter.Write(b)
}

// WriteHeader records the status and sends an HTTP response header with status code.
func (r *responseDupper) WriteHeader(s int) {
	r.Status = s
	r.ResponseWriter.WriteHeader(s)
}

// shortID produces a " unique" 6 bytes long string.
// Do not use as a reliable way to get unique IDs, instead use for things like logging.
func shortID() string {
	b := make([]byte, 6)
	io.ReadFull(rand.Reader, b)
	return base64.RawURLEncoding.EncodeToString(b)
}
