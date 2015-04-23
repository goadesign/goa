package design

import "regexp"

// An action response definition
// Specifies a status, a media type and header definitions
// Header definitions consist of exact values or validating regular expressions
type Response struct {
	Status         int            // Response status code
	MediaType      *MediaType     // Response media type if any
	Headers        Headers        // Response headers
	HeaderPatterns HeaderPatterns // Response headers
}

// Response header values
type Headers map[string]string

// Response header patterns
type HeaderPatterns map[string]*regexp.Regexp

// WithStatus sets the response status
// It returns the response so it can be chained with other WithXXX methods.
func (r *Response) WithStatus(status int) *Response {
	r.Status = status
	return r
}

// WithMediaType sets the response MediaType field.
// It returns the response so it can be chained with other WithXXX methods.
func (r *Response) WithMediaType(m *MediaType) *Response {
	r.MediaType = m
	return r
}

// WithLocation sets the response Location field.
// It returns the response so it can be chained with other WithXXX methods.
func (r *Response) WithLocation(l *regexp.Regexp) *Response {
	return r.WithHeaderPattern("Location", l)
}

// WithHeaderPattern initializes the response Header field if not initialized
// yet and sets the given header with the given value pattern.
// It returns the response so it can be chained with other WithXXX methods.
func (r *Response) WithHeaderPattern(name string, value *regexp.Regexp) *Response {
	if r.HeaderPatterns == nil {
		r.HeaderPatterns = make(HeaderPatterns)
	}
	r.HeaderPatterns[name] = value
	return r
}

// WithHeader initializes the response Header field if not initialized yet and
// sets the given header with the given value.
// It returns the response so it can be chained with other WithXXX methods.
func (r *Response) WithHeader(name string, value string) *Response {
	if r.Headers == nil {
		r.Headers = make(Headers)
	}
	r.Headers[name] = value
	return r
}
