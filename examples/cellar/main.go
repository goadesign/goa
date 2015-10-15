package main

import (
	"github.com/raphael/goa"
	"github.com/raphael/goa/examples/cellar/app"
)

func main() {
	// Create goa application
	api := goa.New("cellar")

	// Setup middleware
	api.Use(goa.Recover())
	api.Use(goa.RequestID())
	api.Use(goa.LogRequest())

	// Mount account controller onto application
	ac := NewAccountController()
	app.MountAccountController(api, ac)

	// Mount bottle controller onto application
	bc := NewBottleController()
	app.MountBottleController(api, bc)

	// Run application, listen on port 8080
	api.Run(":8080")
}
