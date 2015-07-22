package goa

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	log "gopkg.in/inconshreveable/log15.v2"
)

type (
	// Controllers implement the actions of a single resource.
	// Use NewController to create controller objects.
	Controller struct {
		log.Logger                // Controller logger
		Application  *Application // Parent application
		Handlers     Handlers     // Registered action handlers indexed by name
		ErrorHandler ErrorHandler // Controller specific error handler
		Resource     string       // Name of resource controller implements
	}

	// UserHandlers associates action names with their handler.
	Handlers map[string]UserHandler

	// UserHandlers are functions that contain the implementation for controller actions.
	// The function signatures match the corresponding design action definition.
	UserHandler interface{}
)

// ShortID produces a "unique" 6 byte long string.
// Do not use as a reliable way to get unique IDs, instead use for things like logging.
func ShortID() string {
	b := make([]byte, 6)
	io.ReadFull(rand.Reader, b)
	return base64.StdEncoding.EncodeToString(b)
}

// NewController instantiates a new goa controller for the resource with given name.
func NewController(resource string) *Controller {
	return &Controller{Resource: resource}
}

// SetHandlers sets the controller action handlers.
func (c *Controller) SetHandlers(handlers Handlers) {
	c.Handlers = handlers
}

// SetErrorHandler defines an application wide error handler.
// The default error handler returns a 500 status code with the error message in the response body.
// Controllers may override the application error handler.
func (c *Controller) SetErrorHandler(handler ErrorHandler) {
	c.ErrorHandler = handler
}

// actionHandle returns a httprouter handle for the given action definition.
// The generated handler builds an action specific context and calls the user handler.
func (c *Controller) actionHandle(generated HandlerFunc, user interface{}) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// Setup recover
		defer func() {
			if r := recover(); r != nil {
				//c.handleError()
				// TBD CALL handleERROR
				// CHANGE OTHER PLACES (ERRORS) TO SIMPLY RETURN 400
			}
		}()
		// Log started event
		startedAt := time.Now()
		id := ShortID()
		c.Info("started", "id", id, r.Method, r.URL.String())

		// Collect URL and query string parameters
		params := make(map[string]string, len(p))
		for _, param := range p {
			params[param.Key] = param.Value
		}
		q := r.URL.Query()
		query := make(map[string][]string, len(q))
		for name, value := range q {
			query[name] = value
		}

		// Load body if any
		var payload interface{}
		var err error
		if r.ContentLength > 0 {
			decoder := json.NewDecoder(r.Body)
			err = decoder.Decode(&payload)
		}
		ctx := Context{
			Logger:  c.Logger.New("id", id),
			Params:  params,
			Query:   query,
			Payload: payload,
			R:       r,
			W:       w,
			Header:  w.Header(),
		}
		if len(params) > 0 {
			ctx.Debug("params", ToLogCtx(params))
		}
		if len(query) > 0 {
			ctx.Debug("query", ToLogCtxA(query))
		}
		if err != nil {
			ctx.JSON(400, fmt.Errorf("invalid JSON: %s", err))
			goto end
		}
		if r.ContentLength > 0 {
			if mp, ok := payload.(map[string]interface{}); ok {
				ctx.Debug("payload", log.Ctx(mp))
			} else {
				ctx.Debug("payload", "raw", payload)
			}
		}

		// Call user controller handler
		if err := generated(user, &ctx); err != nil {
			c.handleError(&ctx, err)
		}

		// We're done
	end:
		c.Info("completed", "id", id, "status", ctx.RespStatus,
			"bytes", ctx.RespLen, "time", time.Since(startedAt).String())
	}
}

// handleError looks up the error handler (first in the controller then in the application) and
// invokes it.
func (c *Controller) handleError(ctx *Context, actionErr error) {
	ctx.Error(actionErr.Error())
	handler := c.ErrorHandler
	if handler == nil {
		handler = c.Application.ErrorHandler
	}
	handler(ctx, actionErr)
}
