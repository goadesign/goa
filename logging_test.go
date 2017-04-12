package goa_test

import (
	"bytes"
	"log"

	"context"

	"github.com/goadesign/goa"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Info", func() {
	Context("with a nil Log", func() {
		It("doesn't log and doesn't crash", func() {
			Ω(func() { goa.LogInfo(context.Background(), "foo", "bar") }).ShouldNot(Panic())
		})
	})
})

var _ = Describe("Error", func() {
	Context("with a nil Log", func() {
		It("doesn't log and doesn't crash", func() {
			Ω(func() { goa.LogError(context.Background(), "foo", "bar") }).ShouldNot(Panic())
		})
	})
})

var _ = Describe("LogAdapter", func() {
	Context("with a valid Log", func() {
		var logger goa.LogAdapter
		const msg = "message"
		data := []interface{}{"data", "foo"}

		var out bytes.Buffer

		BeforeEach(func() {
			stdlogger := log.New(&out, "", log.LstdFlags)
			logger = goa.NewLogger(stdlogger)
		})

		It("Info logs", func() {
			logger.Info(msg, data...)
			Ω(out.String()).Should(ContainSubstring(msg + " data=foo"))
		})

		It("Error logs", func() {
			logger.Error(msg, data...)
			Ω(out.String()).Should(ContainSubstring(msg + " data=foo"))
		})
	})
})
