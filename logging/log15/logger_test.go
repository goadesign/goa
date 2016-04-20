package goalog15_test

import (
	"github.com/goadesign/goa"
	"github.com/goadesign/goa/logging/log15"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/inconshreveable/log15.v2"
)

type TestHandler struct {
	records []*log15.Record
}

func (h *TestHandler) Log(r *log15.Record) error {
	h.records = append(h.records, r)
	return nil
}

var _ = Describe("goalog15", func() {
	var logger log15.Logger
	var adapter goa.LogAdapter
	var handler *TestHandler
	const msg = "msg"

	BeforeEach(func() {
		logger = log15.New()
		handler = new(TestHandler)
		logger.SetHandler(handler)
	})

	JustBeforeEach(func() {
		adapter = goalog15.New(logger)
		adapter.Info(msg)
	})

	It("adapts info messages", func() {
		Ω(handler.records).Should(HaveLen(1))
		Ω(handler.records[0].Msg).Should(ContainSubstring(msg))
	})
})
