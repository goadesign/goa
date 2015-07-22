package main

import (
	"github.com/raphael/goa"
	log "gopkg.in/inconshreveable/log15.v2"
)

func main() {
	// Setup logger
	goa.Log.SetHandler(log.StdoutHandler)

	// Create "bottles" resource controller
	c := goa.NewController("bottles")

	// Register the resource action handlers
	c.SetHandlers(goa.Handlers{
		"list":   ListBottles,
		"show":   ShowBottle,
		"create": CreateBottle,
		"update": UpdateBottle,
		"delete": DeleteBottle,
		"rate":   RateBottle,
	})

	// Create goa application
	app := goa.New("cellar")

	// Mount controller onto application
	app.Mount(c)

	// Run application, listen on port 8080
	app.Run(":8080")
}
