package goakit_test

import (
	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/go-kit/kit/log"
	"github.com/kyokomi/goa-v1"
	goakit "github.com/kyokomi/goa-v1/logging/kit"
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
