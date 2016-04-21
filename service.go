package goa

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/net/context"
)

const (
	// ErrorMediaIdentifier is the media type identifier used for error responses.
	ErrorMediaIdentifier = "application/vnd.api.error+json"
)

var (
	// MaxRequestBodyLength is the maximum length read from request bodies.
	// Set to 0 to remove the limit altogether.
	MaxRequestBodyLength int64 = 1073741824 // 1 GB
)

type (
	// Service is the data structure supporting goa services.
	// It provides methods for configuring a service and running it.
	// At the basic level a service consists of a set of controllers, each implementing a given
	// resource actions. goagen generates global functions - one per resource - that make it
	// possible to mount the corresponding controller onto a service. A service contains the
	// middleware, not found handler, encoders and muxes shared by all its controllers.
	Service struct {
		// Name of service used for logging, tracing etc.
		Name string
		// Mux is the service request mux
		Mux ServeMux
		// Context is the root context from which all request contexts are derived.
		// Set values in the root context prior to starting the server to make these values
		// available to all request handlers.
		Context context.Context

		finalized             bool                    // Whether controllers have been mounted
		middleware            []Middleware            // Middleware chain
		notFound              Handler                 // Handler of requests that don't match registered mux handlers
		cancel                context.CancelFunc      // Service context cancel signal trigger
		decoderPools          map[string]*decoderPool // Registered decoders for the service
		encoderPools          map[string]*encoderPool // Registered encoders for the service
		encodableContentTypes []string                // List of contentTypes for response negotiation
	}

	// Controller defines the common fields and behavior of generated controllers.
	Controller struct {
		Name    string          // Controller resource name
		Service *Service        // Service that exposes the controller
		Context context.Context // Controller root context

		middleware []Middleware // Controller specific middleware if any
	}

	// Handler defines the request handler signatures.
	Handler func(context.Context, http.ResponseWriter, *http.Request) error

	// Unmarshaler defines the request payload unmarshaler signatures.
	Unmarshaler func(context.Context, *Service, *http.Request) error

	// DecodeFunc is the function that initialize the unmarshaled payload from the request body.
	DecodeFunc func(context.Context, io.ReadCloser, interface{}) error
)

// New instantiates a service with the given name.
func New(name string) *Service {
	var (
		stdlog       = log.New(os.Stderr, "", log.LstdFlags)
		ctx          = WithLogger(context.Background(), NewLogger(stdlog))
		cctx, cancel = context.WithCancel(ctx)
		mux          = NewMux()
		service      = &Service{
			Name:    name,
			Context: cctx,
			Mux:     mux,

			cancel:                cancel,
			decoderPools:          map[string]*decoderPool{},
			encoderPools:          map[string]*encoderPool{},
			encodableContentTypes: []string{},
			notFound: func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
				return ErrNotFound(req.URL.Path)
			},
		}
	)

	mux.HandleNotFound(func(rw http.ResponseWriter, req *http.Request, params url.Values) {
		ctx := NewContext(service.Context, rw, req, params)
		err := service.notFound(ctx, rw, req)
		if !ContextResponse(ctx).Written() {
			service.Send(ctx, 404, err)
		}
	})

	return service
}

// CancelAll sends a cancel signals to all request handlers via the context.
// See https://godoc.org/golang.org/x/net/context for details on how to handle the signal.
func (service *Service) CancelAll() {
	service.cancel()
}

// Use adds a middleware to the service wide middleware chain.
// goa comes with a set of commonly used middleware, see the middleware package.
// Controller specific middleware should be mounted using the Controller struct Use method instead.
func (service *Service) Use(m Middleware) {
	if service.finalized {
		panic("goa: cannot mount middleware after controller")
	}
	service.middleware = append(service.middleware, m)
}

// WithLogger sets the logger used internally by the service and by Log.
func (service *Service) WithLogger(logger LogAdapter) {
	service.Context = WithLogger(service.Context, logger)
}

// LogInfo logs the message and values at odd indeces using the keys at even indeces of the keyvals slice.
func (service *Service) LogInfo(msg string, keyvals ...interface{}) {
	LogInfo(service.Context, msg, keyvals...)
}

// LogError logs the error and values at odd indeces using the keys at even indeces of the keyvals slice.
func (service *Service) LogError(msg string, keyvals ...interface{}) {
	LogError(service.Context, msg, keyvals...)
}

// ListenAndServe starts a HTTP server and sets up a listener on the given host/port.
func (service *Service) ListenAndServe(addr string) error {
	service.LogInfo("listen", "transport", "http", "addr", addr)
	return http.ListenAndServe(addr, service.Mux)
}

// ListenAndServeTLS starts a HTTPS server and sets up a listener on the given host/port.
func (service *Service) ListenAndServeTLS(addr, certFile, keyFile string) error {
	service.LogInfo("listen", "transport", "https", "addr", addr)
	return http.ListenAndServeTLS(addr, certFile, keyFile, service.Mux)
}

// NewController returns a controller for the given resource. This method is mainly intended for
// use by the generated code. User code shouldn't have to call it directly.
func (service *Service) NewController(name string) *Controller {
	return &Controller{
		Name:    name,
		Service: service,
		Context: context.WithValue(service.Context, ctrlKey, name),
	}
}

// Send serializes the given body matching the request Accept header against the service
// encoders. It uses the default service encoder if no match is found.
func (service *Service) Send(ctx context.Context, code int, body interface{}) error {
	r := ContextResponse(ctx)
	if r == nil {
		return fmt.Errorf("no response data in context")
	}
	r.WriteHeader(code)
	return service.EncodeResponse(ctx, body)
}

// ServeFiles create a "FileServer" controller and calls ServerFiles on it.
func (service *Service) ServeFiles(path, filename string) error {
	ctrl := service.NewController("FileServer")
	return ctrl.ServeFiles(path, filename)
}

// finalize wraps the NotFound handler with the final middleware chain.
// Use cannot be called after finalize has.
func (service *Service) finalize() {
	if service.finalized {
		return
	}
	notFound := service.notFound
	handler := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
		if !ContextResponse(ctx).Written() {
			return notFound(ctx, rw, req)
		}
		return nil
	}
	ml := len(service.middleware)
	for i := range service.middleware {
		handler = service.middleware[ml-i-1](handler)
	}
	service.notFound = handler
	service.finalized = true
}

// ServeFiles replies to the request with the contents of the named file or directory. The logic
// for what to do when the filename points to a file vs. a directory is the same as the standard
// http package ServeFile function. The path may end with a wildcard that matches the rest of the
// URL (e.g. *filepath). If it does the matching path is appended to filename to form the full file
// path, so:
//
// 	ServeFiles("/index.html", "/www/data/index.html")
//
// Returns the content of the file "/www/data/index.html" when requests are sent to "/index.html"
// and:
//
//	ServeFiles("/assets/*filepath", "/www/data/assets")
//
// returns the content of the file "/www/data/assets/x/y/z" when requests are sent to
// "/assets/x/y/z".
func (ctrl *Controller) ServeFiles(path, filename string) error {
	if strings.Contains(path, ":") {
		return fmt.Errorf("path may only include wildcards that match the entire end of the URL (e.g. *filepath)")
	}
	LogInfo(ctrl.Context, "mount file", "name", filename, "route", fmt.Sprintf("GET %s", path))
	handler := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
		if !ContextResponse(ctx).Written() {
			return ctrl.fileServer(filename, path)(ctx, rw, req)
		}
		return nil
	}
	chain := ctrl.middleware
	ml := len(chain)
	for i := range chain {
		handler = chain[ml-i-1](handler)
	}
	handle := func(rw http.ResponseWriter, req *http.Request, params url.Values) {
		baseCtx := WithLogContext(ctrl.Context, "action", "serve")
		ctx := NewContext(baseCtx, rw, req, params)
		// Invoke middleware chain, errors should be caught earlier, e.g. by ErrorHandler middleware
		if err := handler(ctx, rw, req); err != nil {
			LogError(ctx, "uncaught error", "err", err)
			respBody := fmt.Sprintf("internal error: %s", err) // Sprintf catches panics
			ctrl.Service.Send(ctx, 500, respBody)
		}
	}
	ctrl.Service.Mux.Handle("GET", path, handle)
	return nil
}

// Use adds a middleware to the controller.
// Service-wide middleware should be added via the Service Use method instead.
func (ctrl *Controller) Use(m Middleware) {
	if ctrl.Service.finalized {
		panic("goa: cannot mount middleware after controller")
	}
	ctrl.middleware = append(ctrl.middleware, m)
}

// MuxHandler wraps a request handler into a MuxHandler. The MuxHandler initializes the request
// context by loading the request state, invokes the handler and in case of error invokes the
// controller (if there is one) or Service error handler.
// This function is intended for the controller generated code. User code should not need to call
// it directly.
func (ctrl *Controller) MuxHandler(name string, hdlr Handler, unm Unmarshaler) MuxHandler {
	// Make sure middleware doesn't get mounted later
	ctrl.Service.finalize()

	// Setup middleware outside of closure
	middleware := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
		if !ContextResponse(ctx).Written() {
			return hdlr(ctx, rw, req)
		}
		return nil
	}
	chain := append(ctrl.Service.middleware, ctrl.middleware...)
	ml := len(chain)
	for i := range chain {
		middleware = chain[ml-i-1](middleware)
	}
	return func(rw http.ResponseWriter, req *http.Request, params url.Values) {
		// Build context
		ctx := NewContext(WithAction(ctrl.Context, name), rw, req, params)

		// Protect against request bodies with unreasonable length
		if MaxRequestBodyLength > 0 {
			req.Body = http.MaxBytesReader(rw, req.Body, MaxRequestBodyLength)
		}

		// Load body if any
		var err error
		if req.ContentLength > 0 && unm != nil {
			err = unm(ctx, ctrl.Service, req)
		}

		// Handle invalid payload
		handler := middleware
		if err != nil {
			handler = func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
				rw.Header().Set("Content-Type", ErrorMediaIdentifier)
				status := 400
				body := ErrInvalidEncoding(err)
				if err.Error() == "http: request body too large" {
					status = 413
					body = ErrRequestBodyTooLarge("body length exceeds %d bytes", MaxRequestBodyLength)
				}
				return ctrl.Service.Send(ctx, status, body)
			}
			for i := range chain {
				handler = chain[ml-i-1](handler)
			}
		}

		// Invoke middleware chain, errors should be caught earlier, e.g. by ErrorHandler middleware
		if err := handler(ctx, ContextResponse(ctx), req); err != nil {
			LogError(ctx, "uncaught error", "err", err)
			respBody := fmt.Sprintf("Internal error: %s", err) // Sprintf catches panics
			ctrl.Service.Send(ctx, 500, respBody)
		}
	}
}

// fileServer returns a handler that serves files under the given filename for the given route path
func (ctrl *Controller) fileServer(filename, path string) Handler {
	var wc string
	if idx := strings.Index(path, "*"); idx > -1 && idx < len(path)-1 {
		wc = path[idx+1:]
	}
	return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
		fname := filename
		if len(wc) > 0 {
			if m, ok := ContextRequest(ctx).Params[wc]; ok {
				fname = filepath.Join(filename, m[0])
			}
		}
		LogInfo(ctrl.Context, "serve file", "name", fname, "route", req.URL.Path)
		dir, name := filepath.Split(fname)
		fs := http.Dir(dir)
		f, err := fs.Open(name)
		if err != nil {
			return ErrInvalidFile(err)
		}
		defer f.Close()
		d, err := f.Stat()
		if err != nil {
			return ErrInvalidFile(err)
		}
		// use contents of index.html for directory, if present
		if d.IsDir() {
			index := strings.TrimSuffix(name, "/") + "/index.html"
			ff, err := fs.Open(index)
			if err == nil {
				defer ff.Close()
				dd, err := ff.Stat()
				if err == nil {
					name = index
					d = dd
					f = ff
				}
			}
		}

		// serveContent will check modification time
		// Still a directory? (we didn't find an index.html file)
		if d.IsDir() {
			return dirList(rw, f)
		}
		http.ServeContent(rw, req, d.Name(), d.ModTime(), f)
		return nil
	}
}

var replacer = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	// "&#34;" is shorter than "&quot;".
	`"`, "&#34;",
	// "&#39;" is shorter than "&apos;" and apos was not in HTML until HTML5.
	"'", "&#39;",
)

func dirList(w http.ResponseWriter, f http.File) error {
	dirs, err := f.Readdir(-1)
	if err != nil {
		return err
	}
	sort.Sort(byName(dirs))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<pre>\n")
	for _, d := range dirs {
		name := d.Name()
		if d.IsDir() {
			name += "/"
		}
		// name may contain '?' or '#', which must be escaped to remain
		// part of the URL path, and not indicate the start of a query
		// string or fragment.
		url := url.URL{Path: name}
		fmt.Fprintf(w, "<a href=\"%s\">%s</a>\n", url.String(), replacer.Replace(name))
	}
	fmt.Fprintf(w, "</pre>\n")
	return nil
}

type byName []os.FileInfo

func (s byName) Len() int           { return len(s) }
func (s byName) Less(i, j int) bool { return s[i].Name() < s[j].Name() }
func (s byName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
