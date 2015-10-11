//************************************************************************//
// cellar: Application Controllers
//
// Generated with codegen v0.0.1, command line:
// $ /home/raphael/go/src/github.com/raphael/goa/examples/cellar/codegen485234072/codegen
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

	h = func(c goa.Context) error {
		ctx, err := NewCreateAccountContext(c)
		if err != nil {
			return err
		}
		return ctrl.Create(ctx)
	}
	app.Router.Handle("POST", "/:accountID/accounts", goa.NewHTTPRouterHandle(app, "Account", h))
	idx++
	logger.Info("handler", "action", idx, "POST", "/:accountID/accounts")

	h = func(c goa.Context) error {
		ctx, err := NewDeleteAccountContext(c)
		if err != nil {
			return err
		}
		return ctrl.Delete(ctx)
	}
	app.Router.Handle("DELETE", "/:accountID/accounts/:id", goa.NewHTTPRouterHandle(app, "Account", h))
	idx++
	logger.Info("handler", "action", idx, "DELETE", "/:accountID/accounts/:id")

	h = func(c goa.Context) error {
		ctx, err := NewShowAccountContext(c)
		if err != nil {
			return err
		}
		return ctrl.Show(ctx)
	}
	app.Router.Handle("GET", "/:accountID/accounts/:id", goa.NewHTTPRouterHandle(app, "Account", h))
	idx++
	logger.Info("handler", "action", idx, "GET", "/:accountID/accounts/:id")

	h = func(c goa.Context) error {
		ctx, err := NewUpdateAccountContext(c)
		if err != nil {
			return err
		}
		return ctrl.Update(ctx)
	}
	app.Router.Handle("PUT", "/:accountID/accounts/:id", goa.NewHTTPRouterHandle(app, "Account", h))
	idx++
	logger.Info("handler", "action", idx, "PUT", "/:accountID/accounts/:id")

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

	h = func(c goa.Context) error {
		ctx, err := NewCreateBottleContext(c)
		if err != nil {
			return err
		}
		return ctrl.Create(ctx)
	}
	app.Router.Handle("POST", "/:accountID/accounts/bottles", goa.NewHTTPRouterHandle(app, "Bottle", h))
	idx++
	logger.Info("handler", "action", idx, "POST", "/:accountID/accounts/bottles")

	h = func(c goa.Context) error {
		ctx, err := NewDeleteBottleContext(c)
		if err != nil {
			return err
		}
		return ctrl.Delete(ctx)
	}
	app.Router.Handle("DELETE", "/:accountID/accounts/bottles/:id", goa.NewHTTPRouterHandle(app, "Bottle", h))
	idx++
	logger.Info("handler", "action", idx, "DELETE", "/:accountID/accounts/bottles/:id")

	h = func(c goa.Context) error {
		ctx, err := NewListBottleContext(c)
		if err != nil {
			return err
		}
		return ctrl.List(ctx)
	}
	app.Router.Handle("GET", "/:accountID/accounts/bottles", goa.NewHTTPRouterHandle(app, "Bottle", h))
	idx++
	logger.Info("handler", "action", idx, "GET", "/:accountID/accounts/bottles")

	h = func(c goa.Context) error {
		ctx, err := NewRateBottleContext(c)
		if err != nil {
			return err
		}
		return ctrl.Rate(ctx)
	}
	app.Router.Handle("PUT", "/:accountID/accounts/bottles/:id/actions/rate", goa.NewHTTPRouterHandle(app, "Bottle", h))
	idx++
	logger.Info("handler", "action", idx, "PUT", "/:accountID/accounts/bottles/:id/actions/rate")

	h = func(c goa.Context) error {
		ctx, err := NewShowBottleContext(c)
		if err != nil {
			return err
		}
		return ctrl.Show(ctx)
	}
	app.Router.Handle("GET", "/:accountID/accounts/bottles/:id", goa.NewHTTPRouterHandle(app, "Bottle", h))
	idx++
	logger.Info("handler", "action", idx, "GET", "/:accountID/accounts/bottles/:id")

	h = func(c goa.Context) error {
		ctx, err := NewUpdateBottleContext(c)
		if err != nil {
			return err
		}
		return ctrl.Update(ctx)
	}
	app.Router.Handle("PATCH", "/:accountID/accounts/bottles/:id", goa.NewHTTPRouterHandle(app, "Bottle", h))
	idx++
	logger.Info("handler", "action", idx, "PATCH", "/:accountID/accounts/bottles/:id")

	logger.Info("mounted")
}
