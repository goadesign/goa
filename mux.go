package goa

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/julienschmidt/httprouter"
)

type (
	// ServeMux is the interface implemented by the goa HTTP request mux.
	// The goa mux allows for routing to different API version controllers serving the same
	// endpoint. The ServeVersion method should be called for each supported version to
	// provide the corresponding mux. Upon receving a HTTP request the ServeMux ServeHTTP
	// method should lookup the desired API version and dispatch the request to the
	// corresponding mux.
	//
	// The default implementation returned by calling NewMux looks up the version in the
	// X-API-Version header and if not found in the api_version querystring value. If no
	// version is found then the top level mux returned by NewMux handles the request. If a
	// version is found but there is no corresponding mux then HandleMissingVersion gets called.
	ServeMux interface {
		VersionMux
		// ServeVersion adds a mux for the given API version.
		// This method is called by the generated code once per API version defined in the design.
		ServeVersion(version string, mux VersionMux)
		// Version returns the mux for the given API version or nil if none.
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

	// defaultMux is the default goa mux.
	defaultMux struct {
		*defaultVersionMux
		muxes map[string]VersionMux
	}

	// defaultVersionMux is the default goa API version specific mux.
	defaultVersionMux struct {
		router *httprouter.Router
	}
)

// NewMux creates a top level mux.
func NewMux() ServeMux {
	return &defaultMux{
		defaultVersionMux: &defaultVersionMux{router: httprouter.New()},
		muxes:             make(map[string]VersionMux),
	}
}

// NewVersionMux creates a version specific mux.
func NewVersionMux() VersionMux {
	return &defaultVersionMux{router: httprouter.New()}
}

// ServeVersion records the mux for a given API version.
func (m *defaultMux) ServeVersion(version string, mux VersionMux) {
	m.muxes[version] = mux
}

// VersionMux returns the mux addressing the given version if any.
func (m *defaultMux) Version(version string) VersionMux {
	if m.muxes == nil {
		return nil
	}
	return m.muxes[version]
}

// HandleMissingVersion handles requests that specify a non-existing API version.
func (m *defaultMux) HandleMissingVersion(rw http.ResponseWriter, req *http.Request, version string) {
	rw.WriteHeader(400)
	resp := TypedError{ID: ErrInvalidVersion, Mesg: fmt.Sprintf("API does not support version %s", version)}
	b, err := json.Marshal(resp)
	if err != nil {
		b = []byte("API does not support version")
	}
	rw.Write(b)
}

// ServeHTTP is the function called back by the underlying HTTP server to handle incoming requests.
func (m *defaultMux) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Optimize the unversionned API case
	if len(m.muxes) == 0 {
		m.router.ServeHTTP(rw, req)
		return
	}
	var mux VersionMux
	version := req.Header.Get("X-API-Version")
	if version == "" {
		version = req.URL.Query().Get("api_version")
	}
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
