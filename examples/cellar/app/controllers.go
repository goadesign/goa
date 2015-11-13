//************************************************************************//
// cellar: Application Controllers
//
// Generated with goagen v0.0.1, command line:
// $ goagen
// --out=/home/raphael/go/src/github.com/raphael/goa/examples/cellar
// --design=github.com/raphael/goa/examples/cellar/design
// --pkg=app
//
// The content of this file is auto-generated, DO NOT MODIFY
//************************************************************************//

package app

import (
	"github.com/raphael/goa"
)

// AccountController is the controller interface for the Account actions.
type AccountController interface {
	Create(*CreateAccountContext) error
	Delete(*DeleteAccountContext) error
	Show(*ShowAccountContext) error
	Update(*UpdateAccountContext) error
}

// MountAccountController "mounts" a Account resource controller on the given application.
func MountAccountController(app *goa.Application, ctrl AccountController) {
	var h goa.Handler
	h = func(c *goa.Context) error {
		ctx, err := NewCreateAccountContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.Create(ctx)
	}
	app.Router.Handle("POST", "/cellar/accounts", app.NewHTTPRouterHandle("Account", "Create", h))
	app.Logger.Info("mount", "ctrl", "Account", "action", "Create", "route", "POST /cellar/accounts")
	h = func(c *goa.Context) error {
		ctx, err := NewDeleteAccountContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.Delete(ctx)
	}
	app.Router.Handle("DELETE", "/cellar/accounts/:accountID", app.NewHTTPRouterHandle("Account", "Delete", h))
	app.Logger.Info("mount", "ctrl", "Account", "action", "Delete", "route", "DELETE /cellar/accounts/:accountID")
	h = func(c *goa.Context) error {
		ctx, err := NewShowAccountContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.Show(ctx)
	}
	app.Router.Handle("GET", "/cellar/accounts/:accountID", app.NewHTTPRouterHandle("Account", "Show", h))
	app.Logger.Info("mount", "ctrl", "Account", "action", "Show", "route", "GET /cellar/accounts/:accountID")
	h = func(c *goa.Context) error {
		ctx, err := NewUpdateAccountContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.Update(ctx)
	}
	app.Router.Handle("PUT", "/cellar/accounts/:accountID", app.NewHTTPRouterHandle("Account", "Update", h))
	app.Logger.Info("mount", "ctrl", "Account", "action", "Update", "route", "PUT /cellar/accounts/:accountID")
}

// BottleController is the controller interface for the Bottle actions.
type BottleController interface {
	Create(*CreateBottleContext) error
	Delete(*DeleteBottleContext) error
	List(*ListBottleContext) error
	Rate(*RateBottleContext) error
	Show(*ShowBottleContext) error
	Update(*UpdateBottleContext) error
}

// MountBottleController "mounts" a Bottle resource controller on the given application.
func MountBottleController(app *goa.Application, ctrl BottleController) {
	var h goa.Handler
	h = func(c *goa.Context) error {
		ctx, err := NewCreateBottleContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.Create(ctx)
	}
	app.Router.Handle("POST", "/cellar/accounts/:accountID/bottles", app.NewHTTPRouterHandle("Bottle", "Create", h))
	app.Logger.Info("mount", "ctrl", "Bottle", "action", "Create", "route", "POST /cellar/accounts/:accountID/bottles")
	h = func(c *goa.Context) error {
		ctx, err := NewDeleteBottleContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.Delete(ctx)
	}
	app.Router.Handle("DELETE", "/cellar/accounts/:accountID/bottles/:bottleID", app.NewHTTPRouterHandle("Bottle", "Delete", h))
	app.Logger.Info("mount", "ctrl", "Bottle", "action", "Delete", "route", "DELETE /cellar/accounts/:accountID/bottles/:bottleID")
	h = func(c *goa.Context) error {
		ctx, err := NewListBottleContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.List(ctx)
	}
	app.Router.Handle("GET", "/cellar/accounts/:accountID/bottles", app.NewHTTPRouterHandle("Bottle", "List", h))
	app.Logger.Info("mount", "ctrl", "Bottle", "action", "List", "route", "GET /cellar/accounts/:accountID/bottles")
	h = func(c *goa.Context) error {
		ctx, err := NewRateBottleContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.Rate(ctx)
	}
	app.Router.Handle("PUT", "/cellar/accounts/:accountID/bottles/:bottleID/actions/rate", app.NewHTTPRouterHandle("Bottle", "Rate", h))
	app.Logger.Info("mount", "ctrl", "Bottle", "action", "Rate", "route", "PUT /cellar/accounts/:accountID/bottles/:bottleID/actions/rate")
	h = func(c *goa.Context) error {
		ctx, err := NewShowBottleContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.Show(ctx)
	}
	app.Router.Handle("GET", "/cellar/accounts/:accountID/bottles/:bottleID", app.NewHTTPRouterHandle("Bottle", "Show", h))
	app.Logger.Info("mount", "ctrl", "Bottle", "action", "Show", "route", "GET /cellar/accounts/:accountID/bottles/:bottleID")
	h = func(c *goa.Context) error {
		ctx, err := NewUpdateBottleContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.Update(ctx)
	}
	app.Router.Handle("PATCH", "/cellar/accounts/:accountID/bottles/:bottleID", app.NewHTTPRouterHandle("Bottle", "Update", h))
	app.Logger.Info("mount", "ctrl", "Bottle", "action", "Update", "route", "PATCH /cellar/accounts/:accountID/bottles/:bottleID")
}
