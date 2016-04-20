package goakit_test

import (
	"bytes"

	"github.com/go-kit/kit/log"
	"github.com/goadesign/goa"
	"github.com/goadesign/goa/logging/kit"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("goakit", func() {
	var buf bytes.Buffer
	var logger log.Logger
	var adapter goa.LogAdapter
	const msg = "msg"

	BeforeEach(func() {
		logger = log.NewLogfmtLogger(&buf)
	})

	JustBeforeEach(func() {
		adapter = goakit.New(logger)
		adapter.Info(msg)
	})

	It("adapts info messages", func() {
		Ω(buf.String()).Should(Equal("lvl=info msg=" + msg + "\n"))
	})
})
