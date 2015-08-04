package goa

import (
	"encoding/json"
	"net/http"

	log "gopkg.in/inconshreveable/log15.v2"
)

// Context is the object that provides access to the underlying HTTP request and response data.
type Context struct {
	log.Logger                     // Context logger
	Params     map[string]string   // URL string parameters
	Query      map[string][]string // Query string parameters
	Payload    interface{}         // Payload (request body) parameters
	R          *http.Request       // Underlying HTTP request
	W          http.ResponseWriter // Underlying HTTP response writer
	Header     http.Header         // Underlying response headers
	RespStatus int                 // HTTP response status code
	RespLen    int                 // Written response Length
}

// Validator is implemented by all data structures that can be validated.
// This includes request and response body data structures.
type Validator interface {
	Validate() error
}

// Get returns the param or query string with the given name and true or an empty string and false
// if there isn't one.
func (c *Context) Get(name string) (string, bool) {
	v, ok := c.Params[name]
	if !ok {
		var vs []string
		vs, ok = c.Query[name]
		if ok {
			v = vs[0]
		}
	}
	if !ok {
		return "", false
	}
	return v, true
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
	c.RespStatus = code
	c.W.WriteHeader(code)
	if _, err := c.W.Write(body); err != nil {
		return err
	}
	c.RespLen = len(body)
	return nil
}

// JSON serializes the given body into JSON and sends a HTTP response with the given code
// and serialized JSON string as body.
func (c *Context) JSON(code int, body interface{}) error {
	js, err := json.Marshal(body)
	if err != nil {
		return err
	}
	return c.Respond(code, js)
}

// RespondBadRequest sends a HTTP response with status code 400 and the given body.
func (c *Context) RespondBadRequest(body string) error {
	return c.Respond(400, []byte(body))
}

// RespondBug sends a HTTP response with status code 500 and the given body.
func (c *Context) RespondBug(body string) error {
	return c.Respond(500, []byte(body))
}
