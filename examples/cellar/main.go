// +build !appengine

package main

import (
	"github.com/raphael/goa"
	"github.com/raphael/goa/examples/cellar/app"
	"github.com/raphael/goa/examples/cellar/controllers"
	"github.com/raphael/goa/examples/cellar/js"
	"github.com/raphael/goa/examples/cellar/schema"
	"github.com/raphael/goa/examples/cellar/swagger"
)

func main() {
	// Create goa service
	service := goa.New("cellar")

	// Setup basic middleware
	service.Use(goa.RequestID())
	service.Use(goa.LogRequest())
	service.Use(goa.Recover())

	// Mount account controller onto service
	ac := controllers.NewAccount(service)
	app.MountAccountController(service, ac)

	// Mount bottle controller onto service
	bc := controllers.NewBottle(service)
	app.MountBottleController(service, bc)

	// Mount Swagger Spec controller onto service
	swagger.MountController(service)

	// Mount JSON Schema controller onto service
	schema.MountController(service)

	// Mount JavaScript example
	js.MountController(service)

	// Run service
	service.ListenAndServe(":8080")
}
