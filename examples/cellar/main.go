package main

import "github.com/raphael/goa"

func main() {
	// Create "bottle" resource controller
	c := goa.NewController("bottle")

	// Register the resource action handlers
	c.SetHandlers(goa.Actions{
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
