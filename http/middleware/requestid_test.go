package middleware_test

import (
	"net/http"
	"testing"

	"goa.design/goa/http/middleware"
)

type (
	requestIDTestHandler struct {
		handler http.HandlerFunc
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
		&requestIDTestCase{"default without header", nil, makeRequest(""),
			func(_ http.ResponseWriter, r *http.Request) {
				id := getRequestID(r)
				if len(id) != 8 {
					t.Errorf("unexpected request ID length: %d != 8", len(id))
				}
			},
		},
		&requestIDTestCase{"default ignores header", nil, makeRequest("ignore this header"),
			func(_ http.ResponseWriter, r *http.Request) {
				id := getRequestID(r)
				if len(id) != 8 {
					t.Errorf("unexpected request ID length: %d != 8", len(id))
				}
			},
		},
		&requestIDTestCase{"accept header",
			[]middleware.RequestIDOption{middleware.UseXRequestIDHeaderOption(true)},
			makeRequest("accepted"),
			func(_ http.ResponseWriter, r *http.Request) {
				id := getRequestID(r)
				if id != "accepted" {
					t.Errorf("unexpected request ID length: %s != accepted", id)
				}
			},
		},
		&requestIDTestCase{"truncate header",
			[]middleware.RequestIDOption{
				middleware.UseXRequestIDHeaderOption(true),
				middleware.XRequestHeaderLimitOption(3),
			},
			makeRequest("too long for length limit"),
			func(_ http.ResponseWriter, r *http.Request) {
				id := getRequestID(r)
				if id != "too" {
					t.Errorf("unexpected request ID length: %s != too", id)
				}
			},
		},
	} {
		middleware.RequestID(tc.options...)(
			&requestIDTestHandler{tc.handler}).ServeHTTP(ignored, tc.request)
	}
}

// ServeHTTP implements http.Handler#ServeHTTP
func (h *requestIDTestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler(w, r)
}
