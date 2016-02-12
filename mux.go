package goa

import (
	"net/http"
	"net/url"
	"regexp"

	"github.com/goadesign/goa/design"
	"github.com/julienschmidt/httprouter"
)

type (
	// ServeMux is the interface implemented by the service request muxes. There is one instance
	// of ServeMux per service version and one for requests targetting no version.
	// It implements http.Handler and makes it possible to register request handlers for
	// specific HTTP methods and request path via the Handle method.
	ServeMux interface {
		http.Handler
		// Handle sets the HandleFunc for a given HTTP method and path.
		Handle(method, path string, handle HandleFunc)
		// Lookup returns the HandleFunc associated with the given HTTP method and path.
		Lookup(method, path string) HandleFunc
	}

	// HandleFunc provides the implementation for an API endpoint.
	// The values include both the querystring and path parameter values.
	HandleFunc func(http.ResponseWriter, *http.Request, url.Values)

	// DefaultMux is the default goa mux. It dispatches requests to the appropriate version mux
	// using a SelectVersionFunc. The default func is DefaultVersionFunc, change it with
	// SelectVersion.
	DefaultMux struct {
		*defaultVersionMux
		SelectVersionFunc SelectVersionFunc
		missingVerFunc    handleMissingVersionFunc
		muxes             map[string]ServeMux
	}

	// SelectVersionFunc is used by the default goa mux to compute the API version targetted by
	// a given request.
	// The default implementation looks for a version as path prefix.
	// Alternate implementations can be set using the DefaultMux SelectVersion method.
	SelectVersionFunc func(*http.Request) string

	// defaultVersionMux is the default API version specific mux implementation.
	defaultVersionMux struct {
		router  *httprouter.Router
		handles map[string]HandleFunc
	}

	// HandleMissingVersionFunc is the signature of the function called back when requests
	// target an unsupported API version.
	handleMissingVersionFunc func(rw http.ResponseWriter, req *http.Request, version string)
)

// NewMux returns the default service mux implementation.
func NewMux(app *Application) ServeMux {
	return &DefaultMux{
		defaultVersionMux: &defaultVersionMux{
			router:  httprouter.New(),
			handles: make(map[string]HandleFunc),
		},
		missingVerFunc: func(rw http.ResponseWriter, req *http.Request, version string) {
			if app.missingVersionHandler != nil {
				app.missingVersionHandler(RootContext, rw, req, version)
			}
		},
		SelectVersionFunc: PathSelectVersionFunc("/:api_version/", "api"),
	}
}

// PathSelectVersionFunc returns a SelectVersionFunc that uses the given path pattern to extract the
// version from the request path. Use the same path pattern given in the DSL to define the API base
// path, e.g. "/api/:api_version".
// If the pattern matches zeroVersion then the empty version is returned (i.e. the unversioned
// controller handles the request).
func PathSelectVersionFunc(pattern, zeroVersion string) SelectVersionFunc {
	rgs := design.WildcardRegex.ReplaceAllLiteralString(pattern, `/([^/]+)`)
	rg := regexp.MustCompile("^" + rgs)
	return func(req *http.Request) (version string) {
		match := rg.FindStringSubmatch(req.URL.Path)
		if len(match) > 1 && match[1] != zeroVersion {
			version = match[1]
		}
		return
	}
}

// HeaderSelectVersionFunc returns a SelectVersionFunc that looks for the version in the header with
// the given name.
func HeaderSelectVersionFunc(header string) SelectVersionFunc {
	return func(req *http.Request) string {
		return req.Header.Get(header)
	}
}

// QuerySelectVersionFunc returns a SelectVersionFunc that looks for the version in the querystring
// with the given key.
func QuerySelectVersionFunc(query string) SelectVersionFunc {
	return func(req *http.Request) string {
		return req.URL.Query().Get(query)
	}
}

// CombineSelectVersionFunc returns a SelectVersionFunc that tries each func passed as argument
// in order and returns the first non-empty string version.
func CombineSelectVersionFunc(funcs ...SelectVersionFunc) SelectVersionFunc {
	return func(req *http.Request) string {
		for _, f := range funcs {
			if version := f(req); version != "" {
				return version
			}
		}
		return ""
	}
}

// SelectVersion sets the func used to compute the API version targetted by a request.
func (m *DefaultMux) SelectVersion(sv SelectVersionFunc) {
	m.SelectVersionFunc = sv
}

// ServeHTTP is the function called back by the underlying HTTP server to handle incoming requests.
func (m *DefaultMux) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Optimize the unversionned API case
	if len(m.muxes) == 0 {
		m.router.ServeHTTP(rw, req)
		return
	}
	var mux ServeMux
	version := m.SelectVersionFunc(req)
	if version == "" {
		mux = m.defaultVersionMux
	} else {
		var ok bool
		mux, ok = m.muxes[version]
		if !ok {
			m.missingVerFunc(rw, req, version)
			return
		}
	}
	mux.ServeHTTP(rw, req)
}

// version returns the mux addressing the given version if any.
func (m *DefaultMux) version(version string) ServeMux {
	if m.muxes == nil {
		m.muxes = make(map[string]ServeMux)
	}
	if mux, ok := m.muxes[version]; ok {
		return mux
	}
	mux := &defaultVersionMux{
		router:  httprouter.New(),
		handles: make(map[string]HandleFunc),
	}
	m.muxes[version] = mux
	return mux
}

// Handle sets the handler for the given verb and path.
func (m *defaultVersionMux) Handle(method, path string, handle HandleFunc) {
	hthandle := func(rw http.ResponseWriter, req *http.Request, htparams httprouter.Params) {
		params := req.URL.Query()
		for _, p := range htparams {
			params.Set(p.Key, p.Value)
		}
		handle(rw, req, params)
	}
	m.handles[method+path] = handle
	m.router.Handle(method, path, hthandle)
}

// Lookup returns the HandleFunc associated with the given method and path.
func (m *defaultVersionMux) Lookup(method, path string) HandleFunc {
	return m.handles[method+path]
}

// ServeHTTP is the function called back by the underlying HTTP server to handle incoming requests.
func (m *defaultVersionMux) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	m.router.ServeHTTP(rw, req)
}
