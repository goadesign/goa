package goa

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/context"
	log "gopkg.in/inconshreveable/log15.v2"
)

type (
	// Service is the interface implemented by all goa services.
	// It provides methods for configuring a service and running it.
	Service interface {
		// Logging methods, configure the log handler using the Logger global variable.
		log.Logger

		// Encoding manages the service decoders and encoders.
		Encoding

		// Name is the name of the goa application.
		Name() string

		// Use adds a middleware to the service-wide middleware chain.
		Use(m Middleware)

		// ErrorHandler returns the currently set error handler, useful for middleware.
		ErrorHandler() ErrorHandler

		// SetErrorHandler registers the service-wide error handler.
		SetErrorHandler(ErrorHandler)

		// SetMissingVersionHandler registers the handler invoked when a request targets a
		// non existant API version.
		SetMissingVersionHandler(MissingVersionHandler)

		// ServeMux returns the service mux.
		ServeMux() ServeMux

		// ListenAndServe starts a HTTP server on the given port.
		ListenAndServe(addr string) error

		// ListenAndServeTLS starts a HTTPS server on the given port.
		ListenAndServeTLS(add, certFile, keyFile string) error

		// ServeFiles replies to the request with the contents of the named file or
		// directory. The logic // for what to do when the filename points to a file vs. a
		// directory is the same as the standard http package ServeFile function. The path
		// may end with a wildcard that matches the rest of the URL (e.g. *filepath). If it
		// does the matching path is appended to filename to form the full file path, so:
		// 	ServeFiles("/index.html", "/www/data/index.html")
		// Returns the content of the file "/www/data/index.html" when requests are sent to
		// "/index.html" and:
		//	ServeFiles("/assets/*filepath", "/www/data/assets")
		// returns the content of the file "/www/data/assets/x/y/z" when requests are sent
		// to "/assets/x/y/z".
		ServeFiles(path, filename string) error

		// Version returns an object that implements ServiceVersion based on the version name.
		// If there is no version registered, it will instantiate a new version.
		Version(name string) ServiceVersion

		// Decode uses registered Decoders to unmarshal a body based on the contentType
		Decode(v interface{}, body io.Reader, contentType string) error

		// NewController returns a controller for the resource with the given name.
		// This method is mainly intended for use by generated code.
		NewController(resName string) Controller
	}

	// ServiceVersion is the interface for interacting with individual service versions.
	ServiceVersion interface {
		// Encoding manages the version decoders and encoders.
		Encoding

		// VersionName returns the version name.
		VersionName() string

		// ServeMux returns the version request mux.
		ServeMux() ServeMux
	}

	// Encoding contains the encoding and decoding support.
	Encoding interface {
		// DecodeRequest uses registered Decoders to unmarshal the request body based on
		// the request "Content-Type" header.
		DecodeRequest(ctx *Context, v interface{}) error

		// EncodeResponse uses registered Encoders to marshal the response body based on the
		// request "Accept" header and writes the result to the http.ResponseWriter.
		EncodeResponse(ctx *Context, v interface{}) error

		// SetDecoder registers a decoder for the given content types.
		// If makeDefault is true then the decoder is used to decode payloads where none of
		// the registered decoders support the request content type.
		SetDecoder(f DecoderFactory, makeDefault bool, contentTypes ...string)

		// SetEncoder registers an encoder for the given content types.
		// If makeDefault is true then the encoder is used to encode bodies where none of
		// the registered encoders match the request "Accept" header.
		SetEncoder(f EncoderFactory, makeDefault bool, contentTypes ...string)
	}

	// Controller is the interface implemented by all goa controllers.
	// A controller implements a given resource actions. There is a one-to-one relationship
	// between designed resources and generated controllers.
	// Controllers may override the service wide error handler and be equipped with controller
	// specific middleware.
	Controller interface {
		log.Logger

		// Use adds a middleware to the controller middleware chain.
		// It is a convenient method for doing append(ctrl.MiddlewareChain(), m)
		Use(Middleware)

		// MiddlewareChain returns the controller middleware chain including the
		// service-wide middleware.
		MiddlewareChain() []Middleware

		// ErrorHandler returns the currently set error handler.
		ErrorHandler() ErrorHandler

		// SetErrorHandler sets the controller specific error handler.
		SetErrorHandler(ErrorHandler)

		// HandleFunc returns a HandleFunc from the given handler
		// name is used solely for logging.
		HandleFunc(name string, h, d Handler) HandleFunc
	}

	// Application represents a goa application. At the basic level an application consists of
	// a set of controllers, each implementing a given resource actions. goagen generates
	// global functions - one per resource - that make it possible to mount the corresponding
	// controller onto an application. An application contains the middleware, logger and error
	// handler shared by all its controllers. Setting up an application might look like:
	//
	//	api := goa.New("my api")
	//	api.Use(SomeMiddleware())
	//	rc := NewResourceController()
	//	rc.Use(SomeOtherMiddleware())
	//	app.MountResourceController(api, rc)
	//	api.ListenAndServe(":80")
	//
	// where NewResourceController returns an object that implements the resource actions as
	// defined by the corresponding interface generated by goagen.
	Application struct {
		log.Logger                                  // Application logger
		*version                                    // Embedded default version
		name                  string                // Application name
		errorHandler          ErrorHandler          // Application error handler
		missingVersionHandler MissingVersionHandler // Missing version handler
		middleware            []Middleware          // Middleware chain
		versions              map[string]*version   // Versions by version string
	}

	// ApplicationController provides the common state and behavior for generated controllers.
	ApplicationController struct {
		log.Logger                // Controller logger
		app          *Application //Application which exposes controller
		errorHandler ErrorHandler // Controller specific error handler if any
		middleware   []Middleware // Controller specific middleware if any
	}

	// Handler defines the controller handler signatures.
	// Controller handlers accept a context and return an error.
	// The context provides typed access to the request and response state. It implements
	// the golang.org/x/net/context package Context interface so that handlers may define
	// deadlines and cancelation signals - see the Timeout middleware as an example.
	// If a controller handler returns an error then the application error handler is invoked
	// with the request context and the error. The error handler is responsible for writing the
	// HTTP response. See DefaultErrorHandler and TerseErrorHandler.
	Handler func(*Context) error

	// ErrorHandler defines the application error handler signature.
	ErrorHandler func(*Context, error)

	// MissingVersionHandler defines the function that handles requests targetting a non
	// existant API version.
	MissingVersionHandler func(*Context, string)

	// DecodeFunc is the function that initialize the unmarshaled payload from the request body.
	DecodeFunc func(*Context, io.ReadCloser, interface{}) error

	// A version represents a service version, identified by a version name. This is where
	// application data that needs to be different per version lives.
	version struct {
		name                  string                  // This is the version string
		mux                   ServeMux                // Request mux
		decoderPools          map[string]*decoderPool // Registered decoders for the service
		encoderPools          map[string]*encoderPool // Registered encoders for the service
		encodableContentTypes []string                // List of contentTypes for response negotiation
	}
)

var (
	// Log is the global logger from which other loggers (e.g. request specific loggers) are
	// derived. Configure it by setting its handler prior to calling New.
	// See https://godoc.org/github.com/inconshreveable/log15
	Log log.Logger

	// RootContext is the root context from which all request contexts are derived.
	// Set values in the root context prior to starting the server to make these values
	// available to all request handlers:
	//
	//	goa.RootContext = goa.RootContext.WithValue(key, value)
	//
	RootContext context.Context

	// cancel is the root context CancelFunc.
	// Call Cancel to send a cancellation signal to all the active request handlers.
	cancel context.CancelFunc
)

// Log to STDOUT by default.
func init() {
	Log = log.New()
	Log.SetHandler(log.StdoutHandler)
	RootContext, cancel = context.WithCancel(context.Background())
}

// New instantiates an application with the given name and default decoders/encoders.
func New(name string) Service {
	app := &Application{
		Logger:                Log.New("app", name),
		name:                  name,
		errorHandler:          DefaultErrorHandler,
		missingVersionHandler: DefaultMissingVersionHandler,
	}
	app.version = &version{
		mux:                   NewMux(app),
		decoderPools:          map[string]*decoderPool{},
		encoderPools:          map[string]*encoderPool{},
		encodableContentTypes: []string{},
	}
	return app
}

// Cancel sends a cancellation signal to all handlers through the action context.
// see https://godoc.org/golang.org/x/net/context for details on how to handle the signal.
func Cancel() {
	cancel()
}

// Name returns the application name.
func (app *Application) Name() string {
	return app.name
}

// Use adds a middleware to the application wide middleware chain.
// See NewMiddleware for wrapping goa and http handlers into goa middleware.
// goa comes with a set of commonly used middleware, see middleware.go.
// Controller specific middleware should be mounted using the Controller type Use method instead.
func (app *Application) Use(m Middleware) {
	app.middleware = append(app.middleware, m)
}

// ErrorHandler returns the currently set error handler.
func (app *Application) ErrorHandler() ErrorHandler {
	return app.errorHandler
}

// SetErrorHandler defines an application wide error handler.
// The default error handler (DefaultErrorHandler) responds with a 500 status code and the error
// message in the response body.
// TerseErrorHandler provides an alternative implementation that does not write the error message
// to the response body for internal errors (e.g. for production).
// Set it with SetErrorHandler(TerseErrorHandler).
// Controller specific error handlers should be set using the Controller type SetErrorHandler
// method instead.
func (app *Application) SetErrorHandler(handler ErrorHandler) {
	app.errorHandler = handler
}

// ServeMux returns the top level service mux.
func (app *Application) ServeMux() ServeMux {
	return app.mux
}

// ListenAndServe starts a HTTP server and sets up a listener on the given host/port.
func (app *Application) ListenAndServe(addr string) error {
	app.Info("listen", "addr", addr)
	return http.ListenAndServe(addr, app.ServeMux())
}

// ListenAndServeTLS starts a HTTPS server and sets up a listener on the given host/port.
func (app *Application) ListenAndServeTLS(addr, certFile, keyFile string) error {
	app.Info("listen ssl", "addr", addr)
	return http.ListenAndServeTLS(addr, certFile, keyFile, app.ServeMux())
}

// ServeFiles replies to the request with the contents of the named file or directory. The logic
// for what to do when the filename points to a file vs. a directory is the same as the standard
// http package ServeFile function. The path may end with a wildcard that matches the rest of the
// URL (e.g. *filepath). If it does the matching path is appended to filename to form the full file
// path, so:
// 	ServeFiles("/index.html", "/www/data/index.html")
// Returns the content of the file "/www/data/index.html" when requests are sent to "/index.html"
// and:
//	ServeFiles("/assets/*filepath", "/www/data/assets")
// returns the content of the file "/www/data/assets/x/y/z" when requests are sent to
// "/assets/x/y/z".
func (app *Application) ServeFiles(path, filename string) error {
	if strings.Contains(path, ":") {
		return fmt.Errorf("path may only include wildcards that match the entire end of the URL (e.g. *filepath)")
	}
	if _, err := os.Stat(filename); err != nil {
		return fmt.Errorf("ServeFiles: %s", err)
	}
	app.Info("mount", "file", filename, "route", fmt.Sprintf("GET %s", path))
	ctrl := app.NewController("FileServer")
	handle := ctrl.HandleFunc("Serve", func(ctx *Context) error {
		fullpath := filename
		params := ctx.GetNames()
		if len(params) > 0 {
			suffix := ctx.Get(params[0])
			fullpath = filepath.Join(fullpath, suffix)
		}
		app.Info("serve", "path", ctx.Request().URL.Path, "filename", fullpath)
		http.ServeFile(ctx, ctx.Request(), fullpath)
		return nil
	}, nil)
	app.ServeMux().Handle("GET", path, handle)
	return nil
}

// Version returns an object that implements ServiceVersion based on the version name.
// If there is no version registered, it will instantiate a new version.
func (app *Application) Version(name string) ServiceVersion {
	if app.versions == nil {
		app.versions = make(map[string]*version, 1)
	}

	ver, ok := app.versions[name]
	if ok {
		return ver
	}
	var verMux ServeMux
	if m, ok := app.mux.(*DefaultMux); ok {
		verMux = m.version(name)
	} else {
		verMux = app.mux
	}
	ver = &version{
		name:                  name,
		mux:                   verMux,
		decoderPools:          map[string]*decoderPool{},
		encoderPools:          map[string]*encoderPool{},
		encodableContentTypes: []string{},
	}
	if app.versions == nil {
		app.versions = make(map[string]*version, 1)
	}
	app.versions[ver.name] = ver

	return ver
}

// SetMissingVersionHandler registers the service missing version handler.
func (app *Application) SetMissingVersionHandler(handler MissingVersionHandler) {
	app.missingVersionHandler = handler
}

// VersionMux returns the version specific mux.
func (ver *version) ServeMux() ServeMux {
	return ver.mux
}

// VersionName returns the version name.
func (ver *version) VersionName() string {
	return ver.name
}

// NewController returns a controller for the given resource. This method is mainly intended for
// use by the generated code. User code shouldn't have to call it directly.
func (app *Application) NewController(resName string) Controller {
	logger := app.New("ctrl", resName)
	return &ApplicationController{
		Logger: logger,
		app:    app,
	}
}

// Use adds a middleware to the controller.
// See NewMiddleware for wrapping goa and http handlers into goa middleware.
// goa comes with a set of commonly used middleware, see middleware.go.
func (ctrl *ApplicationController) Use(m Middleware) {
	ctrl.middleware = append(ctrl.middleware, m)
}

// MiddlewareChain returns the controller middleware chain.
func (ctrl *ApplicationController) MiddlewareChain() []Middleware {
	return append(ctrl.app.middleware, ctrl.middleware...)
}

// ErrorHandler returns the currently set error handler.
func (ctrl *ApplicationController) ErrorHandler() ErrorHandler {
	return ctrl.errorHandler
}

// SetErrorHandler defines a controller specific error handler. When a controller action returns an
// error goa checks whether the controller is equipped with a error handler and if so calls it with
// the error given as argument. If there is no controller error handler then goa calls the
// application wide error handler instead.
func (ctrl *ApplicationController) SetErrorHandler(handler ErrorHandler) {
	ctrl.errorHandler = handler
}

// HandleError invokes the controller error handler or - if there isn't one - the service error
// handler.
func (ctrl *ApplicationController) HandleError(ctx *Context, err error) {
	if ctrl.errorHandler != nil {
		ctrl.errorHandler(ctx, err)
	} else if ctrl.app.errorHandler != nil {
		ctrl.app.errorHandler(ctx, err)
	}
}

// HandleFunc wraps al request handler into a HandleFunc. The HandleFunc initializes the
// request context by loading the request state, invokes the handler and in case of error invokes
// the controller (if there is one) or application error handler.
// This function is intended for the controller generated code. User code should not need to call
// it directly.
func (ctrl *ApplicationController) HandleFunc(name string, h, d Handler) HandleFunc {
	// Setup middleware outside of closure
	middleware := func(ctx *Context) error {
		if !ctx.ResponseWritten() {
			if err := h(ctx); err != nil {
				ctrl.HandleError(ctx, err)
			}
		}
		return nil
	}
	chain := ctrl.MiddlewareChain()
	ml := len(chain)
	for i := range chain {
		middleware = chain[ml-i-1](middleware)
	}
	return func(w http.ResponseWriter, r *http.Request, params url.Values) {
		// Build context
		gctx, cancel := context.WithCancel(RootContext)
		defer cancel() // Signal completion of request to any child goroutine
		ctx := NewContext(gctx, ctrl.app, r, w, params)
		ctx.Logger = ctrl.Logger.New("action", name)

		// Load body if any
		var err error
		if r.ContentLength > 0 && d != nil {
			err = d(ctx)
		}

		// Handle invalid payload
		handler := middleware
		if err != nil {
			handler = func(ctx *Context) error {
				msg := "invalid encoding: " + err.Error()
				ctx.Respond(400, fmt.Sprintf(`{"kind":"invalid request","msg":%q}`, msg))
				return nil
			}
			for i := range chain {
				handler = chain[ml-i-1](handler)
			}
		}

		// Invoke middleware chain
		handler(ctx)

		// Make sure a response is sent back to client.
		if ctx.ResponseStatus() == 0 {
			ctrl.HandleError(ctx, fmt.Errorf("unhandled request"))
		}
	}
}

// DefaultErrorHandler returns a 400 response for request validation errors (instances of
// BadRequestError) and a 500 response for other errors. It writes the error message to the
// response body in both cases.
func DefaultErrorHandler(ctx *Context, e error) {
	status := 500
	if _, ok := e.(*BadRequestError); ok {
		status = 400
	}
	ctx.Respond(status, e.Error())
}

// TerseErrorHandler behaves like DefaultErrorHandler except that it does not write to the response
// body for internal errors.
func TerseErrorHandler(ctx *Context, e error) {
	status := 500
	var body interface{}
	if _, ok := e.(*BadRequestError); ok {
		status = 400
		body = e.Error()
	}
	ctx.Respond(status, body)
}

// DefaultMissingVersionHandler returns a 400 response with a typed error in the body containing
// the name of the version that was targeted by the request.
func DefaultMissingVersionHandler(ctx *Context, version string) {
	resp := TypedError{
		ID:   ErrInvalidVersion,
		Mesg: fmt.Sprintf(`API does not support version %s`, version),
	}
	ctx.Respond(400, resp)
}

// Fatal logs a critical message and exits the process with status code 1.
// This function is meant to be used by initialization code to prevent the application from even
// starting up when something is obviously wrong.
// In particular this function should probably not be used when serving requests.
func Fatal(msg string, ctx ...interface{}) {
	log.Crit(msg, ctx...)
	os.Exit(1)
}
