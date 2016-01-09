package goa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/inconshreveable/log15.v2"
)

type (
	// Client is the command client data structure for all goa service clients.
	Client struct {
		// Logger is the logger used to log client requests.
		log15.Logger
		// Client is the underlying http client.
		*http.Client
		// Signers contains the ordered list of request signers. A signer may add headers,
		// cookies etc. to a request generally to perform auth.
		Signers []Signer
		// Scheme is the HTTP scheme used to make requests to the API host.
		Scheme string
		// Host is the service hostname.
		Host string
		// UserAgent is the user agent set in requests made by the client.
		UserAgent string
		// Dump indicates whether to dump request response.
		Dump bool
	}

	// Signer is the common interface implemented by all signers.
	Signer interface {
		// Sign adds required headers, cookies etc.
		Sign(*http.Request) error
		// RegisterFlags registers the command line flags that defines the values used to
		// initialize the signer.
		RegisterFlags(app *kingpin.Application)
	}

	// BasicSigner implements basic auth.
	BasicSigner struct {
		// Username is the basic auth user.
		Username string
		// Password is err guess what? the basic auth password.
		Password string
	}

	// JWTSigner implements JSON Web Token auth.
	JWTSigner struct {
		// Header is the name of the HTTP header which contains the JWT.
		// The default is "Authentication"
		Header string
		// Format represents the format used to render the JWT.
		// The default is "Bearer %s"
		Format string

		// token stores the actual JWT.
		token string
	}

	// OAuth2Signer enables the use of OAuth2 refresh tokens. It takes care of creating access
	// tokens given a refresh token and a refresh URL as defined in RFC 6749.
	// Note that this signer does not concern itself with generating the initial refresh token,
	// this has to be done prior to using the client.
	// Also it assumes the response of the refresh request response is JSON encoded and of the
	// form:
	// 	{
	//		"access_token":"2YotnFZFEjr1zCsicMWpAA",
	// 		"expires_in":3600,
	// 		"refresh_token":"tGzv3JOkF0XG5Qx2TlKWIA"
	// 	}
	// where the "expires_in" and "refresh_token" properties are optional and additional
	// properties are ignored. If the response contains a "expires_in" property then the signer
	// takes care of making refresh requests prior to the token expiration.
	OAuth2Signer struct {
		// RefreshURLFormat is a format that generates the refresh access token URL given a
		// refresh token.
		RefreshURLFormat string
		// RefreshToken contains the OAuth2 refresh token from which access tokens are
		// created.
		RefreshToken string

		// accessToken is the temporary access token.
		accessToken string
		// expiresAt specifies when to create a new access token.
		expiresAt time.Time
	}
)

var (
	_ Signer = &BasicSigner{}
	_ Signer = &JWTSigner{}
	_ Signer = &OAuth2Signer{}
)

// NewClient create a new API client.
func NewClient() *Client {
	logger := log15.New()
	logger.SetHandler(log15.StdoutHandler)
	return &Client{Logger: logger, Client: http.DefaultClient}
}

// Do wraps the underlying http client Do method and adds logging.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", c.UserAgent)
	startedAt := time.Now()
	id := shortID()
	if c.Dump {
		c.dumpRequest(req)
	} else {
		c.Info("started", "id", id, req.Method, req.URL.String())
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if c.Dump {
		c.dumpResponse(resp)
	} else {
		c.Info("completed", "id", id, "status", resp.StatusCode, "time", time.Since(startedAt).String())
	}
	return resp, err
}

// Sign adds the basic auth header to the request.
func (s *BasicSigner) Sign(req *http.Request) error {
	if s.Username != "" && s.Password != "" {
		req.SetBasicAuth(s.Username, s.Password)
	}
	return nil
}

// RegisterFlags adds the "--user" and "--pass" flags to the client tool.
func (s *BasicSigner) RegisterFlags(app *kingpin.Application) {
	app.Flag("user", "Basic Auth username").StringVar(&s.Username)
	app.Flag("pass", "Basic Auth password").StringVar(&s.Password)
}

// Sign adds the JWT auth header.
func (s *JWTSigner) Sign(req *http.Request) error {
	header := s.Header
	if header == "" {
		header = "Authorization"
	}
	format := s.Format
	if format == "" {
		format = "Bearer %s"
	}
	req.Header.Set(header, fmt.Sprintf(format, s.token))
	return nil
}

// RegisterFlags adds the "--jwt" flag to the client tool.
func (s *JWTSigner) RegisterFlags(app *kingpin.Application) {
	app.Flag("jwt", "JSON web token").StringVar(&s.token)
}

// Sign refreshes the access token if needed and adds the OAuth header.
func (s *OAuth2Signer) Sign(req *http.Request) error {
	if s.expiresAt.Before(time.Now()) {
		if err := s.Refresh(); err != nil {
			return fmt.Errorf("failed to refresh OAuth token: %s", err)
		}
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.accessToken))
	return nil
}

// RegisterFlags adds the "--refreshURL" and "--refreshToken" flags to the client tool.
func (s *OAuth2Signer) RegisterFlags(app *kingpin.Application) {
	app.Flag("refreshURL", "OAuth2 refresh URL format, e.g. https://somewhere.com/token?grant_type=authorization_code&code=%s&client_id=xxx").
		StringVar(&s.RefreshURLFormat)
	app.Flag("refreshToken", "OAuth2 refresh token or authorization code").
		StringVar(&s.RefreshToken)
}

// ouath2RefreshResponse is the data structure representing the interesting subset of a OAuth2
// refresh response.
type oauth2RefreshResponse struct {
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	AccessToken  string `json:"access_token"`
}

// Refresh makes a OAuth2 refresh access token request.
func (s *OAuth2Signer) Refresh() error {
	url := fmt.Sprintf(s.RefreshURLFormat, s.RefreshToken)
	resp, err := http.Post(url, "", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var r oauth2RefreshResponse
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %s", err)
	}
	err = json.Unmarshal(respBody, &r)
	if err != nil {
		return fmt.Errorf("failed to decode refresh request response: %s", err)
	}
	s.accessToken = r.AccessToken
	if r.ExpiresIn > 0 {
		s.expiresAt = time.Now().Add(time.Duration(r.ExpiresIn) * time.Second)
	}
	if r.RefreshToken != "" {
		s.RefreshToken = r.RefreshToken
	}
	return nil
}

// dumpRequest dumps the request.
func (c *Client) dumpRequest(req *http.Request) {
	reqBody, err := dumpReqBody(req)
	if err != nil {
		c.Error("Failed to load request body for dump", "err", err.Error())
	}
	var buffer bytes.Buffer
	buffer.WriteString(req.Method + " " + req.URL.String() + "\n")
	writeHeaders(&buffer, req.Header)
	if reqBody != nil {
		buffer.WriteString("\n")
		buffer.Write(reqBody)
		buffer.WriteString("\n")
	}
	fmt.Fprint(os.Stderr, buffer.String())
}

// dumpResponse dumps the response.
func (c *Client) dumpResponse(resp *http.Response) {
	respBody, err := dumpRespBody(resp)
	if err != nil {
		c.Error("Failed to load response body for dump", "err", err.Error())
	}
	var buffer bytes.Buffer
	buffer.WriteString("==> " + resp.Proto + " " + resp.Status + "\n")
	writeHeaders(&buffer, resp.Header)
	if respBody != nil {
		buffer.WriteString("\n")
		buffer.Write(respBody)
		buffer.WriteString("\n")
	}
	fmt.Fprint(os.Stderr, buffer.String())
}

// writeHeaders is a helper function that writes the given HTTP headers to the given buffer as
// human readable strings. writeHeaders filters out headers that are sensitive.
func writeHeaders(w io.Writer, headers http.Header) {
	filterHeaders(headers, func(name string, value []string) {
		fmt.Fprintf(w, "%s: %s\n", name, strings.Join(value, ", "))
	})
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
var sensitiveHeaders = map[string]bool{
	"Authorization": true,
	"Cookie":        true,
}

func filterHeaders(headers http.Header, iterator headerIterator) {
	for k, v := range headers {
		// Skip sensitive headers
		if sensitiveHeaders[k] {
			continue
		}
		iterator(k, v)
	}
}
