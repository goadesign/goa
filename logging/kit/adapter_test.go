package goakit_test

import (
	"bytes"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/log"
	"github.com/goadesign/goa"
	"github.com/goadesign/goa/logging/kit"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("New", func() {
	var buf bytes.Buffer
	var logger log.Logger
	var adapter goa.LogAdapter

	BeforeEach(func() {
		logger = log.NewLogfmtLogger(&buf)
		adapter = goakit.New(logger)
	})

	It("creates an adapter that logs", func() {
		msg := "msg"
		adapter.Info(msg)
		Ω(buf.String()).Should(Equal("lvl=info msg=" + msg + "\n"))
	})
})

var _ = Describe("FromContext", func() {
	var buf bytes.Buffer
	var logctx *log.Context
	var adapter goa.LogAdapter

	BeforeEach(func() {
		logger := log.NewLogfmtLogger(&buf)
		logctx = log.NewContext(logger)
		adapter = goakit.FromContext(logctx)
	})

	It("creates an adapter that logs", func() {
		msg := "msg"
		adapter.Info(msg)
		Ω(buf.String()).Should(Equal("lvl=info msg=" + msg + "\n"))
	})

	Context("Context", func() {
		var ctx context.Context

		BeforeEach(func() {
			ctx = goa.WithLogger(context.Background(), adapter)
		})

		It("extracts the log context", func() {
			Ω(goakit.Context(ctx)).Should(Equal(logctx))
		})
	})
})
