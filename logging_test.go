package goa_test

import (
	"bytes"
	"log"

	"github.com/goadesign/goa"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/net/context"
)

type LogEntry struct {
	ctx  context.Context
	msg  string
	data []goa.KV
}

type TestLog struct {
	infoEntries  []*LogEntry
	errorEntries []*LogEntry
}

func (l *TestLog) Info(ctx context.Context, msg string, data ...goa.KV) {
	l.infoEntries = append(l.infoEntries, &LogEntry{ctx, msg, data})
}

func (l *TestLog) Error(ctx context.Context, msg string, data ...goa.KV) {
	l.errorEntries = append(l.errorEntries, &LogEntry{ctx, msg, data})
}

var _ = Describe("Info and Error", func() {
	Context("with a nil Log", func() {
		BeforeEach(func() {
			goa.Log = nil
		})

		It("Info doesn't log and doesn't crash", func() {
			Ω(func() { goa.Info(nil, "foo") }).ShouldNot(Panic())
		})

		It("Error doesn't log and doesn't crash", func() {
			Ω(func() { goa.Error(nil, "foo") }).ShouldNot(Panic())
		})
	})

	Context("with a valid Log", func() {
		var testLog *TestLog
		var ctx context.Context
		ctxData := []goa.KV{{"ctxData", true}, {"other", 42}}
		data := []goa.KV{{"data", "foo"}}
		const msg = "message"

		BeforeEach(func() {
			testLog = new(TestLog)
			goa.Log = testLog
			ctx = goa.NewLogContext(nil, ctxData...)
		})

		It("Info collects the context data", func() {
			goa.Info(ctx, msg, data...)
			Ω(testLog.infoEntries).Should(HaveLen(1))
			Ω(testLog.infoEntries[0].ctx).Should(Equal(ctx))
			Ω(testLog.infoEntries[0].msg).Should(Equal(msg))
			Ω(testLog.infoEntries[0].data).Should(Equal(append(ctxData, data...)))
		})

		It("Error collects the context data", func() {
			goa.Error(ctx, msg, data...)
			Ω(testLog.errorEntries).Should(HaveLen(1))
			Ω(testLog.errorEntries[0].ctx).Should(Equal(ctx))
			Ω(testLog.errorEntries[0].msg).Should(Equal(msg))
			Ω(testLog.errorEntries[0].data).Should(Equal(append(ctxData, data...)))
		})
	})
})

var _ = Describe("DefaultLogger", func() {
	var logger *goa.DefaultLogger

	Context("logging to a buffer", func() {
		var buffer *bytes.Buffer
		const key = "key"
		const value = "value"
		const msg = "msg"
		ctx := goa.NewLogContext(nil, goa.KV{key, value})

		BeforeEach(func() {
			buffer = new(bytes.Buffer)
			logger = &goa.DefaultLogger{Logger: log.New(buffer, "", log.LstdFlags)}
			goa.Log = logger
			goa.Info(ctx, msg)
		})

		It("logs all the context", func() {
			Ω(buffer.String()).Should(ContainSubstring(key))
			Ω(buffer.String()).Should(ContainSubstring(value))
			Ω(buffer.String()).Should(ContainSubstring(msg))
		})
	})
})
