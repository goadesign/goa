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

	"golang.org/x/net/context"

	"github.com/goadesign/goa"
)

type (
	// Client is the command client data structure for all goa service clients.
	Client struct {
		// Client is the underlying http client.
		*http.Client
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
// If c is nil the returned client wraps the default http client.
func New(c *http.Client) *Client {
	if c == nil {
		c = http.DefaultClient
	}
	return &Client{Client: c}
}

// Do wraps the underlying http client Do method and adds logging.
// The logger should be in the context.
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", c.UserAgent)
	startedAt := time.Now()
	id := shortID()
	goa.LogInfo(ctx, "started", "id", id, req.Method, req.URL.String())
	if c.Dump {
		c.dumpRequest(ctx, req)
	}
	resp, err := c.Client.Do(req)
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
