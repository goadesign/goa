package rest

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	goa "goa.design/goa.v2"

	"github.com/dimfeld/httptreemux"
)

// FileHandler returns a HTTP handler that serves files under the given filename
// for the given route path. The logic for what to do when the filename points
// to a file vs. a directory is the same as the standard http package ServeFile
// function. The path may end with a wildcard that matches the rest of the URL
// (e.g. *filepath). If it does the matching path is appended to filename to
// form the full file path, so:
//
// 	c.FileHandler("/index.html", "/www/data/index.html")
//
// Returns the content of the file "/www/data/index.html" when requests are sent
// to "/index.html" and:
//
//	c.FileHandler("/assets/*filepath", "/www/data/assets")
//
// returns the content of the file "/www/data/assets/x/y/z" when requests are
// sent to "/assets/x/y/z".
func FileHandler(path, filename string, enc func(http.ResponseWriter, *http.Request) Encoder, logger goa.LogAdapter) http.HandlerFunc {
	var wc string
	if idx := strings.LastIndex(path, "/*"); idx > -1 && idx < len(path)-1 {
		wc = path[idx+2:]
		if strings.Contains(wc, "/") {
			wc = ""
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			fname = filename
			ctx   = r.Context()
		)
		if len(wc) > 0 {
			if m, ok := httptreemux.ContextParams(ctx)[wc]; ok {
				fname = filepath.Join(filename, m)
			}
		}
		dir, name := filepath.Split(fname)
		fs := http.Dir(dir)
		f, err := fs.Open(name)
		if err != nil {
			inv := goa.ErrInvalid("failed to open file: %s", err)
			enc(w, r).Encode(NewErrorResponse(inv))
		}
		defer f.Close()
		d, err := f.Stat()
		if err != nil {
			inv := goa.ErrInvalid("failed to query file: %s", err)
			enc(w, r).Encode(NewErrorResponse(inv))
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
			dirList(w, r, f, enc, logger)
		}
		http.ServeContent(w, r, d.Name(), d.ModTime(), f)
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

func dirList(w http.ResponseWriter, r *http.Request, f http.File, enc func(http.ResponseWriter, *http.Request) Encoder, logger goa.LogAdapter) {
	dirs, err := f.Readdir(-1)
	if err != nil {
		enc(w, r).Encode(err)
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
}

type byName []os.FileInfo

func (s byName) Len() int           { return len(s) }
func (s byName) Less(i, j int) bool { return s[i].Name() < s[j].Name() }
func (s byName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
