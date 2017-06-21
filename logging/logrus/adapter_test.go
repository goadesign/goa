package goalogrus_test

import (
	"bytes"

	"context"

	"github.com/goadesign/goa"
	"github.com/goadesign/goa/logging/logrus"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

var _ = Describe("goalogrus", func() {
	var logger *logrus.Logger
	var adapter goa.LogAdapter
	var buf bytes.Buffer

	BeforeEach(func() {
		logger = logrus.New()
		logger.Out = &buf
		adapter = goalogrus.New(logger)
	})

	It("adapts info messages", func() {
		msg := "msg"
		adapter.Info(msg)
		Ω(buf.String()).Should(ContainSubstring(msg))
	})
})

var _ = Describe("FromEntry", func() {
	var entry *logrus.Entry
	var adapter goa.LogAdapter
	var buf bytes.Buffer

	BeforeEach(func() {
		logger := logrus.New()
		logger.Out = &buf
		entry = logrus.NewEntry(logger)
		adapter = goalogrus.FromEntry(entry)
	})

	It("creates an adapter that logs", func() {
		msg := "msg"
		adapter.Info(msg)
		Ω(buf.String()).Should(ContainSubstring(msg))
	})

	Context("Entry", func() {
		var ctx context.Context

		BeforeEach(func() {
			ctx = goa.WithLogger(context.Background(), adapter)
		})

		It("extracts the log entry", func() {
			Ω(goalogrus.Entry(ctx)).Should(Equal(entry))
		})
	})
})
