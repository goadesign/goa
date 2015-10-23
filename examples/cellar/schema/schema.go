//************************************************************************//
// cellar JSON Hyper-schema
//
// Generated with goagen v0.0.1, command line:
// $ goagen
// --out=/home/raphael/go/src/github.com/raphael/goa/examples/cellar
// --design=github.com/raphael/goa/examples/cellar/design
// --url=http://localhost
//
// The content of this file is auto-generated, DO NOT MODIFY
//************************************************************************//

package schema

import (
	"io/ioutil"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/raphael/goa"
)

// Cached schema
var schema []byte

// MountController mounts the API JSON schema controller under "/schema".
func MountController(app *goa.Application) {
	logger := app.Logger.New("ctrl", "Schema")
	logger.Info("mounting")
	app.Router.GET("/schema", getSchema)
	logger.Info("handler", "action", "Show", "route", "GET /schema")
	logger.Info("mounted")
}

// getSchema is the httprouter handle that returns the API JSON schema.
func getSchema(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	if len(schema) == 0 {
		schema, _ = ioutil.ReadFile("/home/raphael/go/src/github.com/raphael/goa/examples/cellar/schema/schema.json")
	}
	w.Header().Set("Content-Type", "application/schema+json")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(200)
	w.Write(schema)
}
