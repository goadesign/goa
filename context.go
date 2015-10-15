package goa

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/net/context"
	log "gopkg.in/inconshreveable/log15.v2"
)

// Context is the object that provides access to the underlying HTTP request and response state.
// It implements the context.Context interface described at http://blog.golang.org/context.
type Context struct {
	context.Context // A goa context is a golang context
	log.Logger      // Context logger
}

// key is the type used to store internal values in the context.
// Context provides typed accessor methods to these values.
type key int

const (
	reqKey key = iota
	respKey
	paramKey
	queryKey
	payloadKey
	respWrittenKey
	respStatusKey
	respLenKey
)

// AddValue sets the value associated with key in the context.
// The value can be retrieved using the Value method.
// Note that this changes the underlying context.Context object and thus clients holding a reference
// to that won't be able to access the new value. It's probably a bad idea to hold a reference to
// the inner context anyway...
func (c *Context) AddValue(key, val interface{}) {
	c.Context = context.WithValue(c.Context, key, val)
}

// Request returns the underlying HTTP request.
func (c *Context) Request() *http.Request {
	return c.Value(reqKey).(*http.Request)
}

// ResponseWriter returns the raw HTTP response writer.
func (c *Context) ResponseWriter() http.ResponseWriter {
	return c.Value(respKey).(http.ResponseWriter)
}

// ResponseHeader returns the response HTTP header object.
func (c *Context) ResponseHeader() http.Header {
	if rw := c.ResponseWriter(); rw != nil {
		return rw.Header()
	}
	return nil
}

// ResponseWritten returns true if an HTTP response was written.
func (c *Context) ResponseWritten() bool {
	if wr := c.Value(respStatusKey); wr != nil {
		return true
	}
	return false
}

// ResponseStatus returns the response status if it was set via one of the context response
// methods (Respond, JSON, BadRequest, Bug), 0 otherwise.
func (c *Context) ResponseStatus() int {
	if is := c.Value(respStatusKey); is != nil {
		return is.(int)
	}
	return 0
}

// ResponseLength returns the response body length in bytes if the response was written to the
// context via one of the response methods (Respond, JSON, BadRequest, Bug), 0 otherwise.
func (c *Context) ResponseLength() int {
	if is := c.Value(respLenKey); is != nil {
		return is.(int)
	}
	return 0
}

// Get returns the param or query string with the given name and true or an empty string and false
// if there isn't one.
func (c *Context) Get(name string) (string, bool) {
	params := c.Value(paramKey).(map[string]string)
	v, ok := params[name]
	if !ok {
		var vs []string
		query := c.Value(queryKey).(map[string][]string)
		vs, ok = query[name]
		if ok {
			v = strings.Join(vs, ",")
		}
	}
	if !ok {
		return "", false
	}
	return v, true
}

// GetMany returns the query string values with the given name or nil if there aren't any.
func (c *Context) GetMany(name string) []string {
	query := c.Value(queryKey).(map[string][]string)
	return query[name]
}

// Payload returns the deserialized request body or nil if body is empty.
func (c *Context) Payload() interface{} {
	return c.Value(payloadKey)
}

// Respond writes the given HTTP status code and response body.
// This method should only be called once per request.
func (c *Context) Respond(code int, body []byte) error {
	rw := c.ResponseWriter()
	rw.WriteHeader(code)
	if _, err := rw.Write(body); err != nil {
		return err
	}
	c.Context = context.WithValue(c.Context, respWrittenKey, true)
	c.Context = context.WithValue(c.Context, respStatusKey, code)
	c.Context = context.WithValue(c.Context, respLenKey, len(body))
	return nil
}

// JSON serializes the given body into JSON and sends a HTTP response with the given status code
// and JSON as body.
func (c *Context) JSON(code int, body interface{}) error {
	js, err := json.Marshal(body)
	if err != nil {
		return err
	}
	return c.Respond(code, js)
}

// BadRequest sends a HTTP response with status code 400 and the given error as body.
func (c *Context) BadRequest(err *BadRequestError) error {
	return c.Respond(400, []byte(err.Error()))
}

// Bug sends a HTTP response with status code 500 and the given body.
// The body can be set using a format and substituted values a la fmt.Printf.
func (c *Context) Bug(format string, a ...interface{}) error {
	body := fmt.Sprintf(format, a...)
	return c.Respond(500, []byte(body))
}
