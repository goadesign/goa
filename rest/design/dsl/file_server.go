package dsl

import (
	"github.com/goadesign/goa/eval"
	"github.com/goadesign/goa/rest/design"
)

// Files defines a endpoint that serves static assets. The logic for what to do when the
// filename points to a file vs. a directory is the same as the standard http package ServeFile
// function. The path may end with a wildcard that matches the rest of the URL (e.g. *filepath). If
// it does the matching path is appended to filename to form the full file path, so:
//
//     Files("/index.html", "/www/data/index.html")
//
// returns the content of the file "/www/data/index.html" when requests are sent to "/index.html"
// and:
//
//    Files("/assets/*filepath", "/www/data/assets")
//
// returns the content of the file "/www/data/assets/x/y/z" when requests are sent to
// "/assets/x/y/z".
//
// Files may appear in Resource.
//
// Files accepts 2 arguments and an optional DSL. The first argument is the request path which may
// use a wildcard starting with *. The second argument is the path on disk to the files being
// served. The file path may be absolute or relative to the current path of the process.  The DSL
// allows setting a description and documentation.
//
// Example:
//
//    var _ = Resource("bottle", func() {
//        Files("/index.html", "/www/data/index.html", func() {
//            Description("Serve home page")
//            Docs(func() {
//                Description("Download docs")
//                URL("http//cellarapi.com/docs/actions/download")
//            })
//        })
//    })
//
func Files(path, filename string, dsls ...func()) {
	if r, ok := eval.Current().(*design.ResourceExpr); ok {
		server := &design.FileServerExpr{
			Parent:      r,
			RequestPath: path,
			FilePath:    filename,
		}
		if len(dsls) > 0 {
			if !eval.Execute(dsls[0], server) {
				return
			}
		}
		r.FileServers = append(r.FileServers, server)
	}
}
