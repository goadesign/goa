package gzip_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"context"

	"github.com/goadesign/goa"
	gzm "github.com/goadesign/goa/middleware/gzip"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type TestResponseWriter struct {
	ParentHeader http.Header
	Body         []byte
	Status       int
}

func (t *TestResponseWriter) Header() http.Header {
	return t.ParentHeader
}

func (t *TestResponseWriter) Write(b []byte) (int, error) {
	t.Body = append(t.Body, b...)
	return len(b), nil
}

func (t *TestResponseWriter) WriteHeader(s int) {
	t.Status = s
}

var _ = Describe("Gzip", func() {
	var ctx context.Context
	var req *http.Request
	var rw *TestResponseWriter
	payload := map[string]interface{}{"payload": 42}

	BeforeEach(func() {
		var err error
		req, err = http.NewRequest("POST", "/foo/bar", strings.NewReader(`{"payload":42}`))
		req.Header.Set("Accept-Encoding", "gzip")
		Ω(err).ShouldNot(HaveOccurred())
		rw = &TestResponseWriter{ParentHeader: make(http.Header)}

		ctx = goa.NewContext(nil, rw, req, nil)
		goa.ContextRequest(ctx).Payload = payload
	})

	It("encodes response using gzip", func() {
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			resp := goa.ContextResponse(ctx)
			resp.Write([]byte("gzip me!"))
			resp.WriteHeader(http.StatusOK)
			return nil
		}
		t := gzm.Middleware(gzip.BestCompression)(h)
		err := t(ctx, rw, req)
		Ω(err).ShouldNot(HaveOccurred())
		resp := goa.ContextResponse(ctx)
		Ω(resp.Status).Should(Equal(http.StatusOK))

		gzr, err := gzip.NewReader(bytes.NewReader(rw.Body))
		Ω(err).ShouldNot(HaveOccurred())
		buf := bytes.NewBuffer(nil)
		io.Copy(buf, gzr)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(buf.String()).Should(Equal("gzip me!"))
	})

})
