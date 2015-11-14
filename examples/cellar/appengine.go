// +build appengine

package cellar

import (
	"net/http"
	"os"
	"regexp"

	"appengine"

	"github.com/raphael/goa"
	"github.com/raphael/goa/examples/cellar/controllers"
	"gopkg.in/inconshreveable/log15.v2"
)

func init() {
	goa.Log.SetHandler(log15.MultiHandler(
		log15.StreamHandler(os.Stderr, log15.LogfmtFormat()),
		AppEngineLogHandler()),
	)
	service := controllers.New()
	service.Use(AppEngineLogCtx())
	service.Use(goa.CORS(corsPath, "*", "", "", "", "GET, POST, PUT, PATCH, DELETE", ""))
	controllers.Mount(service)
	http.HandleFunc("/", service.HTTPHandler().ServeHTTP)
}

// Format used for logging to AppEngine
var logFormat = log15.JsonFormat()

// Paths that must return CORS headers
var corsPath = regexp.MustCompile(`.*`)

// AppEngineLogHandler sends logs to AppEngine.
// The record must contain the appengine request context.
func AppEngineLogHandler() log15.Handler {
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
