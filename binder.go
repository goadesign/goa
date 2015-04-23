package goa

import (
	"net/http"

	"github.com/raphael/goa/design"
)

type BindFunc func(*http.Request, interface{}) error

func Binder(a *design.Action, s ParamSetterFunc) BindFunc {
}
