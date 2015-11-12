package controllers

import (
	"github.com/raphael/goa"
	"github.com/raphael/goa/examples/cellar/app"
	"github.com/raphael/goa/examples/cellar/schema"
	"github.com/raphael/goa/examples/cellar/swagger"
)

// New creates the goa application.
// This code is here so it can be shared between the standalone version of the app (main.go) and
// the appengine version (appengine.go).
func New() *goa.Application {
	// Create goa application
	api := goa.New("cellar")

	// Setup basic middleware
	api.Use(goa.Recover())
	api.Use(goa.RequestID())

	return api
}

// Mount mounts the controllers onto the app.
// This is done as a separate step to allow clients to register middleware beforehand.
func Mount(api *goa.Application) {
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
}
