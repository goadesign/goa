package xray

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"strings"

	"goa.design/goa/v3/middleware/xray"
)

type (
	// HTTPSegment represents an AWS X-Ray segment document for HTTP services.
	// It wraps the AWS X-Ray segment with the http response writer.
	HTTPSegment struct {
		*xray.Segment
		http.ResponseWriter
	}
)

// RecordRequest traces a request.
//
// It sets Http.Request & Namespace (ex: "remote")
func (s *HTTPSegment) RecordRequest(req *http.Request, namespace string) {
	s.Lock()
	defer s.Unlock()

	if s.HTTP == nil {
		s.HTTP = &xray.HTTP{}
	}

	s.Namespace = namespace
	s.HTTP.Request = requestData(req)
}

// RecordResponse traces a response.
//
// It sets Throttle, Fault, Error and HTTP.Response
func (s *HTTPSegment) RecordResponse(resp *http.Response) {
	s.Lock()
	defer s.Unlock()

	if s.HTTP == nil {
		s.HTTP = &xray.HTTP{}
	}

	s.recordStatusCode(resp.StatusCode)

	s.HTTP.Response = responseData(resp)
}

// WriteHeader records the HTTP response code and calls the corresponding
// ResponseWriter method.
func (s *HTTPSegment) WriteHeader(code int) {
	s.Lock()
	defer s.Unlock()
	if s.HTTP == nil {
		s.HTTP = &xray.HTTP{}
	}
	if s.HTTP.Response == nil {
		s.HTTP.Response = &xray.Response{}
	}
	s.HTTP.Response.Status = code
	s.recordStatusCode(code)
	s.ResponseWriter.WriteHeader(code)
}

// Write records the HTTP response content length and error (if any)
// and calls the corresponding ResponseWriter method.
func (s *HTTPSegment) Write(p []byte) (int, error) {
	s.Lock()
	n, err := s.ResponseWriter.Write(p)
	s.Unlock()
	if err != nil {
		s.RecordError(err)
	}

	s.Lock()
	defer s.Unlock()
	if s.HTTP == nil {
		s.HTTP = &xray.HTTP{}
	}
	if s.HTTP.Response == nil {
		s.HTTP.Response = &xray.Response{}
	}
	s.HTTP.Response.ContentLength = int64(n)
	return n, err
}

// Hijack supports the http.Hijacker interface.
func (s *HTTPSegment) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := s.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("xray: inner ResponseWriter cannot be hijacked: %T", s.ResponseWriter)
}

// recordStatusCode sets Throttle, Fault, Error
//
// It is expected that the mutex has already been locked when calling this method.
func (s *HTTPSegment) recordStatusCode(statusCode int) {
	switch {
	case statusCode == http.StatusTooManyRequests:
		s.Throttle = true
	case statusCode >= 400 && statusCode < 500:
		s.Fault = true
	case statusCode >= 500:
		s.Error = true
	}
}

// requestData creates a Request from a http.Request.
func requestData(req *http.Request) *xray.Request {
	var (
		scheme = "http"
		host   = req.Host
	)
	if len(req.URL.Scheme) > 0 {
		scheme = req.URL.Scheme
	}
	if len(req.URL.Host) > 0 {
		host = req.URL.Host
	}

	return &xray.Request{
		Method:        req.Method,
		URL:           fmt.Sprintf("%s://%s%s", scheme, host, req.URL.Path),
		ClientIP:      getIP(req),
		UserAgent:     req.UserAgent(),
		ContentLength: req.ContentLength,
	}
}

// responseData creates a Response from a http.Response.
func responseData(resp *http.Response) *xray.Response {
	return &xray.Response{
		Status:        resp.StatusCode,
		ContentLength: resp.ContentLength,
	}
}

// getIP implements a heuristic that returns an origin IP address for a request.
func getIP(req *http.Request) string {
	for _, h := range []string{"X-Forwarded-For", "X-Real-Ip"} {
		for _, ip := range strings.Split(req.Header.Get(h), ",") {
			if len(ip) == 0 {
				continue
			}
			realIP := net.ParseIP(strings.Replace(ip, " ", "", -1))
			return realIP.String()
		}
	}

	// not found in header
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return req.RemoteAddr
	}
	return host
}
