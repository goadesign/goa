// +build appengine

package main

import (
	"net/http"
	"os"

	"appengine"

	"github.com/raphael/goa"
	"github.com/raphael/goa/cors"
	"github.com/raphael/goa/examples/cellar/app"
	"github.com/raphael/goa/examples/cellar/controllers"
	"github.com/raphael/goa/examples/cellar/swagger"
	"gopkg.in/inconshreveable/log15.v2"
)

func init() {
	// Configure logging for appengine
	goa.Log.SetHandler(log15.MultiHandler(
		log15.StreamHandler(os.Stderr, log15.LogfmtFormat()),
		AppEngineLogHandler()),
	)

	// Create goa application
	service := goa.New("cellar")

	// Setup CORS to allow for swagger UI.
	spec, err := cors.New(func() {
		cors.Origin("*", func() {
			cors.Resource("*", func() {
				cors.Methods("GET", "POST", "PUT", "PATCH", "DELETE")
			})
		})
	})
	if err != nil {
		panic(err)
	}

	// Setup basic middleware
	service.Use(goa.RequestID())
	service.Use(AppEngineLogCtx())
	service.Use(cors.Middleware(spec))
	service.Use(goa.Recover())

	// Mount account controller onto application
	ac := controllers.NewAccount()
	app.MountAccountController(service, ac)

	// Mount bottle controller onto application
	bc := controllers.NewBottle()
	app.MountBottleController(service, bc)

	// Mount Swagger Spec controller onto application
	swagger.MountController(service)

	// Mount CORS preflight controllers
	cors.MountPreflightController(service, spec)

	// Setup HTTP handler
	http.HandleFunc("/", service.HTTPHandler().ServeHTTP)
}

// AppEngineLogHandler sends logs to AppEngine.
// The record must contain the appengine request context.
func AppEngineLogHandler() log15.Handler {
	logFormat := log15.JsonFormat()
	return log15.FuncHandler(func(r *log15.Record) error {
		var c appengine.Context
		index := 0
		for i, e := range r.Ctx {
			if ct, ok := e.(appengine.Context); ok {
				c = ct
				index = i
				break
			}
		}
		if c == nil {
			// not in the context of a request
			return nil
		}
		r.Ctx = append(r.Ctx[:index-1], r.Ctx[index+1:]...)
		log := string(logFormat.Format(r))
		switch r.Lvl {
		case log15.LvlCrit:
			c.Criticalf(log)
		case log15.LvlError:
			c.Errorf(log)
		case log15.LvlWarn:
			c.Warningf(log)
		case log15.LvlInfo:
			c.Infof(log)
		case log15.LvlDebug:
			c.Debugf(log)
		}
		return nil
	})
}

// AppEngineLogCtx returns a goa middleware that sets the appengine context in the log records.
func AppEngineLogCtx() goa.Middleware {
	return func(h goa.Handler) goa.Handler {
		return func(ctx *goa.Context) error {
			actx := appengine.NewContext(ctx.Request())
			ctx.SetValue(goa.ReqIDKey, appengine.RequestID(actx))
			ctx.Logger = ctx.Logger.New("aeCtx", actx)
			return h(ctx)
		}
	}
}
