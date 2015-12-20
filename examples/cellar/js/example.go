//************************************************************************//
// cellar JavaScript Client Example
//
// Generated with goagen v0.0.1, command line:
// $ goagen
// --out=$(GOPATH)/src/github.com/raphael/goa/examples/cellar
// --design=github.com/raphael/goa/examples/cellar/design
//
// The content of this file is auto-generated, DO NOT MODIFY
//************************************************************************//

package js

import "github.com/raphael/goa"

// MountController mounts the JavaScript example controller under "/js".
func MountController(service goa.Service) {
	// Serve static files under js
	service.ServeFiles("/js/*filepath", "/home/raphael/go/src/github.com/raphael/goa/examples/cellar/js")
	service.Info("mount", "ctrl", "JS", "action", "ServeFiles", "route", "GET /js/*")
}
