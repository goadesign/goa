package rest

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/goadesign/goa"
)

// HTTPServer implements a server which accepts HTTP requests.
type HTTPServer struct {
	// Service which runs the server.
	Service *goa.Service
	// Mux is the server HTTP mux (a.k.a. router).
	Mux ServeMux
}

// ServeFiles replies to the request with the contents of the named file or directory. See
// FileHandler for details.
func (svr *HTTPServer) ServeFiles(path, filename string) error {
	if strings.Contains(path, ":") {
		return fmt.Errorf("path may only include wildcards that match the entire end of the URL (e.g. *filepath)")
	}
	svr.Service.LogAdapter.Info("mount file", "name", filename, "route", fmt.Sprintf("GET %s", path))
	handler := func(rw http.ResponseWriter, req *http.Request) error {
		return FileHandler(path, filename)(rw, req)
	}
	svr.Mux.Handle("GET", path, ctrl.MuxHandler("serve", handler, nil))
	return nil
}
