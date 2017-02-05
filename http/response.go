package http

import "net/http"

// ResponseWriter intercepts the calls to WriteHeader and Write and store the
// response status code and content length so that they can be logged for
// example.
type ResponseWriter struct {
	http.ResponseWriter
	StatusCode    int
	ContentLength int
}

// WrapResponseWriter creates a ResponseWriter from the given http writer.
func WrapResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		StatusCode:     http.StatusOK, // Default value
	}
}

// WriteHeader records the value of the status code before writing it.
func (w *ResponseWriter) WriteHeader(code int) {
	w.StatusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Write computes the written len and stores it in ContentLength.
func (w *ResponseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.ContentLength += n
	return n, err
}
