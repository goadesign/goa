package goa

import (
	"net/http"
	"reflect"
)

type Context struct {
	Action        interface{} // Controller action method
	PathParams    map[string]interface{}
	QueryParams   map[string]interface{}
	PayloadParams interface{}
	R             *http.Request
	W             http.ResponseWriter
}

func (c *Context) Bind() {
}

func (c *Context) IntPathParam(name string) int {
	return 0
}

func (c *Context) Call(ctx reflect.Value) *Response {
	f := reflect.ValueOf(c.Action)
	r := f.Call([]reflect.Value{ctx})
	return r[0].Interface().(*Response)
}
