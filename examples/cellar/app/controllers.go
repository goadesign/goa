//************************************************************************//
// cellar: Application Controllers
//
// Generated with goagen v0.0.1, command line:
// $ goagen
// --out=$(GOPATH)/src/github.com/raphael/goa/examples/cellar
// --design=github.com/raphael/goa/examples/cellar/design
// --pkg=app
//
// The content of this file is auto-generated, DO NOT MODIFY
//************************************************************************//

package app

import (
	"github.com/julienschmidt/httprouter"
	"github.com/raphael/goa"
)

// AccountController is the controller interface for the Account actions.
type AccountController interface {
	goa.Controller
	Create(*CreateAccountContext) error
	Delete(*DeleteAccountContext) error
	Show(*ShowAccountContext) error
	Update(*UpdateAccountContext) error
}

// MountAccountController "mounts" a Account resource controller on the given service.
func MountAccountController(service goa.Service, ctrl AccountController) {
	router := service.HTTPHandler().(*httprouter.Router)
	var h goa.Handler
	h = func(c *goa.Context) error {
		ctx, err := NewCreateAccountContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.Create(ctx)
	}
	router.Handle("POST", "/cellar/accounts", ctrl.NewHTTPRouterHandle("Create", h))
	service.Info("mount", "ctrl", "Account", "action", "Create", "route", "POST /cellar/accounts")
	h = func(c *goa.Context) error {
		ctx, err := NewDeleteAccountContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.Delete(ctx)
	}
	router.Handle("DELETE", "/cellar/accounts/:accountID", ctrl.NewHTTPRouterHandle("Delete", h))
	service.Info("mount", "ctrl", "Account", "action", "Delete", "route", "DELETE /cellar/accounts/:accountID")
	h = func(c *goa.Context) error {
		ctx, err := NewShowAccountContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.Show(ctx)
	}
	router.Handle("GET", "/cellar/accounts/:accountID", ctrl.NewHTTPRouterHandle("Show", h))
	service.Info("mount", "ctrl", "Account", "action", "Show", "route", "GET /cellar/accounts/:accountID")
	h = func(c *goa.Context) error {
		ctx, err := NewUpdateAccountContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.Update(ctx)
	}
	router.Handle("PUT", "/cellar/accounts/:accountID", ctrl.NewHTTPRouterHandle("Update", h))
	service.Info("mount", "ctrl", "Account", "action", "Update", "route", "PUT /cellar/accounts/:accountID")
}

// BottleController is the controller interface for the Bottle actions.
type BottleController interface {
	goa.Controller
	Create(*CreateBottleContext) error
	Delete(*DeleteBottleContext) error
	List(*ListBottleContext) error
	Rate(*RateBottleContext) error
	Show(*ShowBottleContext) error
	Update(*UpdateBottleContext) error
}

// MountBottleController "mounts" a Bottle resource controller on the given service.
func MountBottleController(service goa.Service, ctrl BottleController) {
	router := service.HTTPHandler().(*httprouter.Router)
	var h goa.Handler
	h = func(c *goa.Context) error {
		ctx, err := NewCreateBottleContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.Create(ctx)
	}
	router.Handle("POST", "/cellar/accounts/:accountID/bottles", ctrl.NewHTTPRouterHandle("Create", h))
	service.Info("mount", "ctrl", "Bottle", "action", "Create", "route", "POST /cellar/accounts/:accountID/bottles")
	h = func(c *goa.Context) error {
		ctx, err := NewDeleteBottleContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.Delete(ctx)
	}
	router.Handle("DELETE", "/cellar/accounts/:accountID/bottles/:bottleID", ctrl.NewHTTPRouterHandle("Delete", h))
	service.Info("mount", "ctrl", "Bottle", "action", "Delete", "route", "DELETE /cellar/accounts/:accountID/bottles/:bottleID")
	h = func(c *goa.Context) error {
		ctx, err := NewListBottleContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.List(ctx)
	}
	router.Handle("GET", "/cellar/accounts/:accountID/bottles", ctrl.NewHTTPRouterHandle("List", h))
	service.Info("mount", "ctrl", "Bottle", "action", "List", "route", "GET /cellar/accounts/:accountID/bottles")
	h = func(c *goa.Context) error {
		ctx, err := NewRateBottleContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.Rate(ctx)
	}
	router.Handle("PUT", "/cellar/accounts/:accountID/bottles/:bottleID/actions/rate", ctrl.NewHTTPRouterHandle("Rate", h))
	service.Info("mount", "ctrl", "Bottle", "action", "Rate", "route", "PUT /cellar/accounts/:accountID/bottles/:bottleID/actions/rate")
	h = func(c *goa.Context) error {
		ctx, err := NewShowBottleContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.Show(ctx)
	}
	router.Handle("GET", "/cellar/accounts/:accountID/bottles/:bottleID", ctrl.NewHTTPRouterHandle("Show", h))
	service.Info("mount", "ctrl", "Bottle", "action", "Show", "route", "GET /cellar/accounts/:accountID/bottles/:bottleID")
	h = func(c *goa.Context) error {
		ctx, err := NewUpdateBottleContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.Update(ctx)
	}
	router.Handle("PATCH", "/cellar/accounts/:accountID/bottles/:bottleID", ctrl.NewHTTPRouterHandle("Update", h))
	service.Info("mount", "ctrl", "Bottle", "action", "Update", "route", "PATCH /cellar/accounts/:accountID/bottles/:bottleID")
}
