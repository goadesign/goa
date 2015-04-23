package main

import "github.com/raphael/goa/design"

func NewApplication() *design.Application {
	app := design.NewApplication("digital-ocean", "digital ocean plugin")

	droplets := app.NewResource("droplet", "Droplets are instances")
}
