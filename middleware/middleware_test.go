package middleware_test

import (
	"net/http"
	"net/url"

	"github.com/goadesign/goa"
	"golang.org/x/net/context"
)

// Helper that sets up a "working" service
func newService(logger goa.Logger) *goa.Service {
	service := goa.New("test")
	service.Encoder(goa.NewJSONEncoder, "*/*")
	service.Decoder(goa.NewJSONDecoder, "*/*")
	service.WithLogger(logger)
	return service
}

// Creates a test context
func newContext(service *goa.Service, rw http.ResponseWriter, req *http.Request, params url.Values) context.Context {
	ctrl := service.NewController("test")
	return goa.NewContext(ctrl.Context, rw, req, params)
}

type logEntry struct {
	Msg  string
	Data []interface{}
}

type testLogger struct {
	InfoEntries  []logEntry
	ErrorEntries []logEntry
}

func (t *testLogger) Info(msg string, data ...interface{}) {
	e := logEntry{msg, data}
	t.InfoEntries = append(t.InfoEntries, e)
}

func (t *testLogger) Error(msg string, data ...interface{}) {
	e := logEntry{msg, data}
	t.ErrorEntries = append(t.ErrorEntries, e)
}

type testResponseWriter struct {
	ParentHeader http.Header
	Body         []byte
	Status       int
}

func newTestResponseWriter() *testResponseWriter {
	h := make(http.Header)
	return &testResponseWriter{ParentHeader: h}
}

func (t *testResponseWriter) Header() http.Header {
	return t.ParentHeader
}

func (t *testResponseWriter) Write(b []byte) (int, error) {
	t.Body = append(t.Body, b...)
	return len(b), nil
}

func (t *testResponseWriter) WriteHeader(s int) {
	t.Status = s
}
