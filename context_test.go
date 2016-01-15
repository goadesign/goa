package goa_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/raphael/goa"
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

	Describe("SetResponseWriter", func() {
		var rw http.ResponseWriter

		BeforeEach(func() {
			rw = &TestResponseWriter{Status: 42}
		})

		It("sets the response writer and returns the previous one", func() {
			rwo := ctx.SetResponseWriter(rw)
			Ω(rwo).Should(BeNil())
			rwo = ctx.SetResponseWriter(&TestResponseWriter{Status: 43})
			Ω(rwo).ShouldNot(BeNil())
			Ω(rwo).Should(BeAssignableToTypeOf(&TestResponseWriter{}))
			trw := rwo.(*TestResponseWriter)
			Ω(trw.Status).Should(Equal(42))
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
			p := ctx.Get("foo")
			Ω(p).Should(Equal(""))
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
		var app goa.Service
		const resName = "res"
		const actName = "act"
		var handler, unmarshaler goa.Handler
		const reqBody = `"body"`
		const respStatus = 200
		var respContent = []byte("response")
		var handleFunc goa.HandleFunc
		var rw http.ResponseWriter
		var request *http.Request
		var params url.Values

		BeforeEach(func() {
			app = goa.New(appName)
			handler = func(c *goa.Context) error {
				ctx = c
				c.RespondBytes(respStatus, respContent)
				return nil
			}
			unmarshaler = func(c *goa.Context) error {
				if req := c.Request(); req != nil {
					var payload interface{}
					err := c.Service().Decode(ctx, req.Body, &payload, req.Header.Get("Content-Type"))
					if err != nil {
						return err
					}
					c.SetPayload(payload)
				}
				return nil
			}
			var err error
			reader := strings.NewReader(reqBody)
			request, err = http.NewRequest("POST", "/foo?filters=one&filters=two&filters=three", reader)
			Ω(err).ShouldNot(HaveOccurred())
			rw = new(TestResponseWriter)
			params = url.Values{"id": []string{"42"}, "filters": []string{"one", "two", "three"}}
		})

		JustBeforeEach(func() {
			ctrl := app.NewController(resName)
			handleFunc = ctrl.HandleFunc(actName, handler, unmarshaler)
			handleFunc(rw, request, params)
		})

		Describe("RespondBytes", func() {
			It("sets the context fields", func() {
				Ω(ctx.Request()).Should(Equal(request))
				Ω(ctx.Header()).Should(Equal(rw.Header()))
				Ω(ctx.ResponseStatus()).Should(Equal(respStatus))
				Ω(ctx.ResponseLength()).Should(Equal(len(respContent)))
				p := ctx.Get("id")
				Ω(p).Should(Equal("42"))
				ps := ctx.GetMany("filters")
				Ω(ps).Should(Equal([]string{"one", "two", "three"}))
				var payload string
				err := json.Unmarshal([]byte(reqBody), &payload)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(ctx.Payload()).Should(Equal(payload))
			})
		})

		Context("Respond", func() {
			BeforeEach(func() {
				handler = func(c *goa.Context) error {
					ctx = c
					c.Respond(respStatus, string(respContent))
					return nil
				}
			})

			It("sets the context response fields with the JSON", func() {
				Ω(ctx.ResponseStatus()).Should(Equal(respStatus))
				Ω(ctx.ResponseLength()).Should(Equal(len(respContent) + 3)) // quotes and newline
			})
		})

		Context("BadRequest", func() {
			err := fmt.Errorf("boom")
			var badReq = &goa.BadRequestError{Actual: err}

			BeforeEach(func() {
				handler = func(c *goa.Context) error {
					ctx = c
					c.BadRequest(badReq)
					return nil
				}
			})

			It("responds with 400 and the error body", func() {
				Ω(ctx.ResponseStatus()).Should(Equal(400))
				tw := rw.(*TestResponseWriter)
				Ω(string(tw.Body)).Should(ContainSubstring(err.Error()))
			})
		})
	})

})
