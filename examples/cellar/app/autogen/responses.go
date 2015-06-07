package autogen

import "github.com/raphael/goa"

// Default bad request handler.
// Invoked when an incoming request parameters fail to validate.
// Returns status code 400 and error message in body.
var RespondBadRequest = func(err error, ctx *goa.Context) {
	ctx.RespondBadRequest(err.Error())
}

// Default bad response handler.
// Invoked when an invalid response is sent.
// Returns status code 500 and error message in body.
var RespondBadResponse = func(err error, ctx *goa.Context) {
	ctx.RespondBadResponse(err.Error())
}
