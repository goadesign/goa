//************************************************************************//
// cellar: Application Controllers
//
// Generated with codegen v0.0.1, command line:
// $ /home/raphael/go/src/github.com/raphael/goa/examples/cellar/codegen425613848/codegen
// --out=/home/raphael/go/src/github.com/raphael/goa/examples/cellar
// --design=github.com/raphael/goa/examples/cellar/design
// --force
// --pkg=app
//
// The content of this file is auto-generated, DO NOT MODIFY
//************************************************************************//

package app

import (
	"github.com/raphael/goa"
)

type AccountController interface {
	Create(*CreateAccountContext) error
	Delete(*DeleteAccountContext) error
	Show(*ShowAccountContext) error
	Update(*UpdateAccountContext) error
}

// MountAccountController "mounts" a Account resource controller on the given application.
func MountAccountController(app *goa.Application, ctrl AccountController) {
	idx := 0
	var h goa.Handler
	logger := app.Logger.New("ctrl", "Account")
	logger.Info("mounting")

	h = func(c *goa.Context) error {
		ctx, err := NewCreateAccountContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.Create(ctx)
	}
	app.Router.Handle("POST", "/cellar/accounts", goa.NewHTTPRouterHandle(app, "Account", "Create", h))
	idx++
	logger.Info("handler", "action", "Create", "POST", "/cellar/accounts")

	h = func(c *goa.Context) error {
		ctx, err := NewDeleteAccountContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.Delete(ctx)
	}
	app.Router.Handle("DELETE", "/cellar/accounts/:accountID", goa.NewHTTPRouterHandle(app, "Account", "Delete", h))
	idx++
	logger.Info("handler", "action", "Delete", "DELETE", "/cellar/accounts/:accountID")

	h = func(c *goa.Context) error {
		ctx, err := NewShowAccountContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.Show(ctx)
	}
	app.Router.Handle("GET", "/cellar/accounts/:accountID", goa.NewHTTPRouterHandle(app, "Account", "Show", h))
	idx++
	logger.Info("handler", "action", "Show", "GET", "/cellar/accounts/:accountID")

	h = func(c *goa.Context) error {
		ctx, err := NewUpdateAccountContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.Update(ctx)
	}
	app.Router.Handle("PUT", "/cellar/accounts/:accountID", goa.NewHTTPRouterHandle(app, "Account", "Update", h))
	idx++
	logger.Info("handler", "action", "Update", "PUT", "/cellar/accounts/:accountID")

	logger.Info("mounted")
}

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
	idx := 0
	var h goa.Handler
	logger := app.Logger.New("ctrl", "Bottle")
	logger.Info("mounting")

	h = func(c *goa.Context) error {
		ctx, err := NewCreateBottleContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.Create(ctx)
	}
	app.Router.Handle("POST", "/cellar/accounts/:accountID/bottles", goa.NewHTTPRouterHandle(app, "Bottle", "Create", h))
	idx++
	logger.Info("handler", "action", "Create", "POST", "/cellar/accounts/:accountID/bottles")

	h = func(c *goa.Context) error {
		ctx, err := NewDeleteBottleContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.Delete(ctx)
	}
	app.Router.Handle("DELETE", "/cellar/accounts/:accountID/bottles/:id", goa.NewHTTPRouterHandle(app, "Bottle", "Delete", h))
	idx++
	logger.Info("handler", "action", "Delete", "DELETE", "/cellar/accounts/:accountID/bottles/:id")

	h = func(c *goa.Context) error {
		ctx, err := NewListBottleContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.List(ctx)
	}
	app.Router.Handle("GET", "/cellar/accounts/:accountID/bottles", goa.NewHTTPRouterHandle(app, "Bottle", "List", h))
	idx++
	logger.Info("handler", "action", "List", "GET", "/cellar/accounts/:accountID/bottles")

	h = func(c *goa.Context) error {
		ctx, err := NewRateBottleContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.Rate(ctx)
	}
	app.Router.Handle("PUT", "/cellar/accounts/:accountID/bottles/:id/actions/rate", goa.NewHTTPRouterHandle(app, "Bottle", "Rate", h))
	idx++
	logger.Info("handler", "action", "Rate", "PUT", "/cellar/accounts/:accountID/bottles/:id/actions/rate")

	h = func(c *goa.Context) error {
		ctx, err := NewShowBottleContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.Show(ctx)
	}
	app.Router.Handle("GET", "/cellar/accounts/:accountID/bottles/:id", goa.NewHTTPRouterHandle(app, "Bottle", "Show", h))
	idx++
	logger.Info("handler", "action", "Show", "GET", "/cellar/accounts/:accountID/bottles/:id")

	h = func(c *goa.Context) error {
		ctx, err := NewUpdateBottleContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.Update(ctx)
	}
	app.Router.Handle("PATCH", "/cellar/accounts/:accountID/bottles/:id", goa.NewHTTPRouterHandle(app, "Bottle", "Update", h))
	idx++
	logger.Info("handler", "action", "Update", "PATCH", "/cellar/accounts/:accountID/bottles/:id")

	logger.Info("mounted")
}
