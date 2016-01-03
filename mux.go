package goa

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/julienschmidt/httprouter"
	"github.com/raphael/goa/design"
)

type (
	// ServeMux is the interface implemented by the goa HTTP request mux. The goa package
	// provides a default implementation with DefaultMux.
	//
	// The goa mux allows for routing to controllers serving different API versions. Each
	// version has is own mux accessed via Version. Upon receving a HTTP request the ServeMux
	// ServeHTTP method looks up the targetted API version and dispatches the request to the
	// corresponding mux.
	ServeMux interface {
		VersionMux
		// Version returns the mux for the given API version.
		Version(version string) VersionMux
		// HandleMissingVersion handles requests that specify a non-existing API version.
		HandleMissingVersion(rw http.ResponseWriter, req *http.Request, version string)
	}

	// VersionMux is the interface implemented by API version specific request mux.
	// It implements http.Handler and makes it possible to register request handlers for
	// specific HTTP methods and request path via the Handle method.
	VersionMux interface {
		http.Handler
		// Handle sets the HandleFunc for a given HTTP method and path.
		Handle(method, path string, handle HandleFunc)
	}

	// HandleFunc provides the implementation for an API endpoint.
	// The values include both the querystring and path parameter values.
	HandleFunc func(http.ResponseWriter, *http.Request, url.Values)

	// DefaultMux is the default goa mux. It dispatches requests to the appropriate version mux
	// using a SelectVersionFunc. The default func is DefaultVersionFunc, change it with
	// SelectVersion.
	DefaultMux struct {
		*defaultVersionMux
		selectVersion SelectVersionFunc
		muxes         map[string]VersionMux
	}

	// SelectVersionFunc is used by the default goa mux to compute the API version targetted by
	// a given request.
	// The default implementation looks for a version in the "X-API-Version" header and if
	// not found in the "api_version" querystring value.
	// Alternate implementations can be set using the DefaultMux SelectVersion method.
	SelectVersionFunc func(*http.Request) string

	// defaultVersionMux is the default goa API version specific mux.
	defaultVersionMux struct {
		router *httprouter.Router
	}
)

// NewMux creates a top level mux using the default goa mux implementation.
func NewMux() ServeMux {
	return &DefaultMux{
		defaultVersionMux: &defaultVersionMux{router: httprouter.New()},
		selectVersion:     DefaultSelectVersion,
	}
}

// DefaultSelectVersion is the default SelectVersionFunc used by the goa default mux.
// It looks up the version in the "X-API-Version" header and if not found uses the "api_version"
// querystring value.
func DefaultSelectVersion(req *http.Request) (version string) {
	version = req.Header.Get("X-API-Version")
	if version == "" {
		version = req.URL.Query().Get("api_version")
	}
	return
}

// PathSelectVersionFunc returns a SelectVersionFunc that uses the given path pattern to extract the
// version from the request path. Use the same path pattern given in the DSL to define the API base
// path, e.g. "/api/:version".
func PathSelectVersionFunc(pattern string) SelectVersionFunc {
	rgs := design.WildcardRegex.ReplaceAllLiteralString(pattern, `([^/]+)`)
	rg := regexp.MustCompile("^" + rgs)
	return func(req *http.Request) (version string) {
		match := rg.FindStringSubmatch(req.URL.Path)
		if len(match) > 1 {
			version = match[1]
		}
		return
	}
}

// Version returns the mux addressing the given version if any.
func (m *DefaultMux) Version(version string) VersionMux {
	if m.muxes == nil {
		m.muxes = make(map[string]VersionMux)
	}
	if mux, ok := m.muxes[version]; ok {
		return mux
	}
	mux := &defaultVersionMux{router: httprouter.New()}
	m.muxes[version] = mux
	return mux
}

// SelectVersion sets the func used to compute the API version targetted by a request.
func (m *DefaultMux) SelectVersion(sv SelectVersionFunc) {
	m.selectVersion = sv
}

// HandleMissingVersion handles requests that specify a non-existing API version.
func (m *DefaultMux) HandleMissingVersion(rw http.ResponseWriter, req *http.Request, version string) {
	rw.WriteHeader(400)
	resp := TypedError{ID: ErrInvalidVersion, Mesg: fmt.Sprintf(`API does not support version %s`, version)}
	b, err := json.Marshal(resp)
	if err != nil {
		b = []byte("API does not support version")
	}
	rw.Write(b)
}

// ServeHTTP is the function called back by the underlying HTTP server to handle incoming requests.
func (m *DefaultMux) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Optimize the unversionned API case
	if len(m.muxes) == 0 {
		m.router.ServeHTTP(rw, req)
		return
	}
	var mux VersionMux
	version := m.selectVersion(req)
	if version == "" {
		mux = m.defaultVersionMux
	} else {
		var ok bool
		mux, ok = m.muxes[version]
		if !ok {
			m.HandleMissingVersion(rw, req, version)
			return
		}
	}
	mux.ServeHTTP(rw, req)
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
	m.router.Handle(method, path, hthandle)
}

// ServeHTTP is the function called back by the underlying HTTP server to handle incoming requests.
func (m *defaultVersionMux) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	m.router.ServeHTTP(rw, req)
}
