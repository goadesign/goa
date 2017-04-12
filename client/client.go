package client

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"time"

	"context"

	"github.com/goadesign/goa"
)

type (
	// Doer defines the Do method of the http client.
	Doer interface {
		Do(context.Context, *http.Request) (*http.Response, error)
	}

	// Client is the common client data structure for all goa service clients.
	Client struct {
		// Doer is the underlying http client.
		Doer
		// Scheme overrides the default action scheme.
		Scheme string
		// Host is the service hostname.
		Host string
		// UserAgent is the user agent set in requests made by the client.
		UserAgent string
		// Dump indicates whether to dump request response.
		Dump bool
	}
)

// New creates a new API client that wraps c.
// If c is nil, the returned client wraps http.DefaultClient.
func New(c Doer) *Client {
	if c == nil {
		c = HTTPClientDoer(http.DefaultClient)
	}
	return &Client{Doer: c}
}

// HTTPClientDoer turns a stdlib http.Client into a Doer. Use it to enable to call New() with an http.Client.
func HTTPClientDoer(hc *http.Client) Doer {
	return doFunc(func(_ context.Context, req *http.Request) (*http.Response, error) {
		return hc.Do(req)
	})
}

// doFunc is the type definition of the Doer.Do method. It implements Doer.
type doFunc func(context.Context, *http.Request) (*http.Response, error)

// Do implements Doer.Do
func (f doFunc) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	return f(ctx, req)
}

// Do wraps the underlying http client Do method and adds logging.
// The logger should be in the context.
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	// TODO: setting the request ID should be done via client middleware. For now only set it if the
	// caller provided one in the ctx.
	if ctxreqid := ContextRequestID(ctx); ctxreqid != "" {
		req.Header.Set("X-Request-Id", ctxreqid)
	}
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	startedAt := time.Now()
	ctx, id := ContextWithRequestID(ctx)
	goa.LogInfo(ctx, "started", "id", id, req.Method, req.URL.String())
	if c.Dump {
		c.dumpRequest(ctx, req)
	}
	resp, err := c.Doer.Do(ctx, req)
	if err != nil {
		goa.LogError(ctx, "failed", "err", err)
		return nil, err
	}
	goa.LogInfo(ctx, "completed", "id", id, "status", resp.StatusCode, "time", time.Since(startedAt).String())
	if c.Dump {
		c.dumpResponse(ctx, resp)
	}
	return resp, err
}

// Dump request if needed.
func (c *Client) dumpRequest(ctx context.Context, req *http.Request) {
	reqBody, err := dumpReqBody(req)
	if err != nil {
		goa.LogError(ctx, "Failed to load request body for dump", "err", err.Error())
	}
	goa.LogInfo(ctx, "request headers", headersToSlice(req.Header)...)
	if reqBody != nil {
		goa.LogInfo(ctx, "request", "body", string(reqBody))
	}
}

// dumpResponse dumps the response and the request.
func (c *Client) dumpResponse(ctx context.Context, resp *http.Response) {
	respBody, _ := dumpRespBody(resp)
	goa.LogInfo(ctx, "response headers", headersToSlice(resp.Header)...)
	if respBody != nil {
		goa.LogInfo(ctx, "response", "body", string(respBody))
	}
}

// headersToSlice produces a loggable slice from a HTTP header.
func headersToSlice(header http.Header) []interface{} {
	res := make([]interface{}, 2*len(header))
	i := 0
	for k, v := range header {
		res[i] = k
		if len(v) == 1 {
			res[i+1] = v[0]
		} else {
			res[i+1] = v
		}
		i += 2
	}
	return res
}

// Dump request body, strongly inspired from httputil.DumpRequest
func dumpReqBody(req *http.Request) ([]byte, error) {
	if req.Body == nil {
		return nil, nil
	}
	var save io.ReadCloser
	var err error
	save, req.Body, err = drainBody(req.Body)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	var dest io.Writer = &b
	chunked := len(req.TransferEncoding) > 0 && req.TransferEncoding[0] == "chunked"
	if chunked {
		dest = httputil.NewChunkedWriter(dest)
	}
	_, err = io.Copy(dest, req.Body)
	if chunked {
		dest.(io.Closer).Close()
		io.WriteString(&b, "\r\n")
	}
	req.Body = save
	return b.Bytes(), err
}

// Dump response body, strongly inspired from httputil.DumpResponse
func dumpRespBody(resp *http.Response) ([]byte, error) {
	if resp.Body == nil {
		return nil, nil
	}
	var b bytes.Buffer
	savecl := resp.ContentLength
	var save io.ReadCloser
	var err error
	save, resp.Body, err = drainBody(resp.Body)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(&b, resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body = save
	resp.ContentLength = savecl
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// One of the copies, say from b to r2, could be avoided by using a more
// elaborate trick where the other copy is made during Request/Response.Write.
// This would complicate things too much, given that these functions are for
// debugging only.
func drainBody(b io.ReadCloser) (r1, r2 io.ReadCloser, err error) {
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(b); err != nil {
		return nil, nil, err
	}
	if err = b.Close(); err != nil {
		return nil, nil, err
	}
	return ioutil.NopCloser(&buf), ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
}

// headerIterator is a HTTP header iterator.
type headerIterator func(name string, value []string)

// filterHeaders iterates through the headers skipping hidden headers.
// It calls the given iterator for each header name/value pair. The values are serialized as
// strings.
func filterHeaders(headers http.Header, iterator headerIterator) {
	for k, v := range headers {
		// Skip sensitive headers
		if k == "Authorization" || k == "Cookie" {
			iterator(k, []string{"*****"})
			continue
		}
		iterator(k, v)
	}
}

// shortID produces a "unique" 6 bytes long string.
// Do not use as a reliable way to get unique IDs, instead use for things like logging.
func shortID() string {
	b := make([]byte, 6)
	io.ReadFull(rand.Reader, b)
	return base64.StdEncoding.EncodeToString(b)
}

// clientKey is the private type used to store values in the context.
// It is private to avoid possible collisions with keys used by other packages.
type clientKey int

// ReqIDKey is the context key used to store the request ID value.
const reqIDKey clientKey = 1

// ContextRequestID extracts the Request ID from the context.
func ContextRequestID(ctx context.Context) string {
	var reqID string
	id := ctx.Value(reqIDKey)
	if id != nil {
		reqID = id.(string)
	}
	return reqID
}

// ContextWithRequestID returns ctx and the request ID if it already has one or creates and returns a new context with
// a new request ID.
func ContextWithRequestID(ctx context.Context) (context.Context, string) {
	reqID := ContextRequestID(ctx)
	if reqID == "" {
		reqID = shortID()
		ctx = context.WithValue(ctx, reqIDKey, reqID)
	}
	return ctx, reqID
}

// SetContextRequestID sets a request ID in the given context and returns a new context.
func SetContextRequestID(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, reqIDKey, reqID)
}
