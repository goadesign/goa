package controllers

import (
	"github.com/raphael/goa"
	"github.com/raphael/goa/examples/cellar/app"
	"github.com/raphael/goa/examples/cellar/schema"
	"github.com/raphael/goa/examples/cellar/swagger"
)

// Run sets up the goa application and starts it.
// This code is here so it can be shared between the standalone version of the app (main.go) and
// the appengine version (appengine.go).
func Run(host string) {
	// Create goa application
	api := goa.New("cellar")

	// Setup middleware
	api.Use(goa.Recover())
	api.Use(goa.RequestID())
	api.Use(goa.LogRequest())

	// Mount account controller onto application
	ac := NewAccount()
	app.MountAccountController(api, ac)

	// Mount bottle controller onto application
	bc := NewBottle()
	app.MountBottleController(api, bc)

	// Mount JSON Schema controller onto application
	schema.MountController(api)

	// Mount Swagger Spec controller onto application
	swagger.MountController(api)

	// Run application, listen on port 8080
	api.Run(host)
}
