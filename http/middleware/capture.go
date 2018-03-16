package middleware

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
)

// ResponseCapture is a http.ResponseWriter which captures the response status
// code and content length.
type ResponseCapture struct {
	http.ResponseWriter
	StatusCode    int
	ContentLength int
}

// CaptureResponse creates a ResponseCapture that wraps the given ResponseWriter.
func CaptureResponse(w http.ResponseWriter) *ResponseCapture {
	return &ResponseCapture{ResponseWriter: w}
}

// WriteHeader records the value of the status code before writing it.
func (w *ResponseCapture) WriteHeader(code int) {
	w.StatusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Write computes the written len and stores it in ContentLength.
func (w *ResponseCapture) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.ContentLength += n
	return n, err
}

// Hijack supports the http.Hijacker interface.
func (w *ResponseCapture) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := w.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, fmt.Errorf("response writer does not support hijacking: %T", w.ResponseWriter)
}
