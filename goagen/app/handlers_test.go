package app_test

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/raphael/goa/goagen/app"
)

var _ = Describe("HandlerWriter", func() {
	var writer *app.HandlersWriter
	var filename string
	var newErr error

	JustBeforeEach(func() {
		writer, newErr = app.NewHandlersWriter(filename)
	})

	Context("correctly configured", func() {
		var f *os.File
		BeforeEach(func() {
			f, _ = ioutil.TempFile("", "")
			filename = f.Name()
		})

		AfterEach(func() {
			os.Remove(filename)
		})

		It("NewHandlersWriter creates a writer", func() {
			Ω(newErr).ShouldNot(HaveOccurred())
		})

		Context("with data", func() {
			var actions, verbs, paths, names, contexts []string

			var data []*app.ActionHandlerTemplateData

			BeforeEach(func() {
				actions = nil
				verbs = nil
				paths = nil
				names = nil
				contexts = nil
			})

			JustBeforeEach(func() {
				data = make([]*app.ActionHandlerTemplateData, len(actions))
				for i := 0; i < len(actions); i++ {
					e := &app.ActionHandlerTemplateData{
						Resource: "bottles",
						Action:   actions[i],
						Verb:     verbs[i],
						Path:     paths[i],
						Name:     names[i],
						Context:  contexts[i],
					}
					data[i] = e
				}
			})

			Context("with missing data", func() {
				It("returns an empty string", func() {
					err := writer.Write(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).Should(BeEmpty())
				})
			})

			Context("with a simple handler", func() {
				BeforeEach(func() {
					actions = []string{"list"}
					verbs = []string{"GET"}
					paths = []string{"/accounts/:accountID/bottles"}
					names = []string{"listBottlesHandler"}
					contexts = []string{"ListBottleContext"}
				})

				It("writes the handlers code", func() {
					err := writer.Write(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).ShouldNot(BeEmpty())
					Ω(written).Should(ContainSubstring(simpleHandler))
					Ω(written).Should(ContainSubstring(simpleInit))
				})
			})

			Context("with multiple handlers", func() {
				BeforeEach(func() {
					actions = []string{"list", "show"}
					verbs = []string{"GET", "GET"}
					paths = []string{"/accounts/:accountID/bottles", "/accounts/:accountID/bottles/:id"}
					names = []string{"listBottlesHandler", "showBottlesHandler"}
					contexts = []string{"ListBottleContext", "ShowBottleContext"}
				})

				It("writes the handlers code", func() {
					err := writer.Write(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).ShouldNot(BeEmpty())
					Ω(written).Should(ContainSubstring(multiHandler1))
					Ω(written).Should(ContainSubstring(multiHandler2))
					Ω(written).Should(ContainSubstring(multiInit))
				})
			})
		})
	})
})

const (
	simpleHandler = `func listBottlesHandler(userHandler interface{}) (goa.Handler, error) {
	h, ok := userHandler.(func(c *ListBottleContext) error)
	if !ok {
		return nil, fmt.Errorf("invalid handler signature for action list bottles, expected 'func(c *ListBottleContext) error'")
	}
	return func(c goa.Context) error {
		ctx, err := NewListBottleContext(c)
		if err != nil {
			return err
		}
		return h(ctx)
	}, nil
}
`
	simpleInit = `func init() {
	goa.RegisterHandlers(
		&goa.HandlerFactory{"bottles", "list", "GET", "/accounts/:accountID/bottles", listBottlesHandler},
	)
}
`
	multiHandler1 = `func listBottlesHandler(userHandler interface{}) (goa.Handler, error) {
	h, ok := userHandler.(func(c *ListBottleContext) error)
	if !ok {
		return nil, fmt.Errorf("invalid handler signature for action list bottles, expected 'func(c *ListBottleContext) error'")
	}
	return func(c goa.Context) error {
		ctx, err := NewListBottleContext(c)
		if err != nil {
			return err
		}
		return h(ctx)
	}, nil
}
`
	multiHandler2 = `func showBottlesHandler(userHandler interface{}) (goa.Handler, error) {
	h, ok := userHandler.(func(c *ShowBottleContext) error)
	if !ok {
		return nil, fmt.Errorf("invalid handler signature for action show bottles, expected 'func(c *ShowBottleContext) error'")
	}
	return func(c goa.Context) error {
		ctx, err := NewShowBottleContext(c)
		if err != nil {
			return err
		}
		return h(ctx)
	}, nil
}
`
	multiInit = `func init() {
	goa.RegisterHandlers(
		&goa.HandlerFactory{"bottles", "list", "GET", "/accounts/:accountID/bottles", listBottlesHandler},
		&goa.HandlerFactory{"bottles", "show", "GET", "/accounts/:accountID/bottles/:id", showBottlesHandler},
	)
}
`
)
