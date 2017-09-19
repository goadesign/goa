package gzip_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"

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
		req.Header.Set("Range", "bytes=0-1023")
		Ω(err).ShouldNot(HaveOccurred())
		rw = &TestResponseWriter{ParentHeader: make(http.Header)}

		ctx = goa.NewContext(nil, rw, req, nil)
		goa.ContextRequest(ctx).Payload = payload
	})

	It("encodes response using gzip", func() {
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			resp := goa.ContextResponse(ctx)
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte("gzip me!"))
			return nil
		}
		t := gzm.Middleware(gzip.BestCompression, gzm.MinSize(0))(h)
		err := t(ctx, rw, req)
		Ω(err).ShouldNot(HaveOccurred())
		resp := goa.ContextResponse(ctx)
		Ω(resp.Status).Should(Equal(http.StatusOK))
		Ω(resp.Header().Get("Content-Encoding")).Should(Equal("gzip"))

		gzr, err := gzip.NewReader(bytes.NewReader(rw.Body))
		Ω(err).ShouldNot(HaveOccurred())
		var buf bytes.Buffer
		_, err = io.Copy(&buf, gzr)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(buf.String()).Should(Equal("gzip me!"))
		Ω(resp.Header().Get("Content-Length")).Should(Equal(""))
	})

	It("encodes response using gzip (custom status)", func() {
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			resp := goa.ContextResponse(ctx)
			resp.WriteHeader(http.StatusBadRequest)
			resp.Write([]byte("gzip me!"))
			return nil
		}
		t := gzm.Middleware(gzip.BestCompression, gzm.MinSize(0), gzm.AddStatusCodes(http.StatusBadRequest))(h)
		err := t(ctx, rw, req)
		Ω(err).ShouldNot(HaveOccurred())
		resp := goa.ContextResponse(ctx)
		Ω(resp.Status).Should(Equal(http.StatusBadRequest))
		Ω(resp.Header().Get("Content-Encoding")).Should(Equal("gzip"))

		gzr, err := gzip.NewReader(bytes.NewReader(rw.Body))
		Ω(err).ShouldNot(HaveOccurred())
		var buf bytes.Buffer
		_, err = io.Copy(&buf, gzr)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(buf.String()).Should(Equal("gzip me!"))
	})

	It("encodes response using gzip (all status)", func() {
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			resp := goa.ContextResponse(ctx)
			resp.WriteHeader(http.StatusBadRequest)
			resp.Write([]byte("gzip me!"))
			return nil
		}
		t := gzm.Middleware(gzip.BestCompression, gzm.MinSize(0), gzm.OnlyStatusCodes())(h)
		err := t(ctx, rw, req)
		Ω(err).ShouldNot(HaveOccurred())
		resp := goa.ContextResponse(ctx)
		Ω(resp.Status).Should(Equal(http.StatusBadRequest))
		Ω(resp.Header().Get("Content-Encoding")).Should(Equal("gzip"))

		gzr, err := gzip.NewReader(bytes.NewReader(rw.Body))
		Ω(err).ShouldNot(HaveOccurred())
		var buf bytes.Buffer
		_, err = io.Copy(&buf, gzr)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(buf.String()).Should(Equal("gzip me!"))
	})

	It("encodes response using gzip (custom type)", func() {
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			resp := goa.ContextResponse(ctx)
			resp.Header().Add("Content-Type", "custom/type")
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte("gzip me!"))
			return nil
		}
		t := gzm.Middleware(gzip.BestCompression, gzm.MinSize(0), gzm.AddContentTypes("custom/type"))(h)
		err := t(ctx, rw, req)
		Ω(err).ShouldNot(HaveOccurred())
		resp := goa.ContextResponse(ctx)
		Ω(resp.Status).Should(Equal(http.StatusOK))
		Ω(resp.Header().Get("Content-Encoding")).Should(Equal("gzip"))

		gzr, err := gzip.NewReader(bytes.NewReader(rw.Body))
		Ω(err).ShouldNot(HaveOccurred())
		var buf bytes.Buffer
		_, err = io.Copy(&buf, gzr)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(buf.String()).Should(Equal("gzip me!"))
	})

	It("encodes response using gzip (length check)", func() {
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			resp := goa.ContextResponse(ctx)
			resp.WriteHeader(http.StatusOK)
			// Use multiple writes.
			for i := 0; i < 128; i++ {
				_, err := resp.Write([]byte("gzip me!"))
				if err != nil {
					return err
				}
			}
			return nil
		}
		t := gzm.Middleware(gzip.BestCompression, gzm.MinSize(512))(h)
		err := t(ctx, rw, req)
		Ω(err).ShouldNot(HaveOccurred())
		resp := goa.ContextResponse(ctx)
		Ω(resp.Status).Should(Equal(http.StatusOK))
		Ω(resp.Header().Get("Content-Encoding")).Should(Equal("gzip"))

		gzr, err := gzip.NewReader(bytes.NewReader(rw.Body))
		Ω(err).ShouldNot(HaveOccurred())
		var buf bytes.Buffer
		_, err = io.Copy(&buf, gzr)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(buf.String()).Should(Equal(strings.Repeat("gzip me!", 128)))
	})

	It("removes Accept-Ranges header", func() {
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			resp := goa.ContextResponse(ctx)
			resp.Header().Add("Accept-Ranges", "some value")
			resp.WriteHeader(http.StatusOK)
			// Use multiple writes.
			for i := 0; i < 128; i++ {
				_, err := resp.Write([]byte("gzip me!"))
				if err != nil {
					return err
				}
			}
			return nil
		}
		t := gzm.Middleware(gzip.BestCompression, gzm.MinSize(512))(h)
		err := t(ctx, rw, req)
		Ω(err).ShouldNot(HaveOccurred())
		resp := goa.ContextResponse(ctx)
		Ω(resp.Status).Should(Equal(http.StatusOK))
		Ω(resp.Header().Get("Content-Encoding")).Should(Equal("gzip"))
		Ω(resp.Header().Get("Accept-Ranges")).Should(Equal(""))

		gzr, err := gzip.NewReader(bytes.NewReader(rw.Body))
		Ω(err).ShouldNot(HaveOccurred())
		var buf bytes.Buffer
		_, err = io.Copy(&buf, gzr)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(buf.String()).Should(Equal(strings.Repeat("gzip me!", 128)))
	})

	It("should preserve status code", func() {
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			resp := goa.ContextResponse(ctx)
			resp.WriteHeader(http.StatusConflict)
			// Use multiple writes.
			for i := 0; i < 128; i++ {
				_, err := resp.Write([]byte("gzip me!"))
				if err != nil {
					return err
				}
			}
			return nil
		}
		t := gzm.Middleware(gzip.BestCompression, gzm.MinSize(512), gzm.AddStatusCodes(http.StatusConflict))(h)
		err := t(ctx, rw, req)
		Ω(err).ShouldNot(HaveOccurred())
		resp := goa.ContextResponse(ctx)
		Ω(resp.Status).Should(Equal(http.StatusConflict))
		Ω(resp.Header().Get("Content-Encoding")).Should(Equal("gzip"))
		Ω(resp.Header().Get("Accept-Ranges")).Should(Equal(""))

		gzr, err := gzip.NewReader(bytes.NewReader(rw.Body))
		Ω(err).ShouldNot(HaveOccurred())
		var buf bytes.Buffer
		_, err = io.Copy(&buf, gzr)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(buf.String()).Should(Equal(strings.Repeat("gzip me!", 128)))
	})
})

var _ = Describe("NotGzip", func() {
	var ctx context.Context
	var req *http.Request
	var rw *TestResponseWriter
	payload := map[string]interface{}{"payload": 42}

	BeforeEach(func() {
		var err error
		req, err = http.NewRequest("POST", "/foo/bar", strings.NewReader(`{"payload":42}`))
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("Range", "bytes=0-10")
		Ω(err).ShouldNot(HaveOccurred())
		rw = &TestResponseWriter{ParentHeader: make(http.Header)}

		ctx = goa.NewContext(nil, rw, req, nil)
		goa.ContextRequest(ctx).Payload = payload
	})

	It("does not encode response (already gzipped)", func() {
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			resp := goa.ContextResponse(ctx)
			resp.Header().Set("Content-Type", "gzip")
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte("gzip data"))
			return nil
		}
		t := gzm.Middleware(gzip.BestCompression, gzm.MinSize(0))(h)
		err := t(ctx, rw, req)
		Ω(err).ShouldNot(HaveOccurred())
		resp := goa.ContextResponse(ctx)
		Ω(resp.Status).Should(Equal(http.StatusOK))
		Ω(resp.Header().Get("Content-Encoding")).ShouldNot(Equal("gzip"))

		var buf bytes.Buffer
		_, err = io.Copy(&buf, bytes.NewBuffer(rw.Body))
		Ω(err).ShouldNot(HaveOccurred())
		Ω(buf.String()).Should(Equal("gzip data"))
	})

	It("does not encode response (too small)", func() {
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			resp := goa.ContextResponse(ctx)
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte("gzip me!"))
			return nil
		}
		t := gzm.Middleware(gzip.BestCompression)(h)
		err := t(ctx, rw, req)
		Ω(err).ShouldNot(HaveOccurred())
		resp := goa.ContextResponse(ctx)
		Ω(resp.Status).Should(Equal(http.StatusOK))
		Ω(resp.Header().Get("Content-Encoding")).ShouldNot(Equal("gzip"))

		var buf bytes.Buffer
		n, err := io.Copy(&buf, bytes.NewBuffer(rw.Body))
		Ω(err).ShouldNot(HaveOccurred())
		Ω(buf.String()).Should(Equal("gzip me!"))
		Ω(resp.Header().Get("Content-Length")).Should(Equal(strconv.Itoa(int(n))))
		Ω(resp.Header().Get("Content-Length")).Should(Equal(strconv.Itoa(len("gzip me!"))))
	})

	It("does not encode response (wrong status code)", func() {
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			resp := goa.ContextResponse(ctx)
			resp.WriteHeader(http.StatusBadRequest)
			resp.Write([]byte("gzip me!"))
			return nil
		}
		t := gzm.Middleware(gzip.BestCompression, gzm.MinSize(0))(h)
		err := t(ctx, rw, req)
		Ω(err).ShouldNot(HaveOccurred())
		resp := goa.ContextResponse(ctx)
		Ω(resp.Status).Should(Equal(http.StatusBadRequest))
		Ω(resp.Header().Get("Content-Encoding")).ShouldNot(Equal("gzip"))

		var buf bytes.Buffer
		_, err = io.Copy(&buf, bytes.NewBuffer(rw.Body))
		Ω(err).ShouldNot(HaveOccurred())
		Ω(buf.String()).Should(Equal("gzip me!"))
	})

	It("does not encode response (removed status code)", func() {
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			resp := goa.ContextResponse(ctx)
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte("gzip me!"))
			return nil
		}
		t := gzm.Middleware(gzip.BestCompression, gzm.MinSize(0), gzm.OnlyStatusCodes(http.StatusBadRequest))(h)
		err := t(ctx, rw, req)
		Ω(err).ShouldNot(HaveOccurred())
		resp := goa.ContextResponse(ctx)
		Ω(resp.Status).Should(Equal(http.StatusOK))
		Ω(resp.Header().Get("Content-Encoding")).ShouldNot(Equal("gzip"))

		var buf bytes.Buffer
		_, err = io.Copy(&buf, bytes.NewBuffer(rw.Body))
		Ω(err).ShouldNot(HaveOccurred())
		Ω(buf.String()).Should(Equal("gzip me!"))
	})

	It("does not encode response (unknown content type)", func() {
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			resp := goa.ContextResponse(ctx)
			resp.Header().Add("Content-Type", "unknown/contenttype")
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte("gzip me!"))
			return nil
		}
		t := gzm.Middleware(gzip.BestCompression, gzm.MinSize(0))(h)
		err := t(ctx, rw, req)
		Ω(err).ShouldNot(HaveOccurred())
		resp := goa.ContextResponse(ctx)
		Ω(resp.Status).Should(Equal(http.StatusOK))
		Ω(resp.Header().Get("Content-Encoding")).ShouldNot(Equal("gzip"))

		var buf bytes.Buffer
		_, err = io.Copy(&buf, bytes.NewBuffer(rw.Body))
		Ω(err).ShouldNot(HaveOccurred())
		Ω(buf.String()).Should(Equal("gzip me!"))
	})

	It("does not encode response (removed type)", func() {
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			resp := goa.ContextResponse(ctx)
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte("gzip me!"))
			return nil
		}
		t := gzm.Middleware(gzip.BestCompression, gzm.MinSize(0), gzm.OnlyContentTypes("some/type"))(h)
		err := t(ctx, rw, req)
		Ω(err).ShouldNot(HaveOccurred())
		resp := goa.ContextResponse(ctx)
		Ω(resp.Status).Should(Equal(http.StatusOK))
		Ω(resp.Header().Get("Content-Encoding")).ShouldNot(Equal("gzip"))

		var buf bytes.Buffer
		_, err = io.Copy(&buf, bytes.NewBuffer(rw.Body))
		Ω(err).ShouldNot(HaveOccurred())
		Ω(buf.String()).Should(Equal("gzip me!"))
	})

	It("does not encode response (has Range header)", func() {
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			resp := goa.ContextResponse(ctx)
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte("gzip me!"))
			return nil
		}
		t := gzm.Middleware(gzip.BestCompression, gzm.MinSize(0), gzm.IgnoreRange(false))(h)
		err := t(ctx, rw, req)
		Ω(err).ShouldNot(HaveOccurred())
		resp := goa.ContextResponse(ctx)
		Ω(resp.Status).Should(Equal(http.StatusOK))
		Ω(resp.Header().Get("Content-Encoding")).ShouldNot(Equal("gzip"))

		var buf bytes.Buffer
		_, err = io.Copy(&buf, bytes.NewBuffer(rw.Body))
		Ω(err).ShouldNot(HaveOccurred())
		Ω(buf.String()).Should(Equal("gzip me!"))
	})

	It("should preserve status code", func() {
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			resp := goa.ContextResponse(ctx)
			resp.WriteHeader(http.StatusConflict)
			resp.Write([]byte("gzip me!"))
			return nil
		}
		t := gzm.Middleware(gzip.BestCompression)(h)
		err := t(ctx, rw, req)
		Ω(err).ShouldNot(HaveOccurred())
		resp := goa.ContextResponse(ctx)
		Ω(resp.Status).Should(Equal(http.StatusConflict))
		Ω(resp.Header().Get("Content-Encoding")).ShouldNot(Equal("gzip"))

		var buf bytes.Buffer
		_, err = io.Copy(&buf, bytes.NewBuffer(rw.Body))
		Ω(err).ShouldNot(HaveOccurred())
		Ω(buf.String()).Should(Equal("gzip me!"))
	})

	It("should preserve status code with no body", func() {
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			resp := goa.ContextResponse(ctx)
			resp.WriteHeader(http.StatusConflict)
			return nil
		}
		t := gzm.Middleware(gzip.BestCompression)(h)
		err := t(ctx, rw, req)
		Ω(err).ShouldNot(HaveOccurred())
		resp := goa.ContextResponse(ctx)
		Ω(resp.Status).Should(Equal(http.StatusConflict))
		Ω(resp.Header().Get("Content-Encoding")).ShouldNot(Equal("gzip"))

		var buf bytes.Buffer
		_, err = io.Copy(&buf, bytes.NewBuffer(rw.Body))
		Ω(err).ShouldNot(HaveOccurred())
		Ω(buf.String()).Should(Equal(""))
	})

	It("should default to OK with no code set", func() {
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			resp := goa.ContextResponse(ctx)
			resp.Write([]byte("gzip me!"))
			return nil
		}
		t := gzm.Middleware(gzip.BestCompression)(h)
		err := t(ctx, rw, req)
		Ω(err).ShouldNot(HaveOccurred())
		resp := goa.ContextResponse(ctx)
		Ω(resp.Status).Should(Equal(http.StatusOK))
		Ω(resp.Header().Get("Content-Encoding")).ShouldNot(Equal("gzip"))

		var buf bytes.Buffer
		_, err = io.Copy(&buf, bytes.NewBuffer(rw.Body))
		Ω(err).ShouldNot(HaveOccurred())
		Ω(buf.String()).Should(Equal("gzip me!"))
	})

})

var _ = Describe("NotGzip", func() {
	var ctx context.Context
	var req *http.Request
	var rw *TestResponseWriter
	payload := map[string]interface{}{"payload": 42}

	BeforeEach(func() {
		var err error
		req, err = http.NewRequest("POST", "/foo/bar", strings.NewReader(`{"payload":42}`))
		req.Header.Set("Accept-Encoding", "nothing")
		Ω(err).ShouldNot(HaveOccurred())
		rw = &TestResponseWriter{ParentHeader: make(http.Header)}

		ctx = goa.NewContext(nil, rw, req, nil)
		goa.ContextRequest(ctx).Payload = payload
	})

	It("does not encode response (wrong accept-encoding)", func() {
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			resp := goa.ContextResponse(ctx)
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte("gzip me!"))
			return nil
		}
		t := gzm.Middleware(gzip.BestCompression, gzm.MinSize(0))(h)
		err := t(ctx, rw, req)
		Ω(err).ShouldNot(HaveOccurred())
		resp := goa.ContextResponse(ctx)
		Ω(resp.Status).Should(Equal(http.StatusOK))
		Ω(resp.Header().Get("Content-Encoding")).ShouldNot(Equal("gzip"))

		var buf bytes.Buffer
		_, err = io.Copy(&buf, bytes.NewBuffer(rw.Body))
		Ω(err).ShouldNot(HaveOccurred())
		Ω(buf.String()).Should(Equal("gzip me!"))
	})
})
