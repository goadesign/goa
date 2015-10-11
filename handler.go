package goa

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/julienschmidt/httprouter"
)

// Handler defines a goa controller action handler signature.
// Handlers accept a context and return an error.
// If the error returned is not nil then the controller error handler (if defined) or application
// error handler gets invoked.
type Handler func(Context) error

// NewHTTPRouterHandle returns a httprouter handle which initializes a new context using the HTTP
// request state and calls the given goa handler with it.
func NewHTTPRouterHandle(app *Application, resName string, h Handler) httprouter.Handle {
	log := app.Logger.New("ctrl", resName)
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// Setup recover
		defer func() {
			if r := recover(); r != nil {
				app.Error(fmt.Sprintf("BUG: %v", r))
				w.WriteHeader(500)
			}
		}()
		// Log started event
		startedAt := time.Now()
		id := ShortID()
		log.Info("started", "id", id, r.Method, r.URL.String())

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
		ctx := &ContextData{
			Logger:      log.New("id", id),
			Params:      params,
			Query:       query,
			PayloadData: payload,
			R:           r,
			W:           w,
			HeaderData:  w.Header(),
		}
		if len(params) > 0 {
			ctx.Debug("params", ToLogCtx(params))
		}
		if len(query) > 0 {
			ctx.Debug("query", ToLogCtxA(query))
		}
		if err != nil {
			ctx.Respond(400, []byte(fmt.Sprintf(`{"kind":"invalid request","msg":"invalid JSON: %s"}`, err)))
			goto end
		}
		if r.ContentLength > 0 {
			if mp, ok := payload.(map[string]interface{}); ok {
				ctx.Debug("payload", log15.Ctx(mp))
			} else {
				ctx.Debug("payload", "raw", payload)
			}
		}

		// Call user controller handler
		if err := h(ctx); err != nil {
			app.ErrorHandler(ctx, err)
		}

		// We're done
	end:
		log.Info("completed", "id", id, "status", ctx.RespStatus,
			"bytes", ctx.RespLen, "time", time.Since(startedAt).String())
	}
}

// ShortID produces a "unique" 6 bytes long string.
// Do not use as a reliable way to get unique IDs, instead use for things like logging.
func ShortID() string {
	b := make([]byte, 6)
	io.ReadFull(rand.Reader, b)
	return base64.StdEncoding.EncodeToString(b)
}
