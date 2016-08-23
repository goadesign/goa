package http

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/goadesign/goa"
)

type Server struct {
	Mux    ServeMux
	Logger goa.LogAdapter
}

// ServeFiles replies to the request with the contents of the named file or directory. See
// FileHandler for details.
func (svr *Server) ServeFiles(path, filename string) error {
	if strings.Contains(path, ":") {
		return fmt.Errorf("path may only include wildcards that match the entire end of the URL (e.g. *filepath)")
	}
	svr.Logger.Info("mount file", "name", filename, "route", fmt.Sprintf("GET %s", path))
	handler := func(rw http.ResponseWriter, req *http.Request) error {
		return ctrl.FileHandler(path, filename)(rw, req)
	}
	svr.Mux.Handle("GET", path, ctrl.MuxHandler("serve", handler, nil))
	return nil
}

// Use adds a middleware to the controller.
// Service-wide middleware should be added via the Service Use method instead.
func (ctrl *Controller) Use(m Middleware) {
	ctrl.middleware = append(ctrl.middleware, m)
}

// MuxHandler wraps a request handler into a MuxHandler. The MuxHandler initializes the request
// context by loading the request state, invokes the handler and in case of error invokes the
// controller (if there is one) or Service error handler.
// This function is intended for the controller generated code. User code should not need to call
// it directly.
func (ctrl *Controller) MuxHandler(name string, hdlr Handler, unm Unmarshaler) MuxHandler {
	// Use closure to enable late computation of handlers to ensure all middleware has been
	// registered.
	var handler Handler

	return func(rw http.ResponseWriter, req *http.Request, params url.Values) {
		// Build handler middleware chains on first invocation
		if handler == nil {
			handler = func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
				if !ContextResponse(ctx).Written() {
					return hdlr(ctx, rw, req)
				}
				return nil
			}
			chain := append(ctrl.Service.middleware, ctrl.middleware...)
			ml := len(chain)
			for i := range chain {
				handler = chain[ml-i-1](handler)
			}
		}

		// Build context
		ctx := NewContext(WithAction(ctrl.Context, name), rw, req, params)

		// Protect against request bodies with unreasonable length
		if ctrl.MaxRequestBodyLength > 0 {
			req.Body = http.MaxBytesReader(rw, req.Body, ctrl.MaxRequestBodyLength)
		}

		// Load body if any
		if req.ContentLength > 0 && unm != nil {
			if err := unm(ctx, ctrl.Service, req); err != nil {
				if err.Error() == "http: request body too large" {
					msg := fmt.Sprintf("request body length exceeds %d bytes", ctrl.MaxRequestBodyLength)
					err = ErrRequestBodyTooLarge(msg)
				} else {
					err = ErrBadRequest(err)
				}
				ctx = WithError(ctx, err)
			}
		}

		// Invoke handler
		if err := handler(ctx, ContextResponse(ctx), req); err != nil {
			LogError(ctx, "uncaught error", "err", err)
			respBody := fmt.Sprintf("Internal error: %s", err) // Sprintf catches panics
			ctrl.Service.Send(ctx, 500, respBody)
		}
	}
}

// FileHandler returns a handler that serves files under the given filename for the given route path.
// The logic for what to do when the filename points to a file vs. a directory is the same as the
// standard http package ServeFile function. The path may end with a wildcard that matches the rest
// of the URL (e.g. *filepath). If it does the matching path is appended to filename to form the
// full file path, so:
//
// 	c.FileHandler("/index.html", "/www/data/index.html")
//
// Returns the content of the file "/www/data/index.html" when requests are sent to "/index.html"
// and:
//
//	c.FileHandler("/assets/*filepath", "/www/data/assets")
//
// returns the content of the file "/www/data/assets/x/y/z" when requests are sent to
// "/assets/x/y/z".
func (ctrl *Controller) FileHandler(path, filename string) Handler {
	var wc string
	if idx := strings.LastIndex(path, "/*"); idx > -1 && idx < len(path)-1 {
		wc = path[idx+2:]
		if strings.Contains(wc, "/") {
			wc = ""
		}
	}
	return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
		fname := filename
		if len(wc) > 0 {
			if m, ok := ContextRequest(ctx).Params[wc]; ok {
				fname = filepath.Join(filename, m[0])
			}
		}
		LogInfo(ctx, "serve file", "name", fname, "route", req.URL.Path)
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
