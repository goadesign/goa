package goa_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/raphael/goa"
	"github.com/raphael/goa/support"
	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"
)

var _ = Describe("Context", func() {
	var logger log15.Logger
	var ctx *goa.Context

	BeforeEach(func() {
		gctx := context.Background()
		ctx = &goa.Context{Context: gctx, Logger: logger}
	})

	Describe("SetValue", func() {
		key := "answer"
		val := 42

		BeforeEach(func() {
			ctx.SetValue(key, val)
		})

		It("sets the value in the context.Context", func() {
			v := ctx.Value(key)
			Ω(v).Should(Equal(val))
		})
	})

	Describe("Request", func() {
		It("returns nil if not initialized", func() {
			Ω(ctx.Request()).Should(BeNil())
		})
	})

	Describe("Header", func() {
		It("returns nil if not initialized", func() {
			Ω(ctx.Header()).Should(BeNil())
		})
	})

	Describe("ResponseWritten", func() {
		It("returns false if not initialized", func() {
			Ω(ctx.ResponseWritten()).Should(BeFalse())
		})
	})

	Describe("ResponseStatus", func() {
		It("returns 0 if not initialized", func() {
			Ω(ctx.ResponseStatus()).Should(Equal(0))
		})
	})

	Describe("ResponseLength", func() {
		It("returns 0 if not initialized", func() {
			Ω(ctx.ResponseLength()).Should(Equal(0))
		})
	})

	Describe("Get", func() {
		It(`returns "", false if not initialized`, func() {
			p, ok := ctx.Get("foo")
			Ω(p).Should(Equal(""))
			Ω(ok).Should(BeFalse())
		})
	})

	Describe("GetMany", func() {
		It("returns nil if not initialized", func() {
			Ω(ctx.GetMany("foo")).Should(BeNil())
		})
	})

	Describe("Payload", func() {
		It("returns nil if not initialized", func() {
			Ω(ctx.Payload()).Should(BeNil())
		})
	})

	Context("with a request response", func() {
		const appName = "foo"
		var app *goa.Application
		const resName = "res"
		const actName = "act"
		var handler goa.Handler
		const reqBody = `"body"`
		const respStatus = 200
		var respContent = []byte("response")
		var httpHandle httprouter.Handle
		var rw http.ResponseWriter
		var request *http.Request
		var params httprouter.Params

		BeforeEach(func() {
			app = goa.New(appName)
			handler = func(c *goa.Context) error {
				ctx = c
				c.Respond(respStatus, respContent)
				return nil
			}
			var err error
			reader := strings.NewReader(reqBody)
			request, err = http.NewRequest("POST", "/foo?filters=one&filters=two&filters=three", reader)
			Ω(err).ShouldNot(HaveOccurred())
			rw = new(TestResponseWriter)
			id := httprouter.Param{Key: "id", Value: "42"}
			params = httprouter.Params{id}
		})

		JustBeforeEach(func() {
			httpHandle = app.NewHTTPRouterHandle(resName, actName, handler)
			httpHandle(rw, request, params)
		})

		Describe("Respond", func() {
			It("sets the context fields", func() {
				Ω(ctx.Request()).Should(Equal(request))
				Ω(ctx.Header()).Should(Equal(rw.Header()))
				Ω(ctx.ResponseWritten()).Should(BeTrue())
				Ω(ctx.ResponseStatus()).Should(Equal(respStatus))
				Ω(ctx.ResponseLength()).Should(Equal(len(respContent)))
				p, ok := ctx.Get("id")
				Ω(ok).Should(BeTrue())
				Ω(p).Should(Equal("42"))
				ps := ctx.GetMany("filters")
				Ω(ps).Should(Equal([]string{"one", "two", "three"}))
				var payload string
				err := json.Unmarshal([]byte(reqBody), &payload)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(ctx.Payload()).Should(Equal(payload))
			})
		})

		Context("JSON", func() {
			BeforeEach(func() {
				handler = func(c *goa.Context) error {
					ctx = c
					c.JSON(respStatus, string(respContent))
					return nil
				}
			})

			It("sets the context response fields with the JSON", func() {
				Ω(ctx.ResponseWritten()).Should(BeTrue())
				Ω(ctx.ResponseStatus()).Should(Equal(respStatus))
				Ω(ctx.ResponseLength()).Should(Equal(len(respContent) + 2)) // quotes
			})
		})

		Context("BadRequest", func() {
			err := fmt.Errorf("boom")
			var badReq = &support.BadRequestError{Actual: err}

			BeforeEach(func() {
				handler = func(c *goa.Context) error {
					ctx = c
					c.BadRequest(badReq)
					return nil
				}
			})

			It("responds with 400 and the error body", func() {
				Ω(ctx.ResponseWritten()).Should(BeTrue())
				Ω(ctx.ResponseStatus()).Should(Equal(400))
				tw := rw.(*TestResponseWriter)
				Ω(string(tw.Body)).Should(ContainSubstring(err.Error()))
			})
		})
	})

})
