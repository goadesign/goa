//************************************************************************//
// cellar JSON Hyper-schema
//
// Generated with goagen v0.0.1, command line:
// $ goagen
// --out=$(GOPATH)/src/github.com/raphael/goa/examples/cellar
// --design=github.com/raphael/goa/examples/cellar/design
// --url=http://localhost
//
// The content of this file is auto-generated, DO NOT MODIFY
//************************************************************************//

package schema

import "github.com/raphael/goa"

// MountController mounts the API JSON schema controller under "/schema.json".
func MountController(service goa.Service) {
	service.ServeFiles("/schema.json", "schema/schema.json")
}
