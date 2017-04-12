package gzip

import (
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"context"

	"github.com/goadesign/goa"
)

// These compression constants are copied from the compress/gzip package.
const (
	encodingGzip = "gzip"

	headerAcceptEncoding  = "Accept-Encoding"
	headerContentEncoding = "Content-Encoding"
	headerContentLength   = "Content-Length"
	headerContentType     = "Content-Type"
	headerVary            = "Vary"
	headerSecWebSocketKey = "Sec-WebSocket-Key"
)

// gzipResponseWriter wraps the http.ResponseWriter to provide gzip
// capabilities.
type gzipResponseWriter struct {
	http.ResponseWriter
	gzw *gzip.Writer
}

// Write writes bytes to the gzip.Writer. It will also set the Content-Type
// header using the net/http library content type detection if the Content-Type
// header was not set yet.
func (grw gzipResponseWriter) Write(b []byte) (int, error) {
	if len(grw.Header().Get(headerContentType)) == 0 {
		grw.Header().Set(headerContentType, http.DetectContentType(b))
	}
	return grw.gzw.Write(b)
}

// handler struct contains the ServeHTTP method
type handler struct {
	pool sync.Pool
}

// Middleware encodes the response using Gzip encoding and sets all the appropriate
// headers. If the Content-Type is not set, it will be set by calling
// http.DetectContentType on the data being written.
func Middleware(level int) goa.Middleware {
	gzipPool := sync.Pool{
		New: func() interface{} {
			gz, err := gzip.NewWriterLevel(ioutil.Discard, level)
			if err != nil {
				panic(err)
			}
			return gz
		},
	}
	return func(h goa.Handler) goa.Handler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) (err error) {
			// Skip compression if the client doesn't accept gzip encoding, is
			// requesting a WebSocket or the data is already compressed.
			if !strings.Contains(req.Header.Get(headerAcceptEncoding), encodingGzip) ||
				len(req.Header.Get(headerSecWebSocketKey)) > 0 ||
				req.Header.Get(headerContentEncoding) == encodingGzip {
				return h(ctx, rw, req)
			}

			// Set the appropriate gzip headers.
			resp := goa.ContextResponse(ctx)
			resp.Header().Set(headerContentEncoding, encodingGzip)
			resp.Header().Set(headerVary, headerAcceptEncoding)

			// Retrieve gzip writer from the pool. Reset it to use the ResponseWriter.
			// This allows us to re-use an already allocated buffer rather than
			// allocating a new buffer for every request.
			gz := gzipPool.Get().(*gzip.Writer)

			// Get the original http.ResponseWriter
			w := resp.SwitchWriter(nil)
			// Reset our gzip writer to use the http.ResponseWriter
			gz.Reset(w)

			// Wrap the original http.ResponseWriter with our gzipResponseWriter
			grw := gzipResponseWriter{
				ResponseWriter: w,
				gzw:            gz,
			}

			// Set the new http.ResponseWriter
			resp.SwitchWriter(grw)

			// Call the next handler supplying the gzipResponseWriter instead of
			// the original.
			err = h(ctx, rw, req)
			if err != nil {
				return
			}

			// Delete the content length after we know we have been written to.
			grw.Header().Del(headerContentLength)
			gz.Close()
			gzipPool.Put(gz)
			return
		}
	}
}
