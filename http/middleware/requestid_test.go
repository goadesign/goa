package middleware_test

import (
	"net/http"
	"testing"

	httpm "goa.design/goa/v3/http/middleware"
	"goa.design/goa/v3/middleware"
)

type (
	requestIDTestHandler struct {
		testCaseName string
		handler      http.HandlerFunc
	}

	requestIDTestCase struct {
		name    string
		options []middleware.RequestIDOption
		request *http.Request
		handler http.HandlerFunc
	}
)

func TestRequestID(t *testing.T) {
	var (
		getRequestID = func(r *http.Request) (id string) {
			id, _ = r.Context().Value(middleware.RequestIDKey).(string)
			return
		}
		makeRequest = func(xRequestIDValue string) (r *http.Request) {
			r = &http.Request{
				Header: make(http.Header),
			}
			if len(xRequestIDValue) > 0 {
				r.Header.Set("X-Request-Id", xRequestIDValue)
			}
			return
		}
		ignored http.ResponseWriter
	)
	for _, tc := range []*requestIDTestCase{
		{"default without header", nil, makeRequest(""),
			func(_ http.ResponseWriter, r *http.Request) {
				id := getRequestID(r)
				if len(id) != 8 {
					t.Errorf("%s: unexpected request ID length: %d != 8", r.Header.Get("Test-Case"), len(id))
				}
			},
		},
		{"default ignores header", nil, makeRequest("ignore this header"),
			func(_ http.ResponseWriter, r *http.Request) {
				id := getRequestID(r)
				if len(id) != 8 {
					t.Errorf("%s: unexpected request ID length: %d != 8", r.Header.Get("Test-Case"), len(id))
				}
			},
		},
		{"generate without header",
			[]middleware.RequestIDOption{httpm.UseXRequestIDHeaderOption(true)},
			makeRequest(""),
			func(_ http.ResponseWriter, r *http.Request) {
				id := getRequestID(r)
				if len(id) != 8 {
					t.Errorf("%s: unexpected request ID length: %d != 8", r.Header.Get("Test-Case"), len(id))
				}
			},
		},
		{"accept header",
			[]middleware.RequestIDOption{httpm.UseXRequestIDHeaderOption(true)},
			makeRequest("accept+header"),
			func(_ http.ResponseWriter, r *http.Request) {
				id := getRequestID(r)
				if id != "accept+header" {
					t.Errorf("%s: unexpected request ID length: %s != accept+header", r.Header.Get("Test-Case"), id)
				}
			},
		},
		{"truncate header",
			[]middleware.RequestIDOption{
				httpm.UseXRequestIDHeaderOption(true),
				httpm.XRequestHeaderLimitOption(3),
			},
			makeRequest("too long for length limit"),
			func(_ http.ResponseWriter, r *http.Request) {
				id := getRequestID(r)
				if id != "too" {
					t.Errorf("%s: unexpected request ID length: %s != too", r.Header.Get("Test-Case"), id)
				}
			},
		},
	} {
		httpm.RequestID(tc.options...)(
			&requestIDTestHandler{tc.name, tc.handler}).ServeHTTP(ignored, tc.request)
	}
}

// ServeHTTP implements http.Handler#ServeHTTP
func (h *requestIDTestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("Test-Case", h.testCaseName)
	h.handler(w, r)
}
