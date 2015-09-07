package goa

import (
	"encoding/json"
	"net/http"

	"gopkg.in/inconshreveable/log15.v2"

	log "gopkg.in/inconshreveable/log15.v2"
)

// Context is the interface implemented by all context objects.
// It provides access to the underlying request data and exposes helper methods to create the
// response.
type Context interface {
	// Get returns the param or query string with the given name and true or an empty string
	// and false if there isn't one.
	Get(name string) (string, bool)

	// Env returns the request environment.
	// Environment values are typically created and used by middleware.
	Env() map[string]interface{}

	// Bind loads the request body in the given Validator then calls the Validate method on the
	// unmarshalled object.
	Bind(v Validator) error

	// Respond writes the given HTTP status code and response body.
	// This method can only be called once per request.
	Respond(code int, body []byte) error

	// JSON serializes the given body into JSON and sends a HTTP response with the given code
	// and serialized JSON string as body.
	JSON(code int, body interface{}) error

	// RespondBadRequest sends a HTTP response with status code 400 and the given body.
	RespondBadRequest(body string) error

	// RespondBug sends a HTTP response with status code 500 and the given body.
	RespondBug(body string) error

	// Log is the request specific logger.
	// All log entries are prefixed with the request ID.
	Log() log15.Logger

	// Request returns the underlying HTTP request.
	// In general it should not be necessary to call this function, the request elements should
	// be accessed using the other functions exposed by Context (i.e. Get, Env and Bind)
	Request() *http.Request

	// ResponseWriter returns the underlying HTTP response writer.
	// In general it should not be necessary to call this function, the response can be Written
	// using on the helper functions exposes by Context (i.e. Respond, JSON, RespondError and
	// RespondBug)
	ResponseWriter() http.ResponseWriter

	// New creates a new context with the same environment. Only the Request and ResponseWriter
	// are different (in particular note that Get returns the same value with the new context
	// even if the request has a different URL).
	// This method is used to wrap "legacy" middleware that acts on http.Handler instead of
	// goa.Handler. This means that legacy middleware can't affect the request itself, it can
	// only update the environment.
	New(http.ResponseWriter, *http.Request) Context
}

// ContextData is the object that provides access to the underlying HTTP request and response data.
// It implements the Context interface.
type ContextData struct {
	log.Logger                        // Context logger
	Params     map[string]string      // URL string parameters
	Query      map[string][]string    // Query string parameters
	Payload    interface{}            // Deserialized payload (request body)
	R          *http.Request          // Underlying HTTP request
	W          http.ResponseWriter    // Underlying HTTP response writer
	Header     http.Header            // Underlying response headers
	RespStatus int                    // HTTP response status code
	RespLen    int                    // Written response Length
	EnvData    map[string]interface{} // Env is the request environment used by middleware to stash state.
}

// Validator is implemented by all data structures that can be validated.
// This includes request and response body data structures.
type Validator interface {
	Validate() error
}

// New creates a new context with the same environment. Only the R and W fields are updated.
// This is useful to create contexts out of middlewares that act directly on http.Handler.
func (c *ContextData) New(w http.ResponseWriter, r *http.Request) Context {
	newC := ContextData{
		Logger:     c.Logger,
		Params:     c.Params,
		Query:      c.Query,
		Payload:    c.Payload,
		R:          r,
		W:          w,
		Header:     c.Header,
		RespStatus: c.RespStatus,
		RespLen:    c.RespLen,
		EnvData:    c.EnvData,
	}
	return &newC
}

// Get returns the param or query string with the given name and true or an empty string and false
// if there isn't one.
func (c *ContextData) Get(name string) (string, bool) {
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

// Env returns the request environment.
// Environment values are typically created and used by middleware.
func (c *ContextData) Env() map[string]interface{} {
	return c.EnvData
}

// Bind loads the request body in the given Validator then calls the Validate method on the
// unmarshalled object.
func (c *ContextData) Bind(v Validator) error {
	decoder := json.NewDecoder(c.R.Body)
	if err := decoder.Decode(v); err != nil {
		return err
	}
	return v.Validate()
}

// Respond writes the given HTTP status code and response body.
// This method can only be called once per request.
func (c *ContextData) Respond(code int, body []byte) error {
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
func (c *ContextData) JSON(code int, body interface{}) error {
	js, err := json.Marshal(body)
	if err != nil {
		return err
	}
	return c.Respond(code, js)
}

// RespondBadRequest sends a HTTP response with status code 400 and the given body.
func (c *ContextData) RespondBadRequest(body string) error {
	return c.Respond(400, []byte(body))
}

// RespondBug sends a HTTP response with status code 500 and the given body.
func (c *ContextData) RespondBug(body string) error {
	return c.Respond(500, []byte(body))
}

// Log is the request specific logger.
func (c *ContextData) Log() log15.Logger {
	return c.Logger
}

// Request returns the underlying HTTP request.
// In general it should not be necessary to call this function, the request elements should
// be accessed using the other functions exposed by Context (i.e. Get, Env and Bind)
func (c *ContextData) Request() *http.Request {
	return c.R
}

// ResponseWriter returns the underlying HTTP response writer.
// In general it should not be necessary to call this function, the response can be Written
// using on the helper functions exposes by Context (i.e. Respond, JSON, RespondError and
// RespondBug)
func (c *ContextData) ResponseWriter() http.ResponseWriter {
	return c.W
}
