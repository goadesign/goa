package dsl

import (
	"strings"

	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// Files defines a endpoint that serves static assets via HTTP. The logic for
// what to do when the filename points to a file vs. a directory is the same as
// the standard http package ServeFile function. The path may end with a
// wildcard that matches the rest of the URL (e.g. *filepath). If it does the
// matching path is appended to filename to form the full file path, so:
//
//     Files("/index.html", "/www/data/index.html")
//
// returns the content of the file "/www/data/index.html" when requests are sent
// to "/index.html" and:
//
//    Files("/assets/*filepath", "/www/data/assets")
//
// returns the content of the file "/www/data/assets/x/y/z" when requests are
// sent to "/assets/x/y/z".
//
// Files must appear in Service.
//
// Files accepts 2 arguments and an optional DSL. The first argument is the
// request path which may use a wildcard starting with *. The second argument is
// the path on disk to the files being served. The file path may be absolute or
// relative to the current path of the process.  The DSL allows setting a
// description and documentation.
//
// Example:
//
//    var _ = Service("bottle", func() {
//        Files("/index.html", "/www/data/index.html", func() {
//            Description("Serve home page")
//            Docs(func() {
//                Description("Additional documentation")
//                URL("https://goa.design")
//            })
//        })
//    })
//
func Files(path, filename string, fns ...func()) {
	if len(fns) > 1 {
		eval.ReportError("too many arguments given to Files")
		return
	}
	// Make sure request path starts with a "/" so codegen can rely on it.
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if s, ok := eval.Current().(*expr.ServiceExpr); ok {
		r := expr.Root.API.HTTP.ServiceFor(s)
		server := &expr.HTTPFileServerExpr{
			Service:      r,
			RequestPaths: []string{path},
			FilePath:     filename,
		}
		if len(fns) > 0 {
			eval.Execute(fns[0], server)
		}
		r.FileServers = append(r.FileServers, server)
	}
}
