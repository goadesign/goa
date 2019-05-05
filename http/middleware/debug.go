package middleware

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"

	goahttp "goa.design/goa/v3/http"
	"goa.design/goa/v3/middleware"
)

// responseDupper tees the response to a buffer and a response writer.
type responseDupper struct {
	http.ResponseWriter
	Buffer *bytes.Buffer
	Status int
}

// Debug returns a debug middleware which prints detailed information about
// incoming requests and outgoing responses including all headers, parameters
// and bodies.
func Debug(mux goahttp.Muxer, w io.Writer) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			buf := &bytes.Buffer{}
			// Request ID
			reqID := r.Context().Value(middleware.RequestIDKey)
			if reqID == nil {
				reqID = shortID()
			}

			// Request URL
			buf.WriteString(fmt.Sprintf("> [%s] %s %s", reqID, r.Method, r.URL.String()))

			// Request Headers
			keys := make([]string, len(r.Header))
			i := 0
			for k := range r.Header {
				keys[i] = k
				i++
			}
			sort.Strings(keys)
			for _, k := range keys {
				buf.WriteString(fmt.Sprintf("\n> [%s] %s: %s", reqID, k, strings.Join(r.Header[k], ", ")))
			}

			// Request parameters
			params := mux.Vars(r)
			keys = make([]string, len(params))
			i = 0
			for k := range params {
				keys[i] = k
				i++
			}
			sort.Strings(keys)
			for _, k := range keys {
				buf.WriteString(fmt.Sprintf("\n> [%s] %s: %s", reqID, k, strings.Join(r.Header[k], ", ")))
			}

			// Request body
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				b = []byte("failed to read body: " + err.Error())
			}
			if len(b) > 0 {
				buf.WriteByte('\n')
				lines := strings.Split(string(b), "\n")
				for _, line := range lines {
					buf.WriteString(fmt.Sprintf("[%s] %s\n", reqID, line))
				}
			}
			r.Body = ioutil.NopCloser(bytes.NewBuffer(b))

			dupper := &responseDupper{ResponseWriter: rw, Buffer: &bytes.Buffer{}}
			h.ServeHTTP(dupper, r)

			buf.WriteString(fmt.Sprintf("\n< [%s] %s", reqID, http.StatusText(dupper.Status)))
			keys = make([]string, len(dupper.Header()))
			i = 0
			for k := range dupper.Header() {
				keys[i] = k
				i++
			}
			sort.Strings(keys)
			for _, k := range keys {
				buf.WriteString(fmt.Sprintf("\n< [%s] %s: %s", reqID, k, strings.Join(dupper.Header()[k], ", ")))
			}
			if dupper.Buffer.Len() > 0 {
				buf.WriteByte('\n')
				lines := strings.Split(dupper.Buffer.String(), "\n")
				for _, line := range lines {
					buf.WriteString(fmt.Sprintf("[%s] %s\n", reqID, line))
				}
			}
			buf.WriteByte('\n')
			w.Write(buf.Bytes())
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
