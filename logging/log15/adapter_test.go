package goalog15_test

import (
	"context"

	"github.com/goadesign/goa"
	"github.com/goadesign/goa/logging/log15"
	"github.com/inconshreveable/log15"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type TestHandler struct {
	records []*log15.Record
}

func (h *TestHandler) Log(r *log15.Record) error {
	h.records = append(h.records, r)
	return nil
}

var _ = Describe("New", func() {
	var logger log15.Logger
	var adapter goa.LogAdapter
	var handler *TestHandler

	BeforeEach(func() {
		logger = log15.New()
		handler = new(TestHandler)
		logger.SetHandler(handler)
		adapter = goalog15.New(logger)
	})

	It("creates an adapter that logs", func() {
		msg := "msg"
		adapter.Info(msg)
		Ω(handler.records).Should(HaveLen(1))
		Ω(handler.records[0].Msg).Should(ContainSubstring(msg))
	})

	Context("Logger", func() {
		var ctx context.Context

		BeforeEach(func() {
			ctx = goa.WithLogger(context.Background(), adapter)
		})

		It("extracts the logger", func() {
			Ω(goalog15.Logger(ctx)).Should(Equal(logger))
		})
	})
})
