// +build !appengine

package main

import "github.com/raphael/goa/examples/cellar/controllers"

func main() {
	controllers.Run(":8080")
}
