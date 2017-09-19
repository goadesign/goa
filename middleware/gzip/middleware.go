package gzip

import (
	"bytes"
	"compress/gzip"
	"context"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"

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
	headerRange           = "Range"
	headerAcceptRanges    = "Accept-Ranges"
	headerSecWebSocketKey = "Sec-WebSocket-Key"
)

// gzipResponseWriter wraps the http.ResponseWriter to provide gzip
// capabilities.
type gzipResponseWriter struct {
	http.ResponseWriter
	gzw            *gzip.Writer
	buf            bytes.Buffer
	pool           *sync.Pool
	statusCode     int
	shouldCompress *bool
	o              options
}

// Write writes bytes to the gzip.Writer. It will also set the Content-Type
// header using the net/http library content type detection if the Content-Type
// header was not set yet.
func (grw *gzipResponseWriter) Write(b []byte) (int, error) {
	if len(grw.Header().Get(headerContentType)) == 0 {
		grw.Header().Set(headerContentType, http.DetectContentType(b))
	}

	// If we already decided to gzip, do that.
	if grw.gzw != nil {
		return grw.gzw.Write(b)
	}

	// If we have already decided not to gzip, do that.
	if grw.shouldCompress != nil && !*grw.shouldCompress {
		return grw.ResponseWriter.Write(b)
	}

	// Detect types, check status code.
	if grw.shouldCompress == nil {
		s := grw.o.shouldCompress(grw.Header().Get(headerContentType), grw.statusCode)
		grw.shouldCompress = &s
		if !s {
			grw.ResponseWriter.WriteHeader(grw.statusCode)
			return grw.ResponseWriter.Write(b)
		}
	}

	// Check if length is above minimum,
	// if not save to buffer.
	size := len(b) + grw.buf.Len()
	if size < grw.o.minSize {
		return grw.buf.Write(b)
	}

	// Reset our gzip writer to use the http.ResponseWriter
	// Retrieve gzip writer from the pool. Reset it to use the ResponseWriter.
	// This allows us to re-use an already allocated buffer rather than
	// allocating a new buffer for every request.
	gz := grw.pool.Get().(*gzip.Writer)

	// We must write header now
	grw.Header().Set(headerContentEncoding, encodingGzip)
	grw.Header().Set(headerVary, headerAcceptEncoding)
	grw.Header().Del(headerContentLength)
	grw.Header().Del(headerAcceptRanges)
	grw.ResponseWriter.WriteHeader(grw.statusCode)
	gz.Reset(grw.ResponseWriter)
	grw.gzw = gz

	// Write buffer
	if grw.buf.Len() > 0 {
		_, err := gz.Write(grw.buf.Bytes())
		if err != nil {
			return 0, err
		}
		grw.buf.Reset()
	}
	return gz.Write(b)
}

func (grw *gzipResponseWriter) WriteHeader(n int) {
	grw.statusCode = n
}

type (
	// Option allows to override default parameters.
	Option func(*options) error

	// options contains final options
	options struct {
		ignoreRange  bool
		minSize      int
		contentTypes []string
		statusCodes  map[int]struct{}
	}
)

// defaultContentTypes is the default list of content types for which
// a Handler considers gzip compression. This list originates from the
// file compression.conf within the Apache configuration found at
// https://html5boilerplate.com/
var defaultContentTypes = []string{
	"application/atom+xml",
	"application/font-sfnt",
	"application/javascript",
	"application/json",
	"application/ld+json",
	"application/manifest+json",
	"application/rdf+xml",
	"application/rss+xml",
	"application/schema+json",
	"application/vnd.", // All custom vendor types
	"application/x-font-ttf",
	"application/x-javascript",
	"application/x-web-app-manifest+json",
	"application/xhtml+xml",
	"application/xml",
	"font/eot",
	"font/opentype",
	"image/bmp",
	"image/svg+xml",
	"image/vnd.microsoft.icon",
	"image/x-icon",
	"text/cache-manifest",
	"text/css",
	"text/html",
	"text/javascript",
	"text/plain",
	"text/vcard",
	"text/vnd.rim.location.xloc",
	"text/vtt",
	"text/x-component",
	"text/x-cross-domain-policy",
	"text/xml",
}

// defaultStatusCodes are the status codes that will be compressed.
var defaultStatusCodes = []int{
	http.StatusOK,
	http.StatusCreated,
	http.StatusAccepted,
}

// AddContentTypes allows to specify specific content types to encode.
// Adds to previous content types.
func AddContentTypes(types ...string) Option {
	return func(c *options) error {
		dst := make([]string, len(c.contentTypes)+len(types))
		copy(dst, c.contentTypes)
		copy(dst[len(c.contentTypes):], types)
		c.contentTypes = dst
		return nil
	}
}

// OnlyContentTypes allows to specify specific content types to encode.
// Overrides previous content types.
// no types = ignore content types (always compress).
func OnlyContentTypes(types ...string) Option {
	return func(c *options) error {
		if len(types) == 0 {
			c.contentTypes = nil
			return nil
		}
		c.contentTypes = types
		return nil
	}
}

// AddStatusCodes allows to specify specific content types to encode.
// All content types that has the supplied prefixes are compressed.
func AddStatusCodes(codes ...int) Option {
	return func(c *options) error {
		dst := make(map[int]struct{}, len(c.statusCodes)+len(codes))
		for code := range c.statusCodes {
			dst[code] = struct{}{}
		}
		for _, code := range codes {
			c.statusCodes[code] = struct{}{}
		}
		return nil
	}
}

// OnlyStatusCodes allows to specify specific content types to encode.
// All content types that has the supplied prefixes are compressed.
// No codes = ignore content types (always compress).
func OnlyStatusCodes(codes ...int) Option {
	return func(c *options) error {
		if len(codes) == 0 {
			c.statusCodes = nil
			return nil
		}
		c.statusCodes = make(map[int]struct{}, len(codes))
		for _, code := range codes {
			c.statusCodes[code] = struct{}{}
		}
		return nil
	}
}

// MinSize will set a minimum size for compression.
func MinSize(n int) Option {
	return func(c *options) error {
		if n <= 0 {
			c.minSize = 0
			return nil
		}
		c.minSize = n
		return nil
	}
}

// IgnoreRange will set make the compressor ignore Range requests.
// Range requests are incompatible with compressed content,
// so if this is set to true "Range" headers will be ignored.
// If set to false, compression is disabled for all requests with Range header.
func IgnoreRange(b bool) Option {
	return func(c *options) error {
		c.ignoreRange = b
		return nil
	}
}

// Middleware encodes the response using Gzip encoding and sets all the
// appropriate headers. If the Content-Type is not set, it will be set by
// calling http.DetectContentType on the data being written.
func Middleware(level int, o ...Option) goa.Middleware {
	opts := options{
		ignoreRange:  true,
		minSize:      256,
		contentTypes: defaultContentTypes,
	}
	opts.statusCodes = make(map[int]struct{}, len(defaultStatusCodes))
	for _, v := range defaultStatusCodes {
		opts.statusCodes[v] = struct{}{}
	}
	for _, opt := range o {
		err := opt(&opts)
		if err != nil {
			panic(err)
		}
	}
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
				rw.Header().Get(headerContentEncoding) == encodingGzip ||
				(!opts.ignoreRange && req.Header.Get(headerRange) != "") {
				return h(ctx, rw, req)
			}

			// Set the appropriate gzip headers.
			resp := goa.ContextResponse(ctx)

			// Get the original http.ResponseWriter
			w := resp.SwitchWriter(nil)

			// Wrap the original http.ResponseWriter with our gzipResponseWriter
			grw := &gzipResponseWriter{
				ResponseWriter: w,
				pool:           &gzipPool,
				statusCode:     http.StatusOK,
				o:              opts,
			}

			// Set the new http.ResponseWriter
			resp.SwitchWriter(grw)

			// We cannot do ranges, if possibly gzipped responses.
			req.Header.Del("Range")

			// Call the next handler supplying the gzipResponseWriter instead of
			// the original.
			err = h(ctx, rw, req)
			if err != nil {
				return
			}

			// Check for uncompressed data
			if grw.buf.Len() > 0 {
				w.Header().Set(headerContentLength, strconv.Itoa(grw.buf.Len()))
				w.WriteHeader(grw.statusCode)
				_, err = w.Write(grw.buf.Bytes())
				return
			}

			// Flush compressor.
			if grw.gzw != nil {
				if err = grw.gzw.Close(); err != nil {
					return
				}
				gzipPool.Put(grw.gzw)
				return
			}
			// No writes, set status code.
			if grw.shouldCompress == nil {
				w.WriteHeader(grw.statusCode)
			}
			return
		}
	}
}

// returns true if we've been configured to compress the specific content type.
func (o options) shouldCompress(contentType string, statusCode int) bool {
	// If contentTypes is nil we handle all content types.
	if len(o.contentTypes) > 0 {
		ct := strings.ToLower(contentType)
		ct = strings.Split(ct, ";")[0]
		found := false
		for _, v := range o.contentTypes {
			if strings.HasPrefix(ct, v) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if len(o.statusCodes) > 0 {
		_, ok := o.statusCodes[statusCode]
		if !ok {
			return false
		}
	}

	return true
}
