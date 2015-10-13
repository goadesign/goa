package main

import (
	"github.com/raphael/goa"
	"github.com/raphael/goa/examples/cellar/app"
	log "gopkg.in/inconshreveable/log15.v2"
)

func main() {
	// Setup logger
	goa.Log.SetHandler(log.StdoutHandler)

	// Create goa application
	api := goa.New("cellar")

	// Mount account controller onto application
	ac := NewAccountController()
	app.MountAccountController(api, ac)

	// Mount bottle controller onto application
	bc := NewBottleController()
	app.MountBottleController(api, bc)

	// Run application, listen on port 8080
	api.Run(":8080")
}
