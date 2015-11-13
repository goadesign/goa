//************************************************************************//
// cellar Swagger Spec
//
// Generated with goagen v0.0.1, command line:
// $ goagen
// --out=/home/raphael/go/src/github.com/raphael/goa/examples/cellar
// --design=github.com/raphael/goa/examples/cellar/design
//
// The content of this file is auto-generated, DO NOT MODIFY
//************************************************************************//

package swagger

import (
	"github.com/julienschmidt/httprouter"
	"github.com/raphael/goa"
)

// MountController mounts the swagger spec controller under "/swagger.json".
func MountController(service goa.Service) {
	service.Info("mount", "ctrl", "Swagger", "action", "Show", "route", "GET /swagger.json")
	h := goa.NewHTTPRouterHandle(service, "Swagger", "Show", getSwagger)
	service.HTTPHandler().(*httprouter.Router).Handle("GET", "/swagger.json", h)
}

// getSwagger is the httprouter handle that returns the Swagger spec.
// func getSwagger(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
func getSwagger(ctx *goa.Context) error {
	ctx.Header().Set("Content-Type", "application/swagger+json")
	ctx.Header().Set("Cache-Control", "public, max-age=3600")
	return ctx.Respond(200, []byte(spec))
}

// Generated spec
const spec = `{"swagger":"2.0","info":{"title":"The virtual wine cellar","description":"A basic example of a CRUD API implemented with goa","version":""},"basePath":"/cellar","schemes":["https"],"consumes":["application/json"],"produces":["application/json"],"paths":{"":{"get":{"description":"List all bottles in account optionally filtering by year","operationId":"bottle#list","consumes":["application/json"],"produces":["application/json"],"parameters":[{"name":"accountID","in":"query","required":false,"type":"string"},{"name":"years","in":"query","description":"Filter by years","required":false,"type":"array","items":{"type":"array","items":{"type":"integer"}}}],"responses":{"200":{"description":"","schema":{"$ref":"#/definitions/BottleCollection"}}},"schemes":["https"]}},"/{accountID}":{"put":{"description":"Change account name","operationId":"account#update","consumes":["application/json"],"produces":["application/json"],"parameters":[{"name":"accountID","in":"path","description":"Account ID","required":true,"type":"integer"}],"responses":{"204":{"description":""},"404":{"description":""}},"schemes":["https"]}},"/{bottleID}":{"patch":{"operationId":"bottle#update","consumes":["application/json"],"produces":["application/json"],"parameters":[{"name":"accountID","in":"query","required":false,"type":"string"},{"name":"bottleID","in":"path","required":true,"type":"integer"}],"responses":{"204":{"description":""},"404":{"description":""}},"schemes":["https"]}},"/{bottleID}/actions/rate":{"put":{"operationId":"bottle#rate","consumes":["application/json"],"produces":["application/json"],"parameters":[{"name":"accountID","in":"query","required":false,"type":"string"},{"name":"bottleID","in":"path","required":true,"type":"integer"}],"responses":{"204":{"description":""},"404":{"description":""}},"schemes":["https"]}}},"definitions":{"Account":{"title":"Mediatype identifier: application/vnd.goa.example.account","type":"object","properties":{"created_at":{"type":"string","description":"Date of creation","format":"date-time"},"created_by":{"type":"string","description":"Email of account ownder","format":"email"},"href":{"type":"string","description":"API href of account"},"id":{"type":"integer","description":"ID of account"},"name":{"type":"string","description":"Name of account"}},"description":"A tenant account","required":["name"]},"Bottle":{"title":"Mediatype identifier: application/vnd.goa.example.bottle","type":"object","properties":{"account":{"description":"Account that owns bottle","$ref":"#/definitions/Account"},"characteristics":{"type":"string","minLength":10,"maxLength":300},"color":{"type":"string","enum":["red","white","rose","yellow","sparkling"]},"country":{"type":"string","minLength":2},"created_at":{"type":"string","description":"Date of creation","format":"date-time"},"href":{"type":"string","description":"API href of bottle"},"id":{"type":"integer","description":"ID of bottle"},"name":{"type":"string","minLength":2},"rating":{"type":"integer","description":"Rating of bottle between 1 and 5","minimum":1,"maximum":5},"region":{"type":"string"},"review":{"type":"string","minLength":10,"maxLength":300},"sweetness":{"type":"integer","minimum":1,"maximum":5},"updated_at":{"type":"string","description":"Date of last update","format":"date-time"},"varietal":{"type":"string","minLength":4},"vineyard":{"type":"string","minLength":2},"vintage":{"type":"integer","minimum":1900,"maximum":2020}},"description":"A bottle of wine","required":["account","name","vineyard"]},"BottleCollection":{"title":"Mediatype identifier: application/vnd.goa.example.bottle; type=collection","type":"array","item":{"$ref":"#/definitions/Bottle"}}},"responses":{"Created":{"description":"Resource created","headers":{"Location":{"description":"href to created resource","type":"string"}}}}} `
