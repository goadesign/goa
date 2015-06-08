package goa

import (
	"encoding/json"
	"net/http"
)

// Context is the object that provides access to the underlying HTTP request and response data.
type Context struct {
	Action        interface{}            // Controller action method
	Params        map[string]interface{} // URL and query string parameters
	PayloadParams interface{}            // Payload (request body) parameters
	R             *http.Request          // Underlying HTTP request
	W             http.ResponseWriter    // Underlying HTTP response writer
}

// Validator is implemented by all data structures that can be validated.
// This includes request and response body data structures.
type Validator interface {
	Validate() error
}

// HasParam returns true if the request url defines a parameter with the given name.
func (c *Context) HasParam(name string) bool {
	_, ok := c.Params[name]
	return ok
}

// IntParam returns the bool value of parameter with given name or 0 if not found.
func (c *Context) BoolParam(name string) bool {
	if v, ok := c.Params[name]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// IntParam returns the int value of parameter with given name or 0 if not found.
func (c *Context) IntParam(name string) int {
	if v, ok := c.Params[name]; ok {
		if i, ok := v.(int64); ok {
			return int(i)
		}
	}
	return 0
}

// IntParam returns the string value of parameter with given name or 0 if not found.
func (c *Context) StringParam(name string) string {
	if v, ok := c.Params[name]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// BoolSliceParam returns the bool slice value of parameter with given name or nil if not found.
func (c *Context) BoolSliceParam(name string) []bool {
	var r []bool
	if v, ok := c.Params[name]; ok {
		if s, ok := v.([]interface{}); ok {
			r = make([]bool, len(s))
			for i, val := range s {
				r[i] = val.(bool)
			}
		}
	}
	return r
}

// IntSliceParam returns the int slice value of parameter with given name or nil if not found.
func (c *Context) IntSliceParam(name string) []int {
	var r []int
	if v, ok := c.Params[name]; ok {
		if s, ok := v.([]interface{}); ok {
			r = make([]int, len(s))
			for i, val := range s {
				r[i] = int(val.(int64))
			}
		}
	}
	return r
}

// StringSliceParam returns the string slice value of parameter with given name or nil if not found.
func (c *Context) StringSliceParam(name string) []string {
	var r []string
	if v, ok := c.Params[name]; ok {
		if s, ok := v.([]interface{}); ok {
			r = make([]string, len(s))
			for i, val := range s {
				r[i] = val.(string)
			}
		}
	}
	return r
}

// Bind loads the request body in the given Validator then calls the Validate method on the
// unmarshalled object.
func (c *Context) Bind(v Validator) error {
	decoder := json.NewDecoder(c.R.Body)
	if err := decoder.Decode(v); err != nil {
		return err
	}
	return v.Validate()
}

// Respond writes the given HTTP status code and response body.
// This method can only be called once per request.
func (c *Context) Respond(code int, body []byte) error {
	c.W.WriteHeader(code)
	if _, err := c.W.Write(body); err != nil {
		return err
	}
	return nil
}

// RespondBadRequest sends a HTTP response with status code 400 and the given body.
func (c *Context) RespondBadRequest(body string) error {
	return c.Respond(400, []byte(body))
}

// RespondBadResponse sends a HTTP response with status code 500 and the given body.
func (c *Context) RespondBadResponse(body string) error {
	return c.Respond(500, []byte(body))
}
