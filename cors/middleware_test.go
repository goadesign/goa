package cors_test

import (
	"fmt"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/raphael/goa"
	"github.com/raphael/goa/cors"
	"gopkg.in/inconshreveable/log15.v2"
)

var _ = Describe("Middleware", func() {

	Context("with a running goa app", func() {
		var dsl func()
		var method string
		var path string
		var optionsHandler goa.Handler

		var service *goa.GracefulApplication
		var url string
		portIndex := 1

		JustBeforeEach(func() {
			goa.Log.SetHandler(log15.DiscardHandler())
			service = goa.NewGraceful("").(*goa.GracefulApplication)
			spec, err := cors.New(dsl)
			Ω(err).ShouldNot(HaveOccurred())
			service.Use(cors.Middleware(spec))
			router := service.HTTPHandler().(*httprouter.Router)
			h := func(ctx *goa.Context) error { return ctx.Respond(200, nil) }
			ctrl := service.NewController("test")
			router.Handle(method, path, ctrl.NewHTTPRouterHandle("", h))
			router.Handle("OPTIONS", path, ctrl.NewHTTPRouterHandle("", optionsHandler))
			cors.MountPreflightController(service, spec)
			portIndex++
			port := 54511 + portIndex
			url = fmt.Sprintf("http://localhost:%d", port)
			go service.ListenAndServe(fmt.Sprintf(":%d", port))
			// ugh - does anyone have a better idea? we need to wait for the server
			// to start listening or risk tests failing because sendind requests too
			// early.
			time.Sleep(time.Duration(100) * time.Millisecond)
		})

		AfterEach(func() {
			service.Shutdown()
		})

		Context("handling GET requests", func() {
			BeforeEach(func() {
				method = "GET"
				path = "/"
			})

			It("responds", func() {
				resp, err := http.Get(url)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(resp.StatusCode).Should(Equal(200))
			})

			Context("using CORS that allows the request", func() {
				BeforeEach(func() {
					dsl = func() {
						cors.Origin("http://authorized.com", func() {
							cors.Resource("/", func() {
								cors.Methods("GET")
							})
						})
					}
				})

				It("sets the Acess-Control-Allow-Methods header", func() {
					req, err := http.NewRequest("GET", url, nil)
					Ω(err).ShouldNot(HaveOccurred())
					req.Header.Set("Origin", "http://authorized.com")
					resp, err := http.DefaultClient.Do(req)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(resp.StatusCode).Should(Equal(200))
					Ω(resp.Header).Should(HaveKey("Access-Control-Allow-Methods"))
				})

			})

			Context("using CORS that disallows the request", func() {
				BeforeEach(func() {
					dsl = func() {
						cors.Origin("http://authorized.com", func() {
							cors.Resource("/", func() {
								cors.Methods("POST")
							})
						})
					}
				})

				It("does not set the Acess-Control-Allow-Methods header", func() {
					req, err := http.NewRequest("GET", url, nil)
					Ω(err).ShouldNot(HaveOccurred())
					req.Header.Set("Origin", "http://nonauthorized.com")
					resp, err := http.DefaultClient.Do(req)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(resp.StatusCode).Should(Equal(200))
					Ω(resp.Header).ShouldNot(HaveKey("Access-Control-Allow-Methods"))
				})

			})

			Context("using a CORS preflight request", func() {
				BeforeEach(func() {
					dsl = func() {
						cors.Origin("http://authorized.com", func() {
							cors.Resource("/", func() {
								cors.Methods("GET")
							})
						})
					}
				})

				It("sets the Acess-Control-Allow-Methods header when no OPTION action exists", func() {
					req, err := http.NewRequest("OPTIONS", url, nil)
					Ω(err).ShouldNot(HaveOccurred())
					req.Header.Set("Origin", "http://authorized.com")
					req.Header.Set("Access-Control-Request-Method", "GET")
					resp, err := http.DefaultClient.Do(req)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(resp.StatusCode).Should(Equal(200))
					Ω(resp.Header).Should(HaveKey("Access-Control-Allow-Methods"))
				})

				Context("with an OPTIONS action", func() {
					BeforeEach(func() {
						optionsHandler = func(ctx *goa.Context) error { return ctx.Respond(200, nil) }
					})

					It("sets the Acess-Control-Allow-Methods header when OPTION actions exist", func() {
						req, err := http.NewRequest("OPTIONS", url, nil)
						Ω(err).ShouldNot(HaveOccurred())
						req.Header.Set("Origin", "http://authorized.com")
						req.Header.Set("Access-Control-Request-Method", "GET")
						resp, err := http.DefaultClient.Do(req)
						Ω(err).ShouldNot(HaveOccurred())
						Ω(resp.StatusCode).Should(Equal(200))
						Ω(resp.Header).Should(HaveKey("Access-Control-Allow-Methods"))
					})
				})

			})

		})
	})
})
