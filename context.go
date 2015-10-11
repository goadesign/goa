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
	// Log is the request specific logger.
	// All log entries are prefixed with the request ID.
	log.Logger

	// Get returns the param or query string with the given name and true or an empty string
	// and false if there isn't one.
	// If there is more than one query string value then Get returns the first one. Use GetMany
	// to retrieve all the values instead.
	Get(name string) (string, bool)

	// GetMany returns the query string values with the given name or nil if there aren't any.
	GetMany(name string) []string

	// Env returns the request environment.
	// Environment values are typically created and used by middleware.
	Env() map[string]interface{}

	// Payload returns the raw deserialized request body or nil if the body is empty.
	Payload() interface{}

	// Header returns the underlying request headers.
	Header() http.Header

	// Respond writes the given HTTP status code and response body.
	// This method can only be called once per request.
	Respond(code int, body []byte) error

	// JSON serializes the given body into JSON and sends a HTTP response with the given code
	// and serialized JSON string as body.
	JSON(code int, body interface{}) error

	// Bug sends a HTTP response with status code 400 and the given body.
	BadRequest(body string) error

	// Bug sends a HTTP response with status code 500 and the given body.
	Bug(body string) error

	// Request returns the underlying HTTP request.
	// In general it should not be necessary to call this function, the request elements should
	// be accessed using the other functions exposed by Context (i.e. Get, GetMany, Payload, etc.)
	Request() *http.Request

	// ResponseWriter returns the underlying HTTP response writer.
	// In general it should not be necessary to call this function, the response can be Written
	// using on the helper functions exposes by Context (i.e. Respond, JSON, RespondError and
	// Bug)
	ResponseWriter() http.ResponseWriter

	// Duplicate creates a new context with the same environment. Only the Request and
	// ResponseWriter are different (in particular note that Get and GetMany are unaffected).
	// This method is useful to wrap "legacy" middleware that acts on http.Handler instead of
	// goa.Handler.
	Duplicate(http.ResponseWriter, *http.Request) Context
}

// ContextData is the object that provides access to the underlying HTTP request and response data.
// It implements the Context interface.
type ContextData struct {
	log.Logger                         // Context logger
	Params      map[string]string      // URL string parameters
	Query       map[string][]string    // Query string parameters
	PayloadData interface{}            // Deserialized payload (request body)
	R           *http.Request          // Underlying HTTP request
	W           http.ResponseWriter    // Underlying HTTP response writer
	HeaderData  http.Header            // Underlying response headers
	RespStatus  int                    // HTTP response status code
	RespLen     int                    // Written response Length
	EnvData     map[string]interface{} // Env is the request environment used by middleware to stash state.
}

// Duplicate creates a new context with the same environment. Only the R and W fields are updated.
// This is useful to create contexts out of middlewares that act directly on http.Handler.
func (c *ContextData) Duplicate(w http.ResponseWriter, r *http.Request) Context {
	newC := ContextData{
		Logger:      c.Logger,
		Params:      c.Params,
		Query:       c.Query,
		PayloadData: c.PayloadData,
		R:           r,
		W:           w,
		HeaderData:  c.HeaderData,
		RespStatus:  c.RespStatus,
		RespLen:     c.RespLen,
		EnvData:     c.EnvData,
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

// GetMany returns the query string values with the given name and or if there aren't any.
func (c *ContextData) GetMany(name string) []string {
	return c.Query[name]
}

// Env returns the request environment.
// Environment values are typically created and used by middleware.
func (c *ContextData) Env() map[string]interface{} {
	return c.EnvData
}

// Payload returns the deserialized request body or nil if body is empty.
func (c *ContextData) Payload() interface{} {
	return c.PayloadData
}

// Header returns the underlying request headers.
func (c *ContextData) Header() http.Header {
	return c.HeaderData
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

// BadRequest sends a HTTP response with status code 400 and the given body.
func (c *ContextData) BadRequest(body string) error {
	return c.Respond(400, []byte(body))
}

// Bug sends a HTTP response with status code 500 and the given body.
func (c *ContextData) Bug(body string) error {
	return c.Respond(500, []byte(body))
}

// Log is the request specific logger.
func (c *ContextData) Log() log15.Logger {
	return c.Logger
}

// Request returns the underlying HTTP request.
// In general it should not be necessary to call this function, the request elements should
// be accessed using the other functions exposed by Context (i.e. Get, GetMany, Payload, etc.)
func (c *ContextData) Request() *http.Request {
	return c.R
}

// ResponseWriter returns the underlying HTTP response writer.
// In general it should not be necessary to call this function, the response can be Written
// using on the helper functions exposes by Context (i.e. Respond, JSON, RespondError and
// Bug)
func (c *ContextData) ResponseWriter() http.ResponseWriter {
	return c.W
}
