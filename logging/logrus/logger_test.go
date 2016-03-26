package goalogrus_test

import (
	"bytes"

	"github.com/Sirupsen/logrus"
	"github.com/goadesign/goa"
	"github.com/goadesign/goa/logging/logrus"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("goalogrus", func() {
	var logger *logrus.Logger
	var adapter goa.Logger
	const msg = "msg"
	var buffer *bytes.Buffer

	BeforeEach(func() {
		logger = logrus.New()
		buffer = new(bytes.Buffer)
		logger.Out = buffer
	})

	JustBeforeEach(func() {
		adapter = goalogrus.New(logger)
		adapter.Info(msg)
	})

	It("adapts info messages", func() {
		Ω(buffer.String()).Should(ContainSubstring(msg))
	})
})
