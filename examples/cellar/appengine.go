// +build appengine

package cellar

import (
	"fmt"

	"github.com/raphael/goa"
	"github.com/raphael/goa/examples/cellar/controllers"
	"gopkg.in/inconshreveable/log15.v2"
)

func init() {
	goa.Log.SetHandler(log15.DiscardHandler())
	fmt.Println("7000")
	controllers.Run(":7000")
}
