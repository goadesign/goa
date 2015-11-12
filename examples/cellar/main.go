// +build !appengine

package main

import (
	"github.com/raphael/goa"
	"github.com/raphael/goa/examples/cellar/controllers"
)

func main() {
	api := controllers.New()
	api.Use(goa.LogRequest())
	controllers.Mount(api)
	api.Run(":8080")
}
