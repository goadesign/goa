// +build !appengine

package main

import (
	"github.com/raphael/goa"
	"github.com/raphael/goa/examples/cellar/app"
	"github.com/raphael/goa/examples/cellar/controllers"
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
	ac := controllers.NewAccount()
	app.MountAccountController(service, ac)

	// Mount bottle controller onto service
	bc := controllers.NewBottle()
	app.MountBottleController(service, bc)

	// Mount Swagger Spec controller onto service
	swagger.MountController(service)

	// Run service
	service.ListenAndServe(":8080")
}
