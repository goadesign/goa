package goa

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"gopkg.in/inconshreveable/log15.v2"

	. "github.com/smartystreets/goconvey/convey"
)

func ExampleFilterHeaders() {
	headers := http.Header{}
	for header := range sensitiveHeaders {
		headers.Set(header, "")
	}
	headers.Set("header-not-sensitive-1", "") // header key will be convent to canonical format
	headers.Set("header-not-sensitive-2", "")

	iterator := func(name string, value []string) {
		fmt.Println(name == "Header-Not-Sensitive-1" || name == "Header-Not-Sensitive-2")
	}
	filterHeaders(headers, iterator)

	// Output:
	// true
	// true
}

func ExampleWriteHeaders() {
	headers := http.Header{}
	headers.Add("key", "val1")
	headers.Add("key", "val2")

	buffer := bytes.NewBuffer(nil)
	writeHeaders(buffer, headers)
	fmt.Println(buffer.String())

	// Output:
	// Key: val1, val2
}

type clientFakeLogger struct {
	log15.Logger
	records []string
}

func (l *clientFakeLogger) Info(name string, _ ...interface{}) {
	l.records = append(l.records, name)
}

type clientFakeHTTPHandler struct {
	http.Handler
}

func (clientFakeHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("response"))
}

func TestClient(t *testing.T) {
	Convey("Given a server and a client", t, func() {
		s := httptest.NewServer(clientFakeHTTPHandler{})
		c := NewClient()
		l := &clientFakeLogger{}
		c.Logger = l

		Convey("When request", func() {
			req, _ := http.NewRequest("GET", s.URL, nil)
			resp, _ := c.Do(req)

			Convey("HTTP status code should be 200", func() {
				So(resp.StatusCode, ShouldEqual, 200)
			})

			Convey(`Response body should be "response"`, func() {
				body, _ := ioutil.ReadAll(resp.Body)
				So(string(body), ShouldEqual, "response")
				defer resp.Body.Close()
			})

			Convey(`Log records should resemble []string{"started", "completed"}`, func() {
				So(l.records, ShouldResemble, []string{"started", "completed"})
			})
		})

		Reset(func() {
			s.Close()
		})
	})
}
