// +build !appengine

package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
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
	ac := controllers.NewAccount(service)
	app.MountAccountController(service, ac)

	// Mount bottle controller onto service
	bc := controllers.NewBottle(service)
	app.MountBottleController(service, bc)

	// Mount Swagger Spec controller onto service
	swagger.MountController(service)

	// Serve static files under js
	service.HTTPHandler().(*httprouter.Router).ServeFiles("/index/*filepath", http.Dir("/home/raphael/go/src/github.com/raphael/goa/examples/cellar/js"))

	// Run service
	service.ListenAndServe(":8080")
}
